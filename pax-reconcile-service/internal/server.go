package app

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/emaf-pax/pax-reconcile-service/internal/routes"
	logger "github.com/emaf-pax/pax-reconcile-service/pkg/superlog"
)

type GinEngine struct {
	*gin.Engine
}

func setupGin() GinEngine {
	r := gin.Default()

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"*"}
	corsConfig.ExposeHeaders = []string{"Content-Length"}
	corsConfig.AllowCredentials = true
	r.Use(cors.New(corsConfig))

	return GinEngine{Engine: r}
}

func (r GinEngine) addMiddleware() GinEngine {
	r.Use(logger.GinMiddleware())
	r.Use(gin.Recovery())
	return r
}

func (r GinEngine) addRoutes() GinEngine {
	routes.InitReconciliationRoutes(r.Engine)
	return r
}

func RegisterRoutes() GinEngine {
	r := setupGin()
	r.RedirectFixedPath = false
	r.RedirectTrailingSlash = false
	r.addMiddleware()
	r.addRoutes()
	return r
}

func (r GinEngine) StartServer() {
	if err := r.Run(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
