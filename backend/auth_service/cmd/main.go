package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/Artem09076/dp/backend/auth_service/internal/app"
	"github.com/Artem09076/dp/backend/auth_service/internal/config"
	"github.com/Artem09076/dp/backend/auth_service/internal/logger"
)

func main() {
	cfg := config.LoadConfig()
	log := logger.SetupLogger(cfg.Env)
	appl := app.New(log, cfg.GRPC.Port, cfg.DBPath, cfg.TokenTTL)

	go appl.GRPCServ.Run()
	end := make(chan os.Signal, 1)
	signal.Notify(end, syscall.SIGINT, syscall.SIGTERM)
	log.Info("auth service started")
	<-end
	appl.GRPCServ.Stop()
	log.Info("auth service stopped")
}
