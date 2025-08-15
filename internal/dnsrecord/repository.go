package dnsrecord

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"resty.dev/v3"
)

type Repository struct {
	route53Client *route53.Route53
	httpClient    *resty.Client
	logger        *slog.Logger
}

func (r *Repository) GetExternalIp(ctx context.Context) (string, error) {
	url := "http://checkip.amazonaws.com/"

	response, error := r.httpClient.R().WithContext(ctx).Get(url)
	if error != nil {
		return "", fmt.Errorf("error retrieving external ip: %w", error)
	}

	defer response.Body.Close()

	if !response.IsSuccess() {
		return "", fmt.Errorf("failed to retrieve external ip: %d", response.StatusCode)
	}

	body, error := io.ReadAll(response.Body)
	if error != nil {
		return "", fmt.Errorf("failed to read response body: %w", error)
	}

	ipAddress := string(bytes.Trim(body[:], "\n"))
	if !IsValidIPAddress(ipAddress, r.logger) {
		return "", fmt.Errorf("the parsed value is not a valid ip address: %s", ipAddress)
	}

	return ipAddress, nil
}

func (r *Repository) GetRecord(hostedZoneId string, recordName string) (*route53.ResourceRecord, error) {
	input := &route53.ListResourceRecordSetsInput{
		HostedZoneId:    &hostedZoneId,
		StartRecordName: aws.String(recordName),
		StartRecordType: aws.String("A"),
	}

	response, err := r.route53Client.ListResourceRecordSets(input)
	if err != nil {
		return nil, MapAWSError(err)
	}

	r.logger.Debug("ListResourceRecordSets response", "response", response)

	if len(response.ResourceRecordSets) == 0 {
		return nil, fmt.Errorf("no records found for hosted zone %s and record name %s", hostedZoneId, recordName)
	}

	// Just pick first record for the given name
	resourceRecords := response.ResourceRecordSets[0].ResourceRecords
	if resourceRecords != nil && len(resourceRecords) == 0 {
		return nil, fmt.Errorf("no resource records found for hosted zone %s and record name %s", hostedZoneId, recordName)
	}

	result := resourceRecords[0]
	return result, nil
}

func (r *Repository) UpdateRecord(hostedZoneId string, recordName string, ipAddress string) error {
	input := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String("UPSERT"),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name: aws.String(recordName),
						ResourceRecords: []*route53.ResourceRecord{
							{
								Value: aws.String(ipAddress),
							},
						},
						TTL:  aws.Int64(1800),
						Type: aws.String("A"),
					},
				},
			},
		},
		HostedZoneId: aws.String(hostedZoneId),
	}

	result, err := r.route53Client.ChangeResourceRecordSets(input)
	if err != nil {
		return MapAWSError(err)
	}

	r.logger.Info("Updated record successfully", "recordName", recordName, "ipAddress", ipAddress, "result", result)

	return nil
}

func NewRepository(route53Client *route53.Route53, httpClient *resty.Client, logger *slog.Logger) *Repository {
	return &Repository{
		route53Client: route53Client,
		httpClient:    httpClient,
		logger:        logger,
	}
}
