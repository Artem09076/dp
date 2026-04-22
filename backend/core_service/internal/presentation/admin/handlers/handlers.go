package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"log/slog"

	"github.com/Artem09076/dp/backend/core_service/internal/application/admin"
	apierrors "github.com/Artem09076/dp/backend/core_service/internal/lib/api/errors"
	"github.com/Artem09076/dp/backend/core_service/internal/presentation/admin/dto"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type AdminHandler struct {
	service *admin.AdminService
	log     *slog.Logger
}

func NewAdminHandler(service *admin.AdminService, log *slog.Logger) *AdminHandler {
	return &AdminHandler{
		service: service,
		log:     log,
	}
}

func (h *AdminHandler) GetUnverifiedPerformers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "admin.handlers.GetUnverifiedPerformers"
		log := h.log.With(slog.String("op", op), slog.String("request_id", middleware.GetReqID(r.Context())))

		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

		if page <= 0 {
			page = 1
		}
		if pageSize <= 0 || pageSize > 100 {
			pageSize = 20
		}

		users, total, err := h.service.GetUnverifiedPerformers(r.Context(), page, pageSize)
		if err != nil {
			log.Error("failed to get unverified performers", "error", err)
			h.WriteError(w, r, err)
			return
		}

		response := dto.PaginatedResponse{
			Data:       users,
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: (int(total) + pageSize - 1) / pageSize,
		}
		render.JSON(w, r, response)
	}
}

func (h *AdminHandler) GetUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "admin.handlers.GetUsers"
		log := h.log.With(slog.String("op", op), slog.String("request_id", middleware.GetReqID(r.Context())))

		role := r.URL.Query().Get("role")
		verificationStatus := r.URL.Query().Get("verification_status")
		search := r.URL.Query().Get("search")
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

		if page <= 0 {
			page = 1
		}
		if pageSize <= 0 || pageSize > 100 {
			pageSize = 20
		}

		users, total, err := h.service.GetUsers(r.Context(), role, verificationStatus, search, page, pageSize)
		if err != nil {
			log.Error("failed to get users", "error", err)
			h.WriteError(w, r, err)
			return
		}

		response := dto.PaginatedResponse{
			Data:       users,
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: (int(total) + pageSize - 1) / pageSize,
		}
		render.JSON(w, r, response)
	}
}

func (h *AdminHandler) GetUserByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "admin.handlers.GetUserByID"
		log := h.log.With(slog.String("op", op), slog.String("request_id", middleware.GetReqID(r.Context())))

		userIDStr := chi.URLParam(r, "user_id")
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			h.WriteError(w, r, apierrors.ErrInvalidInput)
			return
		}

		user, err := h.service.GetUserByID(r.Context(), userID)
		if err != nil {
			log.Error("failed to get user", "error", err)
			h.WriteError(w, r, err)
			return
		}

		render.JSON(w, r, user)
	}
}

