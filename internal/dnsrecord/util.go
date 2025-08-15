package dnsrecord

import (
	"bytes"
	"fmt"
	"log/slog"
	"net"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/muonsoft/validation/validate"
)

func IsValidIPAddress(ipAddress string, logger *slog.Logger) bool {
	if validate.IPv4(ipAddress) == nil {
		logger.Debug("Valid IPv4 address", "ipAddress", ipAddress)
		return true
	}

	if validate.IPv6(ipAddress) == nil {
		logger.Debug("Valid IPv6 address", "ipAddress", ipAddress)
		return true
	}

	logger.Debug("Not a valid IP address", "ipAddress", ipAddress)
	return false
}

func MapAWSError(err error) error {
	if err == nil {
		return nil
	}

	if awsError, ok := err.(awserr.Error); ok {
		return fmt.Errorf("AWS Error: Message: %s, Code: %s, Error: %w", awsError.Message(), awsError.Code(), awsError.OrigErr())
	}

	return err
}

func IsEqualIPAddresses(ip1, ip2 string) bool {
	if ip1 == "" || ip2 == "" {
		return false
	}

	return bytes.Equal(net.ParseIP(ip1), net.ParseIP(ip2))
}
