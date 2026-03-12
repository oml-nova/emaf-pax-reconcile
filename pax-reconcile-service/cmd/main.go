package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	app "github.com/emaf-pax/pax-reconcile-service/internal"
	"github.com/emaf-pax/pax-reconcile-service/internal/config"
	database "github.com/emaf-pax/pax-reconcile-service/internal/config/databases"
	s3client "github.com/emaf-pax/pax-reconcile-service/internal/config/s3"
	service "github.com/emaf-pax/pax-reconcile-service/internal/services/reconciliation"
	logger "github.com/emaf-pax/pax-reconcile-service/pkg/superlog"
)

func main() {
	// STEP 1 — Hydrate env vars from remote config service
	err := config.LoadEnvironment()

	// STEP 2 — Init logger BEFORE checking err so panics are logged
	isRemoteLoggerEnabled := strings.ToLower(os.Getenv("REMOTE_LOG")) == "true"
	if isRemoteLoggerEnabled {
		remoteLoggerURL := os.Getenv("LOGGER_HOST")
		logger.NewLogger(
			&logger.ConsoleShipper{},
			&logger.HTTPShipper{URL: remoteLoggerURL},
		)
	} else {
		logger.NewLogger(
			&logger.ConsoleShipper{},
			&logger.FileShipper{FilePath: "app.log"},
		)
	}
	fmt.Println("Remote Logger Enabled:", isRemoteLoggerEnabled)

	if err != nil {
		panic(err.Error())
	}

	// STEP 3 — Connect databases
	err = database.InitMongoDB()
	if err != nil {
		panic(err.Error())
	}

	// STEP 4 — Init S3 client
	err = s3client.InitS3()
	if err != nil {
		panic(err.Error())
	}

	// STEP 5 — Load merchant mappings into memory
	ctx := context.Background()
	err = service.LoadMerchantMappings(ctx)
	if err != nil {
		panic(err.Error())
	}

	// STEP 6 — Start SQS consumer (blocks in goroutine)
	logger.Log().Info("Reconcile service started. Polling for S3 event messages...", nil)

	go func() {
		app.InitSQSConsumer(ctx)
	}()

	// Block forever (consumer runs in background goroutine)
	select {}
}
