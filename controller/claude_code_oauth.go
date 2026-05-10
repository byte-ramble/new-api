package controller

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/relay/channel/claude_code"
	"github.com/QuantumNous/new-api/service"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type claudeCodeOAuthCompleteRequest struct {
	Input string `json:"input"`
}

func claudeCodeOAuthSessionKey(channelID int, field string) string {
	return fmt.Sprintf("claude_code_oauth_%s_%d", field, channelID)
}

// parseClaudeCodeAuthorizationInput 兼容多种回调输入：
//   - "code#state"（Anthropic copy-paste 流程返回值常见格式）
//   - 完整回调 URL：https://console.anthropic.com/oauth/code/callback?code=xxx&state=yyy
//   - 仅查询字符串：code=xxx&state=yyy
//   - 仅 code（state 留空，由调用方决定是否报错）
func parseClaudeCodeAuthorizationInput(input string) (code string, state string, err error) {
	v := strings.TrimSpace(input)
	if v == "" {
		return "", "", errors.New("empty input")
	}
	if strings.Contains(v, "#") {
		parts := strings.SplitN(v, "#", 2)
		code = strings.TrimSpace(parts[0])
		state = strings.TrimSpace(parts[1])
		return code, state, nil
	}
	if strings.Contains(v, "code=") {
		// 优先走 url.Parse 处理完整 URL（含 scheme://host/path?query 形式）。
		// 仅当解析出非空 code 时才采用结果，否则回落到 url.ParseQuery 处理裸查询字符串。
		// 这样 "code=xxx&state=yyy" 这种文档示例形式也能被正确解析（url.Parse 会把整串当 path）。
		if u, parseErr := url.Parse(v); parseErr == nil {
			q := u.Query()
			code = strings.TrimSpace(q.Get("code"))
			state = strings.TrimSpace(q.Get("state"))
			if code != "" {
				return code, state, nil
			}
		}
		if q, parseErr := url.ParseQuery(v); parseErr == nil {
			code = strings.TrimSpace(q.Get("code"))
			state = strings.TrimSpace(q.Get("state"))
			return code, state, nil
		}
	}

	code = v
	return code, "", nil
}

func StartClaudeCodeOAuth(c *gin.Context) {
	startClaudeCodeOAuthWithChannelID(c, 0)
}

func StartClaudeCodeOAuthForChannel(c *gin.Context) {
	channelID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, fmt.Errorf("invalid channel id: %w", err))
		return
	}
	startClaudeCodeOAuthWithChannelID(c, channelID)
}

func startClaudeCodeOAuthWithChannelID(c *gin.Context, channelID int) {
	if channelID > 0 {
		ch, err := model.GetChannelById(channelID, false)
		if err != nil {
			common.ApiError(c, err)
			return
		}
		if ch == nil {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "channel not found"})
			return
		}
		if ch.Type != constant.ChannelTypeClaudeCode {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "channel type is not Claude Code"})
			return
		}
	}

	flow, err := service.CreateClaudeCodeOAuthAuthorizationFlow()
	if err != nil {
		common.ApiError(c, err)
		return
	}

	session := sessions.Default(c)
	session.Set(claudeCodeOAuthSessionKey(channelID, "state"), flow.State)
	session.Set(claudeCodeOAuthSessionKey(channelID, "verifier"), flow.Verifier)
	session.Set(claudeCodeOAuthSessionKey(channelID, "created_at"), time.Now().Unix())
	_ = session.Save()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"authorize_url": flow.AuthorizeURL,
		},
	})
}

func CompleteClaudeCodeOAuth(c *gin.Context) {
	completeClaudeCodeOAuthWithChannelID(c, 0)
}

func CompleteClaudeCodeOAuthForChannel(c *gin.Context) {
	channelID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, fmt.Errorf("invalid channel id: %w", err))
		return
	}
	completeClaudeCodeOAuthWithChannelID(c, channelID)
}

