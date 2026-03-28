package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Artem09076/dp/backend/booking_service/internal/lib/api/response"
	"github.com/Artem09076/dp/backend/booking_service/internal/presentation/booking/dto"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type BookingService interface {
	CreateBooking(ctx context.Context, userID uuid.UUID, bookingData dto.CreateBookingRequest) (uuid.UUID, error)
	CancelBooking(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID) error
	SubmitBooking(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID) error
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
		claims, ok := r.Context().Value("claims").(*jwt.MapClaims)
		if !ok || claims == nil {
			log.Error("Failed to parse claims")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid authentication claims"))
			return
		}
		userID, err := uuid.Parse((*claims)["user_id"].(string))
		if err != nil {
			log.Error("Failed to parse user_id")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid user_id"))
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
			log.Error("internal server error", slog.String("Error", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("internal server error"))
			return
		}
		render.JSON(w, r, dto.CreateBookingResponse{ID: id})
	}
}

func (h *BookingHandler) CancelBooking() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := h.log.With(slog.String("op", "booking.handlers.CancelBooking"))
		claims, ok := r.Context().Value("claims").(*jwt.MapClaims)
		if !ok || claims == nil {
			log.Error("Failed to parse claims")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid authentication claims"))
			return
		}
		userID, err := uuid.Parse((*claims)["user_id"].(string))
		if err != nil {
			log.Error("Failed to parse user_id")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid user_id"))
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
		err = h.bookingService.CancelBooking(r.Context(), userID, bookingID)
		if err != nil {
			log.Error("internal server error", slog.String("Error", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("internal server error"))
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *BookingHandler) SubmitBooking() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := h.log.With(slog.String("op", "booking.handlers.CancelBooking"))
		claims, ok := r.Context().Value("claims").(*jwt.MapClaims)
		if !ok || claims == nil {
			log.Error("Failed to parse claims")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid authentication claims"))
			return
		}
		userID, err := uuid.Parse((*claims)["user_id"].(string))
		if err != nil {
			log.Error("Failed to parse user_id")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid user_id"))
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
		err = h.bookingService.SubmitBooking(r.Context(), userID, bookingID)
		if err != nil {
			log.Error("internal server error", slog.String("Error", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("internal server error"))
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
