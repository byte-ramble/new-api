# OmniRouter Grafana 看板

> 一键导入 Grafana 看板 + 推荐告警规则。配合 OmniRouter `/metrics` 暴露的 `omnirouter_*` 指标用。

## 文件

| 文件 | 用途 |
|---|---|
| [omnirouter-dashboard.json](./omnirouter-dashboard.json) | 主看板（8 panels）一键导入 |
| [alerts.yaml](./alerts.yaml) | Prometheus 告警规则（4 条核心）|

## 一键导入步骤

### 前置条件

1. 已部署 Prometheus 在抓取 OmniRouter `/metrics`（参考 [`prometheus-metrics.md`](../../../observability/prometheus-metrics.md) §抓取配置）
2. 已部署 Grafana 并配置好 Prometheus 作为 datasource
3. Grafana 9.0+ （schema v38）

### 导入

```
Grafana → Dashboards → New → Import
  → Upload JSON file:  omnirouter-dashboard.json
  → 选择 Prometheus datasource
  → Import
```

或者命令行：

```bash
# 假设你已经配好 grafana CLI 环境
curl -X POST http://admin:admin@grafana.omnirouter.org/api/dashboards/db \
  -H 'Content-Type: application/json' \
  -d "{\"dashboard\": $(cat omnirouter-dashboard.json), \"overwrite\": true}"
```

### 第一次打开会看到

- **Total RPS** — 整站 QPS（绿/黄/红阈值 100 / 1000）
- **5xx Error Rate** — 5 分钟错误率（黄 0.5% / 红 1%）
- **In-flight Requests** — 并发请求数（黄 50 / 红 100）
- **Token throughput** — 每秒 token 流转（待 RecordRelayTokens wiring 后才有数据）
- **RPS by Route** — 按路由分解的 QPS
- **Latency Percentiles** — p50 / p95 / p99（5m 滑动窗口）
- **Status Code Distribution** — 200/404/429/5xx 堆叠柱状
- **Upstream Errors by Channel** — 上游错误，按 channel + 原因分解

## 推荐告警

[`alerts.yaml`](./alerts.yaml) 包含 4 条核心告警，复制到 Prometheus 的 `rules.d/` 目录或 Grafana Alert Rules：

1. **HighErrorRate** — 5 分钟内 5xx > 1% 持续 5 分钟
2. **HighLatency** — `/v1/chat/completions` p99 > 10s 持续 5 分钟
3. **InflightSpike** — in-flight > 100 持续 2 分钟（多半上游卡顿）
4. **ChannelErrorBurst** — 任一 channel 错误率突增 > 0.5/s 持续 3 分钟

## 告警通道接 Lark

Grafana → Alerting → Contact points → Webhook → URL 填你的 Lark/飞书自定义机器人 webhook URL。

或者用 OmniRouter 自带的 `service.SendSystemAlert`（见 [`lark-notification.md`](../../../observability/lark-notification.md)）：写个简单的 prometheus → 业务告警端点桥接，让告警走我们的 Lark 卡片格式（更好看，含分类色）。

## 自定义建议

### 增加 panel 想法

- **每分组 RPS / 错误率**（需把 channel label 拓展到 HTTP metrics — Phase 2 改造）
- **缓存命中率随时间变化**（需暴露 cache_hit / total 计数器 — Phase 2）
- **每用户的调用次数 / 消费 top 10**（需暴露 user_id label — 注意基数，建议只暴露 top-N）
- **节省金额可视化**（需暴露 cost_saved_usd 计数器 — Phase 2 商业化打磨）

### 减少 panel 想法

如果你只关心运维不关心业务指标，可以删除"Token throughput"和"Upstream Errors by Channel" panel，dashboard 更紧凑。

## 故障排查

| 症状 | 可能原因 | 解决 |
|---|---|---|
| 所有 panel 都 No Data | Prometheus 没在抓 OmniRouter | 确认 prometheus.yml scrape_config 指向 omnirouter.org/metrics |
| 部分 panel No Data | Token / upstream metrics 还没 wiring | 看 plan 第 11.3 阶段二 — Token 计数 wiring 是 deferred 项 |
| Latency 全是 0 或 NaN | histogram_quantile 上层标签太少 | 改 expr 加 `or vector(0)` 兜底（已加） |
| Datasource 错误 | 导入时没选 Prometheus | 编辑面板的 Data source 字段重新指定 |

## 相关文件

- [`../../../observability/prometheus-metrics.md`](../../../observability/prometheus-metrics.md) — `/metrics` 端点的设计 + 完整指标清单
- [`../../../observability/lark-notification.md`](../../../observability/lark-notification.md) — Lark 告警通道接入
- [`../launch-checklist.md`](../launch-checklist.md) §6 监控告警 — 整体监控部署 SOP
