package operation_setting

import (
	"os"
	"strconv"

	"github.com/QuantumNous/new-api/setting/config"
)

// AffiliateSetting holds the OmniRouter multi-level commission configuration.
//
// Disabled by default — operator opts in by setting Enabled=true (or env
// AFFILIATE_ENABLED=true) AFTER reading omnirouter-docs/operations/affiliate-program.md
// and ensuring legal / TOS coverage of the program.
type AffiliateSetting struct {
	// Master switch. When false, PayCommission is a no-op and any affiliate
	// API endpoints return a feature-disabled error.
	Enabled bool `json:"enabled"`

	// Commission rates, expressed as PERCENTAGE (15.0 means 15%).
	// L1 = direct inviter (the user whose AffCode the new user used).
	// L2 = inviter's inviter (the user who invited the L1).
	// Set L2 = 0 to run a single-level program.
	Level1RatePct float64 `json:"level1_rate_pct"`
	Level2RatePct float64 `json:"level2_rate_pct"`

	// Anti-fraud: maximum commission a single earner can accumulate from
	// L1+L2 events in any rolling 24-hour window. 0 = no cap.
	MaxDailyCommissionRmb float64 `json:"max_daily_commission_rmb"`

	// Withdrawal economics.
	MinWithdrawalRmb     float64 `json:"min_withdrawal_rmb"`     // 0 = no minimum
	WithdrawalFeeRmb     float64 `json:"withdrawal_fee_rmb"`     // flat per-request fee (e.g. ¥2 to cover Alipay handling)
	MaxWithdrawalsPerDay int     `json:"max_withdrawals_per_day"` // anti-spam, 0 = no limit

	// Self-invite policy. When false, the inviter and invitee being the same
	// person (detected by IP / device fingerprint match at registration time)
	// disables commission for that lineage.
	AllowSelfInvite bool `json:"allow_self_invite"`
}

// Conservative defaults.
//   - Disabled out of the box (operator must opt in)
//   - 15% L1 / 5% L2 (industry-typical for AI/SaaS affiliate programs)
//   - ¥1000 daily cap (sane sanity ceiling against fraud)
//   - ¥100 minimum withdrawal (covers payment-channel fees + admin overhead)
var affiliateSetting = AffiliateSetting{
	Enabled:               false,
	Level1RatePct:         15.0,
	Level2RatePct:         5.0,
	MaxDailyCommissionRmb: 1000.0,
	MinWithdrawalRmb:      100.0,
	WithdrawalFeeRmb:      0.0,
	MaxWithdrawalsPerDay:  3,
	AllowSelfInvite:       false,
}

func init() {
	config.GlobalConfig.Register("affiliate_setting", &affiliateSetting)
}

// GetAffiliateSetting returns the live config, applying env overrides each
// call so changes via docker-compose env take effect on restart.
func GetAffiliateSetting() *AffiliateSetting {
	if v := os.Getenv("AFFILIATE_ENABLED"); v != "" {
		affiliateSetting.Enabled = (v == "1" || v == "true" || v == "yes")
	}
	if v := os.Getenv("AFFILIATE_L1_PCT"); v != "" {
		if n, err := strconv.ParseFloat(v, 64); err == nil && n >= 0 && n <= 100 {
			affiliateSetting.Level1RatePct = n
		}
	}
	if v := os.Getenv("AFFILIATE_L2_PCT"); v != "" {
		if n, err := strconv.ParseFloat(v, 64); err == nil && n >= 0 && n <= 100 {
			affiliateSetting.Level2RatePct = n
		}
	}
	if v := os.Getenv("AFFILIATE_DAILY_CAP_RMB"); v != "" {
		if n, err := strconv.ParseFloat(v, 64); err == nil && n >= 0 {
			affiliateSetting.MaxDailyCommissionRmb = n
		}
	}
	if v := os.Getenv("AFFILIATE_MIN_WITHDRAWAL_RMB"); v != "" {
		if n, err := strconv.ParseFloat(v, 64); err == nil && n >= 0 {
			affiliateSetting.MinWithdrawalRmb = n
		}
	}
	return &affiliateSetting
}
