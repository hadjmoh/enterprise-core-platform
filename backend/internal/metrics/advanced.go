package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// ProcessingLatency tracks event processing time histograms
	ProcessingLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "enterprise_event_processing_seconds",
			Help:    "Histogram of event processing latency",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5},
		},
		[]string{"protocol", "tenant_id", "stage"}, // ingestion, enrichment, correlation, storage
	)

	// ErrorOccurrences tracks categorized errors
	ErrorOccurrences = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "enterprise_error_total",
			Help: "Total errors by category and component",
		},
		[]string{"component", "category", "tenant_id"},
	)

	// PipelineSaturation tracks buffer and worker utilization
	PipelineSaturation = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "enterprise_pipeline_saturation_ratio",
			Help: "Current pipeline saturation (0-1)",
		},
		[]string{"component"},
	)

	// EnrichmentSuccess tracks success/failure of enrichment stages
	EnrichmentSuccess = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "enterprise_enrichment_total",
			Help: "Enrichment attempts by type and status",
		},
		[]string{"type", "status"}, // geoip, threat_intel, asset, user | success, failure
	)
)
