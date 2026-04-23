package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Artem09076/dp/backend/core_service/internal/application/admin"
	"github.com/Artem09076/dp/backend/core_service/internal/application/discounts"
	"github.com/Artem09076/dp/backend/core_service/internal/application/profile"
	"github.com/Artem09076/dp/backend/core_service/internal/application/reviews"
	"github.com/Artem09076/dp/backend/core_service/internal/application/services"
	"github.com/Artem09076/dp/backend/core_service/internal/config"
	"github.com/Artem09076/dp/backend/core_service/internal/lib/jwt"
	"github.com/Artem09076/dp/backend/core_service/internal/logger"
	adminhandler "github.com/Artem09076/dp/backend/core_service/internal/presentation/admin/handlers"
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

	profileService := profile.NewProfileService(queries, log, publisher, redisClient)
	profileHandlers := profilehandlers.NewProfileHandler(profileService, log)

	serviceService := services.NewService(queries, log, redisClient)
	serviceHandlers := servicehandlers.NewServiceHandler(serviceService, log)

	discountService := discounts.NewDiscountService(queries, log, redisClient)
	discountHandlers := discounthandlers.NewDiscountsHandler(discountService, log)

	reviewService := reviews.NewReviewService(queries, log, redisClient)
	reviewDiscount := reviewhandlers.NewReviewHandler(reviewService, log)

	adminServcie := admin.NewAdminService(queries, log, redisClient)
	adminHandlers := adminhandler.NewAdminHandler(adminServcie, log)
	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(coremiddleware.CorsMiddleware)
	router.Get("/api/v1/services/search", serviceHandlers.SearchServices())
	router.Get("/api/v1/services/{id}", serviceHandlers.GetService())
	router.Group(func(r chi.Router) {
		r.Use(coremiddleware.NewJWTMiddleware(log, validator))
		r.Get("/api/v1/profile", profileHandlers.GetProfile())
		r.Patch("/api/v1/profile", profileHandlers.PatchProfile())
		r.Delete("/api/v1/profile", profileHandlers.DeleteProfile())
		r.Get("/api/v1/discounts/{id}", discountHandlers.GetDiscount())
		r.Get("/api/v1/services/{id}/discounts", discountHandlers.GetDiscountsByServiceID())
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
		r.Get("/api/v1/admin/performers/unverified", adminHandlers.GetUnverifiedPerformers())
		r.Get("/api/v1/admin/users", adminHandlers.GetUsers())
		r.Get("/api/v1/admin/users/{user_id}", adminHandlers.GetUserByID())
		r.Delete("/api/v1/admin/users/{user_id}", adminHandlers.DeleteUser())
		r.Post("/api/v1/admin/users/verify/batch", adminHandlers.BatchVerifyPerformers())
		r.Get("/api/v1/admin/services", adminHandlers.GetServices())

		r.Get("/api/v1/admin/bookings", adminHandlers.GetBookings())

		r.Get("/api/v1/admin/reviews", adminHandlers.GetReviews())
		r.Delete("/api/v1/admin/reviews/{review_id}", adminHandlers.DeleteReview())

		r.Patch("/api/v1/admin/performers/verification_status", profileHandlers.UpdateVerificationStatus())

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
