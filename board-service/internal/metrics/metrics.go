package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// ==================== Business Metrics ====================

// Board Metrics
var (
	BoardCreatedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "board_created_total",
			Help: "Total number of boards created",
		},
		[]string{"project_id"},
	)

	BoardUpdatedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "board_updated_total",
			Help: "Total number of boards updated",
		},
		[]string{"project_id"},
	)

	BoardDeletedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "board_deleted_total",
			Help: "Total number of boards deleted",
		},
		[]string{"project_id"},
	)

	BoardOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "board_operation_duration_seconds",
			Help:    "Duration of board operations in seconds",
			Buckets: prometheus.DefBuckets, // 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10
		},
		[]string{"operation", "project_id"},
	)
)

// Project Metrics
var (
	ProjectCreatedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "project_created_total",
			Help: "Total number of projects created",
		},
		[]string{"workspace_id"},
	)

	ProjectMemberAddedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "project_member_added_total",
			Help: "Total number of project members added",
		},
		[]string{"project_id", "role"},
	)

	ProjectOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "project_operation_duration_seconds",
			Help:    "Duration of project operations in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)
)

// Comment Metrics
var (
	CommentCreatedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "comment_created_total",
			Help: "Total number of comments created",
		},
		[]string{"board_id"},
	)

	CommentDeletedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "comment_deleted_total",
			Help: "Total number of comments deleted",
		},
		[]string{"board_id"},
	)
)

// Custom Field Metrics
var (
	FieldCreatedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "field_created_total",
			Help: "Total number of custom fields created",
		},
		[]string{"project_id", "field_type"},
	)

	FieldValueSetTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "field_value_set_total",
			Help: "Total number of field values set",
		},
		[]string{"board_id", "field_type"},
	)
)

// ==================== Infrastructure Metrics ====================

// Cache Metrics
var (
	CacheHitTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_hit_total",
			Help: "Total number of cache hits",
		},
		[]string{"cache_type"},
	)

	CacheMissTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_miss_total",
			Help: "Total number of cache misses",
		},
		[]string{"cache_type"},
	)

	CacheOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cache_operation_duration_seconds",
			Help:    "Duration of cache operations in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1}, // Faster buckets for cache
		},
		[]string{"operation", "cache_type"},
	)
)

// Database Metrics
var (
	DatabaseQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "database_query_duration_seconds",
			Help:    "Duration of database queries in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "table"},
	)

	DatabaseErrorTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "database_error_total",
			Help: "Total number of database errors",
		},
		[]string{"operation", "table"},
	)
)

// External Service Metrics
var (
	ExternalServiceRequestTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "external_service_request_total",
			Help: "Total number of external service requests",
		},
		[]string{"service", "method", "status"},
	)

	ExternalServiceDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "external_service_duration_seconds",
			Help:    "Duration of external service calls in seconds",
			Buckets: []float64{0.1, 0.25, 0.5, 1, 2.5, 5, 10}, // External calls may be slower
		},
		[]string{"service", "method"},
	)
)

// ==================== Helper Functions ====================

// RecordDuration records the duration of an operation
// Usage: defer metrics.RecordDuration(time.Now(), metrics.BoardOperationDuration, "create", projectID)
func RecordDuration(start time.Time, histogram *prometheus.HistogramVec, labels ...string) {
	duration := time.Since(start).Seconds()
	histogram.WithLabelValues(labels...).Observe(duration)
}

// RecordCacheHit records a cache hit
func RecordCacheHit(cacheType string) {
	CacheHitTotal.WithLabelValues(cacheType).Inc()
}

// RecordCacheMiss records a cache miss
func RecordCacheMiss(cacheType string) {
	CacheMissTotal.WithLabelValues(cacheType).Inc()
}

// RecordExternalCall records an external service call
func RecordExternalCall(service, method, status string, duration time.Duration) {
	ExternalServiceRequestTotal.WithLabelValues(service, method, status).Inc()
	ExternalServiceDuration.WithLabelValues(service, method).Observe(duration.Seconds())
}

// RecordDatabaseError records a database error
func RecordDatabaseError(operation, table string) {
	DatabaseErrorTotal.WithLabelValues(operation, table).Inc()
}
