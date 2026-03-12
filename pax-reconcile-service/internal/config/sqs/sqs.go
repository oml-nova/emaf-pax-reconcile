package sqs

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	logger "github.com/emaf-pax/pax-reconcile-service/pkg/superlog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type MessageHandler func(message []byte) error

type Service struct {
	client   *sqs.Client
	queueURL string
	handlers map[string]MessageHandler
}

var (
	sqsServiceInstance *Service
	once               sync.Once
)

func GetSQSService() *Service {
	once.Do(func() {
		region := os.Getenv("AWS_REGION")
		queueURL := os.Getenv("SOURCE_QUEUE_URI")

		cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
		if err != nil {
			fmt.Printf("Error creating AWS config: %v\n", err)
			return
		}
		sqsServiceInstance = &Service{
			client:   sqs.NewFromConfig(cfg),
			queueURL: queueURL,
			handlers: make(map[string]MessageHandler),
		}
	})
	return sqsServiceInstance
}

func (s *Service) RegisterHandler(messageType string, handler MessageHandler) {
	s.handlers[messageType] = handler
}

// ReceiveMessages long-polls the queue and returns up to 10 messages.
func (s *Service) ReceiveMessages(ctx context.Context) ([]types.Message, error) {
	out, err := s.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(s.queueURL),
		MaxNumberOfMessages: 1,
		WaitTimeSeconds:     20,
	})
	if err != nil {
		return nil, fmt.Errorf("SQS receive: %w", err)
	}
	return out.Messages, nil
}

// DeleteMessage removes a successfully processed message from the queue.
func (s *Service) DeleteMessage(ctx context.Context, receiptHandle string) {
	_, err := s.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(s.queueURL),
		ReceiptHandle: aws.String(receiptHandle),
	})
	if err != nil {
		logger.Log().Warn("Failed to delete SQS message", map[string]interface{}{"error": err.Error()})
	}
}

// StartProcessing is a blocking long-poll loop that dispatches messages to registered handlers.
func (s *Service) StartProcessing(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		messages, err := s.ReceiveMessages(ctx)
		if err != nil {
			logger.Log().Error("Error receiving SQS messages", map[string]interface{}{"error": err.Error()})
			continue
		}

		for _, msg := range messages {
			var wrapper struct {
				Type    string          `json:"type"`
				Payload json.RawMessage `json:"payload"`
			}

			body := *msg.Body

			// Try typed dispatch first
			if err := json.Unmarshal([]byte(body), &wrapper); err == nil && wrapper.Type != "" {
				handler, ok := s.handlers[wrapper.Type]
				if !ok {
					logger.Log().Error("No handler for SQS message type", map[string]interface{}{"type": wrapper.Type})
					continue
				}
				if err := handler(wrapper.Payload); err != nil {
					logger.Log().Error("Handler failed", map[string]interface{}{"error": err.Error(), "type": wrapper.Type})
					continue
				}
				logger.Log().Debug("SQS message processed successfully", map[string]interface{}{"type": wrapper.Type})
				s.DeleteMessage(ctx, *msg.ReceiptHandle)
				continue
			}

			// Fallback: dispatch raw message to default handler
			if handler, ok := s.handlers["default"]; ok {
				if err := handler([]byte(body)); err != nil {
					logger.Log().Error("Default handler failed", map[string]interface{}{"error": err.Error()})
					continue
				}
				logger.Log().Debug("SQS default message processed successfully", nil)
				s.DeleteMessage(ctx, *msg.ReceiptHandle)
			}
		}
	}
}
