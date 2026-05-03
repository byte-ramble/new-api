package service

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/setting/system_setting"
)

// LarkPayload is the request body for a Lark / Feishu custom-bot webhook.
//
// Reference: https://open.larksuite.com/document/uAjLw4CM/ukTMukTMukTM/bot-v2/add-custom-bot
//
// Two msg_type variants are commonly used:
//   - "text"        → Content = {"text": "..."}            (plain text)
//   - "interactive" → Card    = {...}                      (rich card UI)
//
// We default to "interactive" for richer alerts; downstream callers can build
// raw payloads themselves if they need a different shape.
type LarkPayload struct {
	Timestamp string                 `json:"timestamp,omitempty"` // unix seconds (string), required when sign is set
	Sign      string                 `json:"sign,omitempty"`      // base64( HMAC-SHA256(key="ts\nsecret", msg="") )
	MsgType   string                 `json:"msg_type"`            // "text" | "interactive"
	Content   map[string]string      `json:"content,omitempty"`   // for msg_type=text
	Card      map[string]interface{} `json:"card,omitempty"`      // for msg_type=interactive
}

// genLarkSign computes the Lark/Feishu custom-bot 加签 signature.
//
// Spec (counter-intuitive — the timestamp+secret is the HMAC KEY, message is empty):
//
//	stringToSign := strconv.FormatInt(timestamp, 10) + "\n" + secret
//	digest       := HmacSHA256(key=stringToSign, msg="")
//	signature    := base64.StdEncoding.EncodeToString(digest)
//
// Reference: https://open.larksuite.com/document/uAjLw4CM/ukTMukTMukTM/bot-v2/add-custom-bot#3c6592d6
func genLarkSign(secret string, timestamp int64) string {
	stringToSign := strconv.FormatInt(timestamp, 10) + "\n" + secret
	mac := hmac.New(sha256.New, []byte(stringToSign))
	// message is intentionally empty — per Lark spec
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// SendLarkNotify sends an alert to a Lark / Feishu custom-bot webhook.
//
//   - webhookURL: full URL like https://open.feishu.cn/open-apis/bot/v2/hook/<token>
//   - secret:     optional; required only when 加签 verification is enabled on the bot
//   - data:       standard Notify DTO (placeholders in Content are filled from Values)
//
// Returns an error when the URL fails SSRF validation, the request can't be sent,
// or the webhook responds with a non-2xx status. Caller is responsible for logging.
func SendLarkNotify(webhookURL string, secret string, data dto.Notify) error {
	// Render template: %s placeholders in Content filled from Values
	content := data.Content
	if len(data.Values) > 0 {
		content = fmt.Sprintf(content, data.Values...)
	}

	title := data.Title
	if title == "" {
		title = "OmniRouter Notification"
	}

	payload := LarkPayload{
		MsgType: "interactive",
		Card: map[string]interface{}{
			"config": map[string]interface{}{"wide_screen_mode": true},
			"header": map[string]interface{}{
				"title": map[string]interface{}{
					"tag":     "plain_text",
					"content": title,
				},
				"template": templateForType(data.Type),
			},
			"elements": []map[string]interface{}{
				{
					"tag": "div",
					"text": map[string]interface{}{
						"tag":     "lark_md",
						"content": content,
					},
				},
				{"tag": "hr"},
				{
					"tag": "note",
					"elements": []map[string]interface{}{
						{
							"tag":     "plain_text",
							"content": fmt.Sprintf("type: %s · %s", data.Type, time.Now().Format("2006-01-02 15:04:05")),
						},
					},
				},
			},
		},
	}

	if secret != "" {
		ts := time.Now().Unix()
		payload.Timestamp = strconv.FormatInt(ts, 10)
		payload.Sign = genLarkSign(secret, ts)
	}

	body, err := common.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal lark payload: %w", err)
	}

	// Reuse the SSRF guard already used by the generic webhook channel.
	fetchSetting := system_setting.GetFetchSetting()
	if err := common.ValidateURLWithFetchSetting(
		webhookURL,
		fetchSetting.EnableSSRFProtection,
		fetchSetting.AllowPrivateIp,
		fetchSetting.DomainFilterMode,
		fetchSetting.IpFilterMode,
		fetchSetting.DomainList,
		fetchSetting.IpList,
		fetchSetting.AllowedPorts,
		fetchSetting.ApplyIPFilterForDomain,
	); err != nil {
		return fmt.Errorf("lark webhook url rejected by SSRF guard: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, webhookURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("build lark request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := GetHttpClient()
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send lark request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("lark webhook responded HTTP %d", resp.StatusCode)
	}
	return nil
}

// templateForType maps OmniRouter notify types to Lark card header colors.
// Lark templates: blue / wathet / turquoise / green / yellow / orange / red / carmine / violet / purple / indigo / grey
func templateForType(t string) string {
	switch t {
	case dto.NotifyTypeQuotaExceed:
		return "orange"
	case dto.NotifyTypeChannelUpdate:
		return "blue"
	case dto.NotifyTypeChannelTest:
		return "yellow"
	default:
		return "indigo"
	}
}