func (h *AdminHandler) DeleteUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "admin.handlers.DeleteUser"
		log := h.log.With(slog.String("op", op), slog.String("request_id", middleware.GetReqID(r.Context())))

		userIDStr := chi.URLParam(r, "user_id")
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			h.WriteError(w, r, apierrors.ErrInvalidInput)
			return
		}

		if err := h.service.DeleteUser(r.Context(), userID); err != nil {
			log.Error("failed to delete user", "error", err)
			h.WriteError(w, r, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// BatchVerifyPerformers POST /api/v1/admin/users/verify/batch
func (h *AdminHandler) BatchVerifyPerformers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "admin.handlers.BatchVerifyPerformers"
		log := h.log.With(slog.String("op", op), slog.String("request_id", middleware.GetReqID(r.Context())))

		var req dto.BatchVerifyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.WriteError(w, r, apierrors.ErrInvalidInput)
			return
		}

		if err := h.service.BatchVerifyPerformers(r.Context(), req.UserIDs, req.Status); err != nil {
			log.Error("failed to batch verify performers", "error", err)
			h.WriteError(w, r, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// GetServices GET /api/v1/admin/services
func (h *AdminHandler) GetServices() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "admin.handlers.GetServices"
		log := h.log.With(slog.String("op", op), slog.String("request_id", middleware.GetReqID(r.Context())))

		performerIDStr := r.URL.Query().Get("performer_id")
		var performerID uuid.UUID
		if performerIDStr != "" {
			var err error
			performerID, err = uuid.Parse(performerIDStr)
			if err != nil {
				h.WriteError(w, r, apierrors.ErrInvalidInput)
				return
			}
		}

		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

		if page <= 0 {
			page = 1
		}
		if pageSize <= 0 || pageSize > 100 {
			pageSize = 20
		}

		services, total, err := h.service.GetServices(r.Context(), performerID, page, pageSize)
		if err != nil {
			log.Error("failed to get services", "error", err)
			h.WriteError(w, r, err)
			return
		}

		response := dto.PaginatedResponse{
			Data:       services,
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: (int(total) + pageSize - 1) / pageSize,
		}
		render.JSON(w, r, response)
	}
}

// GetBookings GET /api/v1/admin/bookings
func (h *AdminHandler) GetBookings() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "admin.handlers.GetBookings"
		log := h.log.With(slog.String("op", op), slog.String("request_id", middleware.GetReqID(r.Context())))

		status := r.URL.Query().Get("status")
		clientID := r.URL.Query().Get("client_id")
		performerID := r.URL.Query().Get("performer_id")
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

		if page <= 0 {
			page = 1
		}
		if pageSize <= 0 || pageSize > 100 {
			pageSize = 20
		}

		bookings, total, err := h.service.GetBookings(r.Context(), status, clientID, performerID, page, pageSize)
		if err != nil {
			log.Error("failed to get bookings", "error", err)
			h.WriteError(w, r, err)
			return
		}

		response := dto.PaginatedResponse{
			Data:       bookings,
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: (int(total) + pageSize - 1) / pageSize,
		}
		render.JSON(w, r, response)
	}
}

// GetReviews GET /api/v1/admin/reviews
func (h *AdminHandler) GetReviews() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "admin.handlers.GetReviews"
		log := h.log.With(slog.String("op", op), slog.String("request_id", middleware.GetReqID(r.Context())))

		serviceIDStr := r.URL.Query().Get("service_id")
		var serviceID uuid.UUID
		if serviceIDStr != "" {
			var err error
			serviceID, err = uuid.Parse(serviceIDStr)
			if err != nil {
				h.WriteError(w, r, apierrors.ErrInvalidInput)
				return
			}
		}

		rating, _ := strconv.Atoi(r.URL.Query().Get("rating"))
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

		if page <= 0 {
			page = 1
		}
		if pageSize <= 0 || pageSize > 100 {
			pageSize = 20
		}

		reviews, total, err := h.service.GetReviews(r.Context(), serviceID, rating, page, pageSize)
		if err != nil {
			log.Error("failed to get reviews", "error", err)
			h.WriteError(w, r, err)
			return
		}

		response := dto.PaginatedResponse{
			Data:       reviews,
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: (int(total) + pageSize - 1) / pageSize,
		}
		render.JSON(w, r, response)
	}
}

// DeleteReview DELETE /api/v1/admin/reviews/{review_id}
func (h *AdminHandler) DeleteReview() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "admin.handlers.DeleteReview"
		log := h.log.With(slog.String("op", op), slog.String("request_id", middleware.GetReqID(r.Context())))

		reviewIDStr := chi.URLParam(r, "review_id")
		reviewID, err := uuid.Parse(reviewIDStr)
		if err != nil {
			h.WriteError(w, r, apierrors.ErrInvalidInput)
			return
		}

		if err := h.service.DeleteReview(r.Context(), reviewID); err != nil {
			log.Error("failed to delete review", "error", err)
			h.WriteError(w, r, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *AdminHandler) WriteError(w http.ResponseWriter, r *http.Request, err error) {
	apiErr := apierrors.MapError(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiErr.StatusCode)
	render.JSON(w, r, apierrors.NewErrorResponse(err))
}
