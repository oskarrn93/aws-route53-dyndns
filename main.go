package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"

	"github.com/gregdel/pushover"
	"github.com/joho/godotenv"
	"github.com/muonsoft/validation/validate"
	"github.com/sirupsen/logrus"
)

type envConfig struct {
	hostedZoneId     string
	recordName       string
	logLevel         string
	pushoverApiToken string
	pushoverUserKey  string
}

var log = logrus.New()

func isValidIpAddress(ipAddress string) bool {
	if validate.IPv4(ipAddress) == nil {
		log.WithFields(logrus.Fields{
			"ipAddress": ipAddress,
		}).Debug("Valid ipv4 address")
		return true
	}
	log.WithFields(logrus.Fields{
		"ipAddress": ipAddress,
	}).Info("Invalid ipv4 address")

	if validate.IPv6(ipAddress) == nil {
		log.WithFields(logrus.Fields{
			"ipAddress": ipAddress,
		}).Debug("Valid ipv6 address")
		return true
	}
	log.WithFields(logrus.Fields{
		"ipAddress": ipAddress,
	}).Info("Invalid ipv6 address")

	return false
}

func getExternalIp() string {
	url := "http://checkip.amazonaws.com/"

	response, error := http.Get(url)

	if error != nil {
		log.WithFields(logrus.Fields{
			"error": error,
		}).Fatal("Error retrieving external ip")
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		log.WithFields(logrus.Fields{
			"responseStatusCode": response.StatusCode,
		}).Fatalln("Non-OK HTTP status when retrieving external ip")
	}

	body, error := io.ReadAll(response.Body)

	if error != nil {
		log.WithFields(logrus.Fields{
			"error": error,
		}).Fatal("Error reading body")
	}

	ipAddress := string(bytes.Trim(body[:], "\n"))

	if isValidIpAddress(ipAddress) == false {
		log.WithFields(logrus.Fields{
			"ipAddress": ipAddress,
		}).Fatalln("The parsed value is not a valid ip address")
	}

	return ipAddress
}

func updateRecord(svc *route53.Route53, hostedZoneId string, recordName string, ipAddress string) {
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

	result, error := svc.ChangeResourceRecordSets(input)
	if error != nil {
		if aerror, ok := error.(awserr.Error); ok {
			switch aerror.Code() {
			case route53.ErrCodeNoSuchHostedZone:
				log.Fatalln(route53.ErrCodeNoSuchHostedZone, aerror.Error())
			case route53.ErrCodeNoSuchHealthCheck:
				log.Fatalln(route53.ErrCodeNoSuchHealthCheck, aerror.Error())
			case route53.ErrCodeInvalidChangeBatch:
				log.Fatalln(route53.ErrCodeInvalidChangeBatch, aerror.Error())
			case route53.ErrCodeInvalidInput:
				log.Fatalln(route53.ErrCodeInvalidInput, aerror.Error())
			case route53.ErrCodePriorRequestNotComplete:
				log.Fatalln(route53.ErrCodePriorRequestNotComplete, aerror.Error())
			default:
				log.Fatalln(aerror.Error())
			}
		}

		log.Fatalln(error.Error())
	}

	log.Debug(result)
}

func getResourceRecord(svc *route53.Route53, hostedZoneId string, recordName string, recordType string) *route53.ResourceRecord {
	input := &route53.ListResourceRecordSetsInput{
		HostedZoneId:    &hostedZoneId,
		StartRecordName: aws.String(recordName),
		StartRecordType: aws.String(recordType),
	}

	response, error := svc.ListResourceRecordSets(input)
	if error != nil {
		if aerror, ok := error.(awserr.Error); ok {
			switch aerror.Code() {
			case route53.ErrCodeNoSuchHostedZone:
				log.Fatalln(route53.ErrCodeNoSuchHostedZone, aerror.Error())
			case route53.ErrCodeInvalidInput:
				log.Fatalln(route53.ErrCodeInvalidInput, aerror.Error())
			default:
				log.Fatalln(aerror.Error())
			}
		}

		log.Fatalln(error.Error())
	}

	log.WithField("response", response).Debug("ListResourceRecordSets response")

	if len(response.ResourceRecordSets) == 0 {
		log.Fatalln("No records in resource record sets")
	}

	if len(response.ResourceRecordSets[0].ResourceRecords) == 0 {
		log.Fatalln("No records in resource")
	}

	result := response.ResourceRecordSets[0].ResourceRecords[0]
	return result
}

