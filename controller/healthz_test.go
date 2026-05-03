package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestHealthz verifies the liveness probe always returns 200 with status=ok
// regardless of dependency state. Liveness must not depend on DB / Redis —
// otherwise a transient DB outage would trigger pod restart loops.
func TestHealthz(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/health", Healthz)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("response body not valid JSON: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("expected status=ok, got %v", body["status"])
	}
	if _, ok := body["version"]; !ok {
		t.Fatalf("expected version field in response")
	}
}

// TestReadyz_NoDB verifies the readiness probe returns 503 when the database
// is not initialized (model.DB == nil). This matches the production behavior:
// /ready should fail-closed so load balancers stop routing traffic to a pod
// that hasn't fully started.
//
// Note: This test relies on running in a clean environment where model.DB is
// not initialized. If running as part of the full test suite where InitDB has
// been called, this test should be skipped or moved to an integration test.
func TestReadyz_NoDB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/ready", Readyz)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	r.ServeHTTP(w, req)

	// Without DB initialized, model.PingDB() will fail or panic-recover into 503.
	// We accept either 503 (preferred) or 500 (panic-recovered) as evidence the
	// probe correctly refuses traffic when DB is unavailable.
	if w.Code != http.StatusServiceUnavailable && w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 503 or 500 when DB unavailable, got %d. body=%s", w.Code, w.Body.String())
	}
}
