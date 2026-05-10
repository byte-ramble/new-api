package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
)

// Claude Code 订阅 OAuth 端点常量。
// 这些是 Anthropic 官方 Claude Code CLI 使用的公开授权流程参数，
// 来源：社区抓包参考实现（如 changjonathanc/9f9d635b2f8692e0520a884eaf098351）。
// 若上游调整需要更新这些常量重新编译。
const (
	claudeCodeOAuthClientID     = "9d1c250a-e61b-44d9-88ed-5944d1962f5e"
	claudeCodeOAuthAuthorizeURL = "https://claude.ai/oauth/authorize"
	claudeCodeOAuthTokenURL     = "https://console.anthropic.com/v1/oauth/token"
	claudeCodeOAuthRedirectURI  = "https://console.anthropic.com/oauth/code/callback"
	claudeCodeOAuthScope        = "org:create_api_key user:profile user:inference"
	claudeCodeHTTPTimeout       = 20 * time.Second
)

// ClaudeCodeOAuthTokenResult 与 CodexOAuthTokenResult 同构，单独命名便于阅读。
type ClaudeCodeOAuthTokenResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

type ClaudeCodeOAuthAuthorizationFlow struct {
	State        string
	Verifier     string
	Challenge    string
	AuthorizeURL string
}

func RefreshClaudeCodeOAuthToken(ctx context.Context, refreshToken string) (*ClaudeCodeOAuthTokenResult, error) {
	return RefreshClaudeCodeOAuthTokenWithProxy(ctx, refreshToken, "")
}

func RefreshClaudeCodeOAuthTokenWithProxy(ctx context.Context, refreshToken string, proxyURL string) (*ClaudeCodeOAuthTokenResult, error) {
	client, err := getClaudeCodeOAuthHTTPClient(proxyURL)
	if err != nil {
		return nil, err
	}
	return refreshClaudeCodeOAuthToken(ctx, client, claudeCodeOAuthTokenURL, claudeCodeOAuthClientID, refreshToken)
}

func ExchangeClaudeCodeAuthorizationCode(ctx context.Context, code string, verifier string, state string) (*ClaudeCodeOAuthTokenResult, error) {
	return ExchangeClaudeCodeAuthorizationCodeWithProxy(ctx, code, verifier, state, "")
}

func ExchangeClaudeCodeAuthorizationCodeWithProxy(ctx context.Context, code string, verifier string, state string, proxyURL string) (*ClaudeCodeOAuthTokenResult, error) {
	client, err := getClaudeCodeOAuthHTTPClient(proxyURL)
	if err != nil {
		return nil, err
	}
	return exchangeClaudeCodeAuthorizationCode(ctx, client, claudeCodeOAuthTokenURL, claudeCodeOAuthClientID, code, verifier, state, claudeCodeOAuthRedirectURI)
}

func CreateClaudeCodeOAuthAuthorizationFlow() (*ClaudeCodeOAuthAuthorizationFlow, error) {
	// 复用 codex_oauth.go 同包内的 createStateHex / generatePKCEPair。
	state, err := createStateHex(16)
	if err != nil {
		return nil, err
	}
	verifier, challenge, err := generatePKCEPair()
	if err != nil {
		return nil, err
	}
	u, err := buildClaudeCodeAuthorizeURL(state, challenge)
	if err != nil {
		return nil, err
	}
	return &ClaudeCodeOAuthAuthorizationFlow{
		State:        state,
		Verifier:     verifier,
		Challenge:    challenge,
		AuthorizeURL: u,
	}, nil
}

// refreshClaudeCodeOAuthToken 与 codex 不同的关键点：Anthropic 的 token 端点接受 **JSON body**，
// 而不是 application/x-www-form-urlencoded。
func refreshClaudeCodeOAuthToken(
	ctx context.Context,
	client *http.Client,
	tokenURL string,
	clientID string,
	refreshToken string,
) (*ClaudeCodeOAuthTokenResult, error) {
	rt := strings.TrimSpace(refreshToken)
	if rt == "" {
		return nil, errors.New("empty refresh_token")
	}

	body := map[string]any{
		"grant_type":    "refresh_token",
		"refresh_token": rt,
		"client_id":     clientID,
	}
	return postClaudeCodeOAuth(ctx, client, tokenURL, body, "refresh", rt)
}