func getExistingValueForRecord(svc *route53.Route53, hostedZoneId string, recordName string) string {
	resourceRecord := getResourceRecord(svc, hostedZoneId, recordName, "A")

	ipAddress := *resourceRecord.Value
	return ipAddress
}

func loadEnvironmentVariables() envConfig {
	error := godotenv.Load()
	if error != nil {
		log.Debug("No .env file found")
	}

	hostedZoneId, ok := os.LookupEnv("HOSTED_ZONE_ID")

	if ok == false {
		log.Fatalf("HOSTED_ZONE_ID is missing in .env file")
	}

	recordName, ok := os.LookupEnv("RECORD_NAME")

	if ok == false {
		log.Fatalf("RECORD_NAME is missing in .env file")
	}

	logLevel, ok := os.LookupEnv("LOG_LEVEL")

	if ok == false {
		logLevel = "info" //default to info logger level
	}

	pushoverApiToken, ok := os.LookupEnv("PUSHOVER_API_TOKEN")
	if ok == false {
		log.Debug("PUSHOVER_API_TOKEN not defined in .env file")
	}

	pushoverUserKey, ok := os.LookupEnv("PUSHOVER_USER_KEY")
	if ok == false {
		log.Debug("PUSHOVER_USER_KEY not defined in .env file")
	}

	return envConfig{
		hostedZoneId:     hostedZoneId,
		recordName:       recordName,
		logLevel:         logLevel,
		pushoverApiToken: pushoverApiToken,
		pushoverUserKey:  pushoverUserKey,
	}
}

func initLogger(logLevelString string) {
	log.SetFormatter(&logrus.JSONFormatter{PrettyPrint: true})

	logLevel, error := logrus.ParseLevel(logLevelString)
	if error != nil {
		log.Fatalln("Invalid log level:", logLevelString)
	}

	log.SetLevel(logLevel)
}

func initPushover(apiToken string, userKey string) (*pushover.Pushover, *pushover.Recipient) {
	log.Debug("apiToken", apiToken)
	log.Debug("userKey", userKey)

	if apiToken == "" || userKey == "" {
		log.Debug("Missing required credentials for Pushover")
		return nil, nil
	}

	app := pushover.New(apiToken)

	recipient := pushover.NewRecipient(userKey)

	return app, recipient
}

func sendPushoverNotification(app *pushover.Pushover, recipient *pushover.Recipient, recordName string) {

	log.WithFields(logrus.Fields{
		"app":        app,
		"recipient":  recipient,
		"recordName": recordName,
	}).Debug("sendPushoverNotification")

	if app == nil || recipient == nil {
		log.Debug("Pushover is not configured, will not send push notification")
		return
	}

	message := pushover.NewMessageWithTitle(
		fmt.Sprintf("AWS Route53 DNS record updated for: %s", recordName),
		"DNS record updated",
	)

	response, err := app.SendMessage(message, recipient)
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
		}).Errorln("Failed to send message using Pushover")
	}

	log.WithFields(logrus.Fields{
		"response": response,
	}).Debug("Pushover send message response")
}

func main() {
	log.Debug("Starting script")

	config := loadEnvironmentVariables()
	initLogger(config.logLevel)

	svc := route53.New(session.New())

	ipAddress := getExternalIp()
	log.WithFields(logrus.Fields{
		"ipAddress": ipAddress,
	}).Debug("Retrieved external ip address")

	existingIpAddress := getExistingValueForRecord(svc, config.hostedZoneId, config.recordName)
	log.WithFields(logrus.Fields{
		"existingIpAddress": existingIpAddress,
	}).Debug("Retrieved existing ip address for dns record")

	if bytes.Compare(net.ParseIP(existingIpAddress), net.ParseIP(ipAddress)) == 0 {
		log.Debug("Ip address has not changed, exiting...")
		return
	}

	updateRecord(svc, config.hostedZoneId, config.recordName, ipAddress)

	log.WithFields(logrus.Fields{
		"ipAddress":  ipAddress,
		"recordName": config.recordName,
	}).Info("Record updated")

	pushoverApp, pushoverRecipient := initPushover(config.pushoverApiToken, config.pushoverUserKey)
	sendPushoverNotification(pushoverApp, pushoverRecipient, config.recordName)
}
