# aws-route53-dyndns

This script retrieves the external ip address and updates the A record value for the domain name in AWS Route53. This is used for [Dynamic DNS](https://en.wikipedia.org/wiki/Dynamic_DNS).

## How to

### Development

#### Run

```sh
go run main.go
```

### Docker

The container is running a cron job to run the script every 5 minutes.

```sh
docker-compose up -d
```

## Environment Variables

| Name         | Description                    | Required |
| ------------ | ------------------------------ | -------- |
| hostedZoneId | The Route53 hosted zone id     | ✅       |
| recordName   | The record, e.g. `example.com` | ✅       |
| logLevel     | Verbosity level for logger     | ❌       |

## AWS

This assumes that [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html) is configured for the IAM User.

### IAM Permissions

The script requires the IAM User to have permissions for the following actions

- `route53:ListResourceRecordSets`
- `route53:ChangeResourceRecordSets`
