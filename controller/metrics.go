package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsHandler exposes the Prometheus default registry as a Gin handler.
//
// Security note: the endpoint is unauthenticated — operator must restrict
// access at the ingress / reverse-proxy layer (allow-list of Prometheus
// scrape source IPs, or mount on a private network). This is consistent
// with how /metrics is conventionally exposed.
//
// Toggle: registration of the route is gated by env var METRICS_ENABLED
// (default: true). The metric collection middleware is independent — it
// always runs because the cost is negligible.
func MetricsHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
