package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPPort    string
	GRPCPort    string
	DatabaseURL string
	JWTSecret   string
}

func NewFromEnv() *Config {
	_ = godotenv.Load()

	return &Config{
		HTTPPort:    getEnv("HTTP_PORT", "8001"),
		GRPCPort:    getEnv("GRPC_PORT", "50001"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
