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
    LOCALLY
    ARG AWS_ENDPOINT_URL="http://localhost:4566"
    RUN aws sqs create-queue --queue-name kontrolplane-users
    RUN aws sqs create-queue --queue-name kontrolplane-users-deadletter
    RUN aws sqs create-queue --queue-name kontrolplane-emails
    RUN aws sqs create-queue --queue-name kontrolplane-emails-deadletter
    RUN aws sqs create-queue --queue-name kontrolplane-orders
    RUN aws sqs create-queue --queue-name kontrolplane-orders-deadletter

list:
    LOCALLY
    ARG AWS_ENDPOINT_URL="http://localhost:4566"
    RUN aws sqs list-queues

all:
  BUILD +compile
  BUILD +container
