package booking

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	handlerlib "github.com/Artem09076/dp/backend/booking_service/internal/lib/api/handler"
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
	UpdateBookingTime1(ctx context.Context, arg sqlc.UpdateBookingTime1Params) error
	UpdateBookingTime2(ctx context.Context, arg sqlc.UpdateBookingTime2Params) error
	GetBookingByClientID(ctx context.Context, clientID uuid.UUID) ([]sqlc.Booking, error)
	GetBookingByPerformerID(ctx context.Context, performerID uuid.UUID) ([]sqlc.Booking, error)
	ServiceExists(ctx context.Context, id uuid.UUID) (bool, error)
	IncreaseDiscountUsage(ctx context.Context, id uuid.UUID) error
	DeleteBooking(ctx context.Context, id uuid.UUID) error
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

func (s *BookingService) CheckBookingOwnerships(ctx context.Context, userID uuid.UUID, clientID uuid.UUID, performerID uuid.UUID) bool {
	user, err := s.repo.GetUserById(ctx, userID)
	if err != nil {
		s.log.Error("failed to get user", "error", err)
		return false
	}
	return clientID == userID || performerID == userID || user.Role == "admin"
}

func (s *BookingService) CreateBooking(ctx context.Context, ClientID uuid.UUID, booking dto.CreateBookingRequest) (uuid.UUID, error) {
	user, err := s.repo.GetUserById(ctx, ClientID)
	if err != nil {
		s.log.Error("failed to get user", "error", err)
		return uuid.Nil, err
	}
	if user.Role != "client" {
		return uuid.Nil, handlerlib.ErrForbidden
	}
	ok, err := s.CheckBookingTime(ctx, booking.ServiceID, booking.BookingTime)
	if err != nil {
		s.log.Info(err.Error())
		return uuid.Nil, err
	}
	if !ok {
		return uuid.Nil, handlerlib.ErrTimeBusy
	}
	service, err := s.repo.GetServiceById(ctx, booking.ServiceID)
	if err != nil {
		return uuid.Nil, handlerlib.ErrNotFound
	}

	var discount sqlc.Discount
	var discountId uuid.NullUUID
	if booking.DiscountID != uuid.Nil {
		discountId = uuid.NullUUID{UUID: booking.DiscountID, Valid: true}
		discount, err = s.repo.GetDiscountById(ctx, booking.DiscountID)
		if err != nil {
			return uuid.Nil, handlerlib.ErrNotFound
		}
		if discount.ServiceID != booking.ServiceID || discount.ValidTo.Before(booking.BookingTime) || discount.ValidFrom.After(booking.BookingTime) {
			return uuid.Nil, handlerlib.ErrInvalidInput
		}
		if discount.UsedCount+1 > discount.MaxUses {
			return uuid.Nil, handlerlib.ErrInvalidInput
		}
	}
	finalPrice := int32(service.Price)
	if discountId.Valid {
		finalPrice = s.calculateFinalPrice(int32(service.Price), discount)
	}

	id, err := s.repo.CreateBooking(ctx, sqlc.CreateBookingParams{
		ClientID:    ClientID,
		ServiceID:   booking.ServiceID,
		BasePrice:   int32(service.Price),
		DiscountID:  discountId,
		FinalPrice:  finalPrice,
		BookingTime: booking.BookingTime,
	})
	if err != nil {
		s.log.Error("failed to create booking", "error", err)
		return uuid.Nil, err
	}
	if err := s.repo.IncreaseDiscountUsage(ctx, booking.DiscountID); err != nil {
		return uuid.Nil, err
	}
	go s.publishEvent(BookingCreated, id, service.PerformerID, service.Title, booking.BookingTime)

	return id, nil
}

func (s *BookingService) CancelBooking(ctx context.Context, userID uuid.UUID, booking *sqlc.GetBookingByIDRow) error {
	if !s.CheckBookingOwnerships(ctx, userID, booking.ClientID, booking.PerformerID) {
		return handlerlib.ErrForbidden
	}

	if err := s.repo.CancelBooking(ctx, booking.ID); err != nil {
		s.log.Error("failed to cancel booking", "error", err)
		return err
	}
	go s.publishBoth(BookingCancelled, booking)
	return nil
}

func (s *BookingService) SubmitBooking(ctx context.Context, userID uuid.UUID, booking *sqlc.GetBookingByIDRow) error {
	if booking.Status == "confirmed" {
		return handlerlib.ErrAlreadyDone
	}
	if booking.PerformerID != userID {
		return handlerlib.ErrForbidden
	}

	if err := s.repo.SubmitBooking(ctx, booking.ID); err != nil {
		s.log.Error("failed to submit booking", "error", err)
		return err
	}
	go s.publishEvent(BookingSubmit, booking.ID, booking.ClientID, booking.ServiceTitle, booking.BookingTime)

	return nil
}

