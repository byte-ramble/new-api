# Lark / 飞书告警通道

## 背景

new-api 自带的通知能力（`service/webhook.go`）是 PackyAPI 风格的**通用 webhook**：POST 一个固定 schema（`{type, title, content, values, timestamp}`）+ HMAC-SHA256 签名头。这套对接 PackyAPI / 自建 Worker 很合适，但**不能直接打到飞书机器人**——飞书要求的载荷格式完全不同（`msg_type` + `card` / `content`，加签算法也不同）。

OmniRouter 选用 Lark / 飞书替代 PackyAPI 在用的 QQ 群作为告警 + 客服触达通道（理由：Lark 有自定义机器人、企业用户多、富文本卡片体验好、跨平台）。

## 设计

### 文件结构

| 文件 | 职责 |
|---|---|
| `service/lark_notify.go` | Lark webhook 发送器：构造载荷、签名、SSRF 防护、HTTP 投递 |
| `service/system_alert.go` | 系统级告警分发器：业务代码统一入口，按运行时配置 fan-out 到各通道 |
| `service/lark_notify_test.go` | 单元测试：签名正确性、富卡片结构、签名场景、4xx 失败处理 |

### 调用链

```
业务代码（如 controller/channel-test.go::testChannel）
  └─> service.SendSystemAlert(dto.Notify)            ← 业务统一入口
        └─> 读 LARK_WEBHOOK_URL / LARK_WEBHOOK_SECRET
              └─> service.SendLarkNotify             ← Lark 专用 sender
                    ├─> genLarkSign (HMAC-SHA256)    ← 飞书加签
                    ├─> common.Marshal(LarkPayload)  ← 按 Rule 1 走统一 JSON 包装
                    ├─> common.ValidateURLWithFetchSetting  ← 复用 SSRF 防护
                    └─> service.GetHttpClient().Do(req)
```

### 核心方法

| 方法 | 输入 | 输出 | 副作用 |
|---|---|---|---|
| `service.SendSystemAlert(dto.Notify)` | 通知 DTO | 无（错误内部 swallow） | 读 env、调用 Lark |
| `service.SendLarkNotify(url, secret, dto.Notify)` | webhook URL、密钥、通知 DTO | error | 发 HTTP POST |
| `service.genLarkSign(secret, ts)` | 密钥、unix 秒 | base64 签名 | 无 |
| `service.templateForType(t)` | 通知类型常量 | Lark card 模板色 | 无 |

### LarkPayload schema

```go
type LarkPayload struct {
    Timestamp string                 `json:"timestamp,omitempty"`
    Sign      string                 `json:"sign,omitempty"`
    MsgType   string                 `json:"msg_type"`           // "interactive"
    Content   map[string]string      `json:"content,omitempty"`  // 当 msg_type=text 时使用
    Card      map[string]interface{} `json:"card,omitempty"`     // 富卡片
}
```

实际发送的 interactive card 结构：
```json
{
  "msg_type": "interactive",
  "timestamp": "1715000000",
  "sign": "base64(...)",
  "card": {
    "config": {"wide_screen_mode": true},
    "header": {
      "title": {"tag": "plain_text", "content": "Channel Test Failed"},
      "template": "yellow"
    },
    "elements": [
      {"tag": "div", "text": {"tag": "lark_md", "content": "channel openai-1 failed: timeout"}},
      {"tag": "hr"},
      {"tag": "note", "elements": [{"tag": "plain_text", "content": "type: channel_test · 2026-05-03 12:34:56"}]}
    ]
  }
}
```

### 模板色映射（card header.template）

| 通知类型 | 颜色 | 含义 |
|---|---|---|
| `quota_exceed` | orange | 配额耗尽，需关注但非紧急 |
| `channel_update` | blue | 渠道配置变更，信息性 |
| `channel_test` | yellow | 渠道健康检查异常 |
| 其它 | indigo | 默认 |

### 飞书加签算法（容易踩坑）

飞书的加签算法**反直觉**——`timestamp + "\n" + secret` 是 HMAC 的**密钥**，**消息是空字符串**：

```go
func genLarkSign(secret string, timestamp int64) string {
    stringToSign := strconv.FormatInt(timestamp, 10) + "\n" + secret
    mac := hmac.New(sha256.New, []byte(stringToSign))   // KEY = ts\nsecret
    return base64.StdEncoding.EncodeToString(mac.Sum(nil))  // MSG = empty
}
```

如果按"正常"思路把 `secret` 当 key、`timestamp` 当 message，飞书会返回 `sign 验证失败`。

## 配置

环境变量（MVP 阶段；后续可迁到 system_setting 让管理员后台配置）：

| 变量 | 是否必需 | 说明 |
|---|---|---|
| `LARK_WEBHOOK_URL` | 必需才启用 | 飞书自定义机器人完整 URL |
| `LARK_WEBHOOK_SECRET` | 可选 | 仅当机器人启用"加签"校验时填 |

`docker-compose.yml` 添加示例：
```yaml
environment:
  - LARK_WEBHOOK_URL=https://open.feishu.cn/open-apis/bot/v2/hook/xxxxxxx
  - LARK_WEBHOOK_SECRET=xxxxxxxx   # 可选
```

## 业务接入

任何想触发告警的业务调用：
```go
import "github.com/QuantumNous/new-api/dto"
import "github.com/QuantumNous/new-api/service"

service.SendSystemAlert(dto.NewNotify(
    dto.NotifyTypeChannelTest,
    "Channel Test Failed",
    "channel %s failed: %s",
    []interface{}{ch.Name, errMsg},
))
```

如果未来加 Slack / 钉钉 / 企微 / 邮件，只需在 `service/system_alert.go::SendSystemAlert` 里追加 fan-out 分支，业务调用方零修改。

## 错误策略

- 告警发送失败 → `common.SysError` 打日志 + swallow（**绝不**让告警子系统的故障传播到业务路径）
- SSRF 防护命中 → 返回错误 + 日志，仍 swallow
- HTTP 非 2xx → 返回带状态码的错误，仍 swallow

## 测试

`service/lark_notify_test.go` 覆盖：
1. `TestGenLarkSign` — 签名结构正确、不同 secret/ts 产生不同结果
2. `TestSendLarkNotify_Success` — mock Lark 端点验证 card 结构
3. `TestSendLarkNotify_WithSign` — 带 secret 时 timestamp/sign 字段一致
4. `TestSendLarkNotify_Non2xx` — 4xx 响应被识别为失败

跑：
```bash
cd /Users/maning/Documents/dev/js-dev/new-api
go test ./service/ -run TestGenLarkSign -v
go test ./service/ -run TestSendLarkNotify -v
```

## 后续扩展（不在本次 MVP）

- [ ] 把 LARK_WEBHOOK_URL/SECRET 移到 `setting/system_setting/`，加管理面板 UI
- [ ] 把告警接到具体业务点：channel test 失败、auto-ban 触发、配额阈值告警
- [ ] 加告警频率限制（同一类型 5 分钟内只发一次，避免刷屏）
- [ ] 多通道（Slack / 钉钉 / 企微 / 邮件）
- [ ] 告警自动升级（5 分钟未恢复 → @所有人）
