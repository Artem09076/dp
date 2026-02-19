package grpcapp

import (
	"fmt"
	"log/slog"
	"net"

	authgrpc "github.com/Artem09076/dp/backend/auth_service/internal/grpc"
	"google.golang.org/grpc"
)

type App struct {
	log        *slog.Logger
	gRpcServer *grpc.Server
	port       uint
}

func New(log *slog.Logger, authService authgrpc.Auth, port uint) *App {
	gRpcServer := grpc.NewServer()
	authgrpc.RegisterAuth(gRpcServer, authService)
	return &App{
		log:        log,
		gRpcServer: gRpcServer,
		port:       port,
	}
}

func (a *App) Run() {
	if err := a.Start(); err != nil {
		panic(err)
	}

}

func (a *App) Start() error {
	const op = "grpcapp.Start"
	log := a.log.With(slog.String("op", op))

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %s", op, err)
	}

	log.Info("grpc server start", slog.String("address", listener.Addr().String()))

	if err := a.gRpcServer.Serve(listener); err != nil {
		return fmt.Errorf("%s: %s", op, err)
	}
	return nil
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"
	log := a.log.With(slog.String("op", op))

	log.Info("stopping grpc server")
	a.gRpcServer.GracefulStop()
}
