# OmniRouter 商业上线 Checklist

> 给运营 + 法务 + 技术的统一动作清单。按优先级排，每项标注**谁做、多久、依赖、关键工具**。
>
> 阅读建议：先扫一遍全部，标记你已经有的、还差的。然后按"先决条件"的顺序逐项推进。

---

## 0. 总览：先决条件依赖图

```
公司主体 (1) ──┬─→ 银行账户 (2) ──┬─→ Stripe 收款 (3)
               │                   ├─→ 支付宝/微信商家 (4)
               └─→ 税务 (5)        └─→ 国内对公收款
                                                            ↓
品牌资产 (6) ──→ 域名 + ICP 备案 (7) ──→ HTTPS/CDN (8) ──→ 上线
                                                            ↑
法律文书 (9) ──┬→ 用户协议 ──→ 站内展示 ──→ 注册流程
               ├→ 隐私政策
               └→ 数据合规

监控运维 (10) ─┬→ Lark Webhook ──→ 告警生效
               ├→ Prometheus + Grafana
               └→ 状态页

社群 (11) ─────┬→ TG 频道
               ├→ Lark/飞书群
               └→ 客服邮箱
```

如果你是个人开发者快速 MVP，**最小路径**是：
1 → 3 → 7 → 8 → 11 → 上线（其它后补）

如果你做 toB / 长期运营，**全部走完**才稳。

---

## 1. 法务 / 主体

