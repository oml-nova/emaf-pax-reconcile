package reconciliation

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	s3client "github.com/emaf-pax/pax-reconcile-service/internal/config/s3"
	service "github.com/emaf-pax/pax-reconcile-service/internal/services/reconciliation"
	logger "github.com/emaf-pax/pax-reconcile-service/pkg/superlog"
)

type triggerRequest struct {
	Bucket string `json:"bucket" binding:"required"`
	Key    string `json:"key"    binding:"required"`
}

// TriggerFileProcessing manually triggers EMAF file processing for a given S3 object.
// POST /reconciliation/trigger
// Body: { "bucket": "my-bucket", "key": "settlements/EMAF_20240312.txt" }
func TriggerFileProcessing(c *gin.Context) {
	var req triggerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logger.Log().Info("Manual trigger received", map[string]interface{}{
		"bucket": req.Bucket,
		"key":    req.Key,
	})

	ctx := context.Background()

	body, err := s3client.GetObject(ctx, req.Bucket, req.Key)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":  "failed to download from S3: " + err.Error(),
			"bucket": req.Bucket,
			"key":    req.Key,
		})
		return
	}

	summary, err := service.ProcessFile(ctx, body, req.Key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "processing failed: " + err.Error(),
			"bucket": req.Bucket,
			"key":    req.Key,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "file processed successfully",
		"bucket":  req.Bucket,
		"key":     req.Key,
		"summary": summary,
	})
}

// UploadFileProcessing accepts a multipart file upload and processes it directly.
// POST /reconciliation/upload
// Form: file=<multipart file>
func UploadFileProcessing(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required: " + err.Error()})
		return
	}

	f, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open file: " + err.Error()})
		return
	}
	defer f.Close()

	body, err := io.ReadAll(f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file: " + err.Error()})
		return
	}

	logger.Log().Info("File upload trigger received", map[string]interface{}{
		"filename": fileHeader.Filename,
	})

	ctx := context.Background()

	summary, err := service.ProcessFile(ctx, io.NopCloser(bytes.NewReader(body)), fileHeader.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":    "processing failed: " + err.Error(),
			"filename": fileHeader.Filename,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "file processed successfully",
		"filename": fileHeader.Filename,
		"summary":  summary,
	})
}
