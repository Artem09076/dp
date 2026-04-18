package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	handlerlib "github.com/Artem09076/dp/backend/core_service/internal/lib/api/handler"
	"github.com/Artem09076/dp/backend/core_service/internal/lib/api/response"
	"github.com/Artem09076/dp/backend/core_service/internal/presentation/services/dto"
	sqlc "github.com/Artem09076/dp/backend/core_service/internal/storage/db"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type ServicesService interface {
	CreateService(ctx context.Context, createServiceObject dto.CreateServiceRequest) (*sqlc.Service, error)
	SearchServices(ctx context.Context, query string, page int, limit int) ([]sqlc.Service, error)
	GetService(ctx context.Context, serviceID uuid.UUID) (*sqlc.GetServiceRow, error)
	DeleteService(ctx context.Context, serviceID uuid.UUID) error
	UpdateService(ctx context.Context, serviceID uuid.UUID, updateServiceObject dto.PatchServiceRequest) error
	CheckServiceOwnership(ctx context.Context, userID uuid.UUID, serviceID uuid.UUID) (bool, error)
	GetServices(ctx context.Context, userID uuid.UUID) ([]sqlc.Service, error)
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

func (h *ServiceHandler) convertToServiceResponse(service *sqlc.Service) dto.ServiceResponse {
	resp := dto.ServiceResponse{
		ID:              service.ID.String(),
		PerformerID:     service.PerformerID.String(),
		Title:           service.Title,
		Price:           service.Price,
		DurationMinutes: service.DurationMinutes,
		CreatedAt:       service.CreatedAt.String(),
		UpdatedAt:       service.UpdatedAt.String(),
	}

	if service.Description.Valid {
		resp.Description = &service.Description.String
	}

	return resp
}

func (h *ServiceHandler) CreateService() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "service.handlers.CreateService"
		log := h.log.With(slog.String("op", op))
		userID, err := handlerlib.GetUserIDFromClaims(r.Context())
		if err != nil {
			log.Error(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error(err.Error()))
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

		resp := h.convertToServiceResponse(service)

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}

	}
}

func (h *ServiceHandler) SearchServices() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "service.handlers.SearchServices"
		log := h.log.With(slog.String("op", op))

		query := r.URL.Query().Get("query")
		page := r.URL.Query().Get("page")
		if page == "" {
			page = "1"
		}
		limit := r.URL.Query().Get("limit")
		if limit == "" {
			limit = "10"
		}
		pageInt, err := strconv.Atoi(page)
		if err != nil || pageInt < 1 {
			log.Error("Invalid page parameter", slog.String("page", page))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid page parameter"))
			return
		}

		limitInt, err := strconv.Atoi(limit)
		if err != nil || limitInt < 1 {
			log.Error("Invalid limit parameter", slog.String("limit", limit))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid limit parameter"))
			return
		}

		services, err := h.service.SearchServices(r.Context(), query, pageInt, limitInt)
		if err != nil {
			log.Error("internal server error", slog.String("Error", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("internal server error"))
			return
		}

		responses := make([]dto.ServiceResponse, len(services))
		for i, service := range services {
			responses[i] = dto.ServiceResponse{
				ID:              service.ID.String(),
				PerformerID:     service.PerformerID.String(),
				Title:           service.Title,
				Price:           service.Price,
				DurationMinutes: service.DurationMinutes,
				CreatedAt:       service.CreatedAt.String(),
				UpdatedAt:       service.UpdatedAt.String(),
			}
			if service.Description.Valid {
				responses[i].Description = &service.Description.String
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(responses); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}

	}
}

func (h *ServiceHandler) GetService() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "service.handlers.GetService"
		log := h.log.With(slog.String("op", op))
		serviceIDStr := chi.URLParam(r, "id")
		serviceID, err := uuid.Parse(serviceIDStr)
		if err != nil {
			log.Error("Failed to parse service_id", slog.String("service_id", serviceIDStr))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid service_id"))
			return
		}

		service, err := h.service.GetService(r.Context(), serviceID)
		if err != nil {
			log.Error("internal server error", slog.String("Error", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("internal server error"))
			return
		}

		resp := dto.ServiceResponse{
			ID:              service.ID.String(),
			PerformerID:     service.PerformerID.String(),
			Title:           service.Title,
			Price:           service.Price,
			DurationMinutes: service.DurationMinutes,
			CreatedAt:       service.CreatedAt.String(),
			UpdatedAt:       service.UpdatedAt.String(),
			AverageRating:   service.AverageRating,
		}

		if service.Description.Valid {
			resp.Description = &service.Description.String
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}

	}
}

func (h *ServiceHandler) GetServices() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "service.handlers.DeleteService"
		log := h.log.With(slog.String("op", op))
		w.Header().Set("Content-Type", "application/json")
		userID, err := handlerlib.GetUserIDFromClaims(r.Context())
		if err != nil {
			log.Error(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error(err.Error()))
			return
		}
		services, err := h.service.GetServices(r.Context(), userID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("Internal server error"))
			return
		}
		responses := make([]dto.ServiceResponse, len(services))
		for i, service := range services {
			responses[i] = dto.ServiceResponse{
				ID:              service.ID.String(),
				PerformerID:     service.PerformerID.String(),
				Title:           service.Title,
				Price:           service.Price,
				DurationMinutes: service.DurationMinutes,
				CreatedAt:       service.CreatedAt.String(),
				UpdatedAt:       service.UpdatedAt.String(),
			}
			if service.Description.Valid {
				responses[i].Description = &service.Description.String
			}
		}
		if err := json.NewEncoder(w).Encode(responses); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}

	}
}

func (h *ServiceHandler) DeleteService() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "service.handlers.DeleteService"
		log := h.log.With(slog.String("op", op))
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
			log.Error("Failed to parse service_id", slog.String("service_id", serviceIDStr))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid service_id"))
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
		err = h.service.DeleteService(r.Context(), serviceID)
		if err != nil {
			log.Error("internal server error", slog.String("Error", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("internal server error"))
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *ServiceHandler) PatchService() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "service.handlers.PatchService"
		log := h.log.With(slog.String("op", op))
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
			log.Error("Failed to parse service_id", slog.String("service_id", serviceIDStr))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid service_id"))
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
		var updateServiceObject dto.PatchServiceRequest
		if err := json.NewDecoder(r.Body).Decode(&updateServiceObject); err != nil {
			log.Error("Failed to decode request body", slog.String("Error", err.Error()))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid request body"))
			return
		}

		if err := h.service.UpdateService(r.Context(), serviceID, updateServiceObject); err != nil {
			log.Error("internal server error", slog.String("Error", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("internal server error"))
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
