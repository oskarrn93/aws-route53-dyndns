package httpclient

import (
	"time"

	"resty.dev/v3"
)

func New() *resty.Client {
	client := resty.New()

	client.SetTimeout(10 * time.Second)
	client.SetRetryCount(3)

	// Set custom headers if needed
	client.SetHeader("User-Agent", "oskar/aws-route53-dyndns")

	return client
}
