package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	handlerlib "github.com/Artem09076/dp/backend/core_service/internal/lib/api/handler"
	"github.com/Artem09076/dp/backend/core_service/internal/lib/api/response"
	"github.com/Artem09076/dp/backend/core_service/internal/presentation/reviews/dto"
	sqlc "github.com/Artem09076/dp/backend/core_service/internal/storage/db"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type ReviewService interface {
	CreateReview(ctx context.Context, userID uuid.UUID, req dto.CreateReviewRequest) (*sqlc.Review, error)
	GetReviewByBookingID(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID) (*sqlc.Review, error)
	GetReviewsByServiceID(ctx context.Context, serviceID uuid.UUID, page int, limit int) ([]sqlc.Review, error)
	PatchReview(ctx context.Context, userID uuid.UUID, reviewID uuid.UUID, req dto.PatchReviewRequest) error
	DeleteReview(ctx context.Context, userID uuid.UUID, reviewID uuid.UUID) error
	GetReviewByID(ctx context.Context, reviewID uuid.UUID) (*sqlc.Review, error)
}

type ReviewHandler struct {
	service ReviewService
	log     *slog.Logger
}

func NewReviewHandler(service ReviewService, log *slog.Logger) *ReviewHandler {
	return &ReviewHandler{
		service: service,
		log:     log,
	}
}

func (h *ReviewHandler) convertToReviewResponse(review *sqlc.Review) dto.ReviewResponse {
	resp := dto.ReviewResponse{
		ID:        review.ID.String(),
		BookingID: review.BookingID.String(),
		Rating:    review.Rating,
		CreatedAt: review.CreatedAt.String(),
		UpdatedAt: review.UpdatedAt.String(),
	}

	if review.Comment.Valid {
		resp.Comment = &review.Comment.String
	}

	return resp
}

func (h *ReviewHandler) CreateReview() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := h.log.With("op", "reviews.handlers.CreateReview")
		w.Header().Set("Content-Type", "application/json")
		userID, err := handlerlib.GetUserIDFromClaims(r.Context())
		if err != nil {
			h.handleError(w, r, err)
			return
		}

		var req dto.CreateReviewRequest
		if err := render.DecodeJSON(r.Body, &req); err != nil {
			log.Error("Failed to decode request body", "error", err.Error())
			h.handleError(w, r, handlerlib.ErrInvalidInput)
			return
		}
		review, err := h.service.CreateReview(r.Context(), userID, req)
		if err != nil {
			h.handleError(w, r, err)
			return
		}
		resp := h.convertToReviewResponse(review)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Error("Failed to encode response", "error", err.Error())
			h.handleError(w, r, err)
			return
		}
	}
}

func (h *ReviewHandler) GetReviewByBookingID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := h.log.With("op", "reviews.handlers.GetReviewByBookingID")
		w.Header().Set("Content-Type", "application/json")
		userID, err := handlerlib.GetUserIDFromClaims(r.Context())
		if err != nil {
			h.handleError(w, r, err)
			return
		}
		bookingIDStr := chi.URLParam(r, "bookingID")
		if bookingIDStr == "" {
			h.handleError(w, r, handlerlib.ErrInvalidInput)
			return
		}

		bookingID, err := uuid.Parse(bookingIDStr)
		if err != nil {
			h.handleError(w, r, handlerlib.ErrInvalidInput)
			return
		}

		review, err := h.service.GetReviewByBookingID(r.Context(), userID, bookingID)
		if err != nil {
			h.handleError(w, r, err)
			return
		}
		resp := h.convertToReviewResponse(review)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Error("Failed to encode response", "error", err.Error())
			h.handleError(w, r, err)
			return
		}
	}
}

