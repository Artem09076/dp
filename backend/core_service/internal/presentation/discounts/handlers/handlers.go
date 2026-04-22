package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	apierrors "github.com/Artem09076/dp/backend/core_service/internal/lib/api/errors"

	handlerlib "github.com/Artem09076/dp/backend/core_service/internal/lib/api/handler"
	"github.com/Artem09076/dp/backend/core_service/internal/presentation/discounts/dto"
	sqlc "github.com/Artem09076/dp/backend/core_service/internal/storage/db"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type DiscountsService interface {
	CreateDiscount(ctx context.Context, userID uuid.UUID, serviceID uuid.UUID, discountData sqlc.CreateDiscountParams) (*sqlc.Discount, error)
	CheckServiceOwnership(ctx context.Context, userID uuid.UUID, serviceID uuid.UUID) (bool, error)
	GetDiscount(ctx context.Context, discountID uuid.UUID) (*sqlc.Discount, error)
	UpdateDiscount(ctx context.Context, discountID uuid.UUID, updateDiscountObj dto.PatchDiscountRequest) error
	DeleteDiscount(ctx context.Context, discountID uuid.UUID) error
}
type DiscountsHandler struct {
	service DiscountsService
	log     *slog.Logger
}

func NewDiscountsHandler(service DiscountsService, log *slog.Logger) *DiscountsHandler {
	return &DiscountsHandler{
		service: service,
		log:     log,
	}
}

func (h *DiscountsHandler) CreateDiscount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "discounts.handlers.CreateDiscount"
		w.Header().Set("Content-Type", "application/json")

		userID, err := handlerlib.GetUserIDFromClaims(r.Context())
		if err != nil {
			h.WriteError(w, r, apierrors.ErrUnauthorized, op)
			return
		}

		var body dto.CreateDiscountRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			h.WriteError(w, r, apierrors.ErrInvalidInput, op)
			return
		}

		if body.ValidFrom.After(body.ValidTo) {
			h.WriteError(w, r, apierrors.ErrInvalidInput, op)
			return
		}
		validDiscountType := sqlc.NullDiscoutnType{}

		if err := validDiscountType.Scan(body.Type); err != nil {
			h.WriteError(w, r, apierrors.ErrInvalidInput, op)
			return
		}

		serviceIDStr := chi.URLParam(r, "id")
		serviceID, err := uuid.Parse(serviceIDStr)
		if err != nil {
			h.WriteError(w, r, apierrors.ErrInvalidInput, op)
			return
		}

		param := sqlc.CreateDiscountParams{
			ServiceID: serviceID,
			Type:      validDiscountType.DiscoutnType,
			Value:     int32(body.Value),
			ValidFrom: body.ValidFrom,
			ValidTo:   body.ValidTo,
			MaxUses:   int32(body.MaxUses),
			UsedCount: 0,
		}

		discount, err := h.service.CreateDiscount(r.Context(), userID, serviceID, param)
		if err != nil {
			h.WriteError(w, r, err, op)
			return
		}

		resp := dto.DiscountResponse{
			ID:        discount.ID.String(),
			ServiceID: discount.ServiceID.String(),
			Type:      string(discount.Type),
			Value:     int(discount.Value),
			ValidFrom: discount.ValidFrom,
			ValidTo:   discount.ValidTo,
			MaxUses:   int(discount.MaxUses),
			UsedCount: int(discount.UsedCount),
			CreatedAt: discount.CreatedAt,
		}

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			h.WriteError(w, r, err, op)
			return
		}
	}
}

