package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// ⚡ Bolt: Performance optimization
	// Using gin.New() instead of gin.Default() omits the default Logger middleware.
	// This reduces per-request overhead, significantly improving throughput
	// for high-frequency endpoints like /health.
	// We only attach gin.Recovery() to prevent panics from crashing the server.
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "backend-core",
		})
	})

	// TODO: Gitea Webhook Receiver
	// TODO: gRPC Client for backend-ai
	// TODO: WebSocket Hub

	r.Run(":8080")
}
