package main

import (
	"aws-route53-dyndns/internal/config"
	"aws-route53-dyndns/internal/dnsrecord"
	"aws-route53-dyndns/internal/httpclient"
	"aws-route53-dyndns/internal/logger"
	"aws-route53-dyndns/internal/notification"
	"aws-route53-dyndns/internal/telemetry"
	"context"
	"fmt"
	"os"
	"os/signal"

	awsSdkConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/route53"
)

func main() {
	// Handle application termination signals
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	cfg, err := config.New()
	if err != nil {
		panic(fmt.Errorf("failed to initialize config: %w", err))
	}

	logger := logger.NewLoggerWithLevel(cfg.LogLevel)

	tel, err := telemetry.New(ctx, cfg, logger)
	if err != nil {
		panic(fmt.Errorf("failed to setup telemetry: %w", err))
	}
	defer tel.Shutdown(ctx)

	ctx, span := tel.Tracer.Start(ctx, "Job")
	defer span.End()

	httpClient := httpclient.New()

	awsConfig, err := awsSdkConfig.LoadDefaultConfig(ctx)
	if err != nil {
		tel.Increment(ctx, telemetry.FailedRunsMetric)
		panic(fmt.Errorf("failed to load AWS configuration: %w", err))
	}

	route53Client := route53.NewFromConfig(awsConfig)

	dnsRecordService := dnsrecord.NewService(dnsrecord.NewRepository(route53Client, httpClient, logger))

	ipAddress, err := dnsRecordService.GetExternalIp(ctx)
	if err != nil {
		tel.Increment(ctx, telemetry.FailedRunsMetric)
		panic(fmt.Errorf("failed to retrieve external IP address: %w", err))
	}

	logger.Debug("Retrieved external ip address", "ipAddress", ipAddress)

	existingIpAddress, err := dnsRecordService.GetIpAddressForRecord(ctx, cfg.HostedZoneId, cfg.RecordName)
	if err != nil {
		tel.Increment(ctx, telemetry.FailedRunsMetric)
		panic(fmt.Errorf("failed to retrieve existing DNS record: %w", err))
	}

	logger.Debug("Retrieved existing ip address for dns record", "existingIpAddress", existingIpAddress)

	if dnsrecord.IsEqualIPAddresses(ipAddress, existingIpAddress) {
		logger.Debug("Ip address has not changed")

		tel.Increment(ctx, telemetry.IPAddressNotChangedMetric)
		tel.Increment(ctx, telemetry.SuccessfulRunsMetric)

		logger.Info("Done")
		return
	}

	err = dnsRecordService.UpdateRecord(ctx, cfg.HostedZoneId, cfg.RecordName, ipAddress)
	if err != nil {
		tel.Increment(ctx, telemetry.FailedRunsMetric)
		panic(fmt.Errorf("failed to update DNS record: %w", err))
	}

	logger.Info("Record updated", "ipAddress", ipAddress, "recordName", cfg.RecordName)

	pushoverNotification := notification.NewPushoverNotication(cfg.Pushover, logger)
	pushoverNotification.Send(cfg.RecordName)
	logger.Debug("Notification sent", "recordName", cfg.RecordName)

	tel.Increment(ctx, telemetry.IPAddressChangedMetric)
	tel.Increment(ctx, telemetry.SuccessfulRunsMetric)

	logger.Info("Done")
}
