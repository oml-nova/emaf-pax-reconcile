package reconciliation

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	s3client "github.com/emaf-pax/pax-reconcile-service/internal/config/s3"
	service "github.com/emaf-pax/pax-reconcile-service/internal/services/reconciliation"
	logger "github.com/emaf-pax/pax-reconcile-service/pkg/superlog"
)

// s3Event is the S3 event notification format delivered via SQS.
type s3Event struct {
	Records []struct {
		S3 struct {
			Bucket struct {
				Name string `json:"name"`
			} `json:"bucket"`
			Object struct {
				Key string `json:"key"`
			} `json:"object"`
		} `json:"s3"`
	} `json:"Records"`
}

// HandleS3EventMessage is the SQS consumer handler for S3 event notifications.
// It is thin: parses the event, downloads from S3, and delegates to the service layer.
func HandleS3EventMessage(data []byte) error {
	var event s3Event
	if err := json.Unmarshal(data, &event); err != nil || len(event.Records) == 0 {
		logger.Log().Warn("Skipping non-S3 message", map[string]interface{}{
			"error": fmt.Sprintf("%v", err),
		})
		return nil // return nil to allow deletion of malformed messages
	}

	record := event.Records[0]
	bucket := record.S3.Bucket.Name
	key, _ := url.QueryUnescape(strings.ReplaceAll(record.S3.Object.Key, "+", " "))

	logger.Log().Info("Processing EMAF file", map[string]interface{}{
		"bucket": bucket,
		"key":    key,
	})

	ctx := context.Background()

	body, err := s3client.GetObject(ctx, bucket, key)
	if err != nil {
		return fmt.Errorf("downloading s3://%s/%s: %w", bucket, key, err)
	}

	if err := service.ProcessFile(ctx, body, key); err != nil {
		return fmt.Errorf("processing %s: %w", key, err)
	}

	return nil
}
