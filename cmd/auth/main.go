package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Kovalyovv/auth-service/internal/config"
	deliveryGRPC "github.com/Kovalyovv/auth-service/internal/delivery/grpc"
	"github.com/Kovalyovv/auth-service/internal/delivery/grpc/pb"
	deliveryHTTP "github.com/Kovalyovv/auth-service/internal/delivery/http"
	"github.com/Kovalyovv/auth-service/internal/pkg/jwt"
	"github.com/Kovalyovv/auth-service/internal/repository/postgres"
	"github.com/Kovalyovv/auth-service/internal/usecase"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

func main() {
	cfg := config.NewFromEnv()

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer pool.Close()

	userRepo := postgres.NewUserRepo(pool)
	tokenManager := jwt.NewTokenManager(cfg.JWTSecret)
	authUC := usecase.NewAuthUseCase(userRepo, tokenManager)

	grpcSrv := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcSrv, deliveryGRPC.NewServer(authUC))

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("failed to listen grpc: %v", err)
	}

	go func() {
		log.Printf("gRPC server listening on %s", cfg.GRPCPort)
		if err := grpcSrv.Serve(lis); err != nil {
			log.Fatalf("grpc serve err: %v", err)
		}
	}()

	authHandler := deliveryHTTP.NewAuthHandler(authUC)
	httpSrv := deliveryHTTP.NewServer(cfg, authHandler)

	go func() {
		log.Printf("HTTP server listening on %s", cfg.HTTPPort)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http listen err: %v", err)
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
