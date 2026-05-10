package claude_code

const ChannelName = "claude_code"

// AnthropicBaseURL 是 Claude Code 订阅请求实际打的地址。
const AnthropicBaseURL = "https://api.anthropic.com"

// AnthropicMessagesPath Claude Messages API 路径。
const AnthropicMessagesPath = "/v1/messages"

// AnthropicVersionHeader 缺省 anthropic-version。
const AnthropicVersionHeader = "2023-06-01"

// AnthropicBetaOAuthHeader Claude Code 订阅必需的 anthropic-beta 值。
// 来源：抓包社区参考实现（gist: changjonathanc/9f9d635b2f8692e0520a884eaf098351）。
// 若 Anthropic 调整该值，仅需更新此常量重新编译。
const AnthropicBetaOAuthHeader = "oauth-2025-04-20"

// ModelList Claude Code 订阅当前可用的模型 ID。
// 调研日期 2026-05；后续随官方更新需同步调整。
var ModelList = []string{
	"claude-opus-4-7",
	"claude-sonnet-4-6",
	"claude-opus-4-6",
	"claude-opus-4-5-20251101",
	"claude-haiku-4-5-20251001",
	"claude-sonnet-4-5-20250929",
}
