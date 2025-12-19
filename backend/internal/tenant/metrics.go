package tenant

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// TenantEventsReceived tracks events per tenant
	TenantEventsReceived = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tenant_events_received_total",
			Help: "Total events received per tenant",
		},
		[]string{"tenant_id"},
	)

	// TenantEventsDropped tracks dropped events per tenant
	TenantEventsDropped = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tenant_events_dropped_total",
			Help: "Total events dropped per tenant",
		},
		[]string{"tenant_id", "reason"},
	)

	// TenantStorageUsed tracks storage usage per tenant
	TenantStorageUsed = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "tenant_storage_bytes",
			Help: "Current storage usage in bytes per tenant",
		},
		[]string{"tenant_id"},
	)

	// TenantBufferSize tracks buffer utilization per tenant
	TenantBufferSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "tenant_buffer_size",
			Help: "Current buffer size per tenant",
		},
		[]string{"tenant_id"},
	)

	// TenantRateLimit tracks rate limit hits per tenant
	TenantRateLimit = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tenant_rate_limit_hits_total",
			Help: "Total rate limit hits per tenant",
		},
		[]string{"tenant_id"},
	)

	// TenantQuotaExceeded tracks quota violations per tenant
	TenantQuotaExceeded = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tenant_quota_exceeded_total",
			Help: "Total quota exceeded events per tenant",
		},
		[]string{"tenant_id", "quota_type"},
	)
)
