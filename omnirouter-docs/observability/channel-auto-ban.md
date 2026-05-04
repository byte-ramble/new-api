# 渠道自动禁用增强（Channel Auto-Ban）

## 背景

new-api 上游已经有完善的"自动禁用 + 自动恢复"基础设施（[`service/channel.go`](../../service/channel.go)、[`controller/channel-test.go::AutomaticallyTestChannels`](../../controller/channel-test.go)）。但有 4 个对生产稳定性 / 可观测性影响明显的缺口：

| 上游现状 | 缺口 |
|---|---|
| ✅ `service.DisableChannel` 改 DB 状态 + 站内通知 | ❌ **没有 Lark 告警** —— 运维不知道 |
| ✅ `service.ShouldDisableChannel` 按 status code / 关键词决策 | ❌ **单次触发** —— 一次 5xx blip 就禁用，假阳性高 |
| ✅ `controller.AutomaticallyTestChannels` 定时探活恢复 | ❌ **无 Prometheus 指标** —— 频率不可观测 |
| ✅ 通用错误判断 | ❌ **错误类型不分** —— 5xx / auth / rate-limit 一视同仁 |

## OmniRouter 增量

只补缺口、**不重写上游**。新增 + 修改文件：

| 文件 | 性质 | 用途 |
|---|---|---|
| `service/channel_health.go`（新） | 库 | 错误分类 + 滑动窗口去抖 |
| `service/channel_health_test.go`（新） | 测试 | 6 个单元测试，覆盖分类 + 计数 + reset |
| `service/channel_metrics.go`（新） | 库 | Prometheus counter（避开 import cycle 放 service 包） |
| `service/channel.go`（改） | 业务 | DisableChannel/EnableChannel 增加 Lark + Prom + reset 调用 |
| `controller/relay.go`（改） | 业务 | processChannelError 增加 burst 去抖 |
| `setting/operation_setting/monitor_setting.go`（改） | 配置 | 新增 `DisableBurstThreshold` + `DisableBurstWindowSec` 字段 + env 覆盖 |
| `middleware/metrics.go`（改） | 注释 | 仅留指针，channel 计数器搬到 service 包 |

## 错误分类

低基数 reason bucket，用于 Prom 标签和 Lark 告警分组：

| `ChannelErrorReason` | HTTP code | 示例 |
|---|---|---|
| `auth_failed` | 401 / 403 | API key 失效、过期 |
| `rate_limited` | 429 | 上游限流 |
| `upstream_5xx` | 500-599 | 上游服务 down / 内部错误 |
| `network` | 0（无 HTTP 响应）| DNS 失败 / connect refused / TLS 握手失败 |
| `quota_exhausted` | 余额 / 配额耗尽（业务层） | 上游账单到期 |
| `manual` | — | admin 手动操作 |
| `other` | 其它 | 兜底 |

## 滑动窗口去抖

### 配置

```go
type MonitorSetting struct {
    AutoTestChannelEnabled bool
    AutoTestChannelMinutes float64

    DisableBurstThreshold int  // ≥N errors in window → disable; 0 = legacy single-error trigger
    DisableBurstWindowSec int  // window size in seconds
}
```

环境变量覆盖（也能 admin UI 改）：
```
DISABLE_BURST_THRESHOLD=5      # 默认 0（legacy）
DISABLE_BURST_WINDOW_SEC=60    # 默认 60
```

### 推荐生产配置

| 场景 | threshold | window | 效果 |
|---|---|---|---|
| 默认（向后兼容） | 0 | 60s | 单次 → 立即禁用（legacy） |
| 推荐生产 | 5 | 60s | 1 分钟内 5 次 → 禁用 |
| 保守（误杀少） | 10 | 120s | 2 分钟内 10 次 → 禁用 |
| 激进（响应快） | 3 | 30s | 30 秒内 3 次 → 禁用 |

### 数据结构

```go
var (
    channelHealthMu      sync.Mutex
    channelHealthByID    = make(map[int]*channelErrorRing)
)

type channelErrorRing struct {
    timestamps []time.Time   // 仅保留 window 内的，每次 record 自动 prune
}
```

简单 in-memory map + mutex。多副本部署时每个实例独立计数（不共享窗口），但 DisableChannel 是幂等 DB 写，所以不影响正确性，只是统计窗口分散。

### 调用链

```
controller/relay.go::processChannelError
  ↓ if ShouldDisableChannel(err) && AutoBan
  ↓
  service.ClassifyChannelError(err)         → reasonClass
  service.RecordChannelError(channelID, …)  → returns count_in_window
  ↓
  if threshold > 0 && count < threshold:
      common.SysLog("debounced, not disabling yet")
  else:
      gopool.Go(service.DisableChannel(channelError, msg, reasonClass))
                                ↓
                                model.UpdateChannelStatus  (existing)
                                NotifyRootUser (existing 站内消息)
                                ───── OmniRouter additions ─────
                                service.RecordChannelAutoDisabled(...) → Prom
                                service.ResetChannelHealth(channelID)
                                service.SendSystemAlert(...) → Lark webhook
```

## Lark 告警

每次禁用 + 恢复都发飞书卡片（`service.SendSystemAlert` → `service.SendLarkNotify` → 飞书 interactive card）。

样例：

**禁用告警** （orange template）
```
[Header] 通道「openai-aws-q」（#42）已被禁用
[Body]   Channel `openai-aws-q` (#42) auto-disabled.
         **Reason class:** `upstream_5xx`
         **Detail:** 503 Service Unavailable from https://...
[Footer] type: channel_update · 2026-05-04 14:23:01
```

**恢复告警** （blue template）
```
[Header] 通道「openai-aws-q」（#42）已被启用
[Body]   Channel `openai-aws-q` (#42) recovered and re-enabled.
[Footer] type: channel_update · 2026-05-04 14:35:12
```

启用 Lark 告警：在容器 env 设 `LARK_WEBHOOK_URL`（详见 [lark-notification.md](./lark-notification.md)）。

## Prometheus 指标

新增两个 counter（位于 `service/channel_metrics.go`）：

```
omnirouter_channel_auto_disabled_total{channel="<name>",reason="<class>"}
omnirouter_channel_recovered_total{channel="<name>"}
```

PromQL 范例：

```promql
# 各 channel 禁用频率（1h）
sum by (channel) (rate(omnirouter_channel_auto_disabled_total[1h]))

# 哪种 reason 最多
sum by (reason) (increase(omnirouter_channel_auto_disabled_total[24h]))

# 某通道 flap 频率（disable + recover 加起来）
rate(omnirouter_channel_auto_disabled_total{channel="cc"}[1h])
+ rate(omnirouter_channel_recovered_total{channel="cc"}[1h])
```

## 测试

`service/channel_health_test.go` 6 个测试全 PASS：

- `TestClassifyChannelError` — 9 个 status code → reason 映射
- `TestRecordChannelError_Disabled` — threshold=0 时 short-circuit
- `TestRecordChannelError_Counts` — 累加到 7
- `TestHasReachedBurstThreshold` — 阈值边界
- `TestResetChannelHealth` — reset 后清零
- `TestRecordChannelError_DisabledThresholdReturnsZero` — threshold=0 时 HasReached 始终 true（向后兼容）

```bash
go test ./service/ -run "TestClassifyChannelError|TestRecordChannelError|TestHasReachedBurstThreshold|TestResetChannelHealth" -v
```

## 上线 checklist

- [ ] env 设 `DISABLE_BURST_THRESHOLD=5`、`DISABLE_BURST_WINDOW_SEC=60`
- [ ] env 设 `LARK_WEBHOOK_URL=https://open.feishu.cn/open-apis/bot/v2/hook/xxx`
- [ ] env 设 `LARK_WEBHOOK_SECRET=xxx`（如启用加签）
- [ ] env 设 `CHANNEL_TEST_FREQUENCY=10`（自动恢复探活间隔，分钟）
- [ ] Prometheus 抓取 /metrics（确认能看到 omnirouter_channel_* 指标）
- [ ] Grafana 加 panel "Channel Disable Frequency"（用上面 PromQL）
- [ ] 告警接到 Lark：在 [Grafana 告警](operations/templates/grafana/alerts.yaml) 加 `OmniRouterChannelDisableSpike` 规则
- [ ] 模拟场景测试：手动停一个 channel 的上游 → 看是否触发禁用 + Lark 通知 → 等恢复 → 看是否自动 enable + Lark 通知

## 后续扩展（不在本次 MVP）

- [ ] **per-错误类型不同冷却时长**：auth_failed 应该冷却 60min（key 大概率废），upstream_5xx 应该冷却 5min。当前所有禁用都靠 AutoTestChannelMinutes 统一恢复，不区分 reason。需要扩展 channel 表加 `disabled_until` 字段 + 探活逻辑改造
- [ ] **per-(channel, model) 粒度禁用**：当前是整个 channel 禁用。未来可针对单个模型路径禁用（保留其它路径可用）
- [ ] **Redis 后端的滑动窗口**：当前是 in-memory，多副本各自统计。Redis 可以共享窗口，但需要权衡时延 + 复杂度
- [ ] **admin UI 配置 burst threshold**：目前只能 env / json 改设置；UI 入口能让运营自服务

## 相关文件

- [`../../service/channel.go`](../../service/channel.go) — 上游 disable / enable 主流程
- [`../../service/channel_health.go`](../../service/channel_health.go) — OmniRouter 滑动窗口 + 分类
- [`../../service/channel_metrics.go`](../../service/channel_metrics.go) — Prometheus counter
- [`../../controller/relay.go`](../../controller/relay.go) — 调用入口
- [`./lark-notification.md`](./lark-notification.md) — Lark 告警通道设计
- [`./prometheus-metrics.md`](./prometheus-metrics.md) — Prometheus 指标全清单
- [`../operations/templates/grafana/`](../operations/templates/grafana/) — Grafana 看板（已含 channel error panel，可加 disable panel）
