package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Artem09076/dp/backend/core_service/internal/application/discounts"
	"github.com/Artem09076/dp/backend/core_service/internal/application/profile"
	"github.com/Artem09076/dp/backend/core_service/internal/application/reviews"
	"github.com/Artem09076/dp/backend/core_service/internal/application/services"
	"github.com/Artem09076/dp/backend/core_service/internal/config"
	"github.com/Artem09076/dp/backend/core_service/internal/lib/jwt"
	"github.com/Artem09076/dp/backend/core_service/internal/logger"
	discounthandlers "github.com/Artem09076/dp/backend/core_service/internal/presentation/discounts/handlers"
	coremiddleware "github.com/Artem09076/dp/backend/core_service/internal/presentation/middleware"
	profilehandlers "github.com/Artem09076/dp/backend/core_service/internal/presentation/profile/handlers"
	reviewhandlers "github.com/Artem09076/dp/backend/core_service/internal/presentation/reviews/handlers"
	servicehandlers "github.com/Artem09076/dp/backend/core_service/internal/presentation/services/handlers"
	sqlc "github.com/Artem09076/dp/backend/core_service/internal/storage/db"
	"github.com/Artem09076/dp/backend/core_service/internal/storage/rabbit"
	"github.com/Artem09076/dp/backend/core_service/internal/storage/redis"
	"github.com/go-chi/chi/middleware"
	amqp "github.com/rabbitmq/amqp091-go"

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

	conn, err := amqp.Dial(cfg.RabbitPath)
	if err != nil {
		log.Error("failed init starage", slog.String("error", err.Error()))
		os.Exit(1)
	}

	ch, _ := conn.Channel()
	ch.QueueDeclare("profile_queue", true, false, false, false, nil)
	publisher := rabbit.NewPublisher(ch)

	redisClient, err := redis.NewRedisClient(log, cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		log.Error("failed to connect to Redis", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer redisClient.Close()

	queries := sqlc.New(db)

	router := chi.NewRouter()

	validator := jwt.NewValidator(cfg.TokenSecret, redisClient)

	profileService := profile.NewProfileService(queries, log, publisher)
	profileHandlers := profilehandlers.NewProfileHandler(profileService, log)

	serviceService := services.NewService(queries, log)
	serviceHandlers := servicehandlers.NewServiceHandler(serviceService, log)

	discountService := discounts.NewDiscountService(queries, log)
	discountHandlers := discounthandlers.NewDiscountsHandler(discountService, log)

	reviewService := reviews.NewReviewService(queries, log)
	reviewDiscount := reviewhandlers.NewReviewHandler(reviewService, log)

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
		r.Get("/api/v1/discounts/{id}", discountHandlers.GetDiscount())
		r.Post("/api/v1/reviews", reviewDiscount.CreateReview())
		r.Get("/api/v1/booking/{bookingID}/reviews", reviewDiscount.GetReviewByBookingID())
		r.Get("/api/v1/reviews/{reviewID}", reviewDiscount.GetReviewByID())
		r.Get("/api/v1/service/{serviceID}/reviews", reviewDiscount.GetReviewsByServiceID())
		r.Patch("/api/v1/reviews/{reviewID}", reviewDiscount.PatchReview())
		r.Delete("/api/v1/reviews/{reviewID}", reviewDiscount.DeleteReview())
	})

	router.Group(func(r chi.Router) {
		r.Use(coremiddleware.NewJWTMiddleware(log, validator))
		r.Use(coremiddleware.NewRoleMiddleware(log, []string{"admin", "performer"}))
		r.Post("/api/v1/services", serviceHandlers.CreateService())
		r.Patch("/api/v1/services/{id}", serviceHandlers.PatchService())
		r.Delete("/api/v1/services/{id}", serviceHandlers.DeleteService())
		r.Get("/api/v1/services", serviceHandlers.GetServices())
		r.Post("/api/v1/services/{id}/discounts", discountHandlers.CreateDiscount())
		r.Patch("/api/v1/services/{serviceID}/discounts/{id}", discountHandlers.UpdateDiscount())
		r.Delete("/api/v1/services/{serviceID}/discounts/{id}", discountHandlers.DeleteDiscount())
	})

	router.Group(func(r chi.Router) {
		r.Use(coremiddleware.NewJWTMiddleware(log, validator))
		r.Use(coremiddleware.NewRoleMiddleware(log, []string{"admin"}))
		r.Patch("/api/v1/profile/verification_status", profileHandlers.UpdateVerificationStatus())
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
