package kue

import (
	"time"
)

// docs: https://github.com/aws/aws-sdk-go-v2/blob/service/sqs/v1.37.15/service/sqs/types/enums.go#L58
type Queue struct {
	Url                                   string            `json:"url"`
	Name                                  string            `json:"name"`
	Arn                                   string            `json:"arn"`
	CreatedTimestamp                      string            `json:"created_timestamp"`
	LastModified                          string            `json:"last_modified"`
	DelaySeconds                          string            `json:"delay_seconds"`
	MaxMessageSize                        string            `json:"max_message_size"`
	MessageRetentionPeriod                string            `json:"message_retention_period"`
	ReceiveMessageWaitTime                string            `json:"receive_message_wait_time"`
	VisibilityTimeout                     string            `json:"visibility_timeout"`
	ApproximateNumberOfMessages           string            `json:"approximate_number_of_messages"`
	ApproximateNumberOfMessagesNotVisible string            `json:"approximate_number_of_messages_not_visible"`
	ApproximateNumberOfMessagesDelayed    string            `json:"approximate_number_of_messages_delayed"`
	RedrivePolicy                         string            `json:"redrive_policy,omitempty"`
	RedriveAllowPolicy                    string            `json:"redrive_allow_policy,omitempty"`
	DeadLetterTargetARN                   string            `json:"dead_letter_target_arn"`
	FifoQueue                             string            `json:"fifo_queue"`
	ContentBasedDeduplication             string            `json:"content_based_deduplication,omitempty"`
	DeduplicationScope                    string            `json:"deduplication_scope,omitempty"`
	Tags                                  map[string]string `json:"tags,omitempty"`
}

type DeadLetterAttributes struct {
	RedrivePolicy       string `json:"redrive_policy,omitempty"`
	DeadLetterTargetARN string `json:"dead_letter_target_arn,omitempty"`
}

type Message struct {
	QueueName              string            `json:"queue_name"`
	MessageID              string            `json:"message_id"`
	Body                   string            `json:"body"`
	MD5OfBody              string            `json:"md5_of_body"`
	Attributes             map[string]string `json:"attributes,omitempty"`
	MessageAttributes      map[string]string `json:"message_attributes,omitempty"`
	ReceiptHandle          string            `json:"receipt_handle"`
	FirstReceiveTime       time.Time         `json:"first_receive_time,omitempty"`
	ReceiveCount           int               `json:"receive_count"`
	SentTimestamp          time.Time         `json:"sent_timestamp"`
	DelaySeconds           int               `json:"delay_seconds,omitempty"`
	VisibilityTimeout      int               `json:"visibility_timeout,omitempty"`
	MessageGroupID         string            `json:"message_group_id,omitempty"`
	MessageDeduplicationID string            `json:"message_deduplication_id,omitempty"`
	SequenceNumber         string            `json:"sequence_number,omitempty"`
}
