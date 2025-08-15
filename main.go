package main

import (
	"aws-route53-dyndns/internal/config"
	"aws-route53-dyndns/internal/dnsrecord"
	"aws-route53-dyndns/internal/httpclient"
	"aws-route53-dyndns/internal/logger"
	"aws-route53-dyndns/internal/notification"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	awsSdkConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/route53"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle application termination signals
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		cancel()
	}()

	config, err := config.New()
	if err != nil {
		panic(fmt.Errorf("invalid configuration: %w", err))
	}

	logger := logger.NewLoggerWithLevel(config.LogLevel)
	httpClient := httpclient.New()

	awsConfig, err := awsSdkConfig.LoadDefaultConfig(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to load AWS configuration: %w", err))
	}

	route53Client := route53.NewFromConfig(awsConfig)

	dnsRecordService := dnsrecord.NewService(dnsrecord.NewRepository(route53Client, httpClient, logger))

	ipAddress, err := dnsRecordService.GetExternalIp(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to retrieve external IP address: %w", err))
	}

	logger.Debug("Retrieved external ip address", "ipAddress", ipAddress)

	existingIpAddress, err := dnsRecordService.GetIpAddressForRecord(ctx, config.HostedZoneId, config.RecordName)
	if err != nil {
		panic(fmt.Errorf("failed to retrieve existing DNS record: %w", err))
	}

	logger.Debug("Retrieved existing ip address for dns record", "existingIpAddress", existingIpAddress)

	if dnsrecord.IsEqualIPAddresses(ipAddress, existingIpAddress) {
		logger.Debug("Ip address has not changed")
		logger.Info("Done")
		return
	}

	err = dnsRecordService.UpdateRecord(ctx, config.HostedZoneId, config.RecordName, ipAddress)
	if err != nil {
		panic(fmt.Errorf("failed to update DNS record: %w", err))
	}

	logger.Info("Record updated", "ipAddress", ipAddress, "recordName", config.RecordName)

	pushoverNotification := notification.NewPushoverNotication(config.Pushover, logger)
	pushoverNotification.Send(config.RecordName)
	logger.Debug("Notification sent", "recordName", config.RecordName)

	logger.Info("Done")
}
