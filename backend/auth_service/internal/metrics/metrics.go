package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HttpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	HttpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "auth_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	GrpcRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_grpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"method", "status"},
	)

	GrpcRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "auth_grpc_request_duration_seconds",
			Help:    "gRPC request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	RegistrationsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "auth_registrations_total",
			Help: "Total number of user registrations",
		},
	)

	LoginsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "auth_logins_total",
			Help: "Total number of logins",
		},
	)

	LoginsByType = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_logins_by_type",
			Help: "Number of logins by type (email/inn)",
		},
		[]string{"type"},
	)

	RefreshTokenTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "auth_refresh_token_total",
			Help: "Total number of token refreshes",
		},
	)

	LogoutsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "auth_logouts_total",
			Help: "Total number of logouts",
		},
	)

	AccessTokensIssuedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "auth_access_tokens_issued_total",
			Help: "Total number of access tokens issued",
		},
	)

	RefreshTokensIssuedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "auth_refresh_tokens_issued_total",
			Help: "Total number of refresh tokens issued",
		},
	)

	TokensRevokedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "auth_tokens_revoked_total",
			Help: "Total number of tokens revoked",
		},
	)

	BlacklistedTokensTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "auth_blacklisted_tokens_current",
			Help: "Current number of blacklisted tokens",
		},
	)

	ActiveSessionsGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "auth_active_sessions_total",
			Help: "Current number of active sessions",
		},
	)

	SessionsByDevice = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "auth_sessions_by_device",
			Help: "Number of sessions by device type",
		},
		[]string{"device_type"},
	)

	TokenValidationsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "auth_token_validations_total",
			Help: "Total number of token validations",
		},
	)

	TokenValidationErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_token_validation_errors_total",
			Help: "Total number of token validation errors",
		},
		[]string{"error_type"},
	)

	UsersByRole = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "auth_users_by_role",
			Help: "Number of users by role",
		},
		[]string{"role"},
	)

	DatabaseQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "auth_db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"query", "success"},
	)

	DatabaseErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_db_errors_total",
			Help: "Total number of database errors",
		},
		[]string{"operation"},
	)

	RedisOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_redis_operations_total",
			Help: "Total number of Redis operations",
		},
		[]string{"operation", "status"},
	)

	ActiveGoroutines = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "auth_goroutines_current",
			Help: "Current number of goroutines",
		},
	)
)

func RecordRegistration() {
	RegistrationsTotal.Inc()
}

func RecordLogin(loginType string) {
	LoginsTotal.Inc()
	LoginsByType.WithLabelValues(loginType).Inc()
}

func RecordRefreshToken() {
	RefreshTokenTotal.Inc()
}

func RecordLogout() {
	LogoutsTotal.Inc()
}

func RecordAccessTokenIssued() {
	AccessTokensIssuedTotal.Inc()
}

func RecordRefreshTokenIssued() {
	RefreshTokensIssuedTotal.Inc()
}

func RecordTokenRevoked() {
	TokensRevokedTotal.Inc()
}

func RecordTokenValidation() {
	TokenValidationsTotal.Inc()
}

func RecordTokenValidationError(errorType string) {
	TokenValidationErrorsTotal.WithLabelValues(errorType).Inc()
}

func UpdateActiveSessions(count float64) {
	ActiveSessionsGauge.Set(count)
}

func UpdateBlacklistedTokens(count float64) {
	BlacklistedTokensTotal.Set(count)
}
