package routes

import (
	"github.com/gin-gonic/gin"

	reconciliation "github.com/emaf-pax/pax-reconcile-service/internal/handlers/reconciliation"
)

func InitReconciliationRoutes(r *gin.Engine) {
	group := r.Group("/reconciliation")
	group.POST("/trigger", reconciliation.TriggerFileProcessing)
}
