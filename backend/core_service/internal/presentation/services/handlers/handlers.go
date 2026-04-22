package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	apierrors "github.com/Artem09076/dp/backend/core_service/internal/lib/api/errors"
	handlerlib "github.com/Artem09076/dp/backend/core_service/internal/lib/api/handler"
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

func (h *ServiceHandler) convertToServiceRowResponse(service *sqlc.GetServiceRow) dto.ServiceResponse {
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

	return resp
}

func (h *ServiceHandler) CreateService() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "service.handlers.CreateService"
		w.Header().Set("Content-Type", "application/json")

		userID, err := handlerlib.GetUserIDFromClaims(r.Context())
		if err != nil {
			h.WriteError(w, r, apierrors.ErrUnauthorized, op)
			return
		}
		var body dto.CreateServiceRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			h.WriteError(w, r, apierrors.ErrInvalidInput, op)
			return
		}

		body.PerformerID = userID

		service, err := h.service.CreateService(r.Context(), body)
		if err != nil {
			h.WriteError(w, r, err, op)
			return
		}

		resp := h.convertToServiceResponse(service)

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}

	}
}

func (h *ServiceHandler) SearchServices() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "service.handlers.SearchServices"
		w.Header().Set("Content-Type", "application/json")

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
			h.WriteError(w, r, apierrors.ErrInvalidInput, op)
			return
		}

		limitInt, err := strconv.Atoi(limit)
		if err != nil || limitInt < 1 || limitInt > 100 {
			h.WriteError(w, r, apierrors.ErrInvalidInput, op)
			return
		}

		services, err := h.service.SearchServices(r.Context(), query, pageInt, limitInt)
		if err != nil {
			h.WriteError(w, r, err, op)
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

func (h *ServiceHandler) GetService() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "service.handlers.GetService"
		w.Header().Set("Content-Type", "application/json")

		serviceIDStr := chi.URLParam(r, "id")
		serviceID, err := uuid.Parse(serviceIDStr)
		if err != nil {
			h.WriteError(w, r, apierrors.ErrInvalidInput, op)
		}

		service, err := h.service.GetService(r.Context(), serviceID)
		if err != nil {
			h.WriteError(w, r, err, op)
			return
		}

		resp := h.convertToServiceRowResponse(service)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

func (h *ServiceHandler) GetServices() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "service.handlers.GetServices"
		w.Header().Set("Content-Type", "application/json")

		userID, err := handlerlib.GetUserIDFromClaims(r.Context())
		if err != nil {
			h.WriteError(w, r, apierrors.ErrUnauthorized, op)
			return
		}

		services, err := h.service.GetServices(r.Context(), userID)
		if err != nil {
			h.WriteError(w, r, err, op)
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
		w.Header().Set("Content-Type", "application/json")

		userID, err := handlerlib.GetUserIDFromClaims(r.Context())
		if err != nil {
			h.WriteError(w, r, apierrors.ErrUnauthorized, op)
			return
		}

		serviceIDStr := chi.URLParam(r, "id")
		serviceID, err := uuid.Parse(serviceIDStr)
		if err != nil {
			h.WriteError(w, r, apierrors.ErrInvalidInput, op)
		}

		if ok, err := h.service.CheckServiceOwnership(r.Context(), userID, serviceID); err != nil {
			h.WriteError(w, r, err, op)
			return
		} else if !ok {
			h.WriteError(w, r, apierrors.ErrForbidden, op)
			return
		}

		if err := h.service.DeleteService(r.Context(), serviceID); err != nil {
			h.WriteError(w, r, err, op)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *ServiceHandler) PatchService() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "service.handlers.PatchService"
		w.Header().Set("Content-Type", "application/json")

		userID, err := handlerlib.GetUserIDFromClaims(r.Context())
		if err != nil {
			h.WriteError(w, r, apierrors.ErrUnauthorized, op)
			return
		}

		serviceIDStr := chi.URLParam(r, "id")
		serviceID, err := uuid.Parse(serviceIDStr)
		if err != nil {
			h.WriteError(w, r, apierrors.ErrInvalidInput, op)
		}
		if ok, err := h.service.CheckServiceOwnership(r.Context(), userID, serviceID); err != nil {
			h.WriteError(w, r, err, op)
			return
		} else if !ok {
			h.WriteError(w, r, apierrors.ErrForbidden, op)
			return
		}

		var updateServiceObject dto.PatchServiceRequest
		if err := json.NewDecoder(r.Body).Decode(&updateServiceObject); err != nil {
			h.WriteError(w, r, apierrors.ErrInvalidInput, op)
			return
		}

		if err := h.service.UpdateService(r.Context(), serviceID, updateServiceObject); err != nil {
			h.WriteError(w, r, err, op)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *ServiceHandler) WriteError(w http.ResponseWriter, r *http.Request, err error, op string) {
	h.log.Error(op, slog.String("error", err.Error()))
	apiErr := apierrors.MapError(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiErr.StatusCode)
	render.JSON(w, r, apierrors.NewErrorResponse(err))
}
