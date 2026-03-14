package middleware

import (
	"log/slog"
	"net/http"

	"github.com/Artem09076/dp/backend/core_service/internal/lib/api/response"
	"github.com/go-chi/render"
	"github.com/golang-jwt/jwt/v5"
)

func NewRoleMiddleware(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		op := "middleware.NewJWTMiddleware"

		log := log.With(
			slog.String("op", op),
		)

		fn := func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value("claims").(*jwt.MapClaims)
			if !ok || claims == nil {
				log.Error("Failed to parse claims")
				w.WriteHeader(http.StatusBadRequest)
				render.JSON(w, r, response.Error("invalid authentication claims"))
				return
			}
			role := (*claims)["role"].(string)
			if role != "performer" && role != "admin" {
				w.WriteHeader(http.StatusBadRequest)
				render.JSON(w, r, response.Error("Invalid role"))
				return
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)

	}
}