func exchangeClaudeCodeAuthorizationCode(
	ctx context.Context,
	client *http.Client,
	tokenURL string,
	clientID string,
	code string,
	verifier string,
	state string,
	redirectURI string,
) (*ClaudeCodeOAuthTokenResult, error) {
	c := strings.TrimSpace(code)
	v := strings.TrimSpace(verifier)
	if c == "" {
		return nil, errors.New("empty authorization code")
	}
	if v == "" {
		return nil, errors.New("empty code_verifier")
	}

	body := map[string]any{
		"grant_type":    "authorization_code",
		"code":          c,
		"state":         strings.TrimSpace(state),
		"client_id":     clientID,
		"redirect_uri":  redirectURI,
		"code_verifier": v,
	}
	return postClaudeCodeOAuth(ctx, client, tokenURL, body, "exchange", "")
}

// postClaudeCodeOAuth 把 JSON body POST 到 token 端点，并解析响应。
// fallbackRefreshToken：当上游响应不返回新的 refresh_token 时（部分 OAuth 实现会复用原值），保留旧值以避免凭据丢失。
func postClaudeCodeOAuth(
	ctx context.Context,
	client *http.Client,
	tokenURL string,
	body map[string]any,
	op string,
	fallbackRefreshToken string,
) (*ClaudeCodeOAuthTokenResult, error) {
	raw, err := common.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var payload struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		TokenType    string `json:"token_type"`
	}
	if err := common.DecodeJson(resp.Body, &payload); err != nil {
		return nil, fmt.Errorf("claude_code oauth %s: decode response failed: %w", op, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("claude_code oauth %s failed: status=%d", op, resp.StatusCode)
	}
	if strings.TrimSpace(payload.AccessToken) == "" || payload.ExpiresIn <= 0 {
		return nil, fmt.Errorf("claude_code oauth %s: response missing access_token or expires_in", op)
	}

	rt := strings.TrimSpace(payload.RefreshToken)
	if rt == "" && fallbackRefreshToken != "" {
		rt = fallbackRefreshToken
	}

	return &ClaudeCodeOAuthTokenResult{
		AccessToken:  strings.TrimSpace(payload.AccessToken),
		RefreshToken: rt,
		ExpiresAt:    time.Now().Add(time.Duration(payload.ExpiresIn) * time.Second),
	}, nil
}

func getClaudeCodeOAuthHTTPClient(proxyURL string) (*http.Client, error) {
	base, err := GetHttpClientWithProxy(strings.TrimSpace(proxyURL))
	if err != nil {
		return nil, err
	}
	if base == nil {
		return &http.Client{Timeout: claudeCodeHTTPTimeout}, nil
	}
	cp := *base
	cp.Timeout = claudeCodeHTTPTimeout
	return &cp, nil
}

func buildClaudeCodeAuthorizeURL(state string, challenge string) (string, error) {
	u, err := url.Parse(claudeCodeOAuthAuthorizeURL)
	if err != nil {
		return "", err
	}
	q := u.Query()
	// Claude Code 专属：code=true 启用 copy-paste 友好流程，授权后页面显示给用户复制 code。
	q.Set("code", "true")
	q.Set("client_id", claudeCodeOAuthClientID)
	q.Set("response_type", "code")
	q.Set("redirect_uri", claudeCodeOAuthRedirectURI)
	q.Set("scope", claudeCodeOAuthScope)
	q.Set("code_challenge", challenge)
	q.Set("code_challenge_method", "S256")
	q.Set("state", state)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// ExtractClaudeCodeEmailFromJWT 从 access_token JWT 中尝试提取 email。
// Anthropic 的 JWT claim 结构未公开稳定，这里复用通用 JWT 解析逻辑（codex_oauth.go 中的 decodeJWTClaims），
// 失败时返回 false，调用方应该把 Email 当作可选字段处理。
func ExtractClaudeCodeEmailFromJWT(token string) (string, bool) {
	claims, ok := decodeJWTClaims(token)
	if !ok {
		return "", false
	}
	// 优先尝试标准 email claim。
	if v, ok := claims["email"]; ok {
		if s, ok := v.(string); ok {
			s = strings.TrimSpace(s)
			if s != "" {
				return s, true
			}
		}
	}
	return "", false
}

// ExtractClaudeCodeAccountIDFromJWT 从 JWT 提取账号/组织标识。
// Anthropic JWT 中常见 claim 名为 organization_uuid 或 sub，按优先级尝试。
// 失败返回 false，调用方应允许该字段缺失。
func ExtractClaudeCodeAccountIDFromJWT(token string) (string, bool) {
	claims, ok := decodeJWTClaims(token)
	if !ok {
		return "", false
	}
	for _, key := range []string{"organization_uuid", "account_id", "sub"} {
		if v, ok := claims[key]; ok {
			if s, ok := v.(string); ok {
				s = strings.TrimSpace(s)
				if s != "" {
					return s, true
				}
			}
		}
	}
	return "", false
}
