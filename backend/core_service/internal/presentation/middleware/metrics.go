package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Artem09076/dp/backend/core_service/internal/metrics"
	"github.com/go-chi/chi/v5"
)

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriterWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		ww := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(ww, r)

		duration := time.Since(start).Seconds()

		endpoint := getEndpoint(r)
		method := r.Method
		status := strconv.Itoa(ww.statusCode)

		metrics.HttpRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
		metrics.HttpRequestDuration.WithLabelValues(method, endpoint).Observe(duration)
	})
}

func getEndpoint(r *http.Request) string {
	if routeCtx := chi.RouteContext(r.Context()); routeCtx != nil && routeCtx.RoutePattern() != "" {
		return routeCtx.RoutePattern()
	}

	path := r.URL.Path
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if isUUID(part) {
			parts[i] = ":id"
		}
	}
	return strings.Join(parts, "/")
}

func isUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, c := range s {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return false
			}
		} else if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
			return false
		}
	}
	return true
}
