package http

import (
	"fmt"
	"net/http"

	"github.com/Kovalyovv/auth-service/internal/config"
	"github.com/gin-gonic/gin"
)

func NewServer(cfg *config.Config, handler *AuthHandler) *http.Server {
	router := gin.Default()

	auth := router.Group("/auth")
	{
		auth.POST("/register", handler.Register)
		auth.POST("/login", handler.Login)
	}

	return &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.HTTPPort),
		Handler: router,
	}
}
