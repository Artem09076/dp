package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Artem09076/dp/backend/notification_service/internal/application/dto"
	"github.com/Artem09076/dp/backend/notification_service/internal/application/templates"
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
	case dto.BookingUpdated1:
		return h.HandleUpdate1(event)
	case dto.BookingUpdated2:
		return h.HandleUpdate2(event)

	}

	return nil
}

func (h *BookingHandler) formatDateTime(dateTimeStr string) string {
	h.log.Info("formatting date time", slog.String("input", dateTimeStr))

	dateTimeStr = strings.TrimSpace(dateTimeStr)

	formats := []string{
		"2006-01-02 15:04:05 -0700 -0700",
		"2006-01-02 15:04:05 -0700",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"02.01.2006 15:04",
	}

	var parsedTime time.Time
	var err error

	for _, format := range formats {
		parsedTime, err = time.Parse(format, dateTimeStr)
		if err == nil {
			h.log.Info("successfully parsed time", slog.String("format", format), slog.String("result", parsedTime.String()))
			break
		}
	}

	if err != nil {
		h.log.Info("trying regex extraction", slog.String("error", err.Error()))

		re := regexp.MustCompile(`(\d{4})-(\d{2})-(\d{2})\s+(\d{2}):(\d{2}):(\d{2})`)
		matches := re.FindStringSubmatch(dateTimeStr)

		if len(matches) == 7 {
			year, _ := strconv.Atoi(matches[1])
			month, _ := strconv.Atoi(matches[2])
			day, _ := strconv.Atoi(matches[3])
			hour, _ := strconv.Atoi(matches[4])
			minute, _ := strconv.Atoi(matches[5])
			second, _ := strconv.Atoi(matches[6])

			parsedTime = time.Date(year, time.Month(month), day, hour, minute, second, 0, time.Local)
			h.log.Info("extracted time via regex", slog.String("result", parsedTime.String()))
		} else {
			h.log.Warn("failed to parse time, returning original", slog.String("original", dateTimeStr))
			return dateTimeStr
		}
	}

	formatted := parsedTime.Format("02.01.2006 в 15:04")
	h.log.Info("formatted time", slog.String("result", formatted))
	return formatted
}

func (h *BookingHandler) HandleCancelled(msg dto.BookingEvent) error {
	formattedTime := h.formatDateTime(msg.Time)
	emailData := templates.EmailData{
		Title:        "Бронирование отменено",
		Username:     "Уважаемый клиент",
		ButtonText:   "Найти похожие услуги",
		ButtonURL:    "https://your-service.com/services",
		Year:         time.Now().Year(),
		SupportEmail: "support@your-service.com",
	}

	bookingData := templates.BookingData{
		EmailData:   emailData,
		ServiceName: msg.Service,
		DateTime:    formattedTime,
	}

	body, err := templates.RenderBookingCancelled(bookingData)
	if err != nil {
		h.log.Error("failed to render template", slog.String("error", err.Error()))
		return err
	}

	subject := "❌ Бронирование отменено"
	return h.emailSender.Send(context.Background(), domain.Email{
		To:      msg.Email,
		Subject: subject,
		Body:    body,
		Retries: 2,
	})
}

func (h *BookingHandler) HandleCreated(msg dto.BookingEvent) error {
	formattedTime := h.formatDateTime(msg.Time)
	emailData := templates.EmailData{
		Title:        "Новое бронирование",
		Username:     "Уважаемый партнер",
		ButtonText:   "Подтвердить бронь",
		ButtonURL:    fmt.Sprintf("https://your-service.com/bookings/%s/confirm", msg.BookingID),
		Year:         time.Now().Year(),
		SupportEmail: "support@your-service.com",
	}
	bookingData := templates.BookingData{
		EmailData:   emailData,
		ServiceName: msg.Service,
		DateTime:    formattedTime,
	}

	body, err := templates.RenderBookingCreated(bookingData)
	if err != nil {
		h.log.Error("failed to render template", slog.String("error", err.Error()))
		return err
	}
	subject := "🔔 Новое бронирование услуги"

	return h.emailSender.Send(context.Background(), domain.Email{
		To:      msg.Email,
		Subject: subject,
		Body:    body,
		Retries: 2,
	})
}

func (h *BookingHandler) HandleSubmited(msg dto.BookingEvent) error {
	formattedTime := h.formatDateTime(msg.Time)

	emailData := templates.EmailData{
		Title:        "Бронирование подтверждено",
		Username:     "Уважаемый клиент",
		ButtonText:   "Посмотреть детали",
		ButtonURL:    fmt.Sprintf("https://your-service.com/bookings/%s", msg.BookingID),
		Year:         time.Now().Year(),
		SupportEmail: "support@your-service.com",
	}

	bookingData := templates.BookingData{
		EmailData:   emailData,
		ServiceName: msg.Service,
		DateTime:    formattedTime,
	}

	body, err := templates.RenderBookingSubmit(bookingData)
	if err != nil {
		h.log.Error("failed to render template", slog.String("error", err.Error()))
		return err
	}

	subject := "✅ Бронирование успешно подтверждено"
	return h.emailSender.Send(context.Background(), domain.Email{
		To:      msg.Email,
		Subject: subject,
		Body:    body,
		Retries: 2,
	})
}

func (h *BookingHandler) HandleUpdate1(msg dto.BookingEvent) error {
	formattedTime := h.formatDateTime(msg.Time)
	emailData := templates.EmailData{
		Title:        "Время услуги изменено",
		Username:     "Уважаемый клиент",
		ButtonText:   "Связаться с исполнителем",
		ButtonURL:    "https://your-service.com/messages",
		Year:         time.Now().Year(),
		SupportEmail: "support@your-service.com",
	}

	bookingData := templates.BookingData{
		EmailData:   emailData,
		ServiceName: msg.Service,
		DateTime:    formattedTime,
	}

	body, err := templates.RenderBookingUpdated(bookingData, 1)
	if err != nil {
		h.log.Error("failed to render template", slog.String("error", err.Error()))
		return err
	}

	subject := "🕐 Время услуги было изменено"
	return h.emailSender.Send(context.Background(), domain.Email{
		To:      msg.Email,
		Subject: subject,
		Body:    body,
		Retries: 2,
	})

}

func (h *BookingHandler) HandleUpdate2(msg dto.BookingEvent) error {
	formattedTime := h.formatDateTime(msg.Time)
	emailData := templates.EmailData{
		Title:        "Предложение нового времени",
		Username:     "Уважаемый клиент",
		ButtonText:   "Подтвердить новое время",
		ButtonURL:    fmt.Sprintf("https://your-service.com/bookings/%s/confirm-time", msg.BookingID),
		Year:         time.Now().Year(),
		SupportEmail: "support@your-service.com",
	}

	bookingData := templates.BookingData{
		EmailData:   emailData,
		ServiceName: msg.Service,
		DateTime:    formattedTime,
	}

	body, err := templates.RenderBookingUpdated(bookingData, 2)
	if err != nil {
		h.log.Error("failed to render template", slog.String("error", err.Error()))
		return err
	}

	subject := "🔄 Требуется подтверждение нового времени"
	return h.emailSender.Send(context.Background(), domain.Email{
		To:      msg.Email,
		Subject: subject,
		Body:    body,
		Retries: 2,
	})

}
