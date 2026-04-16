#!/usr/bin/env bash
set -euo pipefail

AWS_ENDPOINT_URL="${AWS_ENDPOINT_URL:-http://localhost:4566}"
REGION="${REGION:-us-east-1}"
ACCOUNT_ID="${ACCOUNT_ID:-000000000000}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

for file in "$SCRIPT_DIR"/queues/*.json; do
  # --- Create deadletter queue ---
  dlq_name=$(jq -r '.deadletter.name // empty' "$file")
  if [ -n "$dlq_name" ]; then
    dlq_attrs=$(jq -c '.deadletter.attributes // {}' "$file")
    if [ "$dlq_attrs" = "{}" ]; then
      aws sqs create-queue --queue-name "$dlq_name"
    else
      aws sqs create-queue --queue-name "$dlq_name" --attributes "$dlq_attrs"
    fi
  fi

  # --- Create main queue ---
  queue_name=$(jq -r '.queue.name' "$file")
  max_receive=$(jq -r '.deadletter.maxReceiveCount // empty' "$file")

  base_attrs=$(jq -c '.queue.attributes // {}' "$file")
  if [ -n "$dlq_name" ] && [ -n "$max_receive" ]; then
    dlq_arn="arn:aws:sqs:${REGION}:${ACCOUNT_ID}:${dlq_name}"
    redrive_policy="{\"deadLetterTargetArn\":\"${dlq_arn}\",\"maxReceiveCount\":\"${max_receive}\"}"
    attrs=$(echo "$base_attrs" | jq -c --arg rp "$redrive_policy" '. + {"RedrivePolicy": $rp}')
  else
    attrs="$base_attrs"
  fi

  if [ "$attrs" = "{}" ]; then
    aws sqs create-queue --queue-name "$queue_name"
  else
    aws sqs create-queue --queue-name "$queue_name" --attributes "$attrs"
  fi

  # --- Send messages to main queue ---
  queue_url="${AWS_ENDPOINT_URL}/${ACCOUNT_ID}/${queue_name}"
  jq -c '.messages // [] | .[]' "$file" | while IFS= read -r msg; do
    msg_body=$(echo "$msg" | jq -c '.body')
    extra_args=()
    group_id=$(echo "$msg" | jq -r '.messageGroupId // empty')
    dedup_id=$(echo "$msg" | jq -r '.messageDeduplicationId // empty')
    [ -n "$group_id" ] && extra_args+=(--message-group-id "$group_id")
    [ -n "$dedup_id" ] && extra_args+=(--message-deduplication-id "$dedup_id")
    aws sqs send-message --queue-url "$queue_url" --message-body "$msg_body" "${extra_args[@]+"${extra_args[@]}"}"
  done

  # --- Send messages to deadletter queue ---
  if [ -n "$dlq_name" ]; then
    dlq_url="${AWS_ENDPOINT_URL}/${ACCOUNT_ID}/${dlq_name}"
    jq -c '.deadletterMessages // [] | .[]' "$file" | while IFS= read -r msg; do
      msg_body=$(echo "$msg" | jq -c '.body')
      extra_args=()
      group_id=$(echo "$msg" | jq -r '.messageGroupId // empty')
      dedup_id=$(echo "$msg" | jq -r '.messageDeduplicationId // empty')
      [ -n "$group_id" ] && extra_args+=(--message-group-id "$group_id")
      [ -n "$dedup_id" ] && extra_args+=(--message-deduplication-id "$dedup_id")
      aws sqs send-message --queue-url "$dlq_url" --message-body "$msg_body" "${extra_args[@]+"${extra_args[@]}"}"
    done
  fi
done
