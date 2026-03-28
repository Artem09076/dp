package usecase

import (
	"context"
	"log/slog"

	"github.com/Artem09076/dp/backend/notification_service/internal/domain"
)

type NotificationUsecase struct {
	emailSender domain.EmailSender
	log         *slog.Logger
}

func NewNotificationUsecase(sender domain.EmailSender, log *slog.Logger) *NotificationUsecase {
	return &NotificationUsecase{
		emailSender: sender,
		log:         log,
	}
}

func (u *NotificationUsecase) SendEmail(ctx context.Context, email domain.Email) error {
	u.log.Info("sending email",
		slog.String("to", email.To),
	)

	return u.emailSender.Send(ctx, email)
}
