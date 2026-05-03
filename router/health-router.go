package router

import (
	"os"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/controller"

	"github.com/gin-gonic/gin"
)

// SetHealthRouter registers Kubernetes / load-balancer style probe endpoints,
// plus the Prometheus /metrics endpoint when enabled.
//
// The endpoints are intentionally:
//   - placed at the root (no /api prefix) — orchestrators expect bare paths
//   - registered without auth, rate-limit, gzip, or i18n middleware — probes
//     must be cheap and dependency-free
//   - exposed under both /health and /healthz (and /ready / /readyz) to cover
//     k8s default probe paths and common LB defaults without configuration
//
// The pre-existing /api/status endpoint (used by the docker-compose healthcheck
// and the admin dashboard) is NOT removed — it serves a different purpose
// (front-end bootstrap config + admin status).
//
// /metrics is gated by env var METRICS_ENABLED (default true). Operator must
// restrict scrape access at the ingress layer.
func SetHealthRouter(router *gin.Engine) {
	router.GET("/health", controller.Healthz)
	router.GET("/healthz", controller.Healthz)
	router.GET("/ready", controller.Readyz)
	router.GET("/readyz", controller.Readyz)

	if metricsEnabled() {
		router.GET("/metrics", controller.MetricsHandler())
		common.SysLog("Prometheus metrics endpoint enabled at /metrics")
	}
}

// metricsEnabled returns true when METRICS_ENABLED is unset or set to a
// truthy value (1/true/yes, case-insensitive). Default is on.
func metricsEnabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("METRICS_ENABLED")))
	if v == "" {
		return true
	}
	switch v {
	case "0", "false", "no", "off":
		return false
	default:
		return true
	}
}
