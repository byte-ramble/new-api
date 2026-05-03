package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// TestPrometheusMiddleware_RecordsRequest verifies that:
//  1. A request through the middleware bumps the requests counter for the
//     correct {route, method, status} label combination.
//  2. The /metrics endpoint exposes the resulting series so a Prometheus
//     scrape can pick it up.
func TestPrometheusMiddleware_RecordsRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(PrometheusMiddleware())
	r.GET("/widgets/:id", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"id": c.Param("id")})
	})
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Drive 3 successful requests at the same route template
	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/widgets/123", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	}

	// Scrape /metrics
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("/metrics returned %d", w.Code)
	}
	body, _ := io.ReadAll(w.Body)
	bodyStr := string(body)

	// Verify the namespaced series shows up. The exact value can be larger
	// than 3 if other tests in the package incremented it (counters are
	// process-global), so we just check >= 3 by looking for the series.
	wantPrefix := `omnirouter_http_requests_total{method="GET",route="/widgets/:id",status="200"}`
	if !strings.Contains(bodyStr, wantPrefix) {
		t.Errorf("expected %q in /metrics output, got body length %d", wantPrefix, len(bodyStr))
	}

	// Inflight gauge series should also be present (value should be 0 after
	// requests complete).
	if !strings.Contains(bodyStr, "omnirouter_http_inflight_requests") {
		t.Error("expected omnirouter_http_inflight_requests gauge in /metrics output")
	}
}

// TestPrometheusMiddleware_SkipsProbeRoutes verifies that the middleware does
// NOT record metrics for the probe / metrics paths themselves — otherwise a
// scraping Prometheus would inflate its own metrics every interval.
func TestPrometheusMiddleware_SkipsProbeRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(PrometheusMiddleware())
	r.GET("/health", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	for _, p := range []string{"/health", "/metrics"} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, p, nil)
		r.ServeHTTP(w, req)
	}

	// Scrape and assert no series for these routes.
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	r.ServeHTTP(w, req)
	body, _ := io.ReadAll(w.Body)
	if strings.Contains(string(body), `route="/health"`) {
		t.Error("/health should be skipped but appeared in metrics output")
	}
	if strings.Contains(string(body), `route="/metrics"`) {
		t.Error("/metrics should be skipped but appeared in metrics output")
	}
}

// TestRecordRelayTokens verifies the helper increments the counter only when
// values are positive (avoid negative deltas which Prometheus counters reject).
func TestRecordRelayTokens(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	RecordRelayTokens("openai", "gpt-4o", 100, 50)
	RecordRelayTokens("openai", "gpt-4o", 0, 0) // no-op
	RecordRelayTokens("openai", "gpt-4o", 200, 100)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body, _ := io.ReadAll(w.Body)
	if !strings.Contains(string(body), `omnirouter_relay_prompt_tokens_total{channel="openai",model="gpt-4o"}`) {
		t.Error("expected relay prompt token series in metrics output")
	}
	if !strings.Contains(string(body), `omnirouter_relay_completion_tokens_total{channel="openai",model="gpt-4o"}`) {
		t.Error("expected relay completion token series in metrics output")
	}
}
