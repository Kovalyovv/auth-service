package http

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	conf "github.com/Kovalyovv/auth-service/internal/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewServer(cfg *conf.Config, handler *AuthHandler) *http.Server {
	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowCredentials = true
	config.AllowOriginFunc = func(origin string) bool {
		u, err := url.Parse(origin)
		if err != nil {
			return false
		}
		hostname := strings.Trim(u.Hostname(), "[]")
		return (hostname == "localhost" || hostname == "0.0.0.0" || hostname == "127.0.0.1" || hostname == "::1") && (u.Port() == "9000" || u.Port() == "9002" || u.Port() == "9001")
	}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	config.MaxAge = 12 * time.Hour

	router.Use(cors.New(config))

	auth := router.Group("/auth")
	{
		auth.POST("/register", handler.Register)
		auth.POST("/login", handler.Login)
		auth.POST("/refresh", handler.Refresh)
	}

	return &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.HTTPPort),
		Handler: router,
	}
}
