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
    RUN aws sqs create-queue --queue-name kontrolplane-users
    RUN aws sqs create-queue --queue-name kontrolplane-users-deadletter
    RUN aws sqs create-queue --queue-name kontrolplane-emails
    RUN aws sqs create-queue --queue-name kontrolplane-emails-deadletter
    RUN aws sqs create-queue --queue-name kontrolplane-orders.fifo --attributes '{"FifoQueue":"true","ContentBasedDeduplication":"true","RedrivePolicy":"{\"deadLetterTargetArn\":\"arn:aws:sqs:'${REGION}':'${ACCOUNT_ID}':kontrolplane-orders-deadletter.fifo\",\"maxReceiveCount\":\"5\"}"}'
    RUN aws sqs create-queue --queue-name kontrolplane-orders-deadletter.fifo --attributes FifoQueue=true,ContentBasedDeduplication=true
    RUN aws sqs create-queue --queue-name kontrolplane-logs
    RUN aws sqs create-queue --queue-name kontrolplane-logs-deadletter

messages:
    LOCALLY
    ARG AWS_ENDPOINT_URL="http://localhost:4566"
    ARG REGION="us-east-1"
    ARG ACCOUNT_ID="000000000000"
    RUN aws sqs send-message --queue-url $AWS_ENDPOINT_URL/$ACCOUNT_ID/kontrolplane-users --message-body '{"userId": "20319", "action": "create", "timestamp": "2025-01-01T12:00:00Z"}'
    RUN aws sqs send-message --queue-url $AWS_ENDPOINT_URL/$ACCOUNT_ID/kontrolplane-users --message-body '{"userId": "23809", "action": "delete", "timestamp": "2025-01-01T12:00:00Z"}'
    RUN aws sqs send-message --queue-url $AWS_ENDPOINT_URL/$ACCOUNT_ID/kontrolplane-users --message-body '{"userId": "21234", "action": "update", "timestamp": "2025-01-01T12:00:00Z"}'
    RUN aws sqs send-message --queue-url $AWS_ENDPOINT_URL/$ACCOUNT_ID/kontrolplane-emails --message-body '{"emailId": "8129034293140", "recipient": "user@example.com", "subject": "Welcome!", "timestamp": "2025-01-01T12:01:00Z"}'
    RUN aws sqs send-message --queue-url $AWS_ENDPOINT_URL/$ACCOUNT_ID/kontrolplane-emails --message-body '{"emailId": "8210480221403", "recipient": "user@example.com", "subject": "Goodbye!", "timestamp": "2025-01-01T12:01:00Z"}'
    RUN aws sqs send-message --queue-url $AWS_ENDPOINT_URL/$ACCOUNT_ID/kontrolplane-orders.fifo --message-body '{"orderId": "1273897912902", "status": "pending", "amount": 99.99, "timestamp": "2025-01-01T12:02:00Z"}' --message-group-id "g-01"
    RUN aws sqs send-message --queue-url $AWS_ENDPOINT_URL/$ACCOUNT_ID/kontrolplane-logs --message-body '{"logId": "fghij", "level": "info", "message": "Order processed successfully", "timestamp": "2025-01-01T12:03:00Z"}'

list:
    LOCALLY
    ARG AWS_ENDPOINT_URL="http://localhost:4566"
    RUN aws sqs list-queues

all:
  BUILD +compile
  BUILD +container
