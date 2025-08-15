package config

import (
	"aws-route53-dyndns/internal/notification"
	"fmt"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

func NewPushover() *notification.PushoverConfig {
	pushoverApiToken := GetOptionalEnvironmentVariable("PUSHOVER_API_TOKEN", "")
	pushoverUserKey := GetOptionalEnvironmentVariable("PUSHOVER_USER_KEY", "")

	if pushoverUserKey == "" || pushoverApiToken == "" {
		slog.Debug("Pushover credentials not provided, notifications will not be sent")
		return nil
	}

	return &notification.PushoverConfig{
		ApiToken: pushoverApiToken,
		UserKey:  pushoverUserKey,
	}
}

type Config struct {
	AwsRegion    string
	HostedZoneId string
	RecordName   string
	LogLevel     string
	Pushover     *notification.PushoverConfig
}

func New() (*Config, error) {
	error := godotenv.Load()
	if error != nil {
		slog.Debug("No .env file found")
	}

	awsRegion, err := GetRequiredEnvironmentVariable("AWS_REGION")
	if err != nil {
		return nil, err
	}

	hostedZoneId, err := GetRequiredEnvironmentVariable("HOSTED_ZONE_ID")
	if err != nil {
		return nil, err
	}

	recordName, err := GetRequiredEnvironmentVariable("RECORD_NAME")
	if err != nil {
		return nil, err
	}

	logLevel := GetOptionalEnvironmentVariable("LOG_LEVEL", "info")

	config := &Config{
		AwsRegion:    awsRegion,
		HostedZoneId: hostedZoneId,
		RecordName:   recordName,
		LogLevel:     logLevel,
		Pushover:     NewPushover(),
	}

	return config, nil
}

func GetRequiredEnvironmentVariable(key string) (string, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return "", fmt.Errorf("environment variable %s not set", key)
	}
	return value, nil
}

func GetOptionalEnvironmentVariable(key string, defaultValue string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}
	return value
}