| # | 任务 | 优先 | 时长 | 谁做 | 依赖 | 工具/链接 |
|---|---|---|---|---|---|---|
| 1.1 | 决定主体形态：个人 / 国内有限公司 / 海外公司 (US/HK/SG) | **必须** | 1 天思考 | 你 | — | 见下方决策表 |
| 1.2 | 注册公司（如需） | **必须** | 国内 7-15 天 / 海外 1-2 天 | 你 / 代办 | 1.1 | [Stripe Atlas (US)](https://stripe.com/atlas)、[Sleek (HK/SG)](https://sleek.com)、国内代办 |
| 1.3 | 公司银行账户开户 | **必须** | 国内 1-2 周 / 海外 1-4 周 | 你 | 1.2 | Mercury (US)、汇丰 (HK) |
| 1.4 | 税务登记 | **必须** | 1-3 天 | 你 / 会计 | 1.2 | — |
| 1.5 | （如海外）考虑 EIN / W-8BEN-E 文档 | 重要 | 1-2 天 | 你 / 会计 | 1.2 | IRS.gov |

**主体决策表**：

| 形态 | 优势 | 劣势 | 建议场景 |
|---|---|---|---|
| **个人开发者**（不注册公司） | 0 成本、快 | 不能接 Stripe（信用卡）、税务麻烦、用户不信任 | 仅做小圈子工具 |
| **国内有限公司** | 接支付宝/微信、ICP 备案 | 海外支付难、国内合规重 | 主要面向国内付费用户 |
| **美国 LLC** (Stripe Atlas) | 1500-2000$ 一站式注册 + Stripe 接入 + 美国银行 | 需要美国税务申报 | **推荐：面向开发者/海外用户** |
| **HK / SG 公司** | 跨境收款方便、税率低 | 注册成本 ~5000 USD | 同时做国内国外 |

---

## 2. 支付

| # | 任务 | 优先 | 时长 | 依赖 | 工具/链接 |
|---|---|---|---|---|---|
| 2.1 | 注册 Stripe 商家账号 | **必须** | 1-3 天审核 | 1.2 1.3 | [stripe.com/register](https://stripe.com/register) |
| 2.2 | Stripe Dashboard 创建商品（按量充值 + Codex 包月） | **必须** | 1 小时 | 2.1 | Stripe Dashboard |
| 2.3 | 配置 Stripe Webhook URL = `https://omnirouter.org/api/stripe/webhook` | **必须** | 30 分钟 | 7.1 8.1 | Stripe Dashboard |
| 2.4 | 拿 `STRIPE_API_KEY` + `STRIPE_WEBHOOK_SECRET` + `STRIPE_PRICE_ID` 写到 OmniRouter 后台 | **必须** | 10 分钟 | 2.2 2.3 | OmniRouter `/console/system-settings` 里的 Stripe 段 |
| 2.5 | Stripe 测试模式跑通一次充值 → webhook → 余额到账 | **必须** | 30 分钟 | 2.4 | Stripe test card `4242...` |
| 2.6 | 申请 Stripe 生产模式（需公司证件 + 银行） | **必须** | 1-2 周审核 | 2.5 | Stripe Dashboard |
| 2.7 | （可选）接易支付 / Creem / Waffo 备线 | 重要 | 半天 | 1.2 | 见 [`controller/topup_*.go`](../../controller/) |
| 2.8 | （可选）国内对接支付宝当面付 / 微信扫码 | 重要 | 国内主体 1-2 周审核 | 1.2 1.3 | 微信支付商户平台 |
| 2.9 | 测试模式手动配 Codex 包月套餐（¥60/月） | **必须** | 2 小时 | 2.4 | OmniRouter 订阅设置 |
| 2.10 | 创建首充 9 折优惠券码（如 `OMNIROUTER-FIRST`） | 重要 | 30 分钟 | 2.4 | OmniRouter 兑换码后台 |

---

## 3. 品牌资产

| # | 任务 | 优先 | 时长 | 工具 |
|---|---|---|---|---|
| 3.1 | OmniRouter Logo 设计（512×512 PNG 透明底） | **必须** | 半天-2 天 | Figma / [logo.com](https://logo.com) AI / 找设计师 |
| 3.2 | 产品介绍单页文案（300-500 字，回答"是什么、给谁用、为什么选我"） | **必须** | 2-4 小时 | 你写 / 我帮改 |
| 3.3 | 三句话定位（用于 Twitter bio、TG 频道简介、Footer） | 重要 | 30 分钟 | 你写 |
| 3.4 | Logo 上传到 CDN / OSS（推荐 Cloudflare R2 免费 10GB） | **必须** | 30 分钟 | [Cloudflare R2](https://developers.cloudflare.com/r2/) |
| 3.5 | 把 CDN URL 配到 OmniRouter 后台 → 设置 → Logo | **必须** | 5 分钟 | OmniRouter `/console/system-settings` |
| 3.6 | （可选）favicon.ico / apple-touch-icon | 重要 | 1 小时 | [favicon.io](https://favicon.io/) |
| 3.7 | （可选）OG image（社交分享卡片 1200×630） | 重要 | 1 小时 | Figma |
| 3.8 | （可选）UI 配色定制（覆盖 Tailwind 主题色） | 可选 | 半天 | 改 `web/default/tailwind.config.js` |

---

## 4. 域名 / DNS / HTTPS

| # | 任务 | 优先 | 时长 | 依赖 | 工具 |
|---|---|---|---|---|---|
| 4.1 | `omnirouter.org` 已注册（你已确认） | ✅ | — | — | — |
| 4.2 | DNS A 记录指向你的服务器 IP（或 CNAME 到云厂商 LB） | **必须** | 10 分钟 + DNS 传播 5-30 分钟 | 服务器 5.x | Cloudflare DNS / 阿里云 DNS |
| 4.3 | （国内用户）ICP 备案 | **必须如目标用户在国内** | **15-30 天** | 1.2 | 阿里云/腾讯云 ICP 备案系统 |
| 4.4 | HTTPS 证书：Let's Encrypt 或 Cloudflare Full | **必须** | 1 小时 | 4.2 | [Caddy](https://caddyserver.com/) 自动签 / Cloudflare 一键开 |
| 4.5 | 反向代理配置（nginx / Caddy） | **必须** | 1 小时 | 4.4 | 见下方反代示例 |
| 4.6 | （推荐）Cloudflare 前置：DDoS 防护 + 全球 CDN + WAF | 重要 | 1-2 小时 | 4.4 | [cloudflare.com](https://cloudflare.com) |
| 4.7 | 子域规划：`api.omnirouter.org`（API）、`docs.omnirouter.org`（文档站）、`status.omnirouter.org`（状态页） | 重要 | 30 分钟 | 4.2 | DNS |

**Caddy 反代示例**（最简）：
```caddy
omnirouter.org {
    reverse_proxy localhost:3000
}

# 把 metrics 锁住
omnirouter.org/metrics {
    @prom_only remote_ip 1.2.3.4   # 只让你的 Prometheus IP 访问
    handle @prom_only {
        reverse_proxy localhost:3000
    }
    handle {
        respond "Forbidden" 403
    }
}
```

---

## 5. 服务器 / 部署

| # | 任务 | 优先 | 时长 | 工具 |
|---|---|---|---|---|
| 5.1 | 选云厂商：DigitalOcean / Linode / 阿里云 / 腾讯云 | **必须** | 1 天选型 | — |
| 5.2 | 起一台机器：4 vCPU / 8GB RAM / 80GB SSD（MVP 起步） | **必须** | 30 分钟 | — |
| 5.3 | 装 Docker + Docker Compose | **必须** | 30 分钟 | `curl -fsSL https://get.docker.com \| sh` |
| 5.4 | git clone 你的 fork → 跑 P3 build 出来的 docker compose | **必须** | 1 小时 | 见 P3 输出（构建中） |
| 5.5 | 跑 [`brand-seed.sql`](./brand-seed.sql) 装上 OmniRouter 品牌 | **必须** | 5 分钟 | psql |
| 5.6 | 备份策略：Postgres dump 每日 cron + S3/R2 上传 | **必须** | 半天 | `pg_dump` + `rclone` |
| 5.7 | 日志收集：docker logs → 文件 → logrotate（或 Loki） | 重要 | 半天 | logrotate |
| 5.8 | 系统监控：Netdata / node_exporter | 重要 | 1 小时 | — |

---

## 6. 监控 / 告警

| # | 任务 | 优先 | 时长 | 依赖 | 操作 |
|---|---|---|---|---|---|
| 6.1 | 创建 Lark/飞书自定义群机器人，拿到 webhook URL | **必须** | 10 分钟 | 11.2 | [飞书机器人指南](https://open.larksuite.com/document/uAjLw4CM/ukTMukTMukTM/bot-v2/add-custom-bot) |
| 6.2 | docker-compose env 加 `LARK_WEBHOOK_URL=...` 让我加的告警生效 | **必须** | 5 分钟 | 6.1 5.4 | 改 env 重启容器 |
| 6.3 | （可选）开启飞书加签校验，加 `LARK_WEBHOOK_SECRET=...` | 重要 | 5 分钟 | 6.1 | — |
| 6.4 | 部署 Prometheus 抓 `https://omnirouter.org/metrics` | **必须** | 1 小时 | 4.5 | Prometheus / Grafana Cloud |
| 6.5 | Grafana 配 OmniRouter 看板（QPS、延迟、错误率、in-flight） | **必须** | 半天 | 6.4 | 见 [prometheus-metrics.md](../observability/prometheus-metrics.md) |
| 6.6 | 配 Grafana → Lark 的告警通道 | **必须** | 30 分钟 | 6.5 6.1 | Grafana 内置 webhook |
| 6.7 | 关键告警阈值：5xx > 1%、p99 延迟 > 5s、in-flight > 100 | **必须** | 1 小时 | 6.5 | 见同上 metrics 文档 |
| 6.8 | （可选）状态页：`status.omnirouter.org` | 重要 | 半天 | 4.7 | [Uptime Kuma](https://github.com/louislam/uptime-kuma) 自托管 / [Atlassian Statuspage](https://www.atlassian.com/software/statuspage) |

---

## 7. 法律 / 合规

| # | 任务 | 优先 | 时长 | 工具 |
|---|---|---|---|---|
| 7.1 | 用户协议（Terms of Service） | **必须** | 半天 | [TermsFeed 生成器](https://www.termsfeed.com/) / 找律师（推荐） |
| 7.2 | 隐私政策（Privacy Policy） | **必须** | 半天 | 同上 |
| 7.3 | 公平使用条款（针对 Codex 包月、Claude Code 防滥用） | **必须** | 1 小时 | 你写，我可帮草拟 |
| 7.4 | Cookie 政策（GDPR 必需） | 重要如有欧洲用户 | 1 小时 | 同上 |
| 7.5 | 把 7.1 7.2 7.3 录入 OmniRouter 后台 → 法律设置 | **必须** | 30 分钟 | OmniRouter 后台 |
| 7.6 | 退款政策明确写出（如"7 天内未使用超 50% 可退"） | **必须** | 1 小时 | 写到用户协议 |
| 7.7 | （海外）DMCA 联系信息 | 重要 | 30 分钟 | — |
| 7.8 | （国内）数据出境合规 | 重要 | 1-2 周 | 询问律师 |

---

## 8. 社群 / 触达

| # | 任务 | 优先 | 时长 | 链接 |
|---|---|---|---|---|
| 8.1 | 注册 Telegram 频道 `@OmniRouter`（公告 + 故障播报） | **必须** | 10 分钟 | t.me |
| 8.2 | 注册 Telegram 群（用户自助 + 客服） | **必须** | 10 分钟 | t.me |
| 8.3 | 创建 Lark / 飞书群（管理员告警接收 + 高净值用户群） | **必须** | 10 分钟 | feishu.cn |
| 8.4 | 客服邮箱 `hi@omnirouter.org` 配置（Cloudflare Email Routing 免费） | **必须** | 30 分钟 | Cloudflare Email Routing |
| 8.5 | （可选）Twitter / X 账号 `@omnirouter` | 重要 | 30 分钟 | x.com |
| 8.6 | （可选）Linux DO / V2EX / Reddit r/LocalLLaMA 帐号 | 重要 | 1 小时 | — |
| 8.7 | 把以上入口都放到站点 Footer + 文档头 | **必须** | 30 分钟 | OmniRouter 后台 + 改 [`brand-seed.sql`](./brand-seed.sql) Footer |

---

## 9. 种子运营

| # | 任务 | 优先 | 时长 |
|---|---|---|---|
| 9.1 | 列 20-50 个种子用户名单（V2EX 大佬、Twitter 开发者、TG 群） | **必须** | 1 天 |
| 9.2 | 给每个种子用户准备**专属优惠码 + 大额免费余额**（如 $10） | **必须** | 1 小时 |
| 9.3 | 写 3 篇推广素材：技术博客（对比 PackyAPI）+ Linux DO 帖子 + Twitter 长推文 | **必须** | 1-2 天 |
| 9.4 | 准备 Claude Code / Codex CLI 配置截图、视频 demo | 重要 | 半天 |
| 9.5 | （阶段二完成后）准备多级返佣公告 + 推广素材 | 重要 | 半天 |

---

## 10. 上线前最终检查（Go-Live Checklist）

按这个顺序当天 ✓：

- [ ] 域名 HTTPS 通了 (`curl -I https://omnirouter.org`)
- [ ] `/health` 200、`/ready` 200（DB+Redis up）
- [ ] `/metrics` 不对外（已锁 IP）
- [ ] 注册流程跑一次（含邮箱验证）
- [ ] 充值跑一次（Stripe test card → 真实模式 → 真扣款 1 元 → webhook → 余额到账 → 退款）
- [ ] 创建 API Key → 用 Claude Code 调一次 → 看日志正确扣费
- [ ] 触发一个故障（关掉一条 channel）→ 验证自动 fallback + Lark 告警
- [ ] 用户协议 / 隐私政策 / 退款政策都能从 Footer 点开
- [ ] 客服邮箱发一封验证（自己发自己收）
- [ ] TG 频道发一条"Hello OmniRouter"
- [ ] 备份脚本跑一次完整流程
- [ ] Grafana 看板有数据 + 告警规则有效
- [ ] 公开页 `/pricing`（阶段二完成后）SEO 检查（`view-source:` 看到完整内容、不依赖 JS）

---

## 11. 优先级速查

如果你时间紧只能挑 10 件先做：

1. ✅ 决定主体形态（1.1）
2. ✅ Stripe 注册（2.1, 1 天审核期同时做其它）
3. ✅ 设计 Logo（3.1，可外包）
4. ✅ DNS 解析 + Caddy + HTTPS（4.2, 4.4, 4.5）
5. ✅ 服务器开机 + Docker + 跑起来（5.1-5.5）
6. ✅ Lark 群机器人（6.1, 6.2，开告警）
7. ✅ 用户协议 + 隐私政策（7.1, 7.2，模板生成）
8. ✅ TG 频道 + 客服邮箱（8.1, 8.4）
9. ✅ 1 篇推广文 + 5-10 个种子用户（9.1, 9.3）
10. ✅ Go-Live 当天清单（10.x）

剩下的等你跑起来再补。

---

## 12. 我能帮你做什么

| 你做的 | 我能帮的 |
|---|---|
| 1.1 决定主体 | 给你 4 种形态详细对比 |
| 2.x Stripe 接入 | 帮你写 webhook 调试命令、测试 cURL |
| 3.2 文案 | 草拟 30 字 / 100 字 / 300 字版本 |
| 4.5 反代 | 完整 Caddy / nginx 配置 |
| 5.x 部署 | 写 systemd 单元 + 备份脚本 |
| 6.5 Grafana 看板 | 给你 Grafana JSON 一键导入 |
| 7.1-7.3 法律文书 | 给你符合中转站业务的草稿（**仅参考，正式文本要找律师**） |
| 9.3 推广素材 | 写技术博客 / 论坛帖子 / Twitter 文案 |
| 10.x Go-Live | 帮你跑自动化检查脚本 |

随时告诉我你卡哪一步，我接手。

---

## 相关文件

- [`brand-setup.md`](./brand-setup.md) — 品牌追加详细操作（Rule 5 合规）
- [`brand-seed.sql`](./brand-seed.sql) — 品牌信息 SQL 种子
- [`model-group-catalog.md`](./model-group-catalog.md) — 28 个分组 + 倍率 JSON
- [`../observability/lark-notification.md`](../observability/lark-notification.md) — Lark 告警技术实现
- [`../observability/prometheus-metrics.md`](../observability/prometheus-metrics.md) — Prometheus 指标 + Grafana 配置
- [`../onboarding/README.md`](../onboarding/README.md) — 客户端接入文档
