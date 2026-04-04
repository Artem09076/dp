package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	handlerlib "github.com/Artem09076/dp/backend/core_service/internal/lib/api/handler"
	"github.com/Artem09076/dp/backend/core_service/internal/lib/api/response"
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
		log := h.log.With(slog.String("op", op))
		w.Header().Set("Content-Type", "application/json")
		var body dto.CreateDiscountRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			log.Error("Failed to decode request body", slog.String("Error", err.Error()))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid request body"))
			return
		}
		if body.ValidFrom.After(body.ValidTo) {
			h.log.Error("Invalid discount type")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid discount time"))
			return
		}
		validDiscountType := sqlc.NullDiscoutnType{}

		if err := validDiscountType.Scan(body.Type); err != nil {
			h.log.Error("Invalid discount type", slog.String("discountType", body.Type), slog.String("Error", err.Error()))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid discount type"))
			return
		}

		userID, err := handlerlib.GetUserIDFromClaims(r.Context())
		if err != nil {
			log.Error(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error(err.Error()))
			return
		}

		serviceIDStr := chi.URLParam(r, "id")
		serviceID, err := uuid.Parse(serviceIDStr)
		if err != nil {
			log.Error("Failed to parse serviceID", slog.String("serviceID", serviceIDStr), slog.String("Error", err.Error()))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid serviceID"))
			return
		}
		if ok, err := h.service.CheckServiceOwnership(r.Context(), userID, serviceID); err != nil {
			log.Error("Failed to check service ownership", slog.String("serviceID", serviceID.String()), slog.String("Error", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to check service ownership"))
			return
		} else if !ok {
			log.Warn("User does not own the service", slog.String("userID", userID.String()), slog.String("serviceID", serviceID.String()))
			w.WriteHeader(http.StatusForbidden)
			render.JSON(w, r, response.Error("you do not have permission to create a discount for this service"))
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
			log.Error("Failed to create discount", slog.String("Error", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to create discount"))
			return
		}

		if err := json.NewEncoder(w).Encode(discount); err != nil {
			log.Error("Failed to encode response", slog.String("Error", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to encode response"))
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
			log.Error("Failed to parse discount ID", slog.String("discountID", discountIDStr), slog.String("Error", err.Error()))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid discount ID"))
			return
		}
		discount, err := h.service.GetDiscount(r.Context(), discountID)
		if err != nil {
			log.Error("Failed to get discount", slog.String("discountID", discountID.String()), slog.String("Error", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to get discount"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(discount); err != nil {
			log.Error("Failed to encode response", slog.String("Error", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to encode response"))
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
			log.Error(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error(err.Error()))
			return
		}

		serviceIDStr := chi.URLParam(r, "serviceID")
		serviceID, err := uuid.Parse(serviceIDStr)
		if err != nil {
			log.Error("Failed to parse serviceID", slog.String("serviceID", serviceIDStr), slog.String("Error", err.Error()))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid serviceID"))
			return
		}
		if ok, err := h.service.CheckServiceOwnership(r.Context(), userID, serviceID); err != nil {
			log.Error("Failed to check service ownership", slog.String("serviceID", serviceID.String()), slog.String("Error", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to check service ownership"))
			return
		} else if !ok {
			log.Warn("User does not own the service", slog.String("userID", userID.String()), slog.String("serviceID", serviceID.String()))
			w.WriteHeader(http.StatusForbidden)
			render.JSON(w, r, response.Error("you do not have permission to create a discount for this service"))
			return
		}
		DiscountIDStr := chi.URLParam(r, "id")
		DiscountID, err := uuid.Parse(DiscountIDStr)
		if err != nil {
			log.Error("Failed to parse discount_id", slog.String("discount_id", DiscountIDStr))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid service_id"))
			return
		}
		var updateDiscountObj dto.PatchDiscountRequest
		if err := json.NewDecoder(r.Body).Decode(&updateDiscountObj); err != nil {
			log.Error("Failed to decode request body", slog.String("Error", err.Error()))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid request body"))
			return
		}
		if err := h.service.UpdateDiscount(r.Context(), DiscountID, updateDiscountObj); err != nil {
			log.Error("internal server error", slog.String("Error", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("internal server error"))
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
			log.Error(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error(err.Error()))
			return
		}
		serviceIDStr := chi.URLParam(r, "serviceID")
		serviceID, err := uuid.Parse(serviceIDStr)
		if err != nil {
			log.Error("Failed to parse serviceID", slog.String("serviceID", serviceIDStr), slog.String("Error", err.Error()))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid serviceID"))
			return
		}
		if ok, err := h.service.CheckServiceOwnership(r.Context(), userID, serviceID); err != nil {
			log.Error("Failed to check service ownership", slog.String("serviceID", serviceID.String()), slog.String("Error", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to check service ownership"))
			return
		} else if !ok {
			log.Warn("User does not own the service", slog.String("userID", userID.String()), slog.String("serviceID", serviceID.String()))
			w.WriteHeader(http.StatusForbidden)
			render.JSON(w, r, response.Error("you do not have permission to create a discount for this service"))
			return
		}
		DiscountIDStr := chi.URLParam(r, "id")
		DiscountID, err := uuid.Parse(DiscountIDStr)
		if err != nil {
			log.Error("Failed to parse discount_id", slog.String("discount_id", DiscountIDStr))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid service_id"))
			return
		}
		if err := h.service.DeleteDiscount(r.Context(), DiscountID); err != nil {
			log.Error("internal server error", slog.String("Error", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("internal server error"))
			return
		}

		w.WriteHeader(http.StatusNoContent)

	}
}
