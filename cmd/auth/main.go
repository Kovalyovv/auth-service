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
	"github.com/Kovalyovv/auth-service/pkg/observability"
	"github.com/Kovalyovv/auth-service/pkg/pb"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

const serviceName = "auth-service"

func main() {
	tp, err := observability.InitTracer(serviceName, "jaeger:4317")
	if err != nil {
		slog.Error("failed to initialize tracer", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			slog.Error("failed to shutdown tracer", "error", err)
		}
	}()

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

	var kaep = keepalive.EnforcementPolicy{
		MinTime:             5 * time.Second,
		PermitWithoutStream: true,
	}
	var kasp = keepalive.ServerParameters{
		Time:    15 * time.Second,
		Timeout: 5 * time.Second,
	}

	grpcSrv := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.KeepaliveEnforcementPolicy(kaep),
		grpc.KeepaliveParams(kasp),
	)
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

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(otelgin.Middleware(serviceName))

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	handler := deliveryHTTP.NewAuthHandler(authUC)
	deliveryHTTP.SetupRoutes(router, handler)
	httpSrv := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: router,
	}

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
