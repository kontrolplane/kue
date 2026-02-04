VERSION 0.8
FROM golang:1.23.5-alpine3.21
WORKDIR /kontrolplane

deps:
    COPY go.mod go.sum ./
    RUN go mod download
    SAVE ARTIFACT go.mod AS LOCAL go.mod
    SAVE ARTIFACT go.sum AS LOCAL go.sum

compile:
    FROM +deps
    COPY main.go .
    COPY cmd/ cmd/
    COPY pkg/ pkg/
    RUN go build -o build/kontrolplane/kue main.go
    SAVE ARTIFACT build/kontrolplane/kue AS LOCAL build/kontrolplane/kue

local:
    LOCALLY
    RUN go build -o build/kontrolplane/kue main.go

container:
    COPY +compile/kue ./kontrolplane/kue
    ENTRYPOINT ["./kontrolplane/kue"]
    ARG tag="latest"
    SAVE IMAGE ghcr.io/kontrolplane/northernlights:${tag}

resources:
  BUILD +queues
  BUILD +messages

queues:
    LOCALLY
    ARG AWS_ENDPOINT_URL="http://localhost:4566"
    ARG REGION="us-east-1"
    ARG ACCOUNT_ID="000000000000"
    RUN aws sqs create-queue --queue-name kontrolplane-users-deadletter
    RUN aws sqs create-queue --queue-name kontrolplane-emails-deadletter
    RUN aws sqs create-queue --queue-name kontrolplane-logs-deadletter
    RUN aws sqs create-queue --queue-name kontrolplane-orders-deadletter.fifo --attributes FifoQueue=true,ContentBasedDeduplication=true
    RUN aws sqs create-queue --queue-name kontrolplane-shipments-deadletter.fifo --attributes FifoQueue=true,ContentBasedDeduplication=true
    RUN aws sqs create-queue --queue-name kontrolplane-users --attributes "{\"RedrivePolicy\":\"{\\\"deadLetterTargetArn\\\":\\\"arn:aws:sqs:${REGION}:${ACCOUNT_ID}:kontrolplane-users-deadletter\\\",\\\"maxReceiveCount\\\":\\\"3\\\"}\"}"
    RUN aws sqs create-queue --queue-name kontrolplane-emails --attributes "{\"RedrivePolicy\":\"{\\\"deadLetterTargetArn\\\":\\\"arn:aws:sqs:${REGION}:${ACCOUNT_ID}:kontrolplane-emails-deadletter\\\",\\\"maxReceiveCount\\\":\\\"3\\\"}\"}"
    RUN aws sqs create-queue --queue-name kontrolplane-logs --attributes "{\"RedrivePolicy\":\"{\\\"deadLetterTargetArn\\\":\\\"arn:aws:sqs:${REGION}:${ACCOUNT_ID}:kontrolplane-logs-deadletter\\\",\\\"maxReceiveCount\\\":\\\"5\\\"}\"}"
    RUN aws sqs create-queue --queue-name kontrolplane-orders.fifo --attributes "{\"FifoQueue\":\"true\",\"ContentBasedDeduplication\":\"true\",\"RedrivePolicy\":\"{\\\"deadLetterTargetArn\\\":\\\"arn:aws:sqs:${REGION}:${ACCOUNT_ID}:kontrolplane-orders-deadletter.fifo\\\",\\\"maxReceiveCount\\\":\\\"5\\\"}\"}"
    RUN aws sqs create-queue --queue-name kontrolplane-shipments.fifo --attributes "{\"FifoQueue\":\"true\",\"ContentBasedDeduplication\":\"true\",\"RedrivePolicy\":\"{\\\"deadLetterTargetArn\\\":\\\"arn:aws:sqs:${REGION}:${ACCOUNT_ID}:kontrolplane-shipments-deadletter.fifo\\\",\\\"maxReceiveCount\\\":\\\"5\\\"}\"}"

