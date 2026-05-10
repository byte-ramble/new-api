package claude_code

import (
	"errors"

	"github.com/QuantumNous/new-api/common"
)

// OAuthKey 描述存储在 Channel.Key 中的 Claude Code 订阅 OAuth 凭据。
// 字段命名与 codex.OAuthKey 保持一致，便于后续维护时对照。
type OAuthKey struct {
	IDToken      string `json:"id_token,omitempty"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`

	AccountID   string `json:"account_id,omitempty"`
	LastRefresh string `json:"last_refresh,omitempty"`
	Email       string `json:"email,omitempty"`
	Type        string `json:"type,omitempty"`    // "claude_code"
	Expired     string `json:"expired,omitempty"` // ISO8601
}

func ParseOAuthKey(raw string) (*OAuthKey, error) {
	if raw == "" {
		return nil, errors.New("claude_code channel: empty oauth key")
	}
	var key OAuthKey
	if err := common.Unmarshal([]byte(raw), &key); err != nil {
		return nil, errors.New("claude_code channel: invalid oauth key json")
	}
	return &key, nil
}
