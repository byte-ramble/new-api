# 在 Claude Code 中使用 OmniRouter

> Claude Code 是 Anthropic 官方推出的命令行 / IDE 编程助手。OmniRouter 提供 Anthropic 协议兼容的中转端点，国内开发者无需翻墙、无需海外卡，人民币付费即可使用 Claude Sonnet / Opus / Haiku。

## 准备工作

1. **注册 OmniRouter**：访问 [omnirouter.org](https://omnirouter.org) → 注册 → 邮箱验证
2. **领取新人福利**：注册后自动到账 **$1 余额**（约够 Claude Code 用 200-500 次小请求）
3. **充值**（可选）：首充 9 折，使用优惠码 `OMNIROUTER-FIRST`
4. **创建 API Key**：登录后点击右上角头像 → API Keys → 新建 Key
   - 推荐分组选 `claude-cc`（官方原价、最稳）或 `claude-aws-q`（5 折、性价比）
   - 把 Key 复制好，sk- 开头

## 快速开始

### macOS / Linux

把以下两行加进 `~/.zshrc` 或 `~/.bashrc`：

```bash
export ANTHROPIC_BASE_URL="https://omnirouter.org"
export ANTHROPIC_AUTH_TOKEN="sk-你的-omnirouter-key"
```

然后：
```bash
source ~/.zshrc          # 或 source ~/.bashrc
claude --version         # 验证 Claude Code 已装
claude                   # 启动交互模式
```

### Windows (PowerShell)

```powershell
[System.Environment]::SetEnvironmentVariable('ANTHROPIC_BASE_URL', 'https://omnirouter.org', 'User')
[System.Environment]::SetEnvironmentVariable('ANTHROPIC_AUTH_TOKEN', 'sk-你的-omnirouter-key', 'User')
```

重启 PowerShell 后生效。

### 临时使用（不写入 shell 配置）

```bash
ANTHROPIC_BASE_URL="https://omnirouter.org" \
ANTHROPIC_AUTH_TOKEN="sk-你的-key" \
claude
```

## 验证

进入 Claude Code 交互后输入：
```
> 你好，请介绍下你自己
```

正常应该返回 Claude 的回答。如果报 401/403，检查 Key 是否正确；如果报 503/timeout，检查网络。

## 模型选择

OmniRouter 后台的 Claude 系列分组：

| 分组 | 性价比 | 稳定性 | 推荐场景 |
|---|---|---|---|
| `claude-cc` | 原价 | ⭐⭐⭐⭐⭐ | 重度生产用、不能容忍偶发失败 |
| `claude-aws` | 8.5 折 | ⭐⭐⭐⭐ | 日常开发 |
| `claude-aws-q` | **5 折** | ⭐⭐⭐ | 实验、刷题、对成本敏感 |
| `claude-vertex` | 8.5 折 | ⭐⭐⭐⭐ | 谷歌生态 |

> 切换分组：在 OmniRouter 后台 → API Keys → 编辑 Key → 改"模型分组"。

## 高级配置

### 显式指定模型

```bash
# 在 Claude Code 中按 / 切换模型
> /model claude-sonnet-4-5
```

支持的模型清单见 [omnirouter.org/models](https://omnirouter.org/models)。

### 自定义模型映射

如果你想让 `claude-sonnet-4-5` 这个名字实际打到 `claude-3-5-sonnet-20241022`，在 OmniRouter 后台 → 模型映射里配置即可，本地 Claude Code 配置不变。

### 跨分组重试

OmniRouter 后台 → API Keys → 开启"跨分组重试"。当主分组（如 `claude-aws-q`）失败时自动 fallback 到备用分组（如 `claude-cc`），用户无感。

## Prompt Caching（节省成本必看）

Claude Code 默认会用 Anthropic 的 prompt caching 机制。OmniRouter 完整支持，并按 cache 命中给折扣（cache_ratio=0.1，命中部分只收 10% 价格）。

- 首次发送大段上下文：按全价计费（cache create）
- 后续 5 分钟内重复相同上下文：仅按 10% 计费（cache read）

实际效果：长 session 的 Claude Code 调用，平均成本能比"无缓存中转"低 40-70%。

## 常见问题

### Q: 报错 `invalid x-api-key`
A: 检查 `ANTHROPIC_AUTH_TOKEN` 而不是 `ANTHROPIC_API_KEY`（注意 Claude Code 用的是 AUTH_TOKEN）。

### Q: 速度比官方慢？
A: OmniRouter 多线路自动选最优。如果仍慢，可在 API Key 配置里指定固定分组绕过。

### Q: 能不能接 Cline / OpenCode 等第三方客户端？
A: `claude-cc` 分组**禁止**第三方客户端（合规要求），命中会限流。可以用 `claude-aws` 或 `claude-aws-q` 分组绕过。详见 [防滥用规则](../operations/anti-abuse.md)（Phase 2 文档）。

### Q: 多少钱？
A: 跟 Anthropic 官网一致；最低分组 5 折。1 元 RMB = 1 美元额度。Claude Sonnet 4 输入 $3/1M token、输出 $15/1M token，5 折分组就是 $1.5 / $7.5。

### Q: Claude Code 怎么装？
A: `npm install -g @anthropic-ai/claude-code`（需要 Node.js 18+），或参考 [Anthropic 官方文档](https://docs.anthropic.com/en/docs/agents-and-tools/claude-code/overview)。

### Q: 能用 .env 文件吗？
A: 可以。在项目根目录新建 `.env`：
```
ANTHROPIC_BASE_URL=https://omnirouter.org
ANTHROPIC_AUTH_TOKEN=sk-...
```
但记得加进 `.gitignore`，避免泄露 Key。

## 配套工具推荐

- [CC-Switch](https://github.com/Calcium-Ion/cc-switch) — 桌面 GUI，多 Key/多中转一键切换
- [Claude Code Docs](https://docs.anthropic.com/en/docs/agents-and-tools/claude-code) — 官方完整文档

## 联系我们

- 站内工单：[omnirouter.org/support](https://omnirouter.org/support)
- Lark / 飞书群：扫描站点 Footer 二维码
- Telegram 频道：@OmniRouter（公告 + 故障播报）
- Issue 反馈：邮件 hi@omnirouter.org
