package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// EventsReceived tracks total events received by protocol
	EventsReceived = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "enterprise_events_received_total",
			Help: "Total number of events received",
		},
		[]string{"protocol"},
	)

	// EventsProcessed tracks successfully processed events
	EventsProcessed = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "enterprise_events_processed_total",
			Help: "Total number of events processed",
		},
	)

	// EventsDropped tracks dropped events
	EventsDropped = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "enterprise_events_dropped_total",
			Help: "Total number of events dropped",
		},
	)

	// BytesReceived tracks total bytes ingested
	BytesReceived = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "enterprise_bytes_received_total",
			Help: "Total bytes received",
		},
		[]string{"protocol"},
	)

	// BufferSize tracks current buffer utilization
	BufferSize = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "enterprise_buffer_size",
			Help: "Current number of events in buffer",
		},
	)

	// ProcessingDuration tracks event processing time
	ProcessingDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "enterprise_processing_duration_seconds",
			Help:    "Event processing duration",
			Buckets: prometheus.DefBuckets,
		},
	)
)
