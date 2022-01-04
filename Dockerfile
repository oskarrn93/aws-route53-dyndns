FROM golang:1.17.5-alpine3.15 AS builder
WORKDIR /app

COPY go.mod .
COPY go.sum .
COPY main.go .

RUN go build

FROM alpine:3.15.0
WORKDIR /app

RUN apk update && apk add dumb-init

COPY --from=builder /app/aws-route53-dyndns /app/aws-route53-dyndns 

# Run the cron job every 5 minutes
RUN echo "*/5  *  *  *  * /app/aws-route53-dyndns 2>&1" >> /etc/crontabs/root

ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD [ "crond", "-l", "2", "-f" ]
