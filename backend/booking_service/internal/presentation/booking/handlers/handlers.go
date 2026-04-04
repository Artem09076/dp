package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	handlerlib "github.com/Artem09076/dp/backend/booking_service/internal/lib/api/handler"
	"github.com/Artem09076/dp/backend/booking_service/internal/lib/api/response"
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

func (h *BookingHandler) CreateBooking() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := h.log.With(slog.String("op", "booking.handlers.CreateBooking"))
		userID, err := handlerlib.GetUserIDFromClaims(r.Context())
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			log.Error(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error(err.Error()))
			return
		}

		var req dto.CreateBookingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Error("Failed to decode request body", slog.String("Error", err.Error()))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid request body"))
			return
		}
		id, err := h.bookingService.CreateBooking(r.Context(), userID, req)

		if err != nil {
			h.handleError(w, r, err)
			return
		}
		render.JSON(w, r, dto.CreateBookingResponse{ID: id})
	}
}

func (h *BookingHandler) CancelBooking() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := h.log.With(slog.String("op", "booking.handlers.CancelBooking"))
		userID, err := handlerlib.GetUserIDFromClaims(r.Context())
		if err != nil {
			log.Error(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error(err.Error()))
			return
		}
		bookingIDStr := chi.URLParam(r, "id")
		bookingID, err := uuid.Parse(bookingIDStr)
		if err != nil {
			log.Error("Failed to parse booking_id", slog.String("Error", err.Error()))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid booking_id"))
			return
		}
		w.Header().Set("Content-Type", "application/json")

		booking, err := h.bookingService.GetBooking(r.Context(), userID, bookingID)
		if err != nil {
			h.handleError(w, r, err)
			return
		}

		err = h.bookingService.CancelBooking(r.Context(), userID, booking)
		if err != nil {
			h.handleError(w, r, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *BookingHandler) SubmitBooking() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := h.log.With(slog.String("op", "booking.handlers.SubmitBooking"))
		userID, err := handlerlib.GetUserIDFromClaims(r.Context())
		if err != nil {
			log.Error(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error(err.Error()))
			return
		}
		bookingIDStr := chi.URLParam(r, "id")
		bookingID, err := uuid.Parse(bookingIDStr)
		if err != nil {
			log.Error("Failed to parse booking_id", slog.String("Error", err.Error()))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid booking_id"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		booking, err := h.bookingService.GetBooking(r.Context(), userID, bookingID)
		if err != nil {
			h.handleError(w, r, err)
			return
		}

		if booking.Status == "cancelled" {
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("You can not submit a cancelled booking"))
			return
		}

		err = h.bookingService.SubmitBooking(r.Context(), userID, booking)
		if err != nil {
			h.handleError(w, r, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *BookingHandler) PatchBooking() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := h.log.With(slog.String("op", "booking.handlers.PatchBooking"))
		userID, err := handlerlib.GetUserIDFromClaims(r.Context())
		if err != nil {
			log.Error(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error(err.Error()))
			return
		}
		bookingIDStr := chi.URLParam(r, "id")
		bookingID, err := uuid.Parse(bookingIDStr)
		if err != nil {
			log.Error("Failed to parse booking_id", slog.String("Error", err.Error()))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid booking_id"))
			return
		}
		var req dto.PatchBookingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Error("Failed to decode request body", slog.String("Error", err.Error()))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid request body"))
			return
		}
		err = h.bookingService.UpdateBooking(r.Context(), userID, bookingID, req)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			h.handleError(w, r, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *BookingHandler) GetBooking() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := h.log.With(slog.String("op", "booking.handlers.GetBooking"))
		log.Info("GetBooking handler called")
		userID, err := handlerlib.GetUserIDFromClaims(r.Context())
		if err != nil {
			log.Error(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error(err.Error()))
			return
		}

		bookingIDStr := chi.URLParam(r, "id")
		bookingID, err := uuid.Parse(bookingIDStr)
		if err != nil {
			log.Error("Failed to parse booking_id", slog.String("booking_id", bookingIDStr))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid booking_id"))
			return
		}

		w.Header().Set("Content-Type", "application/json")

		booking, err := h.bookingService.GetBooking(r.Context(), userID, bookingID)
		if err != nil {
			h.handleError(w, r, err)
			return
		}

		if err := json.NewEncoder(w).Encode(booking); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

func (h *BookingHandler) GetBookings() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := h.log.With(slog.String("op", "booking.handlers.GetBooking"))
		userID, err := handlerlib.GetUserIDFromClaims(r.Context())
		if err != nil {
			log.Error(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error(err.Error()))
			return
		}

		w.Header().Set("Content-Type", "application/json")

		bookings, err := h.bookingService.GetBookings(r.Context(), userID)
		if err != nil {
			h.handleError(w, r, err)
			return
		}

		if err := json.NewEncoder(w).Encode(bookings); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

func (h *BookingHandler) handleError(w http.ResponseWriter, r *http.Request, err error) {
	switch err {
	case handlerlib.ErrNotFound:
		w.WriteHeader(http.StatusNotFound)
	case handlerlib.ErrForbidden:
		w.WriteHeader(http.StatusForbidden)
	case handlerlib.ErrInvalidInput:
		w.WriteHeader(http.StatusBadRequest)
	case handlerlib.ErrTimeBusy:
		w.WriteHeader(http.StatusConflict)
	case handlerlib.ErrAlreadyDone:
		w.WriteHeader(http.StatusConflict)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	render.JSON(w, r, response.Error(err.Error()))
}
