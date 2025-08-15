# aws-route53-dyndns

This script retrieves the external IP address and updates the A record value for the domain name in AWS Route53. It is used for [Dynamic DNS](https://en.wikipedia.org/wiki/Dynamic_DNS).

## How to

### Development

#### Run

```sh
make run
```

#### Test

Run the unit tests:

```sh
make test
```

### Docker

The container runs a cron job to execute the script every 5 minutes.

```sh
make run-docker
```

To view logs:

```sh
docker-compose logs -f
```

## Environment Variables

| Name               | Description                                                                | Required |
| ------------------ | -------------------------------------------------------------------------- | -------- |
| AWS_REGION         | The AWS Region to use                                                      | ✅       |
| HOSTED_ZONE_ID     | The Route53 hosted zone ID                                                 | ✅       |
| RECORD_NAME        | The record, e.g., `example.com`                                            | ✅       |
| LOG_LEVEL          | Verbosity level for logger (e.g., `info`, `debug`)                         | ❌       |
| PUSHOVER_API_TOKEN | App API token to send push notifications via Pushover                      | ❌       |
| PUSHOVER_USER_KEY  | Account user key for the recipient to send push notifications via Pushover | ❌       |

## AWS

This assumes that [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html) is configured for the IAM User.

### IAM Permissions

The script requires the IAM User to have permissions for the following actions:

- `route53:ListResourceRecordSets`
- `route53:ChangeResourceRecordSets`
- `route53:GetHostedZone`

### Additional Notes

- Ensure the IAM User has sufficient permissions to access the specified hosted zone.
- Logs are written to the console and can be configured using the `LOG_LEVEL` environment variable.
- If Pushover notifications are enabled, ensure both `PUSHOVER_API_TOKEN` and `PUSHOVER_USER_KEY` are set.
