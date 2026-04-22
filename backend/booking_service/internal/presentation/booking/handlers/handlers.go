package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	apierrors "github.com/Artem09076/dp/backend/booking_service/internal/lib/api/errors"
	handlerlib "github.com/Artem09076/dp/backend/booking_service/internal/lib/api/handler"

	"github.com/Artem09076/dp/backend/booking_service/internal/presentation/booking/dto"
	sqlc "github.com/Artem09076/dp/backend/booking_service/internal/storage/db"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type BookingService interface {
	CreateBooking(ctx context.Context, userID uuid.UUID, bookingData dto.CreateBookingRequest) (uuid.UUID, error)
	CancelBooking(ctx context.Context, userID uuid.UUID, bookingID *sqlc.GetBookingByIDRow) error
	SubmitBooking(ctx context.Context, userID uuid.UUID, bookingID *sqlc.GetBookingByIDRow) error
	UpdateBooking(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID, bookingData dto.PatchBookingRequest) error
	GetBooking(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID) (*sqlc.GetBookingByIDRow, error)
	GetBookings(ctx context.Context, userID uuid.UUID) ([]sqlc.Booking, error)
	CheckBookingTime(ctx context.Context, serviceID uuid.UUID, bookingTime time.Time) (bool, error)
	ServiceExists(ctx context.Context, serviceID uuid.UUID) (bool, error)
	CheckBookingOwnerships(ctx context.Context, userID uuid.UUID, clientID uuid.UUID, performerID uuid.UUID) bool
	DeleteBooking(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID) error
}

type BookingHandler struct {
	bookingService BookingService
	log            *slog.Logger
}

func NewBookingHandler(bookingService BookingService, log *slog.Logger) *BookingHandler {
	return &BookingHandler{
		bookingService: bookingService,
		log:            log,
	}
}

func (h *BookingHandler) convertToBookingResponse(booking *sqlc.GetBookingByIDRow) dto.BookingResponse {
	resp := dto.BookingResponse{
		ID:           booking.ID.String(),
		ClientID:     booking.ClientID.String(),
		ServiceID:    booking.ServiceID.String(),
		ServiceTitle: booking.ServiceTitle,
		PerformerID:  booking.PerformerID.String(),
		BasePrice:    booking.BasePrice,
		FinalPrice:   booking.FinalPrice,
		BookingTime:  booking.BookingTime,
		Status:       string(booking.Status),
		CreatedAt:    booking.CreatedAt,
		UpdatedAt:    booking.UpdatedAt,
	}

	if booking.DiscountID.Valid {
		discountID := booking.DiscountID.UUID.String()
		resp.DiscountID = &discountID
	}

	if booking.DiscountType.Valid {
		discountType := string(booking.DiscountType.DiscoutnType)
		resp.DiscountType = &discountType
	}

	if booking.DiscountValue.Valid {
		resp.DiscountValue = &booking.DiscountValue.Int32
	}

	return resp
}

func (h *BookingHandler) convertToBookingListResponse(booking *sqlc.Booking) dto.BookingResponse {
	resp := dto.BookingResponse{
		ID:          booking.ID.String(),
		ClientID:    booking.ClientID.String(),
		ServiceID:   booking.ServiceID.String(),
		BasePrice:   booking.BasePrice,
		FinalPrice:  booking.FinalPrice,
		BookingTime: booking.BookingTime,
		Status:      string(booking.Status),
		CreatedAt:   booking.CreatedAt,
		UpdatedAt:   booking.UpdatedAt,
	}

	if booking.DiscountID.Valid {
		discountID := booking.DiscountID.UUID.String()
		resp.DiscountID = &discountID
	}

	return resp
}

func (h *BookingHandler) writeError(w http.ResponseWriter, r *http.Request, err error, op string) {
	h.log.Error(op, slog.String("error", err.Error()))
	apiErr := apierrors.MapError(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiErr.StatusCode)
	render.JSON(w, r, apierrors.NewErrorResponse(err))
}

func (h *BookingHandler) getUserID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	userID, err := handlerlib.GetUserIDFromClaims(r.Context())
	if err != nil {
		h.writeError(w, r, apierrors.ErrUnauthorized, "get_user_id")
		return uuid.Nil, false
	}
	return userID, true
}

