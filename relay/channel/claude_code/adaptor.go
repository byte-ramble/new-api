package claude_code

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/relay/channel"
	"github.com/QuantumNous/new-api/relay/channel/claude"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/types"

	"github.com/gin-gonic/gin"
)

// Adaptor 是 Claude Code 订阅渠道的适配器。
// 与 codex 类似，凭据是 OAuth JSON 而非裸 API Key；
// 与 claude 类似，请求/响应转换沿用 Anthropic Messages 协议。
//
// 所以这里复用 claude 包里的请求/响应处理函数，仅重写认证 header 与请求 URL。
type Adaptor struct {
}

func (a *Adaptor) Init(info *relaycommon.RelayInfo) {
}

func (a *Adaptor) GetChannelName() string {
	return ChannelName
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

// GetRequestURL Claude Code 订阅必须打到 Anthropic 官方域名，忽略 Channel.BaseURL（除非用户显式覆盖）。
func (a *Adaptor) GetRequestURL(info *relaycommon.RelayInfo) (string, error) {
	base := strings.TrimRight(info.ChannelBaseUrl, "/")
	if base == "" {
		base = AnthropicBaseURL
	}
	return fmt.Sprintf("%s%s", base, AnthropicMessagesPath), nil
}

// SetupRequestHeader 用 OAuth Bearer + 必需的 anthropic-beta 替换 claude 适配器默认的 x-api-key。
func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Header, info *relaycommon.RelayInfo) error {
	channel.SetupApiRequestHeader(info, c, req)

	key := strings.TrimSpace(info.ApiKey)
	if !strings.HasPrefix(key, "{") {
		return errors.New("claude_code channel: key must be a JSON object")
	}

	oauthKey, err := ParseOAuthKey(key)
	if err != nil {
		return err
	}

	accessToken := strings.TrimSpace(oauthKey.AccessToken)
	if accessToken == "" {
		return errors.New("claude_code channel: access_token is required")
	}

	req.Set("Authorization", "Bearer "+accessToken)

	// anthropic-version：客户端可覆盖，默认值与 claude 适配器保持一致。
	anthropicVersion := c.Request.Header.Get("anthropic-version")
	if anthropicVersion == "" {
		anthropicVersion = AnthropicVersionHeader
	}
	req.Set("anthropic-version", anthropicVersion)

	// anthropic-beta：Claude Code 订阅 OAuth 必需。
	// 若客户端已带，附加在末尾（用 , 拼）；否则只用我们这一个。
	clientBeta := strings.TrimSpace(c.Request.Header.Get("anthropic-beta"))
	if clientBeta == "" {
		req.Set("anthropic-beta", AnthropicBetaOAuthHeader)
	} else if !strings.Contains(clientBeta, AnthropicBetaOAuthHeader) {
		req.Set("anthropic-beta", clientBeta+","+AnthropicBetaOAuthHeader)
	} else {
		req.Set("anthropic-beta", clientBeta)
	}

	req.Set("Content-Type", "application/json")
	if info.IsStream {
		req.Set("Accept", "text/event-stream")
	} else if req.Get("Accept") == "" {
		req.Set("Accept", "application/json")
	}

	// 显式确保不带 x-api-key（避免被上游优先识别）。
	req.Del("x-api-key")
	return nil
}

// ConvertClaudeRequest 透传，与 claude 适配器一致。
func (a *Adaptor) ConvertClaudeRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.ClaudeRequest) (any, error) {
	return request, nil
}

// ConvertOpenAIRequest 复用 claude 包的 OpenAI -> Claude Messages 转换。
func (a *Adaptor) ConvertOpenAIRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	return claude.RequestOpenAI2ClaudeMessage(c, *request)
}

func (a *Adaptor) ConvertGeminiRequest(*gin.Context, *relaycommon.RelayInfo, *dto.GeminiChatRequest) (any, error) {
	return nil, errors.New("claude_code channel: gemini endpoint not supported")
}

func (a *Adaptor) ConvertAudioRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.AudioRequest) (io.Reader, error) {
	return nil, errors.New("claude_code channel: audio endpoint not supported")
}

func (a *Adaptor) ConvertImageRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.ImageRequest) (any, error) {
	return nil, errors.New("claude_code channel: image endpoint not supported")
}

func (a *Adaptor) ConvertRerankRequest(c *gin.Context, relayMode int, request dto.RerankRequest) (any, error) {
	return nil, errors.New("claude_code channel: rerank endpoint not supported")
}

func (a *Adaptor) ConvertEmbeddingRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.EmbeddingRequest) (any, error) {
	return nil, errors.New("claude_code channel: embedding endpoint not supported")
}

func (a *Adaptor) ConvertOpenAIResponsesRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.OpenAIResponsesRequest) (any, error) {
	return nil, errors.New("claude_code channel: openai responses endpoint not supported")
}

func (a *Adaptor) DoRequest(c *gin.Context, info *relaycommon.RelayInfo, requestBody io.Reader) (any, error) {
	return channel.DoApiRequest(a, c, info, requestBody)
}

// DoResponse 复用 claude 包的 SSE / 非流式处理逻辑。
func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (usage any, err *types.NewAPIError) {
	info.FinalRequestRelayFormat = types.RelayFormatClaude
	if info.IsStream {
		return claude.ClaudeStreamHandler(c, resp, info)
	}
	return claude.ClaudeHandler(c, resp, info)
}
