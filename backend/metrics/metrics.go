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

	// [EN] Room temperature gauge (Xiaomi BLE & Emulated rooms)
	// [RU] Метрика текущей температуры по комнатам
	RoomTemperature = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "room_temperature_celsius",
		Help: "Current room temperature in Celsius",
	}, []string{"room"})

	// [EN] Room humidity gauge
	// [RU] Метрика влажности по комнатам
	RoomHumidity = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "room_humidity_percent",
		Help: "Current room humidity percentage",
	}, []string{"room"})

	// [EN] Target temperature set by user via thermostat slider
	// [RU] Целевая температура, выставленная пользователем ползунком
	RoomTargetTemperature = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "room_target_temperature_celsius",
		Help: "Target room temperature set by user",
	}, []string{"room"})
)