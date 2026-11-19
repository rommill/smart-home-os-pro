package api

import (
	"strconv"
	"time"

	"smart-home/metrics"

	"github.com/gin-gonic/gin"
)

// [EN] PrometheusMiddleware measures HTTP request duration for Prometheus metrics.
// [RU] PrometheusMiddleware измеряет длительность HTTP-запросов для метрик Prometheus.
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method
		path := c.FullPath()

		if path != "" {
			metrics.HTTPRequestDuration.WithLabelValues(method, path, status).Observe(duration)
		}
	}
}