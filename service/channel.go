package service

import (
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/QuantumNous/new-api/types"
)

func formatNotifyType(channelId int, status int) string {
	return fmt.Sprintf("%s_%d_%d", dto.NotifyTypeChannelUpdate, channelId, status)
}

// disable & notify
//
// OmniRouter additions on top of upstream behavior:
//   - On successful DB status update, also fan out to:
//       service.SendSystemAlert  → Lark / Feishu group chat (operations)
//       middleware.RecordChannelAutoDisabled → Prometheus counter
//       service.ResetChannelHealth          → clear sliding-window state
//   - reasonClass (defaulting to ReasonOther when empty) is the low-cardinality
//     bucket for metrics + alert routing. The free-form `reason` string still
//     carries the full upstream error message for human readers.
func DisableChannel(channelError types.ChannelError, reason string, reasonClass ...ChannelErrorReason) {
	common.SysLog(fmt.Sprintf("通道「%s」（#%d）发生错误，准备禁用，原因：%s", channelError.ChannelName, channelError.ChannelId, reason))

	// 检查是否启用自动禁用功能
	if !channelError.AutoBan {
		common.SysLog(fmt.Sprintf("通道「%s」（#%d）未启用自动禁用功能，跳过禁用操作", channelError.ChannelName, channelError.ChannelId))
		return
	}

	success := model.UpdateChannelStatus(channelError.ChannelId, channelError.UsingKey, common.ChannelStatusAutoDisabled, reason)
	if !success {
		return
	}

	subject := fmt.Sprintf("通道「%s」（#%d）已被禁用", channelError.ChannelName, channelError.ChannelId)
	content := fmt.Sprintf("通道「%s」（#%d）已被禁用，原因：%s", channelError.ChannelName, channelError.ChannelId, reason)
	NotifyRootUser(formatNotifyType(channelError.ChannelId, common.ChannelStatusAutoDisabled), subject, content)

	// OmniRouter additions ──────────────────────────────────────
	rc := ReasonOther
	if len(reasonClass) > 0 && reasonClass[0] != "" {
		rc = reasonClass[0]
	}
	// Prometheus counter
	RecordChannelAutoDisabled(channelError.ChannelName, string(rc))
	// Reset sliding-window so a recovered+re-failed channel starts fresh
	ResetChannelHealth(channelError.ChannelId)
	// Lark alert (fire-and-forget; SendSystemAlert swallows errors internally)
	SendSystemAlert(dto.NewNotify(
		dto.NotifyTypeChannelUpdate,
		subject,
		fmt.Sprintf("Channel `%s` (#%d) auto-disabled.\n**Reason class:** `%s`\n**Detail:** %s",
			channelError.ChannelName, channelError.ChannelId, rc, reason),
		nil,
	))
}

func EnableChannel(channelId int, usingKey string, channelName string) {
	success := model.UpdateChannelStatus(channelId, usingKey, common.ChannelStatusEnabled, "")
	if !success {
		return
	}
	subject := fmt.Sprintf("通道「%s」（#%d）已被启用", channelName, channelId)
	content := fmt.Sprintf("通道「%s」（#%d）已被启用", channelName, channelId)
	NotifyRootUser(formatNotifyType(channelId, common.ChannelStatusEnabled), subject, content)

	// OmniRouter additions
	RecordChannelRecovered(channelName)
	ResetChannelHealth(channelId)
	SendSystemAlert(dto.NewNotify(
		dto.NotifyTypeChannelUpdate,
		subject,
		fmt.Sprintf("Channel `%s` (#%d) recovered and re-enabled.", channelName, channelId),
		nil,
	))
}

func ShouldDisableChannel(err *types.NewAPIError) bool {
	if !common.AutomaticDisableChannelEnabled {
		return false
	}
	if err == nil {
		return false
	}
	if types.IsChannelError(err) {
		return true
	}
	if types.IsSkipRetryError(err) {
		return false
	}
	if operation_setting.ShouldDisableByStatusCode(err.StatusCode) {
		return true
	}

	lowerMessage := strings.ToLower(err.Error())
	search, _ := AcSearch(lowerMessage, operation_setting.AutomaticDisableKeywords, true)
	return search
}

func ShouldEnableChannel(newAPIError *types.NewAPIError, status int) bool {
	if !common.AutomaticEnableChannelEnabled {
		return false
	}
	if newAPIError != nil {
		return false
	}
	if status != common.ChannelStatusAutoDisabled {
		return false
	}
	return true
}
