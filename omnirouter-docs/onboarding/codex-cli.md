# 在 Codex CLI 中使用 OmniRouter

> Codex CLI 是 OpenAI 官方推出的命令行编程助手。OmniRouter 提供 OpenAI 协议兼容端点，对 Codex 做了**激进的 prompt cache 优化**，按 PackyAPI 公开数据约 $60 额度可跑 1500-2500 次任务。

## 准备工作

1. 注册 [omnirouter.org](https://omnirouter.org)
2. 新人送 $1（约够 Codex 50-200 次小调用）
3. 创建 API Key，分组选 `openai-codex`（专为 Codex 优化的低成本通道）
4. 装 Codex CLI（参考 [OpenAI 官方文档](https://github.com/openai/codex)）

## 快速开始

### macOS / Linux

```bash
export OPENAI_BASE_URL="https://omnirouter.org/v1"
export OPENAI_API_KEY="sk-你的-omnirouter-key"

codex "把这段函数改成纯函数式风格"
```

或写到 `~/.zshrc` / `~/.bashrc` 做持久化。

### Windows (PowerShell)

```powershell
$env:OPENAI_BASE_URL = "https://omnirouter.org/v1"
$env:OPENAI_API_KEY  = "sk-你的-key"

codex "..."
```

### Codex 配置文件方式

如果你的 Codex CLI 支持 `~/.codex/config.toml`：

```toml
[provider]
name = "openai"
base_url = "https://omnirouter.org/v1"
api_key = "sk-你的-key"

[model]
default = "gpt-5"
```

## 推荐分组

| 分组 | 折扣 | 缓存策略 | 适用 |
|---|---|---|---|
| `openai-codex` | 9 折 | 激进缓存（命中 25%） | **推荐**，Codex CLI 默认选这个 |
| `azure-openai-cache5min` | 8.5 折 | 5 分钟内重复请求免费 | 长 session、反复迭代同一段代码 |
| `openai-official` | 原价 | 默认缓存 | 对稳定性要求最高 |

切换分组：OmniRouter 后台 → API Keys → 编辑 → 模型分组。

## Codex 包月套餐（性价比之选）

OmniRouter 提供 **Codex 包月** 套餐，对标 PackyAPI 的 ¥60/月：

- ¥60 / 月 / 1 枚激活码
- 月内不限次（公平使用条款见 [omnirouter.org/codex-monthly](https://omnirouter.org/codex-monthly)）
- 超出后按 ¥80/枚续购
- 适合日均 30+ 次 Codex 调用的开发者

订阅入口：登录 → 订阅 → 选择 Codex Monthly。

## 缓存优化实测

Codex CLI 的 system prompt + 项目上下文通常 8K-20K token，重复调用时几乎全部命中缓存：

```
首次请求（cache miss）：(15000 prompt + 500 completion × 4) × 6.25 × 0.9 = 95625 quota = $0.19
重复请求（cache hit）：(15000 × 0.25 + 500 × 4) × 6.25 × 0.9 = 32344 quota = $0.065
节省 66%
```

这就是为什么同样 ¥60 充值，OmniRouter 的 Codex 用法能跑出比官方原价多 3-5 倍的请求量。

## 验证

```bash
codex "echo hello world"
```

应该 1-3 秒内返回。如果报错：
- `401`：Key 不对
- `404 model not found`：当前分组不含目标模型，去后台改分组
- `429`：触发限流，看 OmniRouter 后台 → 用量

## 高级用法

### 流式输出

Codex CLI 默认走 SSE 流式。OmniRouter 完整支持，无需额外配置。

### Function Calling / Tool Use

Codex 的 tool use 在 OmniRouter 透传到上游。`openai-codex` 分组完整支持。

### 多模型切换

```bash
codex --model gpt-5 "..."
codex --model o3   "..."
codex --model claude-sonnet-4-5 "..."   # 跨厂商，自动走 OmniRouter 协议互转
```

跨厂商调用：你的 API Key 必须开启**跨分组**或多分组授权（后台配置）。

## 常见问题

### Q: 已订阅 Codex 包月，但同时想用按量付费的 Claude，怎么配？
A: 一个账号可以同时挂多个 API Key（不同 Key 对应不同分组）。订阅余额和按量余额在后台分开展示。

### Q: 包月被检测到滥用怎么办？
A: 包月有公平使用阈值（默认 24h 内 800 次任务）。超出会临时降级到按量计费、不会封号。详见 [omnirouter.org/codex-fair-use](https://omnirouter.org/codex-fair-use)。

### Q: 第三方 IDE 插件（Cursor / Cline）能用 Codex 包月吗？
A: 不能。包月仅限 Codex CLI。其它客户端用按量分组（参考 [Cursor 接入文档](./cursor.md)）。

### Q: 速度对比官方？
A: 国内 (北上广深) 平均 50-150ms 首字延迟（OmniRouter 在国内有节点）；官方 OpenAI 国内裸连 200-2000ms。

### Q: 包月支持退款吗？
A: 7 天内未消耗超过 50% 可申请退款；详见服务条款。

## 配套工具

- [OpenAI Codex CLI Docs](https://github.com/openai/codex)
- [CC-Switch](https://github.com/Calcium-Ion/cc-switch) — 同时管理 Codex / Claude Code / Gemini CLI 多 Key

## 联系

- 站内工单：[omnirouter.org/support](https://omnirouter.org/support)
- Lark/飞书群：站点 Footer 扫码
- Telegram：@OmniRouter
- 邮件：hi@omnirouter.org
