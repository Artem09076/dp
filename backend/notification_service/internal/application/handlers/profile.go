package handlers

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/Artem09076/dp/backend/notification_service/internal/application/dto"
	"github.com/Artem09076/dp/backend/notification_service/internal/domain"
)

type ProfileHandler struct {
	log         *slog.Logger
	emailSender domain.EmailSender
}

func NewProfileHandler(log *slog.Logger, emailSender domain.EmailSender) *ProfileHandler {
	return &ProfileHandler{
		log:         log,
		emailSender: emailSender,
	}
}

func (h *ProfileHandler) Handle(msg []byte) error {
	h.log.Info("sdfghytred")
	var event dto.ProfileEvent
	if err := json.Unmarshal(msg, &event); err != nil {
		return err
	}
	switch event.Event {
	case dto.ProfileVerificationStatusUpdatedSubmit:
		return h.HandleVerificationStatusUpdatedSubmit(event)
	case dto.ProfileVerificationStatusUpdatedReject:
		return h.HandleVerificationStatusUpdatedReject(event)
	}

	return nil
}

func (h *ProfileHandler) HandleVerificationStatusUpdatedSubmit(event dto.ProfileEvent) error {
	subject := "Профиль верифицирован"
	body := "Ваш профиль был успешно верифицирован. Теперь вы можете пользоваться всеми функциями нашего сервиса."
	return h.emailSender.Send(context.Background(), domain.Email{
		To:      event.Email,
		Subject: subject,
		Body:    body,
		Retries: 2,
	})
}

func (h *ProfileHandler) HandleVerificationStatusUpdatedReject(event dto.ProfileEvent) error {
	subject := "Профиль отклонен"
	body := "К сожалению, ваш профиль не прошел верификацию. Пожалуйста, свяжитесь с нашей поддержкой для получения дополнительной информации."
	return h.emailSender.Send(context.Background(), domain.Email{
		To:      event.Email,
		Subject: subject,
		Body:    body,
		Retries: 2,
	})
}
