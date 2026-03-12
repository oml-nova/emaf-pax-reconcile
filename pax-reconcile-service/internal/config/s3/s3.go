package s3

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	s3ClientInstance *s3.Client
	once             sync.Once
)

// InitS3 initialises the S3 client singleton.
func InitS3() error {
	var initErr error
	once.Do(func() {
		region := os.Getenv("AWS_REGION")
		cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
		if err != nil {
			initErr = fmt.Errorf("loading AWS config for S3: %w", err)
			return
		}
		s3ClientInstance = s3.NewFromConfig(cfg)
	})
	return initErr
}

// GetObject downloads an S3 object and returns a reader. Caller must close it.
func GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	if s3ClientInstance == nil {
		return nil, fmt.Errorf("S3 client not initialised — call InitS3() first")
	}
	out, err := s3ClientInstance.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("S3 GetObject s3://%s/%s: %w", bucket, key, err)
	}
	return out.Body, nil
}
