package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Kovalyovv/auth-service/internal/config"
	deliveryGRPC "github.com/Kovalyovv/auth-service/internal/delivery/grpc"
	deliveryHTTP "github.com/Kovalyovv/auth-service/internal/delivery/http"
	"github.com/Kovalyovv/auth-service/internal/pkg/jwt"
	"github.com/Kovalyovv/auth-service/internal/repository/postgres"
	"github.com/Kovalyovv/auth-service/internal/usecase"
	"github.com/Kovalyovv/auth-service/pkg/pb"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.NewFromEnv()

	if cfg.JWTSecret == "" {
		slog.Error("missing critical configuration: JWT_SECRET must be set")
		os.Exit(1)
	}
	if cfg.DatabaseURL == "" {
		slog.Error("missing critical configuration: DATABASE_URL must be set")
		os.Exit(1)
	}

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to db", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	userRepo := postgres.NewUserRepo(pool)
	tokenManager := jwt.NewTokenManager(cfg.JWTSecret)
	authUC := usecase.NewAuthUseCase(userRepo, tokenManager, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	grpcSrv := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcSrv, deliveryGRPC.NewServer(authUC))

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		slog.Error("failed to listen grpc", "error", err)
		os.Exit(1)
	}

	go func() {
		slog.Info("gRPC server listening", "port", cfg.GRPCPort)
		if err := grpcSrv.Serve(lis); err != nil {
			slog.Error("grpc serve err", "error", err)
		}
	}()

	authHandler := deliveryHTTP.NewAuthHandler(authUC)
	httpSrv := deliveryHTTP.NewServer(cfg, authHandler)

	go func() {
		slog.Info("HTTP server listening on", "port", cfg.HTTPPort)
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http listen err", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	grpcSrv.GracefulStop()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(ctx)
}
