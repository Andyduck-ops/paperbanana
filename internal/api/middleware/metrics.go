package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// RequestCounter counts the number of requests by method, path, and status.
	RequestCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "paperbanana_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// RequestDuration tracks the duration of requests.
	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "paperbanana_http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path"},
	)

	// ActiveRequests tracks the number of active requests.
	ActiveRequests = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "paperbanana_http_active_requests",
			Help: "Number of active HTTP requests",
		},
		[]string{"method"},
	)

	// DBQueryDuration tracks the duration of database queries.
	DBQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "paperbanana_db_query_duration_seconds",
			Help:    "Duration of database queries in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
		[]string{"operation"},
	)

	// GenerationCounter counts generations by status.
	GenerationCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "paperbanana_generations_total",
			Help: "Total number of generations",
		},
		[]string{"status", "pipeline_mode"},
	)

	// GenerationDuration tracks the duration of generations.
	GenerationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "paperbanana_generation_duration_seconds",
			Help:    "Duration of generations in seconds",
			Buckets: []float64{1, 5, 10, 30, 60, 120, 300},
		},
		[]string{"pipeline_mode"},
	)
)

// Metrics returns a middleware that collects Prometheus metrics.
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip metrics endpoint itself
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		start := time.Now()
		method := c.Request.Method

		// Increment active requests
		ActiveRequests.WithLabelValues(method).Inc()
		defer ActiveRequests.WithLabelValues(method).Dec()

		// Process request
		c.Next()

		// Record metrics
		duration := time.Since(start).Seconds()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		status := strconv.Itoa(c.Writer.Status())

		RequestCounter.WithLabelValues(method, path, status).Inc()
		RequestDuration.WithLabelValues(method, path).Observe(duration)
	}
}

// RecordDBQuery records a database query duration.
func RecordDBQuery(operation string, duration time.Duration) {
	DBQueryDuration.WithLabelValues(operation).Observe(duration.Seconds())
}

// RecordGeneration records a generation result.
func RecordGeneration(status, pipelineMode string, duration time.Duration) {
	GenerationCounter.WithLabelValues(status, pipelineMode).Inc()
	GenerationDuration.WithLabelValues(pipelineMode).Observe(duration.Seconds())
}