func (h *ReviewHandler) GetReviewsByServiceID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := h.log.With("op", "reviews.handlers.GetReviewsByServiceID")
		w.Header().Set("Content-Type", "application/json")
		serviceIDStr := chi.URLParam(r, "serviceID")
		if serviceIDStr == "" {
			h.handleError(w, r, handlerlib.ErrInvalidInput)
			return
		}

		serviceID, err := uuid.Parse(serviceIDStr)
		if err != nil {
			h.handleError(w, r, handlerlib.ErrInvalidInput)
			return
		}

		page, limit := handlerlib.GetPaginationParams(r.Context())
		reviews, err := h.service.GetReviewsByServiceID(r.Context(), serviceID, page, limit)
		if err != nil {
			h.handleError(w, r, err)
			return
		}
		responses := make([]dto.ReviewResponse, len(reviews))
		for i, review := range reviews {
			responses[i] = dto.ReviewResponse{
				ID:        review.ID.String(),
				BookingID: review.BookingID.String(),
				Rating:    review.Rating,
				CreatedAt: review.CreatedAt.String(),
				UpdatedAt: review.UpdatedAt.String(),
			}
			if review.Comment.Valid {
				responses[i].Comment = &review.Comment.String
			}
		}
		if err := json.NewEncoder(w).Encode(responses); err != nil {
			log.Error("Failed to encode response", "error", err.Error())
			h.handleError(w, r, err)
			return
		}
	}
}

func (h *ReviewHandler) PatchReview() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := h.log.With("op", "reviews.handlers.PatchReview")
		w.Header().Set("Content-Type", "application/json")
		userID, err := handlerlib.GetUserIDFromClaims(r.Context())
		if err != nil {
			h.handleError(w, r, err)
			return
		}
		reviewIDStr := chi.URLParam(r, "reviewID")
		if reviewIDStr == "" {
			h.handleError(w, r, handlerlib.ErrInvalidInput)
			return
		}

		reviewID, err := uuid.Parse(reviewIDStr)
		if err != nil {
			h.handleError(w, r, handlerlib.ErrInvalidInput)
			return
		}

		var req dto.PatchReviewRequest
		if err := render.DecodeJSON(r.Body, &req); err != nil {
			log.Error("Failed to decode request body", "error", err.Error())
			h.handleError(w, r, handlerlib.ErrInvalidInput)
			return
		}

		err = h.service.PatchReview(r.Context(), userID, reviewID, req)
		if err != nil {
			h.handleError(w, r, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *ReviewHandler) DeleteReview() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := h.log.With("op", "reviews.handlers.DeleteReview")
		w.Header().Set("Content-Type", "application/json")
		userID, err := handlerlib.GetUserIDFromClaims(r.Context())
		if err != nil {
			h.handleError(w, r, err)
			return
		}
		reviewIDStr := chi.URLParam(r, "reviewID")
		if reviewIDStr == "" {
			h.handleError(w, r, handlerlib.ErrInvalidInput)
			return
		}

		reviewID, err := uuid.Parse(reviewIDStr)
		if err != nil {
			h.handleError(w, r, handlerlib.ErrInvalidInput)
			return
		}

		err = h.service.DeleteReview(r.Context(), userID, reviewID)
		if err != nil {
			log.Error("Failed to delete review", "error", err.Error())
			h.handleError(w, r, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *ReviewHandler) GetReviewByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := h.log.With("op", "reviews.handlers.GetReviewByID")
		w.Header().Set("Content-Type", "application/json")
		reviewIDStr := chi.URLParam(r, "reviewID")
		if reviewIDStr == "" {
			h.handleError(w, r, handlerlib.ErrInvalidInput)
			return
		}

		reviewID, err := uuid.Parse(reviewIDStr)
		if err != nil {
			h.handleError(w, r, handlerlib.ErrInvalidInput)
			return
		}

		review, err := h.service.GetReviewByID(r.Context(), reviewID)
		if err != nil {
			h.handleError(w, r, err)
			return
		}
		resp := h.convertToReviewResponse(review)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Error("Failed to encode response", "error", err.Error())
			h.handleError(w, r, err)
			return
		}
	}
}
func (h *ReviewHandler) handleError(w http.ResponseWriter, r *http.Request, err error) {
	switch err {
	case handlerlib.ErrNotFound:
		w.WriteHeader(http.StatusNotFound)
	case handlerlib.ErrForbidden:
		w.WriteHeader(http.StatusForbidden)
	case handlerlib.ErrInvalidInput:
		w.WriteHeader(http.StatusBadRequest)
	case handlerlib.ErrAlreadyDone:
		w.WriteHeader(http.StatusConflict)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	render.JSON(w, r, response.Error(err.Error()))
}