messages:
    LOCALLY
    ARG AWS_ENDPOINT_URL="http://localhost:4566"
    ARG REGION="us-east-1"
    ARG ACCOUNT_ID="000000000000"
    # Main queue messages
    RUN aws sqs send-message --queue-url $AWS_ENDPOINT_URL/$ACCOUNT_ID/kontrolplane-users --message-body '{"userId": "20319", "action": "create", "timestamp": "2025-01-01T12:00:00Z"}'
    RUN aws sqs send-message --queue-url $AWS_ENDPOINT_URL/$ACCOUNT_ID/kontrolplane-users --message-body '{"userId": "23809", "action": "delete", "timestamp": "2025-01-01T12:00:00Z"}'
    RUN aws sqs send-message --queue-url $AWS_ENDPOINT_URL/$ACCOUNT_ID/kontrolplane-users --message-body '{"userId": "21234", "action": "update", "timestamp": "2025-01-01T12:00:00Z"}'
    RUN aws sqs send-message --queue-url $AWS_ENDPOINT_URL/$ACCOUNT_ID/kontrolplane-emails --message-body '{"emailId": "8129034293140", "recipient": "user@example.com", "subject": "Welcome!", "timestamp": "2025-01-01T12:01:00Z"}'
    RUN aws sqs send-message --queue-url $AWS_ENDPOINT_URL/$ACCOUNT_ID/kontrolplane-emails --message-body '{"emailId": "8210480221403", "recipient": "user@example.com", "subject": "Goodbye!", "timestamp": "2025-01-01T12:01:00Z"}'
    RUN aws sqs send-message --queue-url $AWS_ENDPOINT_URL/$ACCOUNT_ID/kontrolplane-orders.fifo --message-body '{"orderId": "1273897912902", "status": "pending", "amount": 99.99, "timestamp": "2025-01-01T12:02:00Z"}' --message-group-id "g-01"
    RUN aws sqs send-message --queue-url $AWS_ENDPOINT_URL/$ACCOUNT_ID/kontrolplane-logs --message-body '{"logId": "fghij", "level": "info", "message": "Order processed successfully", "timestamp": "2025-01-01T12:03:00Z"}'
    # DLQ messages (simulating failed processing)
    RUN aws sqs send-message --queue-url $AWS_ENDPOINT_URL/$ACCOUNT_ID/kontrolplane-users-deadletter --message-body '{"userId": "99999", "action": "create", "error": "Database connection timeout", "failedAt": "2025-01-01T11:55:00Z"}'
    RUN aws sqs send-message --queue-url $AWS_ENDPOINT_URL/$ACCOUNT_ID/kontrolplane-users-deadletter --message-body '{"userId": "88888", "action": "update", "error": "Validation failed: invalid email format", "failedAt": "2025-01-01T11:50:00Z"}'
    RUN aws sqs send-message --queue-url $AWS_ENDPOINT_URL/$ACCOUNT_ID/kontrolplane-emails-deadletter --message-body '{"emailId": "5555555555555", "recipient": "invalid@", "error": "Invalid recipient address", "failedAt": "2025-01-01T11:45:00Z"}'
    RUN aws sqs send-message --queue-url $AWS_ENDPOINT_URL/$ACCOUNT_ID/kontrolplane-orders-deadletter.fifo --message-body '{"orderId": "9999999999999", "status": "failed", "error": "Payment gateway unavailable", "failedAt": "2025-01-01T11:40:00Z"}' --message-group-id "dlq-01" --message-deduplication-id "dlq-dedup-001"

list:
    LOCALLY
    ARG AWS_ENDPOINT_URL="http://localhost:4566"
    RUN aws sqs list-queues

vhs:
    LOCALLY
    ARG AWS_ENDPOINT_URL="http://localhost:4566"
    RUN vhs vhs/cassette.tape

all:
  BUILD +compile
  BUILD +container
