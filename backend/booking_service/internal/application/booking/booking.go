package booking

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/Artem09076/dp/backend/booking_service/internal/presentation/booking/dto"
	sqlc "github.com/Artem09076/dp/backend/booking_service/internal/storage/db"
	"github.com/Artem09076/dp/backend/booking_service/internal/storage/rabbit"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type BookingRepository interface {
	WithTx(tx *sql.Tx) *sqlc.Queries
	CreateBooking(ctx context.Context, booking sqlc.CreateBookingParams) (uuid.UUID, error)
	CancelBooking(ctx context.Context, bookingID uuid.UUID) error
	GetServiceById(ctx context.Context, id uuid.UUID) (sqlc.Service, error)
	GetDiscountById(ctx context.Context, id uuid.UUID) (sqlc.Discount, error)
	GetBookingByID(ctx context.Context, id uuid.UUID) (sqlc.GetBookingByIDRow, error)
	GetUserById(ctx context.Context, id uuid.UUID) (sqlc.User, error)
	SubmitBooking(ctx context.Context, id uuid.UUID) error
}
type BookingService struct {
	repo      BookingRepository
	db        *sql.DB
	log       *slog.Logger
	publisher *rabbit.Publisher
}

func NewBookingService(repo BookingRepository, db *sql.DB, log *slog.Logger, publisher *rabbit.Publisher) *BookingService {
	return &BookingService{
		repo:      repo,
		db:        db,
		log:       log,
		publisher: publisher,
	}
}

func (s *BookingService) CreateBooking(ctx context.Context, ClientID uuid.UUID, booking dto.CreateBookingRequest) (uuid.UUID, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return uuid.Nil, err

	}
	defer tx.Rollback()
	qtx := s.repo.WithTx(tx)
	service, err := qtx.GetServiceById(ctx, booking.ServiceID)
	if err != nil {
		s.log.Error("failed to get service", "error", err)
		return uuid.Nil, err
	}

	newStart := booking.BookingTime
	newEnd := newStart.Add(time.Duration(service.DurationMinutes) * time.Minute)

	bookings, err := qtx.GetBookingsForUpdate(ctx, service.ID)
	if err != nil {
		return uuid.Nil, err
	}
	for _, b := range bookings {
		exsitsingStart := b.BookingTime
		exsistingEnd := b.BookingTime.Add(time.Duration(b.DurationMinutes) * time.Minute)
		if newStart.Before(exsistingEnd) && newEnd.After(exsitsingStart) {
			return uuid.Nil, fmt.Errorf("time slot already booked")
		}

	}
	var validDiscountId uuid.NullUUID
	if booking.DiscountID == uuid.Nil {
		validDiscountId = uuid.NullUUID{Valid: false}
	} else {
		validDiscountId = uuid.NullUUID{UUID: booking.DiscountID, Valid: true}
	}

	discount, err := qtx.GetDiscountById(ctx, validDiscountId.UUID)
	if err != nil && validDiscountId.Valid {
		s.log.Error("failed to get discount", "error", err)
		return uuid.Nil, err
	}
	if validDiscountId.Valid && (discount.ServiceID != booking.ServiceID || discount.ValidTo.Before(booking.BookingTime) || discount.ValidFrom.After(booking.BookingTime)) {
		s.log.Error("invalid discount", "discount_id", booking.DiscountID)
		return uuid.Nil, err
	}
	var finalPrice int32
	if discount.Type == "percentage" {
		finalPrice = int32(int32(service.Price) * (1 - discount.Value/100))
	} else {
		finalPrice = int32(service.Price) - discount.Value
	}

	id, err := qtx.CreateBooking(ctx, sqlc.CreateBookingParams{
		ClientID:    ClientID,
		ServiceID:   booking.ServiceID,
		BasePrice:   int32(service.Price),
		DiscountID:  validDiscountId,
		FinalPrice:  finalPrice,
		BookingTime: booking.BookingTime,
	})
	if err != nil {
		s.log.Error("failed to create booking", "error", err)
		return uuid.Nil, err
	}

	performer, err := s.repo.GetUserById(ctx, service.PerformerID)
	if err != nil {
		return uuid.Nil, err
	}

	msg := BookingEvent{
		Event:     BookingCreated,
		BookingID: id.String(),
		Email:     performer.Email,
		Service:   service.Title,
		Time:      booking.BookingTime.String(),
	}
	if err := s.publisher.Publish("booking_queue", msg); err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (s *BookingService) CancelBooking(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID) error {
	user, err := s.repo.GetUserById(ctx, userID)
	if err != nil {
		s.log.Error("failed to get user", "error", err)
		return err
	}
	booking, err := s.repo.GetBookingByID(ctx, bookingID)
	if err != nil {
		s.log.Error("failed to get booking", "error", err)
		return err
	}

	if booking.ClientID != userID && booking.PerformerID != userID && user.Role != "admin" {
		s.log.Error("user is not the owner of the booking", "user_id", userID, "booking_id", bookingID)
		return err
	}

	if err := s.repo.CancelBooking(ctx, bookingID); err != nil {
		s.log.Error("failed to cancel booking", "error", err)
		return err
	}
	msg1 := BookingEvent{
		Event:     BookingCancelled,
		BookingID: bookingID.String(),
		Email:     user.Email,
		Service:   booking.ServiceTitle,
		Time:      booking.BookingTime.String(),
	}
	if err := s.publisher.Publish("booking_queue", msg1); err != nil {
		return err
	}
	var msg2 BookingEvent
	if userID == booking.PerformerID {
		client, err := s.repo.GetUserById(ctx, booking.ClientID)
		if err != nil {
			return err
		}
		msg2 = BookingEvent{
			Event:     BookingCancelled,
			BookingID: bookingID.String(),
			Email:     client.Email,
			Service:   booking.ServiceTitle,
			Time:      booking.BookingTime.String(),
		}
	}
	if userID == booking.ClientID {
		performer, err := s.repo.GetUserById(ctx, booking.PerformerID)
		if err != nil {
			return err
		}
		msg2 = BookingEvent{
			Event:     BookingCancelled,
			BookingID: bookingID.String(),
			Email:     performer.Email,
			Service:   booking.ServiceTitle,
			Time:      booking.BookingTime.String(),
		}
	}
	return s.publisher.Publish("booking_queue", msg2)
}

func (s *BookingService) SubmitBooking(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID) error {
	booking, err := s.repo.GetBookingByID(ctx, bookingID)
	if err != nil {
		s.log.Error("failed to get booking", "error", err)
		return err
	}
	if booking.Status == "confirmed" {
		s.log.Warn("booking is already confirmed", "booking_id", bookingID, "status", booking.Status)
		return nil
	}
	if booking.PerformerID != userID {
		s.log.Error("user is not the performer of the booking", "user_id", userID, "booking_id", bookingID)
		return err
	}
	if err := s.repo.SubmitBooking(ctx, bookingID); err != nil {
		s.log.Error("failed to submit booking", "error", err)
		return err
	}
	client, err := s.repo.GetUserById(ctx, booking.ClientID)
	if err != nil {
		return err
	}
	msg := BookingEvent{
		Event:     BookingSubmit,
		BookingID: bookingID.String(),
		Email:     client.Email,
		Service:   booking.ServiceTitle,
		Time:      booking.BookingTime.String(),
	}

	return s.publisher.Publish("booking_queue", msg)
}
