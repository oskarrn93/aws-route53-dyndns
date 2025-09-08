FROM golang:1.25.1-alpine3.22 AS builder
WORKDIR /app

# Install dependencies
RUN apk update && apk add make

# Copy files
COPY makefile .
COPY go.mod .
COPY go.sum .
COPY *.go .
COPY internal/ internal/


# Download & install dependencies
RUN make install

# Build the application
RUN make build

FROM alpine:3.22
WORKDIR /app

# Default is to run cron job every hour
ARG CRON_JOB_INTERVAL_MINUTE=60

RUN apk update && apk add dumb-init

COPY --from=builder /app/aws-route53-dyndns /app/aws-route53-dyndns 

# Run the cron job every CRON_JOB_INTERVAL minutes
RUN echo "*/$CRON_JOB_INTERVAL_MINUTE  *  *  *  * /app/aws-route53-dyndns 2>&1" >> /etc/crontabs/root

ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD [ "crond", "-l", "2", "-f" ]
