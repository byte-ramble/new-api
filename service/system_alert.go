package service

import (
	"os"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
)

// SendSystemAlert dispatches a system-level alert to all configured channels.
//
// This is the single entry point business logic should use to raise an
// operational alert (channel down, quota threshold, deploy event, etc.).
// It fans out to whichever channels are configured via env vars at runtime,
// so call sites do not need to know about Lark/Slack/email/etc.
//
// Currently supported channels:
//   - Lark / Feishu — env: LARK_WEBHOOK_URL (required), LARK_WEBHOOK_SECRET (optional)
//
// All channel failures are logged and swallowed: an alerting subsystem must
// never propagate errors back into the request path that triggered it.
//
// Future: load these from setting/system_setting so admins can configure them
// from the dashboard without restart. For MVP, env var is sufficient.
func SendSystemAlert(notify dto.Notify) {
	if larkURL := os.Getenv("LARK_WEBHOOK_URL"); larkURL != "" {
		secret := os.Getenv("LARK_WEBHOOK_SECRET")
		if err := SendLarkNotify(larkURL, secret, notify); err != nil {
			common.SysError("[SystemAlert] lark notify failed: " + err.Error())
		}
	}
	// Add slack / dingtalk / wecom / email channels here when needed.
}
