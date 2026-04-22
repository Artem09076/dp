package app

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	_ "github.com/lib/pq"

	gatewayapp "github.com/Artem09076/dp/backend/auth_service/internal/app/gateway"
	grpcapp "github.com/Artem09076/dp/backend/auth_service/internal/app/grpc"
	auth "github.com/Artem09076/dp/backend/auth_service/internal/presentation"
	sqlc "github.com/Artem09076/dp/backend/auth_service/internal/storage/db"
	"github.com/Artem09076/dp/backend/auth_service/internal/storage/redis"
)

type App struct {
	GRPCServ      *grpcapp.App
	gatewayServer *gatewayapp.App
	log           *slog.Logger
}

func New(log *slog.Logger, grpcPort uint, gatewayPort uint, redisAddr, redisPassword string, redisDB int, storagePath string, tokenAccessTTL, tokenRefreshTTL time.Duration, tokenSecret string) *App {
	db, err := sql.Open("postgres", storagePath)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		panic(err)
	}

	queries := sqlc.New(db)

	redisClient, err := redis.NewRedisClient(log, redisAddr, redisPassword, redisDB)
	if err != nil {
		log.Error("failed to connect to redis", "error", err)
		panic(err)
	}

	authService := auth.New(log, queries, redisClient, tokenAccessTTL, tokenRefreshTTL, tokenSecret, 5)
	grpcApp := grpcapp.New(log, authService, []byte(tokenSecret), grpcPort)
	gatewayApp := gatewayapp.New(log, authService, grpcPort, gatewayPort, []byte(tokenSecret))
	return &App{
		GRPCServ:      grpcApp,
		gatewayServer: gatewayApp,
		log:           log,
	}
}

func (a *App) RunGRPCServer() error {
	return a.GRPCServ.Start()
}

func (a *App) RunGatewayServer() error {
	return a.gatewayServer.Start()
}

func (a *App) Stop(ctx context.Context) {
	a.log.Info("stopping servers...")

	if a.gatewayServer != nil {
		a.gatewayServer.Stop(ctx)
	}

	if a.GRPCServ != nil {
		a.GRPCServ.Stop()
	}
}