func (h *BookingHandler) getBookingID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	bookingIDStr := chi.URLParam(r, "id")
	bookingID, err := uuid.Parse(bookingIDStr)
	if err != nil {
		h.writeError(w, r, apierrors.ErrInvalidInput, "parse_booking_id")
		return uuid.Nil, false
	}
	return bookingID, true
}

func (h *BookingHandler) CreateBooking() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		op := "booking.handlers.CreateBooking"
		w.Header().Set("Content-Type", "application/json")
		userID, ok := h.getUserID(w, r)
		if !ok {
			return
		}

		var req dto.CreateBookingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.writeError(w, r, apierrors.ErrInvalidInput, op)
			return
		}
		id, err := h.bookingService.CreateBooking(r.Context(), userID, req)

		if err != nil {
			h.writeError(w, r, err, op)
			return
		}
		render.JSON(w, r, dto.CreateBookingResponse{ID: id})
	}
}

func (h *BookingHandler) CancelBooking() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		op := "booking.handlers.CancelBooking"
		w.Header().Set("Content-Type", "application/json")

		userID, ok := h.getUserID(w, r)
		if !ok {
			return
		}

		bookingID, ok := h.getBookingID(w, r)
		if !ok {
			return
		}

		booking, err := h.bookingService.GetBooking(r.Context(), userID, bookingID)
		if err != nil {
			h.writeError(w, r, err, op)
			return
		}

		err = h.bookingService.CancelBooking(r.Context(), userID, booking)
		if err != nil {
			h.writeError(w, r, err, op)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *BookingHandler) SubmitBooking() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		op := "booking.handlers.SubmitBooking"
		w.Header().Set("Content-Type", "application/json")

		userID, ok := h.getUserID(w, r)
		if !ok {
			return
		}

		bookingID, ok := h.getBookingID(w, r)
		if !ok {
			return
		}

		booking, err := h.bookingService.GetBooking(r.Context(), userID, bookingID)
		if err != nil {
			h.writeError(w, r, err, op)
			return
		}

		if booking.Status == "cancelled" {
			h.writeError(w, r, apierrors.ErrAlreadyDone, op)
			return
		}

		err = h.bookingService.SubmitBooking(r.Context(), userID, booking)
		if err != nil {
			h.writeError(w, r, err, op)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *BookingHandler) PatchBooking() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		op := "booking.handlers.PatchBooking"
		w.Header().Set("Content-Type", "application/json")

		userID, ok := h.getUserID(w, r)
		if !ok {
			return
		}

		bookingID, ok := h.getBookingID(w, r)
		if !ok {
			return
		}

		var req dto.PatchBookingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.writeError(w, r, apierrors.ErrInvalidInput, op)
			return
		}

		if err := h.bookingService.UpdateBooking(r.Context(), userID, bookingID, req); err != nil {
			h.writeError(w, r, err, op)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
func (h *BookingHandler) GetBooking() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		op := "booking.handlers.GetBooking"
		w.Header().Set("Content-Type", "application/json")

		userID, ok := h.getUserID(w, r)
		if !ok {
			return
		}

		bookingID, ok := h.getBookingID(w, r)
		if !ok {
			return
		}

		booking, err := h.bookingService.GetBooking(r.Context(), userID, bookingID)
		if err != nil {
			h.writeError(w, r, err, op)
			return
		}

		resp := h.convertToBookingResponse(booking)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			h.writeError(w, r, err, op)
			return
		}
	}
}

func (h *BookingHandler) GetBookings() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		op := "booking.handlers.GetBookings"
		w.Header().Set("Content-Type", "application/json")

		userID, ok := h.getUserID(w, r)
		if !ok {
			return
		}

		bookings, err := h.bookingService.GetBookings(r.Context(), userID)
		if err != nil {
			h.writeError(w, r, err, op)
			return
		}

		responses := make([]dto.BookingResponse, len(bookings))
		for i, booking := range bookings {
			responses[i] = h.convertToBookingListResponse(&booking)
		}

		if err := json.NewEncoder(w).Encode(responses); err != nil {
			h.writeError(w, r, err, op)
			return
		}
	}
}

func (h *BookingHandler) DeleteBooking() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		op := "booking.handlers.DeleteBooking"
		w.Header().Set("Content-Type", "application/json")

		userID, ok := h.getUserID(w, r)
		if !ok {
			return
		}

		bookingID, ok := h.getBookingID(w, r)
		if !ok {
			return
		}

		if err := h.bookingService.DeleteBooking(r.Context(), userID, bookingID); err != nil {
			h.writeError(w, r, err, op)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
