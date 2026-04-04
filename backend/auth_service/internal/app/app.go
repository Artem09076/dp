package app

import (
	"database/sql"
	"log/slog"
	"time"

	_ "github.com/lib/pq"

	grpcapp "github.com/Artem09076/dp/backend/auth_service/internal/app/grpc"
	auth "github.com/Artem09076/dp/backend/auth_service/internal/presentation"
	sqlc "github.com/Artem09076/dp/backend/auth_service/internal/storage/db"
)

type App struct {
	GRPCServ *grpcapp.App
}

func New(log *slog.Logger, grpcPort uint, storagePath string, tokenTTL time.Duration) *App {
	db, err := sql.Open("postgres", storagePath)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		panic(err)
	}

	queries := sqlc.New(db)

	authService := auth.New(log, queries, tokenTTL)
	grpcApp := grpcapp.New(log, authService, grpcPort)
	return &App{
		GRPCServ: grpcApp,
	}
}
