package routes

import (
	"nudge/internal/handlers"
	"nudge/internal/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRouter configures all application routes
func SetupRouter() *gin.Engine {
	router := gin.Default()

	// Add logging middleware to log all requests and responses
	router.Use(middleware.LoggingMiddleware())

	// Swagger UI (development)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Auth routes (public)
	auth := router.Group("/auth")
	{
		auth.POST("/register", handlers.Register)
		auth.POST("/login", handlers.Login)
	}

	// Task routes (protected)
	tasks := router.Group("/tasks")
	tasks.Use(middleware.AuthMiddleware())
	{
		tasks.POST("", handlers.CreateTask)
		tasks.GET("", handlers.GetTasks)
		tasks.GET("/:id", handlers.GetTask)
		tasks.PUT("/:id", handlers.UpdateTask)
		tasks.DELETE("/:id", handlers.DeleteTask)
	}

	// Recurring task routes (protected)
	recurringTasks := router.Group("/recurring-tasks")
	recurringTasks.Use(middleware.AuthMiddleware())
	{
		recurringTasks.POST("", handlers.CreateRecurringTask)
		recurringTasks.GET("", handlers.GetRecurringTasks)
		recurringTasks.GET("/:id", handlers.GetRecurringTask)
		recurringTasks.PUT("/:id", handlers.UpdateRecurringTask)
		recurringTasks.DELETE("/:id", handlers.DeleteRecurringTask)
	}

	return router
}
