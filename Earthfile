VERSION 0.8
FROM golang:1.26-alpine
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
    RUN AWS_ENDPOINT_URL=$AWS_ENDPOINT_URL AWS_REGION=$REGION AWS_DEFAULT_REGION=$REGION AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test AWS_PAGER="" ACCOUNT_ID=$ACCOUNT_ID bash seed/seed.sh

localstack:
    LOCALLY
    ARG AWS_ENDPOINT_URL="http://localhost:4566"
    RUN docker compose up -d
    RUN for i in $(seq 1 60); do curl -fs $AWS_ENDPOINT_URL/_localstack/health | grep -q '"sqs": "\(available\|running\)"' && exit 0; sleep 1; done; echo "localstack did not become ready" && exit 1

dev:
    LOCALLY
    ARG AWS_ENDPOINT_URL="http://localhost:4566"
    ARG REGION="us-east-1"
    ARG ACCOUNT_ID="000000000000"
    WAIT
        BUILD +localstack
    END
    RUN AWS_ENDPOINT_URL=$AWS_ENDPOINT_URL AWS_REGION=$REGION AWS_DEFAULT_REGION=$REGION AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test AWS_PAGER="" ACCOUNT_ID=$ACCOUNT_ID bash seed/seed.sh
    RUN go build -o build/kontrolplane/kue .

list:
    LOCALLY
    ARG AWS_ENDPOINT_URL="http://localhost:4566"
    ARG REGION="us-east-1"
    RUN AWS_ENDPOINT_URL=$AWS_ENDPOINT_URL AWS_REGION=$REGION AWS_DEFAULT_REGION=$REGION AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test AWS_PAGER="" aws sqs list-queues

vhs:
    LOCALLY
    ARG AWS_ENDPOINT_URL="http://localhost:4566"
    ARG REGION="us-east-1"
    RUN AWS_ENDPOINT_URL=$AWS_ENDPOINT_URL AWS_REGION=$REGION AWS_DEFAULT_REGION=$REGION AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test AWS_PAGER="" vhs vhs/cassette.tape

all:
  BUILD +compile
  BUILD +container
