# 模型分组目录（对标 PackyAPI 27+ 分组）

## 背景与目的

PackyAPI 的核心商业差异化是**精细的模型分组体系**：同一个模型（如 Claude Sonnet）按来源不同切成多个分组（CC 官方原价、AWS 8 折、AWS-Q 5 折），让用户按预算和稳定性自选。这是 OmniRouter 商业化的核心运营动作。

new-api 后台已经支持模型分组（`group_ratio` + 渠道分组），所以**不需要写代码，只需要运营配置**。本文档给出 28 个建议分组的完整清单 + 倍率建议 + 直接可粘贴的 JSON 配置。

> 倍率含义见 plan 第 11.1.1 节。简化记法：
> - `model_ratio`：跟模型走（输入价格 / $0.002）
> - `group_ratio`：跟用户分组走（折扣系数）
> - `cache_ratio`：缓存命中折扣
> - `completion_ratio`：补全相对 prompt 的价格倍数

## 分组矩阵（28 组）

### Claude 系列（5 组）

| 分组 key | 显示名 | 上游 | model_ratio (Sonnet 4) | group_ratio | cache_ratio | 适用场景 |
|---|---|---|---|---|---|---|
| `claude-cc` | Claude 官方 | Anthropic Direct | 1.5 | **1.0** | 0.1 | 官网原价，最稳，给重度用户 |
| `claude-aws` | Claude on AWS Bedrock | AWS Bedrock | 1.5 | **0.85** | 0.1 | 8.5 折，稳定 |
| `claude-aws-q` | Claude AWS Q（低成本） | AWS Bedrock 自建账号 | 1.5 | **0.5** | 0.1 | 5 折，可能偶发 422 |
| `claude-vertex` | Claude on Vertex | GCP Vertex AI | 1.5 | **0.85** | 0.1 | 8.5 折，谷歌生态 |
| `claude-cn` | Claude 国产线（试验） | 第三方代理 | 1.5 | **0.6** | 0.1 | 6 折，仅做实验性接入 |

### OpenAI 系列（6 组）

| 分组 key | 显示名 | 上游 | 默认 model_ratio | group_ratio | cache_ratio | 适用 |
|---|---|---|---|---|---|---|
| `openai-official` | OpenAI 官方 | OpenAI | (按官网价) | **1.0** | 0.25 | 官网原价 |
| `azure-openai` | Azure OpenAI | Azure | (按官网价) | **0.95** | 0.25 | 95 折，企业稳定 |
| `azure-openai-cache5min` | Azure 5 分钟缓存 | Azure | (按官网价) | **0.85** | 0.0 | 缓存命中 0 元，未命中 8.5 折 |
| `openai-codex` | Codex 专用 | OpenAI / 第三方 Codex Proxy | 按 Codex | **0.9** | 0.25 | 9 折，给 Codex CLI 用户 |
| `openrouter` | OpenRouter 聚合 | OpenRouter | 按 OR 价 | **1.05** | 0.25 | OR 加 5% 服务费 |
| `openai-cn` | OpenAI 国产线 | 第三方代理 | 按官网 | **0.7** | 0.25 | 7 折 |

### Google 系列（3 组）

| 分组 key | 显示名 | 上游 | model_ratio (2.5 Pro) | group_ratio | cache_ratio | 适用 |
|---|---|---|---|---|---|---|
| `gemini-aistudio` | Gemini AI Studio | Google AI Studio | 0.625 | **1.0** | 0.25 | 官方 Gemini API |
| `gemini-vertex` | Gemini on Vertex | GCP Vertex | 0.625 | **0.95** | 0.25 | 企业版 |
| `gemini-antigravity` | Antigravity（Gemini IDE） | 第三方/官方 | 0.625 | **0.5** | 0.25 | 5 折，主打开发者 |

### 国产模型（8 组）

| 分组 key | 显示名 | 上游 | 备注 model_ratio 范围 | group_ratio | 适用 |
|---|---|---|---|---|---|
| `deepseek-official` | DeepSeek 官方 | DeepSeek | 0.07 ~ 0.14（极低） | **1.0** | 性价比之王 |
| `deepseek-siliconflow` | DeepSeek via SiliconFlow | SiliconFlow | 同 | **0.95** | 95 折备线 |
| `qwen-dashscope` | 通义千问 | 阿里云 DashScope | 0.4 ~ 1.5 | **1.0** | 国内合规 |
| `glm-zhipu` | 智谱 GLM | 智谱清言 | 0.5 ~ 5 | **1.0** | 国内合规 |
| `kimi-moonshot` | Kimi (Moonshot) | Moonshot | 0.6 ~ 6 | **1.0** | 长上下文 |
| `ernie-baidu` | 文心 ERNIE | 百度千帆 | 0.4 ~ 2 | **1.0** | 国内合规 |
| `spark-iflytek` | 讯飞星火 | 科大讯飞 | 0.5 ~ 3 | **1.0** | 国内合规 |
| `doubao-bytedance` | 豆包 | 火山方舟 | 0.4 ~ 2 | **1.0** | 国内合规 |

### 海外其它（2 组）

| 分组 key | 显示名 | 上游 | model_ratio | group_ratio | 适用 |
|---|---|---|---|---|---|
| `xai-grok` | xAI Grok | xAI | 1.5 ~ 5 | **1.0** | Grok 4 |
| `mistral-cohere` | Mistral / Cohere 聚合 | Mistral / Cohere | 0.5 ~ 4 | **1.0** | 欧洲模型 |

### 多模态生成（4 组，走 task channel）

| 分组 key | 显示名 | 上游 | 计费方式 | 适用 |
|---|---|---|---|---|
| `image-gen` | 图像生成（DALL-E / FLUX / Midjourney） | OpenAI / Replicate / Midjourney | 按张计费（用 billingexpr `img` 变量） | 图片生成 |
| `video-gen` | 视频生成（Sora / Kling / Hailuo / Vidu） | OpenAI / 快手 / Minimax / Vidu | 按秒/张计费 | 视频生成 |
| `audio-gen` | 音乐 / TTS / STT（Suno / OpenAI Audio） | Suno / OpenAI | 按秒计费 | 音频 |
| `embedding-rerank` | Embedding & Rerank | OpenAI / Jina / Cohere | 按 token 计费 | 向量化 |

**合计：5 + 6 + 3 + 8 + 2 + 4 = 28 个分组** ✅ 超过 PackyAPI 27 个

## 配置入口

new-api 后台路径：
- **管理员 → 设置 → 倍率设置 → 分组倍率**：填 group_ratio（JSON 格式）
- **管理员 → 渠道**：每条渠道编辑里"分组"字段填上述 group key（如 `claude-cc`）。同一个 key 可关联多条渠道做负载均衡
- **管理员 → 设置 → 倍率设置 → 模型倍率**：模型级 model_ratio（按 Anthropic / OpenAI 公开价填）
- **管理员 → 设置 → 倍率设置 → 缓存倍率**：cache_ratio JSON

## 分组倍率 JSON（直接粘贴到后台）

```json
{
  "default":           1.0,
  "vip":               0.85,
  "svip":              0.7,

  "claude-cc":         1.0,
  "claude-aws":        0.85,
  "claude-aws-q":      0.5,
  "claude-vertex":     0.85,
  "claude-cn":         0.6,

  "openai-official":           1.0,
  "azure-openai":              0.95,
  "azure-openai-cache5min":    0.85,
  "openai-codex":              0.9,
  "openrouter":                1.05,
  "openai-cn":                 0.7,

  "gemini-aistudio":     1.0,
  "gemini-vertex":       0.95,
  "gemini-antigravity":  0.5,

  "deepseek-official":     1.0,
  "deepseek-siliconflow":  0.95,
  "qwen-dashscope":        1.0,
  "glm-zhipu":             1.0,
  "kimi-moonshot":         1.0,
  "ernie-baidu":           1.0,
  "spark-iflytek":         1.0,
  "doubao-bytedance":      1.0,

  "xai-grok":          1.0,
  "mistral-cohere":    1.0,

  "image-gen":         1.0,
  "video-gen":         1.0,
  "audio-gen":         1.0,
  "embedding-rerank":  1.0
}
```

## 模型倍率 JSON 示例（Claude 系列）

按 Anthropic 官网定价（2026 Q2）换算：基准 = $0.002/1K tokens，所以 ratio = price_per_1k_input / 0.002。

```json
{
  "claude-3-5-sonnet-20241022":  1.5,
  "claude-3-5-haiku-20241022":   0.4,
  "claude-3-opus-20240229":      7.5,
  "claude-sonnet-4-20250514":    1.5,
  "claude-sonnet-4-5":           1.5,
  "claude-opus-4-20250514":      7.5,

  "gpt-4o":                      1.25,
  "gpt-4o-mini":                 0.075,
  "gpt-5":                       6.25,
  "o1":                          7.5,
  "o3":                          1.0,

  "gemini-2.5-pro":              0.625,
  "gemini-2.5-flash":            0.0375,
  "gemini-2.0-flash":            0.0375,

  "deepseek-chat":               0.07,
  "deepseek-reasoner":           0.275
}
```

