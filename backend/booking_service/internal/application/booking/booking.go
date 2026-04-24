package booking

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/Artem09076/dp/backend/booking_service/internal/lib/api/errors"
	"github.com/Artem09076/dp/backend/booking_service/internal/presentation/booking/dto"
	sqlc "github.com/Artem09076/dp/backend/booking_service/internal/storage/db"
	"github.com/Artem09076/dp/backend/booking_service/internal/storage/rabbit"
	"github.com/Artem09076/dp/backend/booking_service/internal/storage/redis"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type TransactionManager interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type BookingRepository interface {
	WithTx(tx *sql.Tx) *sqlc.Queries
	CreateBooking(ctx context.Context, booking sqlc.CreateBookingParams) (uuid.UUID, error)
	CancelBooking(ctx context.Context, bookingID uuid.UUID) error
	GetServiceById(ctx context.Context, id uuid.UUID) (sqlc.Service, error)
	GetDiscountById(ctx context.Context, id uuid.UUID) (sqlc.Discount, error)
	GetBookingByID(ctx context.Context, id uuid.UUID) (sqlc.GetBookingByIDRow, error)
	CompletedBooking(ctx context.Context, id uuid.UUID) error
	GetUserById(ctx context.Context, id uuid.UUID) (sqlc.User, error)
	SubmitBooking(ctx context.Context, id uuid.UUID) error
	UpdateBookingTime1(ctx context.Context, arg sqlc.UpdateBookingTime1Params) error
	UpdateBookingTime2(ctx context.Context, arg sqlc.UpdateBookingTime2Params) error
	GetBookingByClientID(ctx context.Context, clientID uuid.UUID) ([]sqlc.Booking, error)
	GetBookingByPerformerID(ctx context.Context, performerID uuid.UUID) ([]sqlc.Booking, error)
	ServiceExists(ctx context.Context, id uuid.UUID) (bool, error)
	IncreaseDiscountUsage(ctx context.Context, id uuid.UUID) error
	DeleteBooking(ctx context.Context, id uuid.UUID) error
	GetBookingsForUpdate(ctx context.Context, arg sqlc.GetBookingsForUpdateParams) ([]sqlc.GetBookingsForUpdateRow, error)
}

type BookingService struct {
	repo      BookingRepository
	db        TransactionManager
	log       *slog.Logger
	publisher rabbit.PublisherInterface
	redis     *redis.RedisClient
}

func NewBookingService(repo BookingRepository, db TransactionManager, log *slog.Logger, publisher rabbit.PublisherInterface, redis *redis.RedisClient) *BookingService {
	return &BookingService{
		repo:      repo,
		db:        db,
		log:       log,
		publisher: publisher,
		redis:     redis,
	}
}

