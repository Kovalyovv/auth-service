package http

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, handler *AuthHandler) {
	// CORS middleware can be applied here or in main.go. Let's keep it here.
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:9000", "http://127.0.0.1:9000", "http://[::1]:9000", "http://0.0.0.0:9000"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	auth := router.Group("/auth")
	{
		auth.POST("/register", handler.Register)
		auth.POST("/login", handler.Login)
		auth.POST("/refresh", handler.Refresh)
	}
}
