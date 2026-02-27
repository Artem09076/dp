package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Artem09076/dp/backend/core_service/internal/lib/api/response"
	"github.com/Artem09076/dp/backend/core_service/internal/presentation/services/dto"
	sqlc "github.com/Artem09076/dp/backend/core_service/internal/storage/db"
	"github.com/go-chi/render"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type ServicesService interface {
	CreateService(ctx context.Context, createServiceObject dto.CreateServiceRequest) (*sqlc.Service, error)
}

type ServiceHandler struct {
	service ServicesService
	log     *slog.Logger
}

func NewServiceHandler(service ServicesService, log *slog.Logger) *ServiceHandler {
	return &ServiceHandler{
		service: service,
		log:     log,
	}
}

func (h *ServiceHandler) CreateService() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "service.handlers.CreateService"
		log := h.log.With(slog.String("op", op))
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
		var body dto.CreateServiceRequest

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "can't decode JSON body", http.StatusBadRequest)
			return
		}

		body.PerformerID = userID

		service, err := h.service.CreateService(r.Context(), body)
		if err != nil {
			log.Error("internal server error", slog.String("Error", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("internal server error"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(service); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}

	}
}
