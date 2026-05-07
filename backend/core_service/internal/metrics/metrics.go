package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HttpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "core_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	HttpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "core_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	UsersDeletedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "core_users_deleted_total",
			Help: "Total number of users deleted",
		},
	)

	PerformersVerifiedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "core_performers_verified_total",
			Help: "Total number of performers verified",
		},
	)

	PerformersRejectedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "core_performers_rejected_total",
			Help: "Total number of performers rejected",
		},
	)

	ServicesCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "core_services_created_total",
			Help: "Total number of services created",
		},
	)

	ServicesDeletedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "core_services_deleted_total",
			Help: "Total number of services deleted",
		},
	)

	ReviewsCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "core_reviews_created_total",
			Help: "Total number of reviews created",
		},
	)

	ReviewsDeletedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "core_reviews_deleted_total",
			Help: "Total number of reviews deleted",
		},
	)

	DiscountsCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "core_discounts_created_total",
			Help: "Total number of discounts created",
		},
	)

	DiscountsDeletedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "core_discounts_deleted_total",
			Help: "Total number of discounts deleted",
		},
	)

	ActiveGoroutines = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "core_goroutines_current",
			Help: "Current number of goroutines",
		},
	)
)

func RecordUserDeleted() {
	UsersDeletedTotal.Inc()
}

func RecordPerformerVerified() {
	PerformersVerifiedTotal.Inc()
}

func RecordPerformerRejected() {
	PerformersRejectedTotal.Inc()
}

func RecordServiceCreated() {
	ServicesCreatedTotal.Inc()
}

func RecordServiceDeleted() {
	ServicesDeletedTotal.Inc()
}

func RecordReviewCreated() {
	ReviewsCreatedTotal.Inc()
}

func RecordReviewDeleted() {
	ReviewsDeletedTotal.Inc()
}

func RecordDiscountCreated() {
	DiscountsCreatedTotal.Inc()
}

func RecordDiscountDeleted() {
	DiscountsDeletedTotal.Inc()
}
