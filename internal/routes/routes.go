package routes

import (
	"nudge/internal/handlers"

	"github.com/gin-gonic/gin"
)

// SetupRouter configures all application routes
func SetupRouter() *gin.Engine {
	router := gin.Default()

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Auth routes
	auth := router.Group("/auth")
	{
		auth.POST("/register", handlers.Register)
		auth.POST("/login", handlers.Login)
	}

	return router
}
