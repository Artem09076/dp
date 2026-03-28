package main

import (
	"log/slog"
	"os"

	"github.com/Artem09076/dp/backend/notification_service/internal/application/handlers"
	"github.com/Artem09076/dp/backend/notification_service/internal/config"
	"github.com/Artem09076/dp/backend/notification_service/internal/infrastructure/email"
	"github.com/Artem09076/dp/backend/notification_service/internal/logger"
	"github.com/Artem09076/dp/backend/notification_service/internal/storage/rabbit"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {

	cfg := config.LoadConfig()
	log := logger.SetupLogger(cfg.Env)

	emailSender := email.NewSMTPSender(
		log,
		cfg.SMTP.SMTPHost,
		cfg.SMTP.SMTPPort,
		cfg.SMTP.SMTPUser,
		cfg.SMTP.SMTPPass,
		cfg.SMTP.SMTPFrom,
	)

	conn, err := amqp.Dial(cfg.RabbitURL)
	if err != nil {
		log.Error("failed init starage", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer conn.Close()
	ch, _ := conn.Channel()
	ch.QueueDeclare("booking_queue", true, false, false, false, nil)

	consumer := rabbit.NewConsumer(ch, log)
	handler := handlers.NewBookingHandler(log, emailSender)
	log.Info("Starting consume")
	if err := consumer.Consume("booking_queue", handler.Handle); err != nil {
		log.Error("failed to start consumer", slog.String("error", err.Error()))
		os.Exit(1)
	}
	select {}

}
