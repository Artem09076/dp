package app

import (
	"database/sql"
	"log/slog"
	"time"

	_ "github.com/lib/pq"

	grpcapp "github.com/Artem09076/dp/backend/auth_service/internal/app/grpc"
	auth "github.com/Artem09076/dp/backend/auth_service/internal/presentation"
	sqlc "github.com/Artem09076/dp/backend/auth_service/internal/storage/db"
	"github.com/Artem09076/dp/backend/auth_service/internal/storage/redis"
)

type App struct {
	GRPCServ *grpcapp.App
}

func New(log *slog.Logger, grpcPort uint, redisAddr, redisPassword string, redisDB int, storagePath string, tokenAccessTTL, tokenRefreshTTL time.Duration, tokenSecret string) *App {
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
	return &App{
		GRPCServ: grpcApp,
	}
}
