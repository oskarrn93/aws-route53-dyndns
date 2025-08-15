package dnsrecord

import (
	"log/slog"
	"net"

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

func IsEqualIPAddresses(ip1, ip2 string) bool {
	if ip1 == "" || ip2 == "" {
		return false
	}

	return net.IP.Equal(net.ParseIP(ip1), net.ParseIP(ip2))
}
