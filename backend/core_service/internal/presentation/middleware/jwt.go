package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/Artem09076/dp/backend/core_service/internal/lib/api/response"
	"github.com/Artem09076/dp/backend/core_service/internal/lib/jwt"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

func NewJWTMiddleware(log *slog.Logger, tokenValidator *jwt.JWTValidator) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		op := "middleware.NewJWTMiddleware"

		log := log.With(
			slog.String("op", op),
		)

		fn := func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				w.WriteHeader(http.StatusUnauthorized)
				render.JSON(w, r, response.Error("User not authorized"))
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				w.WriteHeader(http.StatusUnauthorized)
				render.JSON(w, r, response.Error("Baerer token required"))
				return
			}

			claims, err := tokenValidator.Validate(tokenString)
			if err != nil {
				log.Info(err.Error())
				render.JSON(w, r, response.Error(err.Error()))
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			userID, ok := (*claims)["user_id"].(string)
			if !ok {
				log.Info("user_id not found in token claims")
				render.JSON(w, r, response.Error("Invalid token claims"))
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			role, ok := (*claims)["role"].(string)
			if !ok {
				log.Info("user_id not found in token claims")
				render.JSON(w, r, response.Error("Invalid token claims"))
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), "user_id", userID)
			ctx = context.WithValue(ctx, "role", role)

			entry := log.With(
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
				slog.String("request_id", middleware.GetReqID(r.Context())),
			)
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			t1 := time.Now()
			defer func() {
				entry.Info("request completed",
					slog.Int("status", ww.Status()),
					slog.Int("bytes", ww.BytesWritten()),
					slog.String("duration", time.Since(t1).String()),
				)
			}()

			next.ServeHTTP(ww, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}
