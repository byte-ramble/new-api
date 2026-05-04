# OmniRouter 项目交付总览 + 30 天 Roadmap

> 给未来的你 / 团队 / 接手人一份"看完就知道现状 + 下一步做什么"的归档。
>
> **最后更新**：2026-05-04
> **代码 fork**：[byte-ramble/new-api](https://github.com/byte-ramble/new-api)
> **当前状态**：LIVE 在 docker compose（localhost:3000），技术 MVP 完成，**未上线公网**
> **距离 ¥1 真实营收**：约 **2-3 个月外部依赖 + 1 周工程 + ¥20-50k**

---

## 一、目前已交付（16 commits）

### 代码改动汇总

| 文件类别 | 数量 | 备注 |
|---|---|---|
| 新增 Go 代码 | 18 个 | controller/middleware/service/model 各 layer |
| 修改 Go 代码 | ~10 个 | 多为 hook 接入（topup webhooks, router） |
| 新增前端 TS/TSX | 12 个 | features/affiliate/ + features/pricing/ 增量 |
| 修改前端 TS/TSX | 5 个 | i18n 增 79 keys、main.tsx 加 HelmetProvider |
| 新增 Docker / Compose / Dockerfile | 3 个 | omnirouter 专属 ship 链路 |
| 新增文档 (omnirouter-docs/) | 25+ 个 | 设计文档 + 运营 SOP + 用户文档 + 模板 |
| 新增 Logo 资产 | 6 个 SVG | 4 概念 + wordmark + lockup |
| **总变动行数** | **~7000+ lines** | 含测试与文档 |

### 功能模块（按 Phase 划分）

#### ✅ Phase 1.0 · 可观测性基线（P1.0 / `ac8ae549`）
- `/health`、`/ready`、`/healthz`、`/readyz` k8s 风格探针
- Prometheus `/metrics` 端点 + 65 个 `omnirouter_*` series
- Lark / 飞书告警通道（`SendSystemAlert`，含飞书加签算法）
- 全局 PrometheusMiddleware

#### ✅ Phase 1.1 · 模型广场公开页（P1.1 / 6 commits）
- 4 个新 React 组件（SeoHead / GroupShowcase / CacheDiscountExplainer / PricingCalculator）
- 24 个 zh+en i18n 翻译
- react-helmet-async 加 SEO meta + OG / Twitter card
- 价格计算器：用户输入用量 → 月费 + 对比官方原价 + 节省金额

#### ✅ Phase 1.2 · 渠道自动禁用（P1.2 / `3e3bd7c7`）
- 滑动窗口去抖（避免单次 5xx 误杀）
- 错误分类（auth_failed / rate_limited / 5xx / network / quota / manual / other）
- DisableChannel + EnableChannel hook 接 Lark 告警 + Prometheus counter
- 6/6 单元测试通过

#### ✅ Phase 1.3 · 多级邀请返佣（P1.3 / 3 commits）
- 3 个新数据库表（`affiliate_accounts`、`commission_logs`、`withdrawals`）
- L1/L2 百分比返佣 + 24h 累计上限风控 + 自邀检测
- 5 个充值通道全 wire（epay / Stripe / Creem / Waffo / Waffo-Pancake）
- 7 个 REST API 端点（用户 4 + admin 3）
- 用户"我的推广"页 + admin 提现审核页
- 7/7 单元测试通过、55 个 zh+en i18n 翻译

#### ✅ Phase 2 · 运营 SOP + 模板（P2 / 4 commits）
- `launch-checklist.md` — 200+ 项上线 checklist
- `brand-copy.md` — 中英 / 三长度 / 各社交 / 推广素材 / 公告模板
- `reverse-proxy.md` — Caddy + nginx + Cloudflare 生产配置
- `legal/` 4 篇法律文书草稿（TOS / 隐私 / 退款 / 公平使用 + disclaimer）
- `grafana/` 看板 JSON + 告警规则
- `brand/` 4 logo 概念 + wordmark + lockup + designer brief

#### ✅ Phase 3 · 部署链路（`9bc2d75a`）
- `Dockerfile.omnirouter` — slim 多阶段（无前端 build，复用 host dist）
- `docker-compose.omnirouter.yml` — 一键起 PG + Redis + new-api
- 已构建镜像 `omnirouter/new-api:phase1-mvp`（271MB）

---

## 二、当前 LIVE 状态（localhost:3000）

容器跑了 5 小时无问题：

```
omnirouter        Up 5 hours (healthy)    omnirouter/new-api:phase1-mvp
omnirouter-pg     Up 13 hours (healthy)   postgres:15-alpine
omnirouter-redis  Up 13 hours (healthy)   redis:7-alpine
```

**访问入口**：

| URL | 状态 | 说明 |
|---|---|---|
| `http://localhost:3000/health` | 200 | k8s 探针，process alive |
| `http://localhost:3000/ready` | 200 | DB + Redis 都 up |
| `http://localhost:3000/metrics` | 200 | 65 行 omnirouter_* series |
| `http://localhost:3000/pricing` | 200 | 模型广场公开页（无需登录）|
| `http://localhost:3000/affiliate` | SPA → 401 redirect | 我的推广（需登录）|
| `http://localhost:3000/affiliate/admin` | SPA → 401 | 审核（需 admin）|
| `http://localhost:3000/api/affiliate/*` | 401 | 7 个 REST 端点全注册 |

**品牌**：SystemName=OmniRouter / ServerAddr=https://omnirouter.org / Logo URL 占位 CDN

---

## 三、距离上线还差什么

> 上一轮的 8 维度成熟度评估：核心引擎 80% / 商业化代码 70% / 运维 20% / 法律 5% / 支付 15% / 品牌 15% / 真实运营 0%

按"是否阻塞上线"分级：

### 🔴 Hard Blockers（不做就上不了线）

1. **公司主体注册**（1-4 周）
   - 推荐：Stripe Atlas（约 $500，2-3 周拿到 LLC + Mercury 银行账户）
   - 国内主体：另算 ICP 备案（15-30 天）
2. **域名 → 服务器 → HTTPS 部署**（1 天工程）
   - DigitalOcean / Linode / 阿里云开机
   - DNS 解析 omnirouter.org → 服务器 IP
   - Caddy 自动签 Let's Encrypt
   - `docker compose -f docker-compose.omnirouter.yml up -d`
3. **Stripe 生产模式审核**（1-2 周等审核）
   - 主体到位 + 银行账户 + 网站可访问 → 提交
4. **法律文书律师 review**（1-3 周）— **这就是为什么需要律师**：
   - AI draft 的 TOS / 隐私 / 退款 / 公平使用是**起步草稿**，不是法律建议
   - 哪些条款在你的目标司法辖区可执行 / 哪些被认为不公平/无效，AI 不知道
   - 跨境数据流（中国主体调海外 OpenAI）的合规边界，AI 给不了准确建议
   - 出事时（用户告你 / chargeback / 监管问询）你需要一个能站在你身边代表你的人
   - 国内主体起步约 ¥3-8k 一次 review；海外（美国 LLC）¥10-30k；找熟悉 SaaS / 跨境的
5. **基础备份自动化**（半天工程）
   - Postgres `pg_dump` daily cron + 推到 S3/R2
   - 一炸数据库 = 公司直接死

### 🟡 Soft Blockers（不做能上但很危险）

6. **Logo 真实资产** — 现有 4 个 SVG 概念稿可直接用，或精修后用
7. **状态页**（Uptime Kuma 自托管，半天）
8. **Lark webhook 真接好** —— env 里 `LARK_WEBHOOK_URL=` 留空了，10 分钟事
9. **TG 频道 + 飞书群注册**（半天）
10. **Stripe 测试 → 真模式跑通**（充 ¥1 → webhook → 余额 → 退款，1 小时）

### 🟢 Quality Issues（影响留存 + 长期）

11. 测试覆盖率（当前 ~7%，目标 ≥30%）
12. CI/CD GitHub Actions
13. 多副本部署（限流换 Redis 后端）
14. OpenTelemetry 链路追踪
15. SEO 真做（Rsbuild prerender SSG）
16. 进阶 anti-fraud（IP/设备指纹）
17. 报表导出（PDF/Excel）
18. 日志冷热分离 + 归档
19. mobile UX 真测试
20. 运营 admin dashboard

---

## 四、30 天 Roadmap（"开门"路径）

按这个顺序推进，30 天后能跑能收钱。

### Week 1：确权 + 平行启动审核

| 周一 | 决定主体形态：US LLC（推荐 toC AI 用户）/ 国内有限公司 / HK 公司 |
| 周二 | 注册 Stripe Atlas（$500） — 提交后回家等 |
| 周三 | 找律师挂号（询价 + 预约）— 推荐看 Linux DO / 即刻 / Twitter 上做过 SaaS 出海的律师 |
| 周四 | 服务器开机：DigitalOcean Premium AMD 4vCPU/8GB（$48/mo）/ 北京阿里云（¥320/mo） |
| 周五 | DNS：omnirouter.org A 记录 → 服务器 IP；Cloudflare 套一层（免费） |

**Week 1 现金支出**：~¥4-5k（Stripe Atlas + 1 个月服务器 + 律师定金）
**Week 1 你时间**：~10-15 小时

### Week 2：部署 + 法律 review 同步

| 周一 | git clone byte-ramble/new-api → 服务器 → `docker compose up -d` |
| 周二 | Caddy 配置 + HTTPS 自动签 + 反代到 :3000；测 `https://omnirouter.org/health` |
| 周三 | 跑 brand-seed.sql；浏览器看 SystemName=OmniRouter 生效 |
| 周四 | 备份脚本：`pg_dump` daily → R2/S3 + 7 天保留 |
| 周五 | 注册飞书机器人 webhook → docker env 加 `LARK_WEBHOOK_URL`；故意停一下 PG 验证告警通 |

**Week 2 你时间**：~15-20 小时
**收：律师 review 4 篇法律文书的初稿（你这周拿到改稿）**

### Week 3：法律落地 + Logo + Stripe 通过

| 周一-周二 | 跟律师 1-2 轮迭代法律文书 → 终稿 → 上传到 admin → 注册流程显示同意勾选 |
| 周三 | Logo：选 brand/ 里的 Concept B（OmniBot）→ Fiverr 找设计师精修 ¥300-500 OR 自己 Figma 调 |
| 周四 | Stripe 通常 Week 3 通过：拿到生产 secret → docker env 替换 → 测试模式充 ¥1 → 真模式充 ¥1 → 余额到账 → 退款测试 |
| 周五 | 注册 TG 频道 `@OmniRouter` + 创建飞书用户群；Footer 加入口 |

**Week 3 你时间**：~10-15 小时
**Week 3 现金**：律师尾款（¥3-8k）+ Logo 设计费（¥300-2k）

### Week 4：种子 + 软启动

| 周一 | 列 30-50 个种子用户名单（V2EX / Linux DO / Twitter / 群里的开发者） |
| 周二 | 给每人发：「OmniRouter 即将上线，你是种子用户，专属 ¥10 余额 + 8 折优惠码」 |
| 周三 | 写 1 篇知乎/V2EX 长帖（用 brand-copy.md 的草稿改）+ 1 条 Twitter 长推文 |
| 周四 | 软启动：站点正式开放；监控 Lark 告警群；前 24h 每 2h 看一次 |
| 周五 | 收第一波反馈 → 修第一波 bug → 第一笔 ¥1 真实交易 |

**Week 4 你时间**：~20-30 小时（运营沟通密集）
**Week 4 期望成果**：30-50 注册 / 5-10 充值 / ¥500-3000 GMV

### Day 31+ 常规节奏

- **每周**：1 篇内容（V2EX / 知乎 / Twitter）+ 1 次种子用户私聊
- **每月**：上游模型价格 review + Lark 告警复盘 + 邀请返佣派发对账
- **每季度**：律师 review TOS 是否需更新；监管要求合规审计
- **6 个月里程碑**：MAU 500-1000 / 月 GMV ¥3-5 万 / 月利润 ¥0-1 万（多数还在投入期）
- **12 个月里程碑**：MAU 3000-5000 / 月 GMV ¥30-100 万 / 开始盈利或决定再投资 / 决定做企业版

---

## 五、资源需求总览

### 现金成本

| 类别 | 一次性 | 每月 |
|---|---|---|
| Stripe Atlas（公司注册） | $500 (~¥3.5k) | — |
| 律师 review 4 篇文书 | ¥3-15k | — |
| Logo 精修（可选） | ¥300-3k | — |
| 域名（已有） | — | — |
| 服务器（4vCPU/8GB） | — | ¥300-400 |
| 数据库备份 R2/S3 | — | ¥50-100 |
| Cloudflare（免费够用） | — | ¥0 |
| 邮件（Cloudflare Email Routing 免费） | — | ¥0 |
| Stripe 手续费 | — | 收入的 2.9%+$0.30 |
| **30 天总计** | **¥7-22k** | **¥350-500/mo** |

### 你的时间投入

| 阶段 | 每周时间 |
|---|---|
| Week 1-4（启动） | 15-30 小时/周 |
| Month 2-3（早期运营） | 10-20 小时/周（看用户增长） |
| Month 4+（稳定运营） | 5-15 小时/周（自动化大部分）|

### 工程时间（剩余）

按 plan 第 11.4 节"稳健商业化路线"原估 1.5-2 人月——**我们已经做完代码部分**。剩下的工程时间：

- 服务器部署 + Caddy + 备份：1 天
- Stripe 真模式集成微调：0.5 天
- Lark webhook 真接 + 故障演练：0.5 天
- Bug 修复（前 30 天必有）：分散 3-5 天
- **总计：~1 周**

---

## 六、命运之轮 — 三种结局

### 🟢 顺利路径（30%）
按 30 天 roadmap 推进 → Day 30 软启动 → 月 1 末有第一批种子用户开始用 → 月 3 末 MAU 200-500 / 月 GMV ¥1-3 万 → 月 6 月入 ¥3-10 万 → 决定全职做或保持副业

### 🟡 拖延路径（50% 最常见）
Week 1 顺利但 Week 2-3 卡在律师 / Stripe 审核 / 上游 API 联调 → 6-8 周才软启动 → 错过启动势头 → 用户不增长 → 月 3 决定停下来反思

### 🔴 失败路径（20%）
公司主体注册卡住 / 律师 review 太严格 / Stripe 拒绝 / 第一次重大故障没处理好导致种子用户集体跑路 / 监管问题 → 月 1-2 内放弃

**最大的差异化因素**：你能不能 **每周拿出 15-30 小时** 真投入。技术 done 了，剩下都是销售 / 运营 / 沟通 / 等待。如果你工作太忙没法投入，建议找一个 co-founder 负责运营或干脆把代码当作品集 ship 到 GitHub 写博客分享别图直接变现。

---

## 七、检查点 / KPI

### Day 7（你这周末）
- [ ] Stripe Atlas 注册提交
- [ ] 律师约定 + 给了草稿
- [ ] 服务器开机 + 跑 docker compose
- [ ] DNS 解析生效（`dig omnirouter.org` 返回服务器 IP）

### Day 14
- [ ] HTTPS 通（`curl -I https://omnirouter.org/health` 200）
- [ ] 备份脚本跑过一次完整流程
- [ ] Lark 告警通了（手动触发一次测试）
- [ ] 律师文书 v1 review 完

### Day 21
- [ ] Stripe 生产模式通过
- [ ] 自己充 ¥1 测试 → 余额到账 → 退款成功
- [ ] Logo 终稿
- [ ] 法律文书 v2 终稿 + 上传后台 + 注册流程展示

### Day 30
- [ ] 软启动 + 30+ 注册
- [ ] 5+ 真实充值
- [ ] 0 P0 故障
- [ ] 第一笔多级返佣触发 + Lark 告警通知到群

### Day 60
- [ ] MAU 200+
- [ ] 月 GMV ¥5k+
- [ ] 30+ 推广素材（V2EX / 知乎 / TG）
- [ ] 第一篇深度博客（"国内开发者用 Claude Code 完整指南"）

### Day 90
- [ ] MAU 500+
- [ ] 月 GMV ¥10k+
- [ ] 决定：全职 ALL-IN / 保持副业 / 招第一个员工
- [ ] 阶段 3 工程开工：anti-fraud / 报表 / OpenTelemetry

---

## 八、为什么需要律师（直接答上一轮的问题）

**因为 AI draft 的法律文书是脚手架不是产品。** 关键风险：

1. **管辖权争议**：中国主体面对海外用户用海外模型，出事在哪打官司？AI 默认可能写"China courts"或"US courts"——错位的话用户告你你赢不了
2. **责任上限的有效性**：我写的是 "max(¥1000, 6mo spend)"——但中国《消费者权益保护法》对部分服务有强制性"惩罚性赔偿"，约定可能被法院判无效
3. **数据跨境合规**：你的用户在中国，数据流到 OpenAI（美国），这构成跨境数据传输——PIPL 第 38-43 条对此有具体的合规要求（个人同意、数据出境合规评估、签 standard contract clauses 等），AI draft 没有覆盖到位
4. **生成式 AI 服务管理暂行办法**（2023 年 8 月生效，国内）：你作为 AI 服务提供者要承担"生成内容真实性、准确性、合法性"的部分责任——AI draft 写"我们不为内容版权背书"在国内可能不合规
5. **退款政策可执行性**：我写的"7 天 50% 已消费不可退"在 EU 可能违反 14 天 cooling-off 期；在加州可能违反消费者权益法。AI 不知道你的目标市场细节
6. **chargeback 防御**：信用卡用户走 Stripe chargeback，你的 TOS 决定能不能反诉成功
7. **企业客户合同**：第一个企业客户的法务一定会改你的 TOS 加 SLA、IP 归属、保密条款——你需要一个会改合同不被对方 dominant 的人
8. **监管问询应对**：如果某天网信办 / 公安局 / FBI 来要数据，你的隐私政策决定你**必须**给什么、**可以拒**什么——给错了赔钱，拒错了被盯上

**律师值多少钱**：¥3-15k 一次 review 防一次 ¥10-100 万的官司或封站。投入产出比超划算。

**找谁**：抖音 / 小红书 / 即刻搜 "SaaS 律师 / 出海律师 / 互联网律师"；Linux DO 上有专门做这块的 DAO；如果 toB，找北京/上海 一线律所中等级别合伙人；如果 toC + 海外，找 SF / NY 的 SaaS startup 律所（U Counsel 等）

---

## 九、把代码、文档、Logo 全部归档到一个地方

GitHub fork：https://github.com/byte-ramble/new-api

关键文件路径：
- 代码改动：commit `ac8ae549` … `c89efe14`（17 commits）
- 设计文档：`omnirouter-docs/observability/`、`omnirouter-docs/operations/`
- 用户文档：`omnirouter-docs/onboarding/`
- 法律模板：`omnirouter-docs/operations/templates/legal/`
- Grafana 看板：`omnirouter-docs/operations/templates/grafana/`
- Logo 资产：`omnirouter-docs/operations/templates/brand/`
- 部署链路：`docker-compose.omnirouter.yml` + `Dockerfile.omnirouter`
- 本文件：`omnirouter-docs/STATUS.md`

启用 OmniRouter 自研功能的 env 清单：
```yaml
- AFFILIATE_ENABLED=true
- AFFILIATE_L1_PCT=15
- AFFILIATE_L2_PCT=5
- AFFILIATE_DAILY_CAP_RMB=1000
- DISABLE_BURST_THRESHOLD=5
- DISABLE_BURST_WINDOW_SEC=60
- LARK_WEBHOOK_URL=https://open.feishu.cn/open-apis/bot/v2/hook/xxx
- LARK_WEBHOOK_SECRET=xxx  # 可选，加签校验时用
- METRICS_ENABLED=true
```

---

## 十、最后的话

**你已经拥有 PackyAPI 80% 的代码 + 2 倍的文档 + 一份给设计师的 brief + 一个 Logo 起点。**

**但 80% 的代码 ≠ 80% 的成功。** 真实世界里：
- 代码 done 不等于产品 done
- 产品 done 不等于上线 done
- 上线 done 不等于有用户
- 有用户不等于赚钱

剩下的事，工程能帮你的不多了。要继续往前走，靠的是：
- 你愿不愿意每周花 15+ 小时
- 你愿不愿意花 ¥10k+ 给律师 / 设计师 / 服务器
- 你愿不愿意忍受 Week 2-3 的等待焦虑（Stripe / 律师 / 备案）
- 你愿不愿意第一次故障凌晨 3 点爬起来处理
- 你愿不愿意把"国内 AI 中转"这件事当成接下来 1-2 年的主线任务

**如果是，照 30 天 roadmap 第一步：今天开始决定主体形态。**

**如果不是，至少把代码 fork ship 到 GitHub 写一篇技术博客**——它是真的能跑的好东西，分享出去对开发者圈子有价值。

🚀 OmniRouter Engineering 完成。Operations 接力。