func completeClaudeCodeOAuthWithChannelID(c *gin.Context, channelID int) {
	req := claudeCodeOAuthCompleteRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiError(c, err)
		return
	}

	code, state, err := parseClaudeCodeAuthorizationInput(req.Input)
	if err != nil {
		common.SysError("failed to parse claude_code authorization input: " + err.Error())
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "解析授权信息失败，请检查输入格式"})
		return
	}
	if strings.TrimSpace(code) == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "missing authorization code"})
		return
	}
	if strings.TrimSpace(state) == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "missing state in input"})
		return
	}

	channelProxy := ""
	if channelID > 0 {
		ch, err := model.GetChannelById(channelID, false)
		if err != nil {
			common.ApiError(c, err)
			return
		}
		if ch == nil {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "channel not found"})
			return
		}
		if ch.Type != constant.ChannelTypeClaudeCode {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "channel type is not Claude Code"})
			return
		}
		channelProxy = ch.GetSetting().Proxy
	}

	session := sessions.Default(c)
	expectedState, _ := session.Get(claudeCodeOAuthSessionKey(channelID, "state")).(string)
	verifier, _ := session.Get(claudeCodeOAuthSessionKey(channelID, "verifier")).(string)
	if strings.TrimSpace(expectedState) == "" || strings.TrimSpace(verifier) == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "oauth flow not started or session expired"})
		return
	}
	if state != expectedState {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "state mismatch"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	tokenRes, err := service.ExchangeClaudeCodeAuthorizationCodeWithProxy(ctx, code, verifier, state, channelProxy)
	if err != nil {
		common.SysError("failed to exchange claude_code authorization code: " + err.Error())
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "授权码交换失败，请重试"})
		return
	}

	// account_id / email 是可选字段（Anthropic JWT 里可能不存在），失败容忍。
	accountID, _ := service.ExtractClaudeCodeAccountIDFromJWT(tokenRes.AccessToken)
	email, _ := service.ExtractClaudeCodeEmailFromJWT(tokenRes.AccessToken)

	key := claude_code.OAuthKey{
		AccessToken:  tokenRes.AccessToken,
		RefreshToken: tokenRes.RefreshToken,
		AccountID:    accountID,
		LastRefresh:  time.Now().Format(time.RFC3339),
		Expired:      tokenRes.ExpiresAt.Format(time.RFC3339),
		Email:        email,
		Type:         "claude_code",
	}
	encoded, err := common.Marshal(key)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	session.Delete(claudeCodeOAuthSessionKey(channelID, "state"))
	session.Delete(claudeCodeOAuthSessionKey(channelID, "verifier"))
	session.Delete(claudeCodeOAuthSessionKey(channelID, "created_at"))
	_ = session.Save()

	if channelID > 0 {
		if err := model.DB.Model(&model.Channel{}).Where("id = ?", channelID).Update("key", string(encoded)).Error; err != nil {
			common.ApiError(c, err)
			return
		}
		model.InitChannelCache()
		service.ResetProxyClientCache()
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "saved",
			"data": gin.H{
				"channel_id":   channelID,
				"account_id":   accountID,
				"email":        email,
				"expires_at":   key.Expired,
				"last_refresh": key.LastRefresh,
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "generated",
		"data": gin.H{
			"key":          string(encoded),
			"account_id":   accountID,
			"email":        email,
			"expires_at":   key.Expired,
			"last_refresh": key.LastRefresh,
		},
	})
}

// RefreshClaudeCodeChannelCredentialController 手动触发某个渠道的凭据刷新（管理员操作）。
func RefreshClaudeCodeChannelCredentialController(c *gin.Context) {
	channelID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, fmt.Errorf("invalid channel id: %w", err))
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	newKey, _, err := service.RefreshClaudeCodeChannelCredential(ctx, channelID, service.ClaudeCodeCredentialRefreshOptions{ResetCaches: true})
	if err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "refreshed",
		"data": gin.H{
			"channel_id":   channelID,
			"account_id":   newKey.AccountID,
			"email":        newKey.Email,
			"expires_at":   newKey.Expired,
			"last_refresh": newKey.LastRefresh,
		},
	})
}
