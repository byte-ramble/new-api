# Prometheus 指标 (/metrics)

## 背景

Plan 第 5 节列了"Prometheus metrics 端点"为 ❌ 缺失，是商业化必需。
本次实现把 `/metrics` 暴露给 Prometheus 抓取，配合 Grafana 做监控看板与告警。

复用：项目原本已经有 `middleware/stats.go`（只统计 active connections，仅挂在 relay 路由上）和 `middleware/log.go` 的请求日志，但都不是 Prometheus 格式，无法被现代监控栈直接消费。

## 设计

### 文件结构

| 文件 | 职责 |
|---|---|
| `middleware/metrics.go` | Prometheus 中间件 + 指标定义 + Token/错误记录 helper |
| `controller/metrics.go` | `/metrics` HTTP handler（包装 promhttp.Handler） |
| `router/health-router.go`（修改） | 注册 `/metrics` 路由（通过 `METRICS_ENABLED` env 控制） |
| `main.go`（修改） | 全局挂载 `middleware.PrometheusMiddleware()` |
| `go.mod`（修改） | 把 `prometheus/client_golang` 从 indirect 提到 direct |

### 调用链

```
请求进来
  └─> server.Use(middleware.PrometheusMiddleware())   ← 全局中间件
        ├─> 跳过 /health, /healthz, /ready, /readyz, /metrics
        ├─> httpInflight.Inc / Dec
        ├─> c.Next() (走完全部 handler)
        └─> 记录 httpRequestsTotal / httpRequestDuration

Prometheus 抓取
  └─> GET /metrics
        └─> router.SetHealthRouter 注册的路由
              └─> controller.MetricsHandler
                    └─> promhttp.Handler() (default registry)
```

### 暴露的指标

| 指标 | 类型 | 标签 | 含义 |
|---|---|---|---|
| `omnirouter_http_requests_total` | Counter | route, method, status | HTTP 请求总数 |
| `omnirouter_http_request_duration_seconds` | Histogram | route, method | 请求延迟（5ms..20s 共 12 个桶） |
| `omnirouter_http_inflight_requests` | Gauge | — | 当前 in-flight 请求数 |
| `omnirouter_relay_prompt_tokens_total` | Counter | channel, model | 累计 prompt tokens（待业务点接入） |
| `omnirouter_relay_completion_tokens_total` | Counter | channel, model | 累计 completion tokens（待业务点接入） |
| `omnirouter_relay_upstream_errors_total` | Counter | channel, reason | 上游错误（reason 低基数：timeout/5xx/rate_limited 等） |

> 命名规范：`<namespace>_<subsystem>_<name>_<unit>`，全部小写，单复数符合 Prometheus best practice。

### 关键设计决定

1. **直接用默认 registry**：Pyroscope 已经在用，混在一起没问题；自定义 registry 会带来管理负担。
2. **promauto 在包 init 注册**：无需调用 setup 函数；测试里多次 import 同一包不会重复注册。
3. **跳过探针路径**：避免 Prometheus 拉取自身把 /metrics 计入 metrics（自我引用）。
4. **路由模板而非具体 URL**：`c.FullPath()` 给出 `/v1/chat/completions` 而非 `/v1/chat/completions?stream=true&token=xxx`，控制 label 基数。
5. **404 归一到 `unknown`**：未注册路由不会产生海量 label。
6. **Token 计数器的 wiring 推迟**：函数已暴露（`RecordRelayTokens`、`RecordRelayUpstreamError`），但本次不接入 relay 流程（需要小心找 token 统计的真实点）。Phase 2 再 wiring。
7. **`/metrics` 默认开启**：env `METRICS_ENABLED` 默认 true；显式 false/0/no/off 才关。生产部署应在 ingress 层限制访问源 IP。

### Prometheus 抓取配置示例

```yaml
scrape_configs:
  - job_name: omnirouter
    metrics_path: /metrics
    scrape_interval: 15s
    static_configs:
      - targets: ['omnirouter.org:3000']
```

### Grafana 告警示例（PromQL）

| 名称 | 表达式 | 含义 |
|---|---|---|
| 5xx 错误率 > 1% | `sum(rate(omnirouter_http_requests_total{status=~"5.."}[5m])) / sum(rate(omnirouter_http_requests_total[5m])) > 0.01` | 5 分钟内 5xx 占比超过 1% |
| P99 延迟 > 5s | `histogram_quantile(0.99, rate(omnirouter_http_request_duration_seconds_bucket{route="/v1/chat/completions"}[5m])) > 5` | Chat 接口 p99 超过 5 秒 |
| 上游错误激增 | `sum(rate(omnirouter_relay_upstream_errors_total[5m])) by (channel) > 0.5` | 任一渠道每秒上游错误数 > 0.5 |
| 并发请求堆积 | `omnirouter_http_inflight_requests > 100` | 同时处理的请求数过多 |

可结合 Lark webhook（见 [lark-notification.md](./lark-notification.md)）做告警分发。

## 测试

`middleware/metrics_test.go` 覆盖：
1. `TestPrometheusMiddleware_RecordsRequest` — 请求被正确计入 counter，`/metrics` 暴露目标 series
2. `TestPrometheusMiddleware_SkipsProbeRoutes` — 探针/metrics 自身不被记录
3. `TestRecordRelayTokens` — Token 记录函数正确累加，零值不写入

跑：
```bash
cd /Users/maning/Documents/dev/js-dev/new-api
go test ./middleware/ -run TestPrometheus -v
go test ./middleware/ -run TestRecordRelayTokens -v
```

## 后续扩展（不在本次 MVP）

- [ ] 把 `RecordRelayTokens` 接到 `relay/relay-text.go::PostConsumeQuota` 等真实 token 结算点
- [ ] 把 `RecordRelayUpstreamError` 接到 channel 异常处理路径
- [ ] 加用户/Token/分组维度的指标（注意基数控制）
- [ ] 暴露 Go runtime metrics（已被 promhttp.Handler 默认包含 GC/goroutines/memstats）
- [ ] 集成 OpenTelemetry 做链路追踪（plan 第 5 节"链路追踪"）
- [ ] /metrics 端点加 IP allow-list 中间件（生产场景）
