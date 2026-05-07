package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HttpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "booking_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	HttpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "booking_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	BookingsCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "booking_created_total",
			Help: "Total number of bookings created",
		},
	)

	BookingsCancelledTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "booking_cancelled_total",
			Help: "Total number of bookings cancelled",
		},
	)

	BookingsCompletedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "booking_completed_total",
			Help: "Total number of bookings completed",
		},
	)

	BookingsSubmittedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "booking_submitted_total",
			Help: "Total number of bookings submitted",
		},
	)

	ActiveBookingsGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "booking_active_total",
			Help: "Current number of active bookings (pending + confirmed)",
		},
	)

	BookingsByStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "booking_by_status",
			Help: "Number of bookings by status",
		},
		[]string{"status"},
	)

	DatabaseQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "booking_db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"query", "success"},
	)

	DatabaseErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "booking_db_errors_total",
			Help: "Total number of database errors",
		},
		[]string{"operation"},
	)

	CacheHitsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "booking_cache_hits_total",
			Help: "Total number of cache hits",
		},
	)

	CacheMissesTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "booking_cache_misses_total",
			Help: "Total number of cache misses",
		},
	)

	QueueEventsPublishedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "booking_queue_events_published_total",
			Help: "Total number of events published to queue",
		},
		[]string{"event_type"},
	)

	QueueErrorsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "booking_queue_errors_total",
			Help: "Total number of queue operation errors",
		},
	)

	ActiveGoroutines = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "booking_goroutines_current",
			Help: "Current number of goroutines",
		},
	)
)

func UpdateBookingStatusMetrics(bookings []struct{ Status string }) {
	statusCount := make(map[string]int)
	for _, b := range bookings {
		statusCount[b.Status]++
	}

	for status, count := range statusCount {
		BookingsByStatus.WithLabelValues(status).Set(float64(count))
	}

	pending := statusCount["pending"]
	confirmed := statusCount["confirmed"]
	ActiveBookingsGauge.Set(float64(pending + confirmed))
}

func RecordBookingCreated() {
	BookingsCreatedTotal.Inc()
}

func RecordBookingCancelled() {
	BookingsCancelledTotal.Inc()
}

func RecordBookingCompleted() {
	BookingsCompletedTotal.Inc()
}

func RecordBookingSubmitted() {
	BookingsSubmittedTotal.Inc()
}
