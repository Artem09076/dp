package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Artem09076/dp/backend/booking_service/internal/application/booking"
	"github.com/Artem09076/dp/backend/booking_service/internal/config"
	"github.com/Artem09076/dp/backend/booking_service/internal/lib/jwt"
	"github.com/Artem09076/dp/backend/booking_service/internal/logger"
	bookinghandlers "github.com/Artem09076/dp/backend/booking_service/internal/presentation/booking/handlers"
	bookingmiddleware "github.com/Artem09076/dp/backend/booking_service/internal/presentation/middleware"
	sqlc "github.com/Artem09076/dp/backend/booking_service/internal/storage/db"
	"github.com/Artem09076/dp/backend/booking_service/internal/storage/rabbit"
	"github.com/Artem09076/dp/backend/booking_service/internal/storage/redis"
	amqp "github.com/rabbitmq/amqp091-go"

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

	conn, err := amqp.Dial(cfg.RabbitPath)
	if err != nil {
		log.Error("failed init starage", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer conn.Close()
	ch, _ := conn.Channel()
	ch.QueueDeclare("booking_queue", true, false, false, false, nil)
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

	bookingService := booking.NewBookingService(queries, db, log, publisher, redisClient)
	bookingHandlers := bookinghandlers.NewBookingHandler(bookingService, log)

	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(bookingmiddleware.CorsMiddleware)

	router.Group(func(r chi.Router) {
		r.Use(bookingmiddleware.NewJWTMiddleware(log, validator))
		r.Post("/api/v1/bookings", bookingHandlers.CreateBooking())
		r.Patch("/api/v1/bookings/cancel/{id}", bookingHandlers.CancelBooking())
		r.Patch("/api/v1/bookings/submit/{id}", bookingHandlers.SubmitBooking())
		r.Patch("/api/v1/bookings/completed/{id}", bookingHandlers.CompleteBooking())
		r.Patch("/api/v1/bookings/{id}", bookingHandlers.PatchBooking())
		r.Get("/api/v1/bookings/{id}", bookingHandlers.GetBooking())
		r.Get("/api/v1/bookings", bookingHandlers.GetBookings())
		r.Delete("/api/v1/bookings/{id}", bookingHandlers.DeleteBooking())

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
