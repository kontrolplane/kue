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
    RUN aws sqs create-queue --queue-name kontrolplane-payments-deadletter.fifo --attributes FifoQueue=true,ContentBasedDeduplication=true
    RUN aws sqs create-queue --queue-name kontrolplane-notifications-deadletter
    RUN aws sqs create-queue --queue-name kontrolplane-analytics-deadletter
    RUN aws sqs create-queue --queue-name kontrolplane-inventory-deadletter.fifo --attributes FifoQueue=true,ContentBasedDeduplication=true
    # Create main queues with DLQ references
    RUN aws sqs create-queue --queue-name kontrolplane-users --attributes "{\"RedrivePolicy\":\"{\\\"deadLetterTargetArn\\\":\\\"arn:aws:sqs:${REGION}:${ACCOUNT_ID}:kontrolplane-users-deadletter\\\",\\\"maxReceiveCount\\\":\\\"3\\\"}\"}"
    RUN aws sqs create-queue --queue-name kontrolplane-emails --attributes "{\"RedrivePolicy\":\"{\\\"deadLetterTargetArn\\\":\\\"arn:aws:sqs:${REGION}:${ACCOUNT_ID}:kontrolplane-emails-deadletter\\\",\\\"maxReceiveCount\\\":\\\"3\\\"}\"}"
    RUN aws sqs create-queue --queue-name kontrolplane-logs --attributes "{\"RedrivePolicy\":\"{\\\"deadLetterTargetArn\\\":\\\"arn:aws:sqs:${REGION}:${ACCOUNT_ID}:kontrolplane-logs-deadletter\\\",\\\"maxReceiveCount\\\":\\\"5\\\"}\"}"
    RUN aws sqs create-queue --queue-name kontrolplane-orders.fifo --attributes "{\"FifoQueue\":\"true\",\"ContentBasedDeduplication\":\"true\",\"RedrivePolicy\":\"{\\\"deadLetterTargetArn\\\":\\\"arn:aws:sqs:${REGION}:${ACCOUNT_ID}:kontrolplane-orders-deadletter.fifo\\\",\\\"maxReceiveCount\\\":\\\"5\\\"}\"}"
    RUN aws sqs create-queue --queue-name kontrolplane-shipments.fifo --attributes "{\"FifoQueue\":\"true\",\"ContentBasedDeduplication\":\"true\",\"RedrivePolicy\":\"{\\\"deadLetterTargetArn\\\":\\\"arn:aws:sqs:${REGION}:${ACCOUNT_ID}:kontrolplane-shipments-deadletter.fifo\\\",\\\"maxReceiveCount\\\":\\\"5\\\"}\"}"
    RUN aws sqs create-queue --queue-name kontrolplane-payments.fifo --attributes "{\"FifoQueue\":\"true\",\"ContentBasedDeduplication\":\"true\",\"RedrivePolicy\":\"{\\\"deadLetterTargetArn\\\":\\\"arn:aws:sqs:${REGION}:${ACCOUNT_ID}:kontrolplane-payments-deadletter.fifo\\\",\\\"maxReceiveCount\\\":\\\"5\\\"}\"}"
    RUN aws sqs create-queue --queue-name kontrolplane-notifications --attributes "{\"RedrivePolicy\":\"{\\\"deadLetterTargetArn\\\":\\\"arn:aws:sqs:${REGION}:${ACCOUNT_ID}:kontrolplane-notifications-deadletter\\\",\\\"maxReceiveCount\\\":\\\"3\\\"}\"}"
    RUN aws sqs create-queue --queue-name kontrolplane-analytics --attributes "{\"RedrivePolicy\":\"{\\\"deadLetterTargetArn\\\":\\\"arn:aws:sqs:${REGION}:${ACCOUNT_ID}:kontrolplane-analytics-deadletter\\\",\\\"maxReceiveCount\\\":\\\"5\\\"}\"}"
    RUN aws sqs create-queue --queue-name kontrolplane-inventory.fifo --attributes "{\"FifoQueue\":\"true\",\"ContentBasedDeduplication\":\"true\",\"RedrivePolicy\":\"{\\\"deadLetterTargetArn\\\":\\\"arn:aws:sqs:${REGION}:${ACCOUNT_ID}:kontrolplane-inventory-deadletter.fifo\\\",\\\"maxReceiveCount\\\":\\\"5\\\"}\"}"

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
    RUN aws sqs send-message --queue-url $AWS_ENDPOINT_URL/$ACCOUNT_ID/kontrolplane-users-deadletter --message-body '{"userId": "99999", "action": "create", "error": "Database connection timeout", "failedAt": "2025-01-01T11:55:00Z"}'
    RUN aws sqs send-message --queue-url $AWS_ENDPOINT_URL/$ACCOUNT_ID/kontrolplane-users-deadletter --message-body '{"userId": "88888", "action": "update", "error": "Validation failed: invalid email format", "failedAt": "2025-01-01T11:50:00Z"}'
    RUN aws sqs send-message --queue-url $AWS_ENDPOINT_URL/$ACCOUNT_ID/kontrolplane-emails-deadletter --message-body '{"emailId": "5555555555555", "recipient": "invalid@", "error": "Invalid recipient address", "failedAt": "2025-01-01T11:45:00Z"}'
    RUN aws sqs send-message --queue-url $AWS_ENDPOINT_URL/$ACCOUNT_ID/kontrolplane-orders-deadletter.fifo --message-body '{"orderId": "9999999999999", "status": "failed", "error": "Payment gateway unavailable", "failedAt": "2025-01-01T11:40:00Z"}' --message-group-id "dlq-01" --message-deduplication-id "dlq-dedup-001"
    # Payments messages
    RUN aws sqs send-message --queue-url $AWS_ENDPOINT_URL/$ACCOUNT_ID/kontrolplane-payments.fifo --message-body '{"paymentId": "pay_8a7b6c5d4e3f", "orderId": "1273897912902", "amount": 99.99, "currency": "USD", "status": "pending", "timestamp": "2025-01-01T12:05:00Z"}' --message-group-id "payments-01"
    RUN aws sqs send-message --queue-url $AWS_ENDPOINT_URL/$ACCOUNT_ID/kontrolplane-payments.fifo --message-body '{"paymentId": "pay_1f2e3d4c5b6a", "orderId": "1273897912903", "amount": 249.50, "currency": "USD", "status": "processing", "timestamp": "2025-01-01T12:06:00Z"}' --message-group-id "payments-01"
    RUN aws sqs send-message --queue-url $AWS_ENDPOINT_URL/$ACCOUNT_ID/kontrolplane-payments.fifo --message-body '{"paymentId": "pay_9z8y7x6w5v4u", "orderId": "1273897912904", "amount": 75.00, "currency": "EUR", "status": "pending", "timestamp": "2025-01-01T12:07:00Z"}' --message-group-id "payments-02"
    RUN aws sqs send-message --queue-url $AWS_ENDPOINT_URL/$ACCOUNT_ID/kontrolplane-payments-deadletter.fifo --message-body '{"paymentId": "pay_failed001", "orderId": "1273897900001", "amount": 500.00, "currency": "USD", "error": "Card declined", "failedAt": "2025-01-01T11:30:00Z"}' --message-group-id "dlq-payments" --message-deduplication-id "dlq-pay-001"
    # Shipments messages
    RUN aws sqs send-message --queue-url $AWS_ENDPOINT_URL/$ACCOUNT_ID/kontrolplane-shipments.fifo --message-body '{"shipmentId": "ship_001", "orderId": "1273897912902", "carrier": "FedEx", "trackingNumber": "FX123456789", "status": "in_transit", "timestamp": "2025-01-01T14:00:00Z"}' --message-group-id "shipments-01"
    RUN aws sqs send-message --queue-url $AWS_ENDPOINT_URL/$ACCOUNT_ID/kontrolplane-shipments.fifo --message-body '{"shipmentId": "ship_002", "orderId": "1273897912903", "carrier": "UPS", "trackingNumber": "1Z999AA10123456784", "status": "label_created", "timestamp": "2025-01-01T14:30:00Z"}' --message-group-id "shipments-01"
    # Notifications messages
    RUN aws sqs send-message --queue-url $AWS_ENDPOINT_URL/$ACCOUNT_ID/kontrolplane-notifications --message-body '{"notificationId": "notif_001", "userId": "20319", "type": "push", "title": "Order Shipped!", "body": "Your order is on its way", "timestamp": "2025-01-01T15:00:00Z"}'
    RUN aws sqs send-message --queue-url $AWS_ENDPOINT_URL/$ACCOUNT_ID/kontrolplane-notifications --message-body '{"notificationId": "notif_002", "userId": "23809", "type": "sms", "phone": "+1234567890", "message": "Your verification code is 123456", "timestamp": "2025-01-01T15:05:00Z"}'
    RUN aws sqs send-message --queue-url $AWS_ENDPOINT_URL/$ACCOUNT_ID/kontrolplane-notifications --message-body '{"notificationId": "notif_003", "userId": "21234", "type": "email", "subject": "Password Reset", "timestamp": "2025-01-01T15:10:00Z"}'
    # Analytics messages
    RUN aws sqs send-message --queue-url $AWS_ENDPOINT_URL/$ACCOUNT_ID/kontrolplane-analytics --message-body '{"eventId": "evt_001", "eventType": "page_view", "userId": "20319", "page": "/products/123", "timestamp": "2025-01-01T16:00:00Z"}'
    RUN aws sqs send-message --queue-url $AWS_ENDPOINT_URL/$ACCOUNT_ID/kontrolplane-analytics --message-body '{"eventId": "evt_002", "eventType": "add_to_cart", "userId": "20319", "productId": "prod_456", "quantity": 2, "timestamp": "2025-01-01T16:01:00Z"}'
    RUN aws sqs send-message --queue-url $AWS_ENDPOINT_URL/$ACCOUNT_ID/kontrolplane-analytics --message-body '{"eventId": "evt_003", "eventType": "checkout_started", "userId": "20319", "cartTotal": 99.99, "timestamp": "2025-01-01T16:02:00Z"}'
    RUN aws sqs send-message --queue-url $AWS_ENDPOINT_URL/$ACCOUNT_ID/kontrolplane-analytics --message-body '{"eventId": "evt_004", "eventType": "purchase_completed", "userId": "20319", "orderId": "1273897912902", "timestamp": "2025-01-01T16:05:00Z"}'
    # Inventory messages
    RUN aws sqs send-message --queue-url $AWS_ENDPOINT_URL/$ACCOUNT_ID/kontrolplane-inventory.fifo --message-body '{"inventoryId": "inv_001", "productId": "prod_456", "warehouseId": "wh_east", "action": "decrement", "quantity": 2, "timestamp": "2025-01-01T16:10:00Z"}' --message-group-id "inventory-01"
    RUN aws sqs send-message --queue-url $AWS_ENDPOINT_URL/$ACCOUNT_ID/kontrolplane-inventory.fifo --message-body '{"inventoryId": "inv_002", "productId": "prod_789", "warehouseId": "wh_west", "action": "restock", "quantity": 100, "timestamp": "2025-01-01T16:15:00Z"}' --message-group-id "inventory-01"
    # More logs messages
    RUN aws sqs send-message --queue-url $AWS_ENDPOINT_URL/$ACCOUNT_ID/kontrolplane-logs --message-body '{"logId": "log_001", "level": "warn", "service": "payment-service", "message": "High latency detected", "latencyMs": 2500, "timestamp": "2025-01-01T12:10:00Z"}'
    RUN aws sqs send-message --queue-url $AWS_ENDPOINT_URL/$ACCOUNT_ID/kontrolplane-logs --message-body '{"logId": "log_002", "level": "error", "service": "inventory-service", "message": "Failed to connect to database", "retryCount": 3, "timestamp": "2025-01-01T12:15:00Z"}'
    RUN aws sqs send-message --queue-url $AWS_ENDPOINT_URL/$ACCOUNT_ID/kontrolplane-logs --message-body '{"logId": "log_003", "level": "info", "service": "api-gateway", "message": "Request processed", "requestId": "req_abc123", "timestamp": "2025-01-01T12:20:00Z"}'

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
