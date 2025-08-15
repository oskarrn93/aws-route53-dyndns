package dnsrecord

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"resty.dev/v3"
)

type Repository struct {
	route53Client *route53.Client
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
		return "", fmt.Errorf("failed to retrieve external ip: %d", response.StatusCode())
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

func (r *Repository) GetRecord(ctx context.Context, hostedZoneId string, recordName string) (types.ResourceRecord, error) {
	input := &route53.ListResourceRecordSetsInput{
		HostedZoneId:    &hostedZoneId,
		StartRecordName: aws.String(recordName),
		StartRecordType: types.RRTypeA,
	}

	response, err := r.route53Client.ListResourceRecordSets(ctx, input)
	if err != nil {
		return types.ResourceRecord{}, err
	}

	r.logger.Debug("ListResourceRecordSets response", "response", response)

	if len(response.ResourceRecordSets) == 0 {
		return types.ResourceRecord{}, fmt.Errorf("no records found for hosted zone %s and record name %s", hostedZoneId, recordName)
	}

	// Just pick first record for the given name
	resourceRecords := response.ResourceRecordSets[0].ResourceRecords
	if resourceRecords != nil && len(resourceRecords) == 0 {
		return types.ResourceRecord{}, fmt.Errorf("no resource records found for hosted zone %s and record name %s", hostedZoneId, recordName)
	}

	result := resourceRecords[0]
	return result, nil
}

func (r *Repository) UpdateRecord(ctx context.Context, hostedZoneId string, recordName string, ipAddress string) error {
	input := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &types.ChangeBatch{
			Changes: []types.Change{
				{
					Action: types.ChangeActionUpsert,
					ResourceRecordSet: &types.ResourceRecordSet{
						Name: aws.String(recordName),
						ResourceRecords: []types.ResourceRecord{
							{
								Value: aws.String(ipAddress),
							},
						},
						TTL:  aws.Int64(1800),
						Type: types.RRTypeA,
					},
				},
			},
		},
		HostedZoneId: aws.String(hostedZoneId),
	}

	result, err := r.route53Client.ChangeResourceRecordSets(ctx, input)
	if err != nil {
		return err
	}

	r.logger.Info("Updated record successfully", "recordName", recordName, "ipAddress", ipAddress, "result", result)

	return nil
}

func NewRepository(route53Client *route53.Client, httpClient *resty.Client, logger *slog.Logger) *Repository {
	return &Repository{
		route53Client: route53Client,
		httpClient:    httpClient,
		logger:        logger,
	}
}
