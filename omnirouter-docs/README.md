# omnirouter-docs

OmniRouter 商业化中转站基于 [new-api](https://github.com/QuantumNous/new-api) 二次开发的设计文档目录。

> ⚠️ 本目录仅放 OmniRouter 自研增量的设计文档，**不动 new-api 上游代码与文档**。
> 这样做是为了减少未来 rebase 上游主分支时的冲突（参见 plan 第 9 节"upstream 同步成本"）。

## 项目信息

- **品牌**：OmniRouter
- **主域名**：omnirouter.org
- **底座**：[new-api](https://github.com/QuantumNous/new-api)（保留原 nеw-аρi / QuаntumΝоuѕ 标识，按 CLAUDE.md Rule 5）

## 目录索引

| 文件 | 主题 | 关联代码 |
|---|---|---|
| [observability/health-endpoints.md](observability/health-endpoints.md) | k8s 风格存活/就绪探针 | `controller/healthz.go`, `router/health-router.go` |
| [observability/lark-notification.md](observability/lark-notification.md) | Lark/飞书告警通道 + 系统告警分发器 | `service/lark_notify.go`, `service/system_alert.go` |
| [observability/prometheus-metrics.md](observability/prometheus-metrics.md) | Prometheus 指标 + /metrics 端点 | `middleware/metrics.go`, `controller/metrics.go` |
| [operations/model-group-catalog.md](operations/model-group-catalog.md) | 28 个模型分组 + 倍率建议（对标 PackyAPI） | 运营配置（不动代码） |
| [operations/brand-setup.md](operations/brand-setup.md) | OmniRouter 品牌追加配置指南（Rule 5 合规） | 运营配置（不动代码） |
| [operations/brand-seed.sql](operations/brand-seed.sql) | 品牌信息 SQL 种子（PG/MySQL/SQLite 三库适配） | 运营配置 |
| [onboarding/README.md](onboarding/README.md) | 客户端接入总览 + 端点速查 | 用户文档 |
| [onboarding/claude-code.md](onboarding/claude-code.md) | Claude Code 接入 OmniRouter | 用户文档 |
| [onboarding/codex-cli.md](onboarding/codex-cli.md) | Codex CLI 接入 OmniRouter（含包月套餐） | 用户文档 |
| [onboarding/gemini-cli.md](onboarding/gemini-cli.md) | Gemini CLI 接入 OmniRouter | 用户文档 |
| [onboarding/cursor.md](onboarding/cursor.md) | Cursor IDE 接入 OmniRouter | 用户文档 |

> 后续每加一个商业化模块，在这里追加一行索引。
