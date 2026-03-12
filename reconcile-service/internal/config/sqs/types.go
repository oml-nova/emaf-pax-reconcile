package sqs

// ConsumerEventTypes for SQS message dispatch.
type ConsumerEventTypes string

const (
	ConsumerEventS3Notification ConsumerEventTypes = "S3_NOTIFICATION"
)
