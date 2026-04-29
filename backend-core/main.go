package main

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"service": "backend-core",
		})
	})

	// TODO: Gitea Webhook Receiver
	// TODO: gRPC Client for backend-ai
	// TODO: WebSocket Hub

	r.Run(":8080")
}
