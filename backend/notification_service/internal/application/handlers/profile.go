package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/Artem09076/dp/backend/notification_service/internal/application/dto"
	"github.com/Artem09076/dp/backend/notification_service/internal/application/templates"
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
	emailData := templates.EmailData{
		Title:        "Профиль успешно верифицирован",
		Username:     event.Name,
		ButtonText:   "Перейти в личный кабинет",
		ButtonURL:    "https://your-service.com/dashboard",
		Year:         time.Now().Year(),
		SupportEmail: "support@your-service.com",
	}

	profileData := templates.ProfileData{
		EmailData:          emailData,
		VerificationStatus: "success",
		Reason:             "",
	}

	body, err := templates.RenderProfileVerificationSuccess(profileData)
	if err != nil {
		h.log.Error("failed to render template", slog.String("error", err.Error()))
		return err
	}

	subject := "🎉 Поздравляем! Ваш профиль успешно верифицирован"
	return h.emailSender.Send(context.Background(), domain.Email{
		To:      event.Email,
		Subject: subject,
		Body:    body,
		Retries: 2,
	})
}

func (h *ProfileHandler) HandleVerificationStatusUpdatedReject(event dto.ProfileEvent) error {
	reason := "Номер ИП, предоставленное вами уже закрыто или выглядит подозрительно"
	emailData := templates.EmailData{
		Title:        "Верификация профиля отклонена",
		Username:     event.Name,
		ButtonText:   "Попробовать снова",
		ButtonURL:    "https://your-service.com/verification",
		Year:         time.Now().Year(),
		SupportEmail: "support@your-service.com",
	}

	profileData := templates.ProfileData{
		EmailData:          emailData,
		VerificationStatus: "rejected",
		Reason:             reason,
	}

	body, err := templates.RenderProfileVerificationReject(profileData)
	if err != nil {
		h.log.Error("failed to render template", slog.String("error", err.Error()))
		return err
	}

	subject := "⚠️ Верификация профиля отклонена"
	return h.emailSender.Send(context.Background(), domain.Email{
		To:      event.Email,
		Subject: subject,
		Body:    body,
		Retries: 2,
	})
}
