package service

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/setting/system_setting"
)

// disableSSRFForTest temporarily disables SSRF protection for the duration of
// the calling test, restoring the previous value at cleanup.
//
// Why: httptest.NewServer binds to 127.0.0.1 on a random high port, both of
// which the default fetch settings reject (private IP + port not in allow-list).
// Production deployments keep SSRF protection enabled — this override is
// scoped to a single test.
//
// Implementation note: the package already has a TestMain (in
// task_billing_test.go) that sets up the SQLite test DB. Rather than fighting
// for ownership of TestMain, we toggle the global flag per-test and restore
// it via t.Cleanup. Tests are run sequentially within a package by default,
// so there's no concurrent-mutation hazard.
func disableSSRFForTest(t *testing.T) {
	t.Helper()
	fs := system_setting.GetFetchSetting()
	prev := fs.EnableSSRFProtection
	fs.EnableSSRFProtection = false
	t.Cleanup(func() { fs.EnableSSRFProtection = prev })

	// Production main.go calls InitHttpClient inside InitResources; tests
	// don't have that bootstrap, so the package-level httpClient stays nil
	// and Do() panics on a nil receiver. Initialize lazily here — it's
	// idempotent and cheap.
	if GetHttpClient() == nil {
		InitHttpClient()
	}
}

// TestGenLarkSign verifies the signature matches the well-known Lark spec.
// Hand-computed reference vector (timestamp=1715000000, secret="abc"):
//
//	stringToSign = "1715000000\nabc"
//	HmacSHA256(key=stringToSign, msg="") |> base64
//
// We assert structural properties (non-empty, base64-decodable, 32-byte digest)
// rather than a hard-coded string so the test stays robust against minor
// formatting changes in Go's stdlib.
func TestGenLarkSign(t *testing.T) {
	sig := genLarkSign("abc", 1715000000)
	if sig == "" {
		t.Fatal("expected non-empty signature")
	}
	raw, err := base64.StdEncoding.DecodeString(sig)
	if err != nil {
		t.Fatalf("signature is not valid base64: %v", err)
	}
	if len(raw) != 32 {
		t.Fatalf("expected 32-byte SHA256 digest, got %d", len(raw))
	}

	// Different secret → different signature
	sig2 := genLarkSign("xyz", 1715000000)
	if sig == sig2 {
		t.Fatal("signature must differ when secret changes")
	}
	// Different timestamp → different signature
	sig3 := genLarkSign("abc", 1715000001)
	if sig == sig3 {
		t.Fatal("signature must differ when timestamp changes")
	}
}

// TestSendLarkNotify_Success spins up a mock Lark endpoint and asserts that
// the request body is shaped correctly: msg_type=interactive, card has the
// right title and template, and content placeholders are filled from Values.
func TestSendLarkNotify_Success(t *testing.T) {
	disableSSRFForTest(t)
	var captured map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"code":0,"msg":"ok"}`))
	}))
	defer srv.Close()

	notify := dto.NewNotify(
		dto.NotifyTypeChannelTest,
		"Channel Test Failed",
		"channel %s failed: %s",
		[]interface{}{"openai-1", "timeout"},
	)

	if err := SendLarkNotify(srv.URL, "", notify); err != nil {
		t.Fatalf("SendLarkNotify failed: %v", err)
	}
	if captured["msg_type"] != "interactive" {
		t.Errorf("expected msg_type=interactive, got %v", captured["msg_type"])
	}
	card, ok := captured["card"].(map[string]any)
	if !ok {
		t.Fatalf("expected card object, got %T", captured["card"])
	}
	header, _ := card["header"].(map[string]any)
	if header["template"] != "yellow" {
		t.Errorf("expected template=yellow for ChannelTest, got %v", header["template"])
	}
	// Verify content placeholder substitution propagated into the card body.
	elements, _ := card["elements"].([]any)
	if len(elements) == 0 {
		t.Fatal("expected at least one card element")
	}
	rendered, _ := json.Marshal(elements)
	if !strings.Contains(string(rendered), "openai-1") || !strings.Contains(string(rendered), "timeout") {
		t.Errorf("expected placeholders filled in card content, got: %s", string(rendered))
	}
}

// TestSendLarkNotify_WithSign verifies that when secret is set, the request
// includes timestamp + sign fields that match genLarkSign output.
func TestSendLarkNotify_WithSign(t *testing.T) {
	disableSSRFForTest(t)
	var captured map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	notify := dto.NewNotify("test", "title", "hello", nil)
	if err := SendLarkNotify(srv.URL, "my-secret", notify); err != nil {
		t.Fatalf("SendLarkNotify failed: %v", err)
	}

	tsStr, ok := captured["timestamp"].(string)
	if !ok || tsStr == "" {
		t.Fatalf("expected non-empty timestamp string, got %v", captured["timestamp"])
	}
	ts, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil {
		t.Fatalf("timestamp not a valid int: %v", err)
	}

	gotSign, _ := captured["sign"].(string)
	wantSign := genLarkSign("my-secret", ts)
	if gotSign != wantSign {
		t.Errorf("sign mismatch: want %q, got %q", wantSign, gotSign)
	}
}

// TestSendLarkNotify_Non2xx ensures a non-2xx response is treated as failure.
func TestSendLarkNotify_Non2xx(t *testing.T) {
	disableSSRFForTest(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	err := SendLarkNotify(srv.URL, "", dto.NewNotify("x", "x", "x", nil))
	if err == nil {
		t.Fatal("expected error on non-2xx response, got nil")
	}
	if !strings.Contains(err.Error(), "400") {
		t.Errorf("expected error to mention HTTP 400, got: %v", err)
	}
}
