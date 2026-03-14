package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Artem09076/dp/backend/core_service/internal/application/profile"
	"github.com/Artem09076/dp/backend/core_service/internal/application/services"
	"github.com/Artem09076/dp/backend/core_service/internal/config"
	"github.com/Artem09076/dp/backend/core_service/internal/lib/jwt"
	"github.com/Artem09076/dp/backend/core_service/internal/logger"
	coremiddleware "github.com/Artem09076/dp/backend/core_service/internal/presentation/middleware"
	profilehandlers "github.com/Artem09076/dp/backend/core_service/internal/presentation/profile/handlers"
	servicehandlers "github.com/Artem09076/dp/backend/core_service/internal/presentation/services/handlers"
	sqlc "github.com/Artem09076/dp/backend/core_service/internal/storage/db"
	"github.com/go-chi/chi/middleware"

	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.LoadConfig()

	log := logger.SetupLogger(cfg.Env)

	db, err := sql.Open("postgres", cfg.DBPath)
	if err != nil {
		log.Error("failed init starage", slog.String("error", err.Error()))
		os.Exit(1)
	}

	queries := sqlc.New(db)

	router := chi.NewRouter()

	validator := jwt.NewValidator("secret", queries)

	profileService := profile.NewProfileService(queries, log)
	profileHandlers := profilehandlers.NewProfileHandler(profileService, log)

	serviceService := services.NewService(queries, log)
	serviceHandlers := servicehandlers.NewServiceHandler(serviceService, log)

	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Group(func(r chi.Router) {
		r.Use(coremiddleware.NewJWTMiddleware(log, validator))
		r.Get("/api/v1/profile", profileHandlers.GetProfile())
		r.Patch("/api/v1/profile", profileHandlers.PatchProfile())
		r.Delete("/api/v1/profile", profileHandlers.DeleteProfile())
		r.Get("/api/v1/services", serviceHandlers.SearchServices())
		r.Get("/api/v1/services/{id}", serviceHandlers.GetService())
	})

	router.Group(func(r chi.Router) {
		r.Use(coremiddleware.NewJWTMiddleware(log, validator))
		r.Use(coremiddleware.NewRoleMiddleware(log))
		r.Post("/api/v1/services", serviceHandlers.CreateService())
		r.Patch("/api/v1/services/{id}", serviceHandlers.PatchService())
		r.Delete("/api/v1/services/{id}", serviceHandlers.DeleteService())
	})

	done := make(chan os.Signal, 1)

	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	srv := &http.Server{
		Addr:         cfg.HTTP.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTP.Timeout,
		WriteTimeout: cfg.HTTP.Timeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Error("Failed to start server")
		}
	}()
	log.Info("Starting server", slog.String("address", cfg.HTTP.Address))
	<-done
	log.Info("Stopping server")
	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.Timeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Failed to stop server")
	}

}
