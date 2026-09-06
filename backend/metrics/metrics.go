package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// [EN] Counter for total processed telemetry messages
	// [RU] Счётчик общего количества обработанных сообщений телеметрии
	ProcessedTelemetryTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "smart_home_telemetry_processed_total",
		Help: "The total number of processed telemetry messages from devices",
	})

	// [EN] Histogram for measuring HTTP request durations in seconds
	// [RU] Гистограмма для измерения длительности HTTP-запросов в секундах
	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "smart_home_http_request_duration_seconds",
		Help:    "Duration of HTTP requests in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "endpoint", "status"})
)