# 在 Cursor 中使用 OmniRouter

> [Cursor](https://cursor.com) 是当前最流行的 AI 优先 IDE。OmniRouter 提供 OpenAI / Anthropic 兼容端点，可以让 Cursor 用上 Claude / GPT / Gemini 等多家模型，同时享受国内访问 + 人民币付费。

## 准备工作

1. 注册 [omnirouter.org](https://omnirouter.org)
2. 创建 API Key，**强烈推荐分组选 `openai-official` 或 `claude-aws`**（Cursor 调用频率高，建议用稳定分组）
3. 装 Cursor（[官网下载](https://cursor.com/download)）

> ⚠️ **不要选 `claude-cc` 分组**：CC 分组仅限 Claude Code 官方客户端，第三方 IDE 调用会被识别并限流。

## 快速开始

打开 Cursor → `Cmd/Ctrl + ,` 打开 Settings → **Models** 一栏：

### 配 OpenAI 兼容端点（GPT 系列）

```
OpenAI API Key:     sk-你的-omnirouter-key
Override OpenAI Base URL: https://omnirouter.org/v1
```

勾选要启用的模型（GPT-4o, GPT-5, o1 等）→ Verify。

### 配 Anthropic 兼容端点（Claude 系列）

Cursor 没有原生 Anthropic 配置框，需要走 OpenAI 兼容协议：
- 在 OmniRouter 创建另一个 Key，分组选 `claude-aws`
- 配置同上，模型名直接填 `claude-sonnet-4-5` 等
- OmniRouter 自动做协议转换（OpenAI Chat → Anthropic Messages）

### 配 Gemini

同上方法，base URL 用 `https://omnirouter.org/v1beta/openai`（Gemini 的 OpenAI 兼容端点），模型填 `gemini-2.5-pro`。

## 验证

Cursor 设置页 → Verify 按钮 → 应显示 ✓ Verified。
然后 `Cmd/Ctrl + L` 打开 chat，问一个简单问题，能正常回答即成功。

## 推荐分组矩阵

| 用途 | 分组 | 月均成本估算（中度使用） |
|---|---|---|
| Cursor Chat 主力 | `openai-official` 或 `claude-aws` | ¥80-200 |
| Cursor Tab 自动补全 | `gemini-antigravity`（5 折，速度快） | ¥30-80 |
| 偶尔重型推理 | `openai-official`（o1/o3） | 按使用 |

## Cursor 特殊配置

### Cursor Tab（自动补全）

Cursor Tab 调用频率非常高（每按一下键可能就发请求）。建议：
- 用低成本分组（`gemini-antigravity` 或 `deepseek-official`）
- 在 OmniRouter 后台为该 Key 单独设限速（避免意外烧钱）

### Cursor Composer（多文件改写）

Composer 上下文很大（几十 K token），适合用支持 prompt caching 的模型：
- `claude-aws-q` + Anthropic 缓存（命中 10%）
- `azure-openai-cache5min`（5 分钟内重复请求免费）

### MCP / 扩展

Cursor 的 MCP 调用如果走自定义 model provider，配同样的 base URL 即可。

## 常见问题

### Q: Verify 失败，提示 401
A: API Key 错了，或者 base URL 后面多/少 `/v1`。OpenAI 兼容路径必须是 `/v1`，Gemini 兼容是 `/v1beta/openai`。

### Q: Verify 通过但实际调用 404 model
A: 你的 API Key 当前分组不包含目标模型。后台 → API Keys → 改分组。

### Q: Cursor Tab 烧钱太快
A: 上面说过 — 用 `deepseek-official`（model_ratio=0.07，几乎免费）+ 在后台设单 Key RPM 限制。

### Q: 用 Claude 时报 `tool_use` 格式错误
A: OmniRouter 做了 OpenAI ↔ Anthropic 协议转换，但极少数复杂 tool_use 场景可能失败。建议直接用 Anthropic 协议（如果 Cursor 版本支持）或换 `gpt-5` 等 OpenAI 原生 tool use 模型。

### Q: Cursor 限制 base URL 一定要 https？
A: 是的。OmniRouter 默认 https://omnirouter.org，开箱即用。

### Q: 能不能用 Cursor 的官方订阅 + OmniRouter 双轨？
A: 可以。Cursor 默认有官方订阅；OmniRouter 配置在 "Override" 选项里，开关切换即可。

## 与官方订阅对比

| 维度 | Cursor 官方订阅 ($20/月) | OmniRouter 按量 |
|---|---|---|
| 价格 | $20 固定 | 按用量，平均 $5-30 |
| 速度 | 国外节点，国内 200-1000ms | 国内节点 50-150ms |
| 模型选择 | Cursor 选好的几个 | OmniRouter 的全部 28+ 分组 |
| 上下文 | 有限制（pro 级别） | 模型原生上限（如 Gemini 1M） |
| 付款 | 美元卡 | 人民币 |

## 配套工具

- [Cursor 官方文档](https://docs.cursor.com)
- [CC-Switch](https://github.com/Calcium-Ion/cc-switch) — 切换不同 OmniRouter Key

## 联系

- 站内工单：[omnirouter.org/support](https://omnirouter.org/support)
- Lark/飞书群：站点 Footer
- Telegram：@OmniRouter
- 邮件：hi@omnirouter.org
