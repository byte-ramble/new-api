# 客户端接入文档

OmniRouter 提供 OpenAI / Anthropic / Gemini 三种协议兼容端点，几乎所有主流 AI 客户端都能直接对接。

## 端点速查

| 协议 | base URL | 用法举例 |
|---|---|---|
| OpenAI | `https://omnirouter.org/v1` | GPT、Codex CLI、Cursor |
| Anthropic | `https://omnirouter.org` | Claude Code（注意没有 `/v1`） |
| Gemini | `https://omnirouter.org/v1beta` | Gemini CLI |
| Gemini (OpenAI 兼容) | `https://omnirouter.org/v1beta/openai` | 想用 OpenAI SDK 调 Gemini |

## 分客户端教程

| 客户端 | 文档 | 推荐分组 |
|---|---|---|
| Claude Code | [claude-code.md](./claude-code.md) | `claude-cc` / `claude-aws-q` |
| Codex CLI | [codex-cli.md](./codex-cli.md) | `openai-codex`（含包月） |
| Gemini CLI | [gemini-cli.md](./gemini-cli.md) | `gemini-antigravity`（5 折） |
| Cursor | [cursor.md](./cursor.md) | `openai-official` / `claude-aws` |

## 通用提示

### API Key 获取

登录 [omnirouter.org](https://omnirouter.org) → 右上角头像 → API Keys → 新建。

### 分组选择决策树

```
是 Claude Code 官方客户端？
  └─ 是 → claude-cc（最稳）或 claude-aws-q（5 折）
  └─ 否 → 是第三方 IDE / 自研工具？
       └─ 是 → claude-aws（不要选 claude-cc，会被识别限流）
       └─ 否 → 按厂商选对应分组
```

### 故障自查

1. **401 Unauthorized**：Key 错了，或 base URL 不对
2. **404 model not found**：当前分组不含目标模型
3. **429 Too Many Requests**：触发了限速，去后台看 Token 设置
4. **5xx**：上游问题，OmniRouter 会自动 fallback 到备用渠道；持续异常请提工单

### 日志查询

每次调用都在 OmniRouter 后台 → 日志 里有详细记录（请求时间、模型、token 数、消耗、状态）。

## 想接入未列出的客户端？

90% 的 AI 客户端都支持自定义 API endpoint，只要它支持 OpenAI 协议或 Anthropic 协议即可对接 OmniRouter。
通常做法：找客户端的 "Custom API" 或 "Base URL" 配置，填上面的端点 + 你的 OmniRouter Key 即可。

如果对接遇到问题，欢迎提工单：[omnirouter.org/support](https://omnirouter.org/support)
