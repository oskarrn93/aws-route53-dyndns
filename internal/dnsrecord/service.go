package dnsrecord

import (
	"context"
	"fmt"
)

type Service struct {
	repository *Repository
}

func (s *Service) GetExternalIp(ctx context.Context) (string, error) {
	return s.repository.GetExternalIp(ctx)
}

func (s *Service) GetIpAddressForRecord(hostedZoneId, recordName string) (string, error) {
	existingRecord, err := s.repository.GetRecord(hostedZoneId, recordName)
	if err != nil {
		return "", fmt.Errorf("Failed to retrieve existing DNS record: %w", err)
	}

	ipAddress := existingRecord.Value

	if ipAddress == nil || *ipAddress == "" {
		return "", fmt.Errorf("Record has no value when an IP address is expected")
	}

	if !IsValidIPAddress(*ipAddress, s.repository.logger) {
		return "", fmt.Errorf("Invalid IP address in DNS record: %s", *ipAddress)
	}

	return *ipAddress, nil
}

func (s *Service) UpdateRecord(hostedZoneId, recordName, ipAddress string) error {
	if !IsValidIPAddress(ipAddress, s.repository.logger) {
		return fmt.Errorf("Invalid IP address: %s", ipAddress)
	}

	err := s.repository.UpdateRecord(hostedZoneId, recordName, ipAddress)
	if err != nil {
		return fmt.Errorf("Failed to update DNS record: %w", err)
	}

	s.repository.logger.Info("Record updated", "ipAddress", ipAddress, "recordName", recordName)
	return nil
}

func NewService(repository *Repository) *Service {
	return &Service{
		repository: repository,
	}
}
