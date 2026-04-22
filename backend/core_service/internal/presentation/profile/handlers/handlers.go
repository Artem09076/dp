package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	apierrors "github.com/Artem09076/dp/backend/core_service/internal/lib/api/errors"
	handlerlib "github.com/Artem09076/dp/backend/core_service/internal/lib/api/handler"
	"github.com/Artem09076/dp/backend/core_service/internal/presentation/profile/dto"
	sqlc "github.com/Artem09076/dp/backend/core_service/internal/storage/db"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type ProfileService interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*sqlc.GetProfileRow, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, updateProfileObject dto.PatchProfileRequest) error
	DeleteProfile(ctx context.Context, userID uuid.UUID) error
	UpdateVerificationStatus(ctx context.Context, userID uuid.UUID, verificationStatus string) error
}

type ProfileHandler struct {
	profileService ProfileService
	log            *slog.Logger
}

func NewProfileHandler(profileService ProfileService, log *slog.Logger) *ProfileHandler {
	return &ProfileHandler{
		profileService: profileService,
		log:            log,
	}
}

func (h *ProfileHandler) GetProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "profile.handlers.GetProfile"
		w.Header().Set("Content-Type", "application/json")

		userID, err := handlerlib.GetUserIDFromClaims(r.Context())
		if err != nil {
			h.WriteError(w, r, apierrors.ErrUnauthorized, op)
			return
		}

		profile, err := h.profileService.GetProfile(r.Context(), userID)
		if err != nil {
			h.WriteError(w, r, err, op)
			return
		}

		resp := dto.ProfileResponse{
			Name:               profile.Name,
			Email:              profile.Email,
			Role:               string(profile.Role),
			VerificationStatus: string(profile.VerificationStatus),
		}

		if profile.Inn.Valid {
			resp.Inn = &profile.Inn.String
		}
		if profile.BusinessType.Valid {
			businessType := string(profile.BusinessType.BusinessType)
			resp.BusinessType = &businessType
		}

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			h.WriteError(w, r, err, op)
			return
		}
	}
}

func (h *ProfileHandler) PatchProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "profile.handlers.PatchProfile"
		w.Header().Set("Content-Type", "application/json")

		userID, err := handlerlib.GetUserIDFromClaims(r.Context())
		if err != nil {
			h.WriteError(w, r, apierrors.ErrUnauthorized, op)
			return
		}

		var body dto.PatchProfileRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			h.WriteError(w, r, apierrors.ErrInvalidInput, op)
			return
		}

		if err := h.profileService.UpdateProfile(r.Context(), userID, body); err != nil {
			h.WriteError(w, r, err, op)
			return
		}

		resp := dto.PatchProfileResponse{Status: "updated"}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			h.WriteError(w, r, err, op)
			return
		}
	}
}

func (h *ProfileHandler) DeleteProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "profile.handlers.DeleteProfile"
		w.Header().Set("Content-Type", "application/json")

		userID, err := handlerlib.GetUserIDFromClaims(r.Context())
		if err != nil {
			h.WriteError(w, r, apierrors.ErrUnauthorized, op)
			return
		}
		if err := h.profileService.DeleteProfile(r.Context(), userID); err != nil {
			h.WriteError(w, r, err, op)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *ProfileHandler) UpdateVerificationStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "profile.handlers.UpdateVerificationStatus"
		w.Header().Set("Content-Type", "application/json")

		var body dto.UpdateVerificationStatusRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			h.WriteError(w, r, apierrors.ErrInvalidInput, op)
			return
		}

		if err := h.profileService.UpdateVerificationStatus(r.Context(), body.UserID, body.VerificationStatus); err != nil {
			h.WriteError(w, r, err, op)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *ProfileHandler) WriteError(w http.ResponseWriter, r *http.Request, err error, op string) {
	h.log.Error(op, slog.String("error", err.Error()))
	apiErr := apierrors.MapError(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiErr.StatusCode)
	render.JSON(w, r, apierrors.NewErrorResponse(err))
}
