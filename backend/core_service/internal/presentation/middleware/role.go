package middleware

import (
	"log/slog"
	"net/http"
	"slices"

	"github.com/Artem09076/dp/backend/core_service/internal/lib/api/response"
	"github.com/go-chi/render"
)

func NewRoleMiddleware(log *slog.Logger, allowedRoles []string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			role := r.Context().Value("role").(string)
			w.Header().Set("Content-Type", "application/json")
			if !slices.Contains(allowedRoles, role) {
				w.WriteHeader(http.StatusBadRequest)
				render.JSON(w, r, response.Error("Invalid role"))
				return
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)

	}
}
