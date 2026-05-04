package service

import (
	"sync"
	"time"

	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/QuantumNous/new-api/types"
)

// ============================================================================
// channel_health.go — sliding-window error de-bouncer for channel auto-disable.
//
// Problem: upstream new-api flips a channel to "auto disabled" the FIRST
// time service.ShouldDisableChannel returns true. A single transient blip
// (one timeout, one 5xx during an upstream redeploy) is enough to take a
// healthy channel out of rotation. False-positives are operationally
// expensive — the recovery probe loop runs every AutoTestChannelMinutes
// (default 10 min), so a flap costs minutes of unnecessary downtime.
//
// Solution: keep an in-memory ring of recent error timestamps per channel.
// Only let the legacy disable trigger fire when the burst threshold is met
// (e.g. 5 errors in 60s).
//
// Settings live in operation_setting.MonitorSetting:
//   DisableBurstThreshold (0 disables the de-bouncer; legacy behavior)
//   DisableBurstWindowSec
//
// Notes:
//   - In-memory only. Multi-replica deployments will each track their own
//     window; that's fine — disable is an idempotent global write to DB.
//   - sync.RWMutex around a plain map is simpler than sync.Map for this
//     access pattern (read + mutate per call) and the contention is bounded
//     by the number of distinct channel IDs.
// ============================================================================

// ChannelErrorReason is a low-cardinality bucket suitable for Prometheus
// labels and Lark alert routing. Add buckets sparingly — every distinct
// value creates new metric series.
type ChannelErrorReason string

const (
	ReasonAuthFailed      ChannelErrorReason = "auth_failed"
	ReasonRateLimited     ChannelErrorReason = "rate_limited"
	ReasonUpstream5xx     ChannelErrorReason = "upstream_5xx"
	ReasonNetwork         ChannelErrorReason = "network"
	ReasonQuotaExhausted  ChannelErrorReason = "quota_exhausted"
	ReasonManual          ChannelErrorReason = "manual"
	ReasonOther           ChannelErrorReason = "other"
)

// ClassifyChannelError maps a NewAPIError into one of the low-cardinality
// reason buckets above. Used by both the sliding window and the alert/metric
// emission paths.
func ClassifyChannelError(err *types.NewAPIError) ChannelErrorReason {
	if err == nil {
		return ReasonOther
	}
	code := err.StatusCode
	switch {
	case code == 401 || code == 403:
		return ReasonAuthFailed
	case code == 429:
		return ReasonRateLimited
	case code >= 500 && code <= 599:
		return ReasonUpstream5xx
	case code == 0:
		// No HTTP response (DNS / connect refused / TLS handshake / context cancel)
		return ReasonNetwork
	default:
		return ReasonOther
	}
}

// channelErrorRing holds recent error timestamps for one channel. Append-only
// during normal operation; pruned on every Record() call so the slice never
// grows unbounded as long as the threshold check happens at recording time.
type channelErrorRing struct {
	timestamps []time.Time
}

var (
	channelHealthMu      sync.Mutex
	channelHealthByID    = make(map[int]*channelErrorRing)
)

// RecordChannelError appends a single failure timestamp for the given channel.
// Old entries (outside the configured window) are pruned in the same call so
// the underlying slice stays bounded.
//
// Returns the number of failures within the active window AFTER recording —
// callers can compare against the burst threshold to decide whether to actually
// trigger DisableChannel. Threshold value 0 means "no de-bounce" and callers
// should bypass this entirely (legacy behavior).
func RecordChannelError(channelID int, _ ChannelErrorReason) int {
	cfg := operation_setting.GetMonitorSetting()
	if cfg.DisableBurstThreshold <= 0 {
		// De-bouncer disabled — nothing to track. Returning 0 prevents
		// callers that mis-use the return value from triggering anything.
		return 0
	}
	windowSec := cfg.DisableBurstWindowSec
	if windowSec <= 0 {
		windowSec = 60 // safety net
	}
	now := time.Now()
	cutoff := now.Add(-time.Duration(windowSec) * time.Second)

	channelHealthMu.Lock()
	defer channelHealthMu.Unlock()

	ring, ok := channelHealthByID[channelID]
	if !ok {
		ring = &channelErrorRing{}
		channelHealthByID[channelID] = ring
	}

	// Prune entries older than the window
	keep := ring.timestamps[:0]
	for _, t := range ring.timestamps {
		if t.After(cutoff) {
			keep = append(keep, t)
		}
	}
	keep = append(keep, now)
	ring.timestamps = keep

	return len(keep)
}

// ResetChannelHealth wipes the recorded error history for a channel. Call this
// after EnableChannel succeeds so a recovered channel starts with a clean
// slate (otherwise leftover timestamps from before the disable could
// immediately re-trigger the threshold).
func ResetChannelHealth(channelID int) {
	channelHealthMu.Lock()
	defer channelHealthMu.Unlock()
	delete(channelHealthByID, channelID)
}

// HasReachedBurstThreshold is a convenience for callers that have already
// recorded an error and want a separate decision step. Equivalent to
// `RecordChannelError(...) >= threshold` but does not append a new entry —
// useful in dry-run / inspection paths.
func HasReachedBurstThreshold(channelID int) bool {
	cfg := operation_setting.GetMonitorSetting()
	if cfg.DisableBurstThreshold <= 0 {
		return true // de-bouncer disabled → always allow
	}
	windowSec := cfg.DisableBurstWindowSec
	if windowSec <= 0 {
		windowSec = 60
	}
	cutoff := time.Now().Add(-time.Duration(windowSec) * time.Second)

	channelHealthMu.Lock()
	defer channelHealthMu.Unlock()

	ring, ok := channelHealthByID[channelID]
	if !ok {
		return false
	}
	count := 0
	for _, t := range ring.timestamps {
		if t.After(cutoff) {
			count++
		}
	}
	return count >= cfg.DisableBurstThreshold
}
