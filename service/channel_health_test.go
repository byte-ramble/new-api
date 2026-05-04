package service

import (
	"testing"

	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/QuantumNous/new-api/types"
)

// withBurstSettings is a tiny helper that overrides the global monitor
// settings for the duration of a test. Restores previous values via t.Cleanup
// so test order doesn't leak state.
func withBurstSettings(t *testing.T, threshold, windowSec int) {
	t.Helper()
	cfg := operation_setting.GetMonitorSetting()
	prevT := cfg.DisableBurstThreshold
	prevW := cfg.DisableBurstWindowSec
	cfg.DisableBurstThreshold = threshold
	cfg.DisableBurstWindowSec = windowSec
	t.Cleanup(func() {
		cfg.DisableBurstThreshold = prevT
		cfg.DisableBurstWindowSec = prevW
	})
	// also reset any state left by previous tests
	ResetChannelHealth(-9999)
}

func TestClassifyChannelError(t *testing.T) {
	cases := []struct {
		name string
		code int
		want ChannelErrorReason
	}{
		{"401", 401, ReasonAuthFailed},
		{"403", 403, ReasonAuthFailed},
		{"429", 429, ReasonRateLimited},
		{"500", 500, ReasonUpstream5xx},
		{"503", 503, ReasonUpstream5xx},
		{"599", 599, ReasonUpstream5xx},
		{"network (0)", 0, ReasonNetwork},
		{"400", 400, ReasonOther},
		{"404", 404, ReasonOther},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := &types.NewAPIError{StatusCode: c.code}
			if got := ClassifyChannelError(err); got != c.want {
				t.Errorf("StatusCode=%d: want %q, got %q", c.code, c.want, got)
			}
		})
	}
	// nil shouldn't panic
	if got := ClassifyChannelError(nil); got != ReasonOther {
		t.Errorf("nil err: want %q, got %q", ReasonOther, got)
	}
}

func TestRecordChannelError_Disabled(t *testing.T) {
	withBurstSettings(t, 0, 60) // disabled
	got := RecordChannelError(101, ReasonUpstream5xx)
	if got != 0 {
		t.Errorf("debouncer disabled: want 0, got %d", got)
	}
}

func TestRecordChannelError_Counts(t *testing.T) {
	withBurstSettings(t, 5, 60)
	const ch = 202
	ResetChannelHealth(ch)
	for i := 1; i <= 7; i++ {
		got := RecordChannelError(ch, ReasonUpstream5xx)
		if got != i {
			t.Fatalf("after %d records: want %d, got %d", i, i, got)
		}
	}
}

func TestHasReachedBurstThreshold(t *testing.T) {
	withBurstSettings(t, 5, 60)
	const ch = 303
	ResetChannelHealth(ch)
	if HasReachedBurstThreshold(ch) {
		t.Fatal("fresh channel: should not have reached threshold")
	}
	for i := 0; i < 4; i++ {
		RecordChannelError(ch, ReasonUpstream5xx)
	}
	if HasReachedBurstThreshold(ch) {
		t.Fatal("4 errors < 5: should not have reached threshold")
	}
	RecordChannelError(ch, ReasonUpstream5xx)
	if !HasReachedBurstThreshold(ch) {
		t.Fatal("5 errors == 5: should have reached threshold")
	}
}

func TestResetChannelHealth(t *testing.T) {
	withBurstSettings(t, 3, 60)
	const ch = 404
	for i := 0; i < 3; i++ {
		RecordChannelError(ch, ReasonUpstream5xx)
	}
	if !HasReachedBurstThreshold(ch) {
		t.Fatal("setup: should be at threshold")
	}
	ResetChannelHealth(ch)
	if HasReachedBurstThreshold(ch) {
		t.Fatal("after reset: should be back below threshold")
	}
}

// TestRecordChannelError_DisabledThresholdReturnsZero ensures the function
// short-circuits cleanly when threshold==0 (legacy mode) so callers can't
// accidentally trigger debounced behavior on a single error.
func TestRecordChannelError_DisabledThresholdReturnsZero(t *testing.T) {
	withBurstSettings(t, 0, 60)
	for i := 0; i < 10; i++ {
		got := RecordChannelError(505, ReasonUpstream5xx)
		if got != 0 {
			t.Fatalf("threshold=0: want always 0, got %d", got)
		}
	}
	if !HasReachedBurstThreshold(505) {
		t.Error("threshold=0: HasReachedBurstThreshold should always allow (return true), got false")
	}
}
