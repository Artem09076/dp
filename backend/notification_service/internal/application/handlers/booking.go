package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/Artem09076/dp/backend/notification_service/internal/application/dto"
	"github.com/Artem09076/dp/backend/notification_service/internal/domain"
)

type BookingHandler struct {
	emailSender domain.EmailSender
	log         *slog.Logger
}

func NewBookingHandler(log *slog.Logger, sender domain.EmailSender) *BookingHandler {
	return &BookingHandler{
		log:         log,
		emailSender: sender,
	}
}

func (h *BookingHandler) Handle(msg []byte) error {
	var event dto.BookingEvent
	if err := json.Unmarshal(msg, &event); err != nil {
		return err
	}
	switch event.Event {
	case dto.BookingCancelled:
		return h.HandleCancelled(event)
	case dto.BookingCreated:
		return h.HandleCreated(event)
	case dto.BookingSubmit:
		return h.HandleSubmited(event)

	}

	return nil
}

func (h *BookingHandler) HandleCancelled(msg dto.BookingEvent) error {
	subject := "Бронирование услуги отменено"
	body := fmt.Sprintf(`
	Ваше бронирование услуги было отменено

	Услуга: %s
	Дата: %s
	`, msg.Service, msg.Time)
	return h.emailSender.Send(context.Background(), domain.Email{
		To:      msg.Email,
		Subject: subject,
		Body:    body,
		Retries: 2,
	})
}

func (h *BookingHandler) HandleCreated(msg dto.BookingEvent) error {
	subject := "Бронирование услуги создано"
	body := fmt.Sprintf(`
	Ваша услуга заинтересовала кого-то.
	Если вы в это время готовы оказать услугу, подтвердите бронь.

	Услуга: %s
	Дата: %s
	`, msg.Service, msg.Time)
	h.log.Info("sadfghjk")
	return h.emailSender.Send(context.Background(), domain.Email{
		To:      msg.Email,
		Subject: subject,
		Body:    body,
		Retries: 2,
	})
}

func (h *BookingHandler) HandleSubmited(msg dto.BookingEvent) error {
	subject := "Бронирование услуги подтверждено"
	body := fmt.Sprintf(`
	Ваша бронь услуги была подтверждена.
	Услуга: %s
	Дата: %s
	`, msg.Service, msg.Time)
	h.log.Info("sadfghjk")
	return h.emailSender.Send(context.Background(), domain.Email{
		To:      msg.Email,
		Subject: subject,
		Body:    body,
		Retries: 2,
	})
}
