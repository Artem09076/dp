package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Artem09076/dp/backend/core_service/internal/lib/api/response"
	"github.com/Artem09076/dp/backend/core_service/internal/presentation/profile/dto"
	sqlc "github.com/Artem09076/dp/backend/core_service/internal/storage/db"
	"github.com/go-chi/render"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type ProfileService interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*sqlc.GetProfileRow, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, updateProfileObject dto.PatchProfileRequest) error
	DeleteProfile(ctx context.Context, userID uuid.UUID) error
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
		profile, err := h.profileService.GetProfile(r.Context(), userID)
		if err != nil {
			log.Error("internal server error", slog.String("Error", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("internal server error"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(profile); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

func (h *ProfileHandler) PatchProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "profile.handlers.PatchProfile"
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

		var body dto.PatchProfileRequest

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "can't decode JSON body", http.StatusBadRequest)
			return
		}

		if err := h.profileService.UpdateProfile(r.Context(), userID, body); err != nil {
			log.Error("internal server error", slog.String("Error", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("internal server error"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		resp := dto.PatchProfileResponse{Status: "updated"}
		render.JSON(w, r, resp)
		return
	}
}

func (h *ProfileHandler) DeleteProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "profile.handlers.DeleteProfile"

		log := h.log.With(
			slog.String("op", op),
		)

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

		err = h.profileService.DeleteProfile(r.Context(), userID)
		if err != nil {
			log.Error("failed to delete profile",
				slog.String("error", err.Error()),
			)

			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("internal server error"))
			return
		}

		w.Header().Set("Content-Type", "application/json")

		w.WriteHeader(http.StatusNoContent)
		return
	}
}