func (h *DiscountsHandler) GetDiscount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "discounts.handlers.GetDiscount"
		log := h.log.With(slog.String("op", op))
		discountIDStr := chi.URLParam(r, "id")
		discountID, err := uuid.Parse(discountIDStr)
		if err != nil {
			h.WriteError(w, r, apierrors.ErrInvalidInput, op)
			return
		}
		discount, err := h.service.GetDiscount(r.Context(), discountID)
		if err != nil {
			h.WriteError(w, r, err, op)
			return
		}

		resp := dto.DiscountResponse{
			ID:        discount.ID.String(),
			ServiceID: discount.ServiceID.String(),
			Type:      string(discount.Type),
			Value:     int(discount.Value),
			ValidFrom: discount.ValidFrom,
			ValidTo:   discount.ValidTo,
			MaxUses:   int(discount.MaxUses),
			UsedCount: int(discount.UsedCount),
			CreatedAt: discount.CreatedAt,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Info(err.Error())
			h.WriteError(w, r, err, op)
			return
		}
	}
}

func (h *DiscountsHandler) UpdateDiscount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "discounts.handlers.UpdateDiscount"
		log := h.log.With(slog.String("op", op))
		userID, err := handlerlib.GetUserIDFromClaims(r.Context())
		if err != nil {
			h.WriteError(w, r, apierrors.ErrUnauthorized, op)
			return
		}

		serviceIDStr := chi.URLParam(r, "serviceID")
		serviceID, err := uuid.Parse(serviceIDStr)
		if err != nil {
			h.WriteError(w, r, apierrors.ErrInvalidInput, op)
			return
		}
		if ok, err := h.service.CheckServiceOwnership(r.Context(), userID, serviceID); err != nil {
			h.WriteError(w, r, apierrors.ErrInvalidInput, op)
			return
		} else if !ok {
			h.WriteError(w, r, apierrors.ErrInvalidInput, op)
			return
		}
		DiscountIDStr := chi.URLParam(r, "id")
		DiscountID, err := uuid.Parse(DiscountIDStr)
		if err != nil {
			h.WriteError(w, r, apierrors.ErrInvalidInput, op)
			return
		}
		var updateDiscountObj dto.PatchDiscountRequest
		if err := json.NewDecoder(r.Body).Decode(&updateDiscountObj); err != nil {
			h.WriteError(w, r, apierrors.ErrInvalidInput, op)
			return
		}
		if err := h.service.UpdateDiscount(r.Context(), DiscountID, updateDiscountObj); err != nil {
			log.Info(err.Error())
			h.WriteError(w, r, err, op)
			return
		}

		w.WriteHeader(http.StatusNoContent)

	}
}

func (h *DiscountsHandler) DeleteDiscount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "discounts.handlers.DeleteDiscount"
		log := h.log.With(slog.String("op", op))
		userID, err := handlerlib.GetUserIDFromClaims(r.Context())
		if err != nil {
			h.WriteError(w, r, apierrors.ErrUnauthorized, op)
			return
		}
		serviceIDStr := chi.URLParam(r, "serviceID")
		serviceID, err := uuid.Parse(serviceIDStr)
		if err != nil {
			h.WriteError(w, r, apierrors.ErrInvalidInput, op)
			return
		}
		if ok, err := h.service.CheckServiceOwnership(r.Context(), userID, serviceID); err != nil {
			h.WriteError(w, r, apierrors.ErrInvalidInput, op)
			return
		} else if !ok {
			h.WriteError(w, r, apierrors.ErrInvalidInput, op)
			return
		}
		DiscountIDStr := chi.URLParam(r, "id")
		DiscountID, err := uuid.Parse(DiscountIDStr)
		if err != nil {
			h.WriteError(w, r, apierrors.ErrInvalidInput, op)
			return
		}
		if err := h.service.DeleteDiscount(r.Context(), DiscountID); err != nil {
			log.Info(err.Error())
			h.WriteError(w, r, err, op)
			return
		}

		w.WriteHeader(http.StatusNoContent)

	}
}

func (h *DiscountsHandler) WriteError(w http.ResponseWriter, r *http.Request, err error, op string) {
	apiErr := apierrors.MapError(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiErr.StatusCode)
	render.JSON(w, r, apierrors.NewErrorResponse(err))
}
