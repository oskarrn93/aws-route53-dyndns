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

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
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
		panic(fmt.Errorf("Invalid configuration: %w", err))
	}

	logger := logger.NewLoggerWithLevel(config.LogLevel)
	route53Client := route53.New(session.New())
	httpClient := httpclient.New()

	dnsRecordService := dnsrecord.NewService(dnsrecord.NewRepository(route53Client, httpClient, logger))

	ipAddress, err := dnsRecordService.GetExternalIp(ctx)
	if err != nil {
		panic(fmt.Errorf("Failed to retrieve external IP address: %w", err))
	}

	logger.Debug("Retrieved external ip address", "ipAddress", ipAddress)

	existingIpAddress, err := dnsRecordService.GetIpAddressForRecord(config.HostedZoneId, config.RecordName)
	if err != nil {
		panic(fmt.Errorf("Failed to retrieve existing DNS record: %w", err))
	}

	logger.Debug("Retrieved existing ip address for dns record", "existingIpAddress", existingIpAddress)

	if dnsrecord.IsEqualIPAddresses(ipAddress, existingIpAddress) {
		logger.Debug("Ip address has not changed")
		logger.Info("Done")
		return
	}

	err = dnsRecordService.UpdateRecord(config.HostedZoneId, config.RecordName, ipAddress)
	if err != nil {
		panic(fmt.Errorf("Failed to update DNS record: %w", err))
	}

	logger.Info("Record updated", "ipAddress", ipAddress, "recordName", config.RecordName)

	pushoverNotification := notification.NewPushoverNotication(config.Pushover, logger)
	pushoverNotification.Send(config.RecordName)
	logger.Debug("Notification sent", "recordName", config.RecordName)

	logger.Info("Done")
}