func (s *BookingService) GetUserById(ctx context.Context, id uuid.UUID) (*sqlc.User, error) {
	cachedUser, err := s.redis.GetUser(ctx, id.String())
	if err != nil {
		s.log.Warn("failed to get user from cache", "error", err)
	}

	if cachedUser != nil {
		userData, _ := json.Marshal(cachedUser)
		var user sqlc.User
		if err := json.Unmarshal(userData, &user); err == nil {
			s.log.Debug("user retrieved from cache", "user_id", id)
			return &user, nil
		}
	}

	user, err := s.repo.GetUserById(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.redis.SetUser(ctx, id.String(), user, 15*time.Minute); err != nil {
		s.log.Warn("failed to cache user", "error", err)
	}

	return &user, nil
}

func (s *BookingService) GetServiceById(ctx context.Context, id uuid.UUID) (*sqlc.Service, error) {
	cachedService, err := s.redis.GetService(ctx, id.String())
	if err != nil {
		s.log.Warn("failed to get service from cache", "error", err)
	}

	if cachedService != nil {
		serviceData, _ := json.Marshal(cachedService)
		var service sqlc.Service
		if err := json.Unmarshal(serviceData, &service); err == nil {
			s.log.Debug("service retrieved from cache", "service_id", id)
			return &service, nil
		}
	}

	service, err := s.repo.GetServiceById(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.redis.SetService(ctx, id.String(), service, 1*time.Hour); err != nil {
		s.log.Warn("failed to cache service", "error", err)
	}

	return &service, nil
}

func (s *BookingService) CheckBookingTime(ctx context.Context, serviceID uuid.UUID, bookingTime time.Time) (bool, error) {
	service, err := s.GetServiceById(ctx, serviceID)
	if err != nil {
		return false, errors.ErrNotFound
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	qtx := s.repo.WithTx(tx)

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

	return len(bookings) == 0, nil
}

func (s *BookingService) GetBooking(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID) (*sqlc.GetBookingByIDRow, error) {
	cachedBooking, err := s.redis.GetBooking(ctx, bookingID.String())
	if err != nil {
		s.log.Warn("failed to get booking from cache", "error", err)
	}

	if cachedBooking != nil {
		var booking sqlc.GetBookingByIDRow
		if err := json.Unmarshal(cachedBooking, &booking); err == nil {
			if !s.CheckBookingOwnerships(ctx, userID, booking.ClientID, booking.PerformerID) {
				return nil, errors.ErrForbidden
			}
			s.log.Debug("booking retrieved from cache", "booking_id", bookingID)
			return &booking, nil
		}
	}

	res, err := s.repo.GetBookingByID(ctx, bookingID)
	if err != nil {
		return nil, errors.ErrNotFound
	}

	if !s.CheckBookingOwnerships(ctx, userID, res.ClientID, res.PerformerID) {
		return nil, errors.ErrForbidden
	}

	if err := s.redis.SetBooking(ctx, bookingID.String(), res, 5*time.Minute); err != nil {
		s.log.Warn("failed to cache booking", "error", err)
	}

	return &res, nil
}

func (s *BookingService) GetBookings(ctx context.Context, userID uuid.UUID) ([]sqlc.Booking, error) {
	user, err := s.GetUserById(ctx, userID)
	if err != nil {
		return nil, err
	}

	role := string(user.Role)

	cachedBookings, err := s.redis.GetUserBookings(ctx, userID.String(), role)
	if err != nil {
		s.log.Warn("failed to get user bookings from cache", "error", err)
	}

	if cachedBookings != nil {
		var bookings []sqlc.Booking
		if err := json.Unmarshal(cachedBookings, &bookings); err == nil {
			s.log.Debug("user bookings retrieved from cache", "user_id", userID, "role", role)
			return bookings, nil
		}
	}

	var bookings []sqlc.Booking
	switch user.Role {
	case "performer":
		bookings, err = s.repo.GetBookingByPerformerID(ctx, userID)
	case "client":
		bookings, err = s.repo.GetBookingByClientID(ctx, userID)
	default:
		return nil, errors.ErrForbidden
	}

	if err != nil {
		return nil, err
	}

	if err := s.redis.SetUserBookings(ctx, userID.String(), role, bookings, 5*time.Minute); err != nil {
		s.log.Warn("failed to cache user bookings", "error", err)
	}

	return bookings, nil
}

func (s *BookingService) invalidateCaches(ctx context.Context, bookingID uuid.UUID, clientID, performerID uuid.UUID) {
	if err := s.redis.InvalidateBooking(ctx, bookingID.String()); err != nil {
		s.log.Warn("failed to invalidate booking cache", "error", err)
	}

	if err := s.redis.InvalidateUserBookings(ctx, clientID.String()); err != nil {
		s.log.Warn("failed to invalidate client bookings cache", "error", err)
	}

	if err := s.redis.InvalidateUserBookings(ctx, performerID.String()); err != nil {
		s.log.Warn("failed to invalidate performer bookings cache", "error", err)
	}
}

func (s *BookingService) CreateBooking(ctx context.Context, ClientID uuid.UUID, booking dto.CreateBookingRequest) (uuid.UUID, error) {
	user, err := s.GetUserById(ctx, ClientID)
	if err != nil {
		s.log.Error("failed to get user", "error", err)
		return uuid.Nil, err
	}
	if user.Role != "client" {
		return uuid.Nil, errors.ErrForbidden
	}

	ok, err := s.CheckBookingTime(ctx, booking.ServiceID, booking.BookingTime)
	if err != nil {
		s.log.Info(err.Error())
		return uuid.Nil, err
	}
	if !ok {
		return uuid.Nil, errors.ErrTimeBusy
	}

	service, err := s.GetServiceById(ctx, booking.ServiceID)
	if err != nil {
		return uuid.Nil, errors.ErrNotFound
	}

	var discount sqlc.Discount
	var discountId uuid.NullUUID
	if booking.DiscountID != uuid.Nil {
		discountId = uuid.NullUUID{UUID: booking.DiscountID, Valid: true}
		discount, err = s.repo.GetDiscountById(ctx, booking.DiscountID)
		if err != nil {
			return uuid.Nil, errors.ErrNotFound
		}
		if discount.ServiceID != booking.ServiceID || discount.ValidTo.Before(booking.BookingTime) || discount.ValidFrom.After(booking.BookingTime) {
			return uuid.Nil, errors.ErrInvalidInput
		}
		if discount.UsedCount+1 > discount.MaxUses {
			return uuid.Nil, errors.ErrInvalidInput
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

	go s.invalidateCaches(context.Background(), id, ClientID, service.PerformerID)

	go s.publishEvent(BookingCreated, id, service.PerformerID, service.Title, booking.BookingTime)

	return id, nil
}

func (s *BookingService) CancelBooking(ctx context.Context, userID uuid.UUID, booking *sqlc.GetBookingByIDRow) error {
	if !s.CheckBookingOwnerships(ctx, userID, booking.ClientID, booking.PerformerID) {
		return errors.ErrForbidden
	}

	if err := s.repo.CancelBooking(ctx, booking.ID); err != nil {
		s.log.Error("failed to cancel booking", "error", err)
		return err
	}

	go s.invalidateCaches(context.Background(), booking.ID, booking.ClientID, booking.PerformerID)

	go s.publishBoth(BookingCancelled, booking)
	return nil
}

func (s *BookingService) SubmitBooking(ctx context.Context, userID uuid.UUID, booking *sqlc.GetBookingByIDRow) error {
	if booking.Status == "confirmed" {
		return errors.ErrAlreadyDone
	}
	if booking.PerformerID != userID {
		return errors.ErrForbidden
	}

	if err := s.repo.SubmitBooking(ctx, booking.ID); err != nil {
		s.log.Error("failed to submit booking", "error", err)
		return err
	}

	go s.invalidateCaches(context.Background(), booking.ID, booking.ClientID, booking.PerformerID)

	go s.publishEvent(BookingSubmit, booking.ID, booking.ClientID, booking.ServiceTitle, booking.BookingTime)

	return nil
}

func (s *BookingService) CompleteBooking(ctx context.Context, userID uuid.UUID, booking *sqlc.GetBookingByIDRow) error {
	if booking.Status == "completed" {
		return errors.ErrAlreadyDone
	}
	if !s.CheckBookingOwnerships(ctx, userID, booking.ClientID, booking.PerformerID) {
		return errors.ErrForbidden
	}

	if err := s.repo.CompletedBooking(ctx, booking.ID); err != nil {
		s.log.Error("failed to submit booking", "error", err)
		return err
	}

	go s.invalidateCaches(context.Background(), booking.ID, booking.ClientID, booking.PerformerID)

	return nil
}

func (s *BookingService) UpdateBooking(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID, bookingData dto.PatchBookingRequest) error {
	booking, err := s.GetBooking(ctx, userID, bookingID)
	if err != nil {
		s.log.Error("failed to get booking", "error", err)
		return errors.ErrNotFound
	}

	if !s.CheckBookingOwnerships(ctx, userID, booking.ClientID, booking.PerformerID) {
		return errors.ErrForbidden
	}

	var bookingTime time.Time
	if !bookingData.BookingTime.IsZero() {
		ok, err := s.CheckBookingTime(ctx, booking.ServiceID, bookingData.BookingTime)
		if err != nil {
			return err
		}
		if !ok {
			return errors.ErrTimeBusy
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

		go s.invalidateCaches(context.Background(), bookingID, booking.ClientID, booking.PerformerID)

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

		go s.invalidateCaches(context.Background(), bookingID, booking.ClientID, booking.PerformerID)

		go s.publishEvent(BookingUpdated2, booking.ID, booking.PerformerID, booking.ServiceTitle, req.BookingTime)
	}

	return nil
}

func (s *BookingService) DeleteBooking(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID) error {
	booking, err := s.GetBooking(ctx, userID, bookingID)
	if err != nil {
		s.log.Error("failed to get booking", "error", err)
		return errors.ErrNotFound
	}

	if !s.CheckBookingOwnerships(ctx, userID, booking.ClientID, booking.PerformerID) {
		return errors.ErrForbidden
	}

	if err := s.repo.DeleteBooking(ctx, bookingID); err != nil {
		return err
	}

	go s.invalidateCaches(context.Background(), bookingID, booking.ClientID, booking.PerformerID)

	return nil
}

func (s *BookingService) CheckBookingOwnerships(ctx context.Context, userID uuid.UUID, clientID uuid.UUID, performerID uuid.UUID) bool {
	user, err := s.GetUserById(ctx, userID)
	if err != nil {
		s.log.Error("failed to get user", "error", err)
		return false
	}
	return clientID == userID || performerID == userID || user.Role == "admin"
}

func (s *BookingService) ServiceExists(ctx context.Context, serviceID uuid.UUID) (bool, error) {
	ok, err := s.repo.ServiceExists(ctx, serviceID)
	if err != nil {
		return false, err
	}
	return ok, nil
}

func (s *BookingService) publishEvent(event BookingEventType, bookingID uuid.UUID, userID uuid.UUID, serviceTitle string, bookingTime time.Time) {
	user, err := s.GetUserById(context.Background(), userID)
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
