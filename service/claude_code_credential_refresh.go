package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
)

type ClaudeCodeCredentialRefreshOptions struct {
	ResetCaches bool
}

// ClaudeCodeOAuthKey 与 relay/channel/claude_code.OAuthKey 同构，
// 单独在 service 包定义以避免 service -> relay 包循环依赖（与 codex 模式一致）。
type ClaudeCodeOAuthKey struct {
	IDToken      string `json:"id_token,omitempty"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`

	AccountID   string `json:"account_id,omitempty"`
	LastRefresh string `json:"last_refresh,omitempty"`
	Email       string `json:"email,omitempty"`
	Type        string `json:"type,omitempty"`
	Expired     string `json:"expired,omitempty"`
}

func parseClaudeCodeOAuthKey(raw string) (*ClaudeCodeOAuthKey, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, errors.New("claude_code channel: empty oauth key")
	}
	var key ClaudeCodeOAuthKey
	if err := common.Unmarshal([]byte(raw), &key); err != nil {
		return nil, errors.New("claude_code channel: invalid oauth key json")
	}
	return &key, nil
}

func RefreshClaudeCodeChannelCredential(ctx context.Context, channelID int, opts ClaudeCodeCredentialRefreshOptions) (*ClaudeCodeOAuthKey, *model.Channel, error) {
	ch, err := model.GetChannelById(channelID, true)
	if err != nil {
		return nil, nil, err
	}
	if ch == nil {
		return nil, nil, fmt.Errorf("channel not found")
	}
	if ch.Type != constant.ChannelTypeClaudeCode {
		return nil, nil, fmt.Errorf("channel type is not Claude Code")
	}

	oauthKey, err := parseClaudeCodeOAuthKey(strings.TrimSpace(ch.Key))
	if err != nil {
		return nil, nil, err
	}
	if strings.TrimSpace(oauthKey.RefreshToken) == "" {
		return nil, nil, fmt.Errorf("claude_code channel: refresh_token is required to refresh credential")
	}

	refreshCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	res, err := RefreshClaudeCodeOAuthTokenWithProxy(refreshCtx, oauthKey.RefreshToken, ch.GetSetting().Proxy)
	if err != nil {
		return nil, nil, err
	}

	oauthKey.AccessToken = res.AccessToken
	if strings.TrimSpace(res.RefreshToken) != "" {
		oauthKey.RefreshToken = res.RefreshToken
	}
	oauthKey.LastRefresh = time.Now().Format(time.RFC3339)
	oauthKey.Expired = res.ExpiresAt.Format(time.RFC3339)
	if strings.TrimSpace(oauthKey.Type) == "" {
		oauthKey.Type = "claude_code"
	}

	// AccountID / Email 从 JWT 尝试提取（失败容忍，订阅 OAuth 这两个字段不一定存在）。
	if strings.TrimSpace(oauthKey.AccountID) == "" {
		if accountID, ok := ExtractClaudeCodeAccountIDFromJWT(oauthKey.AccessToken); ok {
			oauthKey.AccountID = accountID
		}
	}
	if strings.TrimSpace(oauthKey.Email) == "" {
		if email, ok := ExtractClaudeCodeEmailFromJWT(oauthKey.AccessToken); ok {
			oauthKey.Email = email
		}
	}

	encoded, err := common.Marshal(oauthKey)
	if err != nil {
		return nil, nil, err
	}

	if err := model.DB.Model(&model.Channel{}).Where("id = ?", ch.Id).Update("key", string(encoded)).Error; err != nil {
		return nil, nil, err
	}

	if opts.ResetCaches {
		model.InitChannelCache()
		ResetProxyClientCache()
	}

	return oauthKey, ch, nil
}
