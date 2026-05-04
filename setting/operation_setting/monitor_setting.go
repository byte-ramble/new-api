package operation_setting

import (
	"os"
	"strconv"

	"github.com/QuantumNous/new-api/setting/config"
)

type MonitorSetting struct {
	AutoTestChannelEnabled bool    `json:"auto_test_channel_enabled"`
	AutoTestChannelMinutes float64 `json:"auto_test_channel_minutes"`

	// OmniRouter additions: sliding-window de-bounce for auto-disable.
	//
	// Without this, a single transient 5xx / timeout from upstream can
	// flip a perfectly healthy channel into "auto disabled" — that's an
	// over-eager false positive, especially during minor upstream blips.
	//
	// When DisableBurstThreshold > 0, the channel must accumulate at
	// least N classifiable errors within DisableBurstWindowSec seconds
	// before service.DisableChannel actually executes. Threshold of 0
	// preserves the legacy single-error-disables behavior.
	//
	// Recommended starting points (for production):
	//   DisableBurstThreshold=5, DisableBurstWindowSec=60
	// (i.e. 5 errors in any rolling 60s window → disable)
	DisableBurstThreshold int `json:"disable_burst_threshold"`
	DisableBurstWindowSec int `json:"disable_burst_window_sec"`
}

// 默认配置
var monitorSetting = MonitorSetting{
	AutoTestChannelEnabled: false,
	AutoTestChannelMinutes: 10,
	// Default 0 → legacy behavior (single error disables). Operators can
	// opt-in to debounce via env vars or admin settings without changing
	// any existing call sites.
	DisableBurstThreshold: 0,
	DisableBurstWindowSec: 60,
}

func init() {
	// 注册到全局配置管理器
	config.GlobalConfig.Register("monitor_setting", &monitorSetting)
}

func GetMonitorSetting() *MonitorSetting {
	if os.Getenv("CHANNEL_TEST_FREQUENCY") != "" {
		frequency, err := strconv.Atoi(os.Getenv("CHANNEL_TEST_FREQUENCY"))
		if err == nil && frequency > 0 {
			monitorSetting.AutoTestChannelEnabled = true
			monitorSetting.AutoTestChannelMinutes = float64(frequency)
		}
	}
	// Burst threshold env override (OmniRouter)
	if v := os.Getenv("DISABLE_BURST_THRESHOLD"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			monitorSetting.DisableBurstThreshold = n
		}
	}
	if v := os.Getenv("DISABLE_BURST_WINDOW_SEC"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			monitorSetting.DisableBurstWindowSec = n
		}
	}
	return &monitorSetting
}