## 补全倍率（completion_ratio）建议

| 模型族 | completion_ratio | 含义 |
|---|---|---|
| GPT 系列（含 4o, 5, o1, o3） | 4.0 | OpenAI 官方比例 |
| Claude 系列 | 5.0 | Anthropic 官方比例 |
| Gemini 系列 | 4.0 | Google 默认 |
| DeepSeek | 4.0 | 官方 |
| 其它国产 | 1.0 ~ 4.0 | 按各家官网 |

## 缓存倍率（cache_ratio）建议

| 模型族 | cache_ratio | 含义 |
|---|---|---|
| OpenAI（cached input） | 0.25 | 命中时只收 25% 价格 |
| Anthropic（cache read） | 0.10 | 命中时只收 10% 价格 |
| Gemini（cached content） | 0.25 | 命中时 25% |
| Azure OpenAI（5 分钟缓存） | 0.0 | 命中免费（已经在 group_ratio 上扣过了） |

## 真实成本测算（举例）

### 案例 1：Claude Code 用户（claude-aws-q 分组）

调用 Claude Sonnet 4 一次（1000 prompt + 500 completion，无缓存）：
```
消耗 = (1000 + 500 × 5.0) × 1.5 × 0.5
     = 3500 × 0.75
     = 2625 quota
     = 2625 / 500000 = $0.00525 USD
按 1 元 = 1 美元额度（PackyAPI 风格） = ¥0.00525
```

### 案例 2：Codex 缓存命中场景（azure-openai-cache5min）

调用 GPT-5 一次（5000 prompt 全部命中缓存 + 500 completion）：
```
prompt 部分（缓存命中） = 5000 × 6.25 × 0    × 0.85 = 0
completion 部分          = 500  × 6.25 × 4   × 0.85 = 10625 quota
合计 = 10625 quota = $0.02125
对比 OpenAI 官网原价 = (5000 × 6.25 + 500 × 6.25 × 4) × 1.0 = 43750 quota = $0.0875
节省 75% — 这就是 PackyAPI Codex 包月能跑 1500-2500 次任务的底气
```

## 渠道（Channel）配置建议

每个分组建议配 **2-3 条渠道** 做冗余：
- 主渠道：`priority=10`，正常承载流量
- 备渠道：`priority=5`，主渠道挂了自动切
- 灰度渠道：`priority=1`，用 weight 分极小流量做新模型探活

每条渠道开 `auto_ban=true`，配合后续要做的"渠道自动禁用规则明确化"（plan 第 11.3 阶段二第 14 项）。

## 价格策略对标 PackyAPI

PackyAPI 公开规则：
- **充值比例**：1 元人民币 = 1 美元额度
- **基准价**：对标 Claude / OpenAI 官网（即 group_ratio = 1.0）
- **最低折扣**：5 折（即某分组 group_ratio = 0.5）
- **新人福利**：注册送 $1 + 首充 9 折

OmniRouter 建议沿用，差异化在：
- **国产模型聚合**：把 8 个国产分组都做齐，PackyAPI 只主打海外
- **企业团队版**：见 plan 第 11.5 节"错位竞争方向"

## 测试清单（运营验收）

部署完成后，每个分组至少跑一次端到端：
1. 创建测试 Token 绑定该分组
2. 用 curl 或 Claude Code/Codex 实际调用一次
3. 看后台日志：扣费金额是否符合 model_ratio × group_ratio × token 计算
4. 看渠道页：流量是否正确路由到目标渠道
5. 模拟主渠道 down，看是否切到备渠道

## 相关文件

- `setting/ratio_setting/model_ratio.go` — 模型倍率（管理员后台 JSON 编辑）
- `setting/ratio_setting/group_ratio.go` — 分组倍率
- `setting/ratio_setting/cache_ratio.go` — 缓存倍率
- `model/channel.go` — 渠道模型（含 `Group` 字段）
- `pkg/billingexpr/expr.md` — 复杂计费走表达式系统

## 后续运营动作

- [ ] 跟进上游各模型涨价/降价，每月更新 model_ratio
- [ ] 监控 [Lark 告警](../observability/lark-notification.md)：渠道异常率高时自动切灰度
- [ ] 用 [Prometheus 指标](../observability/prometheus-metrics.md) 看每个分组的 RPM、错误率，做容量规划
- [ ] 模型广场公开页（plan 第 11.3 阶段二第 11 项）：把这张表渲染成访客可看的页面
