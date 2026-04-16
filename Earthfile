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

seed:
    LOCALLY
    ARG AWS_ENDPOINT_URL="http://localhost:4566"
    ARG REGION="us-east-1"
    ARG ACCOUNT_ID="000000000000"
    RUN AWS_ENDPOINT_URL=$AWS_ENDPOINT_URL REGION=$REGION ACCOUNT_ID=$ACCOUNT_ID bash seed/seed.sh

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
