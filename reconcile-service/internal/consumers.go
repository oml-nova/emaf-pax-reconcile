package app

import (
	"context"

	"github.com/emaf-pax/reconcile-service/internal/config/sqs"
	reconciliation "github.com/emaf-pax/reconcile-service/internal/handlers/reconciliation"
	logger "github.com/emaf-pax/reconcile-service/pkg/superlog"
)

// InitSQSConsumer registers all SQS message handlers and starts the processing loop.
func InitSQSConsumer(ctx context.Context) {
	sqsService := sqs.GetSQSService()

	// Register the default handler for S3 event notifications.
	// Messages arrive as raw S3 events (not typed wrappers), so we use "default".
	sqsService.RegisterHandler("default", reconciliation.HandleS3EventMessage)

	logger.Log().Info("SQS Consumer Initialized", nil)

	// StartProcessing blocks — run in a goroutine from main.
	sqsService.StartProcessing(ctx)
}
