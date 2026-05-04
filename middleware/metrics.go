package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Prometheus metrics for OmniRouter.
//
// Naming follows the Prometheus convention: `<namespace>_<subsystem>_<name>_<unit>`.
// Namespace is "omnirouter" so series are easy to filter when scraped alongside
// other workloads.
//
// All metrics are registered with the default registry via `promauto`, which
// runs at package init — there's no need to call any setup function.
var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "omnirouter",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total HTTP requests partitioned by route, method and status.",
		},
		[]string{"route", "method", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "omnirouter",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "HTTP request latency, partitioned by route and method.",
			// 5ms..20s, 12 buckets — covers fast admin calls and long-poll relay.
			Buckets: prometheus.ExponentialBuckets(0.005, 2, 12),
		},
		[]string{"route", "method"},
	)

	httpInflight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "omnirouter",
			Subsystem: "http",
			Name:      "inflight_requests",
			Help:      "Number of currently in-flight HTTP requests.",
		},
	)

	relayPromptTokensTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "omnirouter",
			Subsystem: "relay",
			Name:      "prompt_tokens_total",
			Help:      "Cumulative prompt tokens by channel and model.",
		},
		[]string{"channel", "model"},
	)

	relayCompletionTokensTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "omnirouter",
			Subsystem: "relay",
			Name:      "completion_tokens_total",
			Help:      "Cumulative completion tokens by channel and model.",
		},
		[]string{"channel", "model"},
	)

	relayUpstreamErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "omnirouter",
			Subsystem: "relay",
			Name:      "upstream_errors_total",
			Help:      "Errors returned by upstream providers, by channel and reason.",
		},
		[]string{"channel", "reason"},
	)
	// NOTE: channel auto-disable / recover metrics live in service/channel_metrics.go.
	// They were originally placed here, but the service package needs to call
	// them — and middleware imports service (via auth.go) — so colocating
	// them with the service code avoids the import cycle.
)

// metricsSkipPaths bypasses metric recording for routes that would only add
// noise (probes, the metrics endpoint itself).
var metricsSkipPaths = map[string]struct{}{
	"/metrics": {},
	"/health":  {},
	"/healthz": {},
	"/ready":   {},
	"/readyz":  {},
}

// PrometheusMiddleware records per-request HTTP metrics.
// Mount once at the engine root (in main.go) — it inspects c.FullPath() so it
// records the route template (e.g. "/v1/chat/completions") rather than the
// concrete URL, keeping cardinality bounded.
//
// Routes without a registered template (404s) get bucket "unknown" so we
// don't blow up label cardinality with arbitrary user input paths.
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.FullPath()
		if _, skip := metricsSkipPaths[path]; skip {
			c.Next()
			return
		}

		httpInflight.Inc()
		defer httpInflight.Dec()

		start := time.Now()
		c.Next()
		duration := time.Since(start).Seconds()

		route := path
		if route == "" {
			route = "unknown"
		}
		status := strconv.Itoa(c.Writer.Status())
		httpRequestsTotal.WithLabelValues(route, c.Request.Method, status).Inc()
		httpRequestDuration.WithLabelValues(route, c.Request.Method).Observe(duration)
	}
}

// RecordRelayTokens records token usage for a successful relay call.
// Should be called from the relay layer after upstream tokens are tallied.
//
// Wiring this into relay/* is intentionally deferred to a later iteration —
// the function is exported so the integration is a one-line change at each
// callsite without churning the metrics package itself.
func RecordRelayTokens(channel, model string, promptTokens, completionTokens int) {
	if promptTokens > 0 {
		relayPromptTokensTotal.WithLabelValues(channel, model).Add(float64(promptTokens))
	}
	if completionTokens > 0 {
		relayCompletionTokensTotal.WithLabelValues(channel, model).Add(float64(completionTokens))
	}
}

// RecordRelayUpstreamError records a failed upstream relay attempt.
// `reason` should be a low-cardinality bucket (e.g. "timeout", "5xx", "rate_limited").
func RecordRelayUpstreamError(channel, reason string) {
	relayUpstreamErrorsTotal.WithLabelValues(channel, reason).Inc()
}

// (channel auto-disable / recover counters: see service/channel_metrics.go)