func (s *BookingService) UpdateBooking(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID, bookingData dto.PatchBookingRequest) error {
	booking, err := s.repo.GetBookingByID(ctx, bookingID)

	if err != nil {
		s.log.Error("failed to get booking", "error", err)
		return handlerlib.ErrNotFound
	}

	if !s.CheckBookingOwnerships(ctx, userID, booking.ClientID, booking.PerformerID) {
		return handlerlib.ErrForbidden
	}
	var bookingTime time.Time
	if !bookingData.BookingTime.IsZero() {
		ok, err := s.CheckBookingTime(ctx, booking.ServiceID, bookingData.BookingTime)
		if err != nil {
			return err
		}
		if !ok {
			return handlerlib.ErrTimeBusy
		}
		bookingTime = bookingData.BookingTime
	}

	if userID == booking.PerformerID {
		req := sqlc.UpdateBookingTime1Params{
			BookingTime: bookingTime,
			ID:          bookingID,
		}
		if err := s.repo.UpdateBookingTime1(ctx, req); err != nil {
			s.log.Error("failed to update booking", "error", err)
			return err
		}

		go s.publishEvent(BookingUpdated1, booking.ID, booking.ClientID, booking.ServiceTitle, req.BookingTime)
	} else if userID == booking.ClientID {
		req := sqlc.UpdateBookingTime2Params{
			BookingTime: bookingTime,
			ID:          bookingID,
		}
		if err := s.repo.UpdateBookingTime2(ctx, req); err != nil {
			s.log.Error("failed to update booking", "error", err)
			return err
		}
		go s.publishEvent(BookingUpdated2, booking.ID, booking.PerformerID, booking.ServiceTitle, req.BookingTime)
	}

	return nil
}

func (s *BookingService) DeleteBooking(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID) error {
	booking, err := s.repo.GetBookingByID(ctx, bookingID)

	if err != nil {
		s.log.Error("failed to get booking", "error", err)
		return handlerlib.ErrNotFound
	}

	if !s.CheckBookingOwnerships(ctx, userID, booking.ClientID, booking.PerformerID) {
		return handlerlib.ErrForbidden
	}
	if err := s.repo.DeleteBooking(ctx, bookingID); err != nil {
		return err
	}

	return nil
}

func (s *BookingService) CheckBookingTime(ctx context.Context, serviceID uuid.UUID, bookingTime time.Time) (bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err

	}
	defer tx.Rollback()
	qtx := s.repo.WithTx(tx)

	service, err := qtx.GetServiceById(ctx, serviceID)
	if err != nil {
		s.log.Error("failed to get service", "error", err)
		return false, err
	}

	newStart := bookingTime
	newEnd := newStart.Add(time.Duration(service.DurationMinutes) * time.Minute)

	bookings, err := qtx.GetBookingsForUpdate(ctx, sqlc.GetBookingsForUpdateParams{
		PerformerID:   service.PerformerID,
		BookingTime:   newStart,
		BookingTime_2: newEnd,
	})
	if err != nil {
		return false, err
	}
	if len(bookings) > 0 {
		return false, handlerlib.ErrTimeBusy
	}
	return true, nil
}

func (s *BookingService) GetBooking(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID) (*sqlc.GetBookingByIDRow, error) {
	res, err := s.repo.GetBookingByID(ctx, bookingID)
	if err != nil {
		return nil, handlerlib.ErrNotFound
	}
	if !s.CheckBookingOwnerships(ctx, userID, res.ClientID, res.PerformerID) {
		return nil, handlerlib.ErrForbidden
	}
	return &res, nil
}

func (s *BookingService) ServiceExists(ctx context.Context, serviceID uuid.UUID) (bool, error) {
	ok, err := s.repo.ServiceExists(ctx, serviceID)
	if err != nil {
		return false, err
	}
	return ok, nil
}

func (s *BookingService) GetBookings(ctx context.Context, userID uuid.UUID) ([]sqlc.Booking, error) {
	user, err := s.repo.GetUserById(ctx, userID)
	if err != nil {
		return nil, err
	}
	switch user.Role {
	case "performer":
		return s.repo.GetBookingByPerformerID(ctx, userID)
	case "client":
		return s.repo.GetBookingByClientID(ctx, userID)
	default:
		return nil, handlerlib.ErrForbidden
	}
}

func (s *BookingService) publishEvent(event BookingEventType, bookingID uuid.UUID, userID uuid.UUID, serviceTitle string, bookingTime time.Time) {
	user, err := s.repo.GetUserById(context.Background(), userID)
	if err != nil {
		return
	}
	msg := BookingEvent{
		Event:     event,
		BookingID: bookingID.String(),
		Email:     user.Email,
		Service:   serviceTitle,
		Time:      bookingTime.String(),
	}
	s.publisher.Publish("booking_queue", msg)
}

func (s *BookingService) publishBoth(event BookingEventType, booking *sqlc.GetBookingByIDRow) {
	s.publishEvent(event, booking.ID, booking.ClientID, booking.ServiceTitle, booking.BookingTime)
	s.publishEvent(event, booking.ID, booking.PerformerID, booking.ServiceTitle, booking.BookingTime)
}

func (s *BookingService) calculateFinalPrice(basePrice int32, discount sqlc.Discount) int32 {
	if discount.Type == sqlc.DiscoutnTypePercentage {
		return int32(float64(basePrice) * (1.0 - float64(discount.Value)/100.0))
	} else {
		return basePrice - discount.Value
	}
}
