# 在 Gemini CLI 中使用 OmniRouter

> Gemini CLI 是 Google 官方推出的 Gemini 命令行工具。OmniRouter 提供 Gemini API 兼容端点（`/v1beta`），国内用户无需翻墙即可使用 Gemini 2.5 Pro / Flash。

## 准备工作

1. 注册 [omnirouter.org](https://omnirouter.org) → 新人送 $1
2. 创建 API Key，分组选：
   - `gemini-aistudio`（原价）
   - `gemini-vertex`（8.5 折，企业级）
   - **`gemini-antigravity`（5 折，推荐性价比）**
3. 装 Gemini CLI：`npm install -g @google/gemini-cli`（或参考 [官方仓库](https://github.com/google-gemini/gemini-cli)）

## 快速开始

### macOS / Linux

```bash
export GEMINI_API_KEY="sk-你的-omnirouter-key"
export GEMINI_BASE_URL="https://omnirouter.org/v1beta"   # 注意 /v1beta 而非 /v1

gemini "解释一下这段代码"
```

写到 `~/.zshrc` / `~/.bashrc` 持久化。

### Windows

```powershell
$env:GEMINI_API_KEY  = "sk-你的-key"
$env:GEMINI_BASE_URL = "https://omnirouter.org/v1beta"
gemini "..."
```

### 配置文件方式

`~/.gemini/settings.json`：
```json
{
  "apiKey": "sk-你的-key",
  "baseUrl": "https://omnirouter.org/v1beta",
  "model": "gemini-2.5-pro"
}
```

## 推荐模型

| 模型 | 输入价 | 输出价 | OmniRouter 5 折后 | 适用 |
|---|---|---|---|---|
| `gemini-2.5-pro` | $1.25/1M | $10/1M | $0.625 / $5 | 复杂推理 |
| `gemini-2.5-flash` | $0.075/1M | $0.30/1M | $0.0375 / $0.15 | 高频小请求 |
| `gemini-2.0-flash` | $0.075/1M | $0.30/1M | $0.0375 / $0.15 | 兼容老应用 |

切换：`gemini --model gemini-2.5-flash "..."`

## 验证

```bash
gemini "Hello, who are you?"
```

正常应返回 Gemini 自我介绍。

## 高级用法

### 多模态输入（图像）

```bash
gemini --image ./screenshot.png "这张图里有什么？"
```

OmniRouter 完整支持 Gemini 的 inline_data 字段（图像、PDF、音频）。

### 长上下文（1M token）

Gemini 2.5 Pro 支持 1M token 上下文。OmniRouter 透传无截断：

```bash
gemini --file ./big-codebase.md "总结这个代码库的架构"
```

### Function Calling

```bash
gemini --tools weather.json "今天东京天气如何？"
```

OmniRouter 透传 tool 定义到上游 Gemini。

## 缓存

Gemini 的 `cached_content` 机制 OmniRouter 完整支持，命中部分按 cache_ratio=0.25 计费（即 25% 价格）。
长上下文场景节省成本明显。

## 常见问题

### Q: `GEMINI_BASE_URL` 怎么填，要不要带 `/v1beta`？
A: **必须带 `/v1beta`**（`https://omnirouter.org/v1beta`）。Gemini 官方 API 路径就是这个版本。

### Q: 报错 `404 Not Found`
A: 多半是 base URL 漏了 `/v1beta` 或者多写了 `/`。

### Q: 报错 `model not found`
A: 你的 API Key 当前分组不含该模型。后台 → API Keys → 编辑 → 改分组（推荐 `gemini-antigravity` 全模型可用）。

### Q: 流式返回有时断开？
A: 网络问题，可在 base URL 前加你的反代。OmniRouter 自身的流式 99.5%+ 可用率（见 [omnirouter.org/status](https://omnirouter.org/status)）。

### Q: 跟官方价对比？
A: `gemini-aistudio` 跟官方完全一致；`gemini-antigravity` 5 折是 OmniRouter 自有低成本通道。

### Q: 能用 OpenAI 协议调 Gemini 吗？
A: 可以。Gemini 也有 OpenAI 兼容端点：
```
OPENAI_BASE_URL=https://omnirouter.org/v1beta/openai
OPENAI_API_KEY=sk-你的-key
```
然后用任意 OpenAI SDK 调用，模型名填 `gemini-2.5-pro`。

## 配套工具

- [Gemini CLI 官方仓库](https://github.com/google-gemini/gemini-cli)
- [Gemini API 官方文档](https://ai.google.dev/gemini-api/docs)
- [CC-Switch](https://github.com/Calcium-Ion/cc-switch) — 多 CLI 统一管理

## 联系

- 站内工单：[omnirouter.org/support](https://omnirouter.org/support)
- Lark/飞书群：站点 Footer
- Telegram：@OmniRouter
- 邮件：hi@omnirouter.org
