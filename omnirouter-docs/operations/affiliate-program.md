# 多级邀请返佣（Affiliate Program）

## 背景

new-api 上游已经有简单的邀请系统（`User.AffCode` / `InviterId` / `AffQuota` / `AffHistoryQuota`），但只是"被邀请人**注册** → 邀请人拿固定 quota 奖励"。这套对个人开发者圈子的轻拉新够用，对商业 SaaS **拉新引擎**不够：

| 维度 | 上游已有 | OmniRouter 需要 |
|---|---|---|
| 触发时点 | **注册即奖** | **每次充值** 都要分润 |
| 计费方式 | 固定额度 | **按充值金额百分比** |
| 层级 | 1 级 | **2 级**（朋友拉朋友也分润） |
| 奖励单位 | quota（站内余额） | **现金 RMB**（可外部提现） |
| 提现 | 仅站内 quota → balance 转账 | **支付宝 / 微信 / 银行卡 / 余额** 多通道 |
| 风控 | 无 | 自邀检测、24h 累计上限、最低提现 |
| 审计 | 仅 `User.AffHistoryQuota` 累加 | **完整 ledger**（每笔事件 + 反向追溯） |

**两套并存**：保留上游 AffQuota（注册赠送），新增 OmniRouter 多级返佣（充值分润）。
- AffQuota = "谢谢拉新，给你 $1 试用"
- Commission = "TA 这个月充了 ¥100？你拿 ¥15。下个月还会有。" — 经常性收入，留存威力大很多

## 数据模型

3 张新表（`model/`），全部走 GORM AutoMigrate（兼容 SQLite/MySQL/PostgreSQL）。

### `affiliate_accounts` — 用户佣金账户摘要（denormalized）

| 字段 | 类型 | 说明 |
|---|---|---|
| `user_id` | int (PK) | 关联 users.id |
| `balance_rmb` | float64 | **可提现余额**（RMB） |
| `total_earned_rmb` | float64 | 历史累计赚到 |
| `total_withdrawn_rmb` | float64 | 历史累计提现 |
| `last_earned_at` | int64 | 最后一次入账时间 |

为什么 denormalize：每次刷"我的推广"页都 `SUM(commission_logs)` 太重；这张表是 hot read。
**真值在 `commission_logs`**，可以从中重建。

### `commission_logs` — 每笔佣金事件（source of truth）

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | int (PK) | |
| `earner_id` | int (idx) | 收佣人 user.id |
| `source_user_id` | int (idx) | 触发充值的 user.id |
| `level` | int | 1 (直邀) / 2 (二级) |
| `topup_amount_rmb` | float64 | 触发充值金额 |
| `commission_rmb` | float64 | 实际入账 |
| `rate_pct` | float64 | 入账时的费率（防止后续改设置追溯影响） |
| `status` | string | `paid` / `frozen` / `reversed` |
| `ref_id` | *int | 反向冲销时指向原 log |
| `topup_channel` | string | `epay` / `stripe` / `creem` / ... |
| `note` | string | 自由文本 |
| `created_at` | int64 (idx) | |

### `withdrawals` — 提现申请

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | int (PK) | |
| `user_id` | int (idx) | |
| `amount_rmb` | float64 | 申请金额（毛） |
| `fee_rmb` | float64 | 平台手续费 |
| `net_rmb` | float64 | 实际到账（amount - fee） |
| `method` | string | `alipay` / `wechat` / `bank` / `balance` |
| `account` | string | 收款账号 |
| `status` | string (idx) | `pending` / `approved` / `rejected` / `reversed` |
| `user_note` / `admin_note` | string | 双向备注 |
| `processed_by` / `processed_at` | int / int64 | admin 审核记录 |
| `created_at` / `updated_at` | int64 | |

### 状态机

**Withdrawal**:
```
pending ──admin approve──→ approved (terminal)
pending ──admin reject──→ rejected (refunds locked balance, terminal)
approved ──admin reverse──→ reversed (rare, e.g. payout bounced)
```

**CommissionLog**:
```
paid (default on insert)
paid ──fraud detected──→ frozen
paid ──refund/reversal──→ reversed (creates new log row with negative amount, ref_id → original)
```

## 配置

`setting/operation_setting/affiliate_setting.go`：

| Key | Env override | Default | 说明 |
|---|---|---|---|
| `Enabled` | `AFFILIATE_ENABLED` | **false** | 主开关 |
| `Level1RatePct` | `AFFILIATE_L1_PCT` | 15.0 | L1 直邀返佣比例（%） |
| `Level2RatePct` | `AFFILIATE_L2_PCT` | 5.0 | L2 二级返佣比例（%）；设 0 = 单级模式 |
| `MaxDailyCommissionRmb` | `AFFILIATE_DAILY_CAP_RMB` | 1000 | 单人 24h 累计佣金上限（防刷）；0 = 不限 |
| `MinWithdrawalRmb` | `AFFILIATE_MIN_WITHDRAWAL_RMB` | 100 | 最低单笔提现 |
| `WithdrawalFeeRmb` | — | 0 | 平台收的手续费 |
| `MaxWithdrawalsPerDay` | — | 3 | 单人每日提现申请上限 |
| `AllowSelfInvite` | — | false | 自邀（同一人/IP/设备）是否给佣金 |

**默认 disabled**——operator 阅读本文档 + 法律文书后通过 env 开启。

## 分润流程

### 触发

任意充值通道成功 → 调 `service.PayCommission(userId, topupRmb, channel)`。已 wire 的 callsite：
- ✅ `controller/topup.go::EpayCallback` (易支付) — line ~398

待 wire（next turn）：
- ⏳ `controller/topup_creem.go::CreemWebhook`
- ⏳ `controller/topup_stripe.go::StripeWebhook`
- ⏳ `controller/topup_waffo.go::WaffoWebhook`
- ⏳ `controller/topup_waffo_pancake.go::WaffoPancakeWebhook`

注：兑换码（`controller/user.go::TopUp`）**不接** commission——兑换码本身可能是邀请奖励，避免循环刷。

### 算法

```
PayCommission(userId, topupRmb, channel):
  1. 读 cfg = AffiliateSetting; if !Enabled || topupRmb <= 0: return
  2. 读 user = GetUserById(userId); if user.InviterId == 0: return
  3. 读 l1 = GetUserById(user.InviterId)
     if cfg.Level2RatePct > 0 && l1.InviterId 合法:
       l2InviterId = l1.InviterId
  4. 在 tx 内:
     a. payOneLevel(L1=user.InviterId, topupRmb × L1Pct/100)
     b. if l2InviterId > 0: payOneLevel(L2=l2InviterId, topupRmb × L2Pct/100)

  payOneLevel:
    1. commission = round2(topupRmb × ratePct / 100)
    2. 查 SumCommissionPaidToday(earnerId); 若 + commission > MaxDailyCommissionRmb: 跳过 + log
    3. GetOrCreateAffiliateAccount(earnerId)
    4. CreditAffiliateAccount(earnerId, commission) → balance += commission
    5. InsertCommissionLog(earnerId, sourceUserId, level, topup, commission, rate, channel, status='paid')
```

**关键设计决定**：
- **读出 tx，写入 tx**：`GetUserById` 在 tx 外做（gorm.Transaction 持有独占连接，SQLite 测试环境只 1 conn → 嵌套读会自死锁）
- **L2 失败不影响 L1**：L2 出错时 log + return nil 保留 L1 入账；不让 L2 错把 L1 也回滚
- **rate 入 ledger**：commission_log 存当时的 rate_pct，不存只引用 settings——避免后续改费率追溯影响历史
- **round2 用 math.Round**：float64 不能精确表示所有小数（如 1.005），最差 ¥0.01 漂移；如要精确建议换 shopspring/decimal

### 提现流程

```
RequestWithdrawal(userId, amount, method, account, note):
  validation: enabled / amount > 0 / >= min / method in {alipay,wechat,bank,balance} / account != ""
  tx:
    DebitAffiliateAccountForWithdrawal(userId, amount)   // balance -= amount，conditional update：not enough → error
    InsertWithdrawal(pending row)
  return Withdrawal

ApproveWithdrawal(id, adminId, note):
  MarkWithdrawalApproved(id, adminId, note)              // pending → approved，admin 自己去外部支付
  // balance 已经 debit 过，不用再动

RejectWithdrawal(id, adminId, note):
  tx:
    MarkWithdrawalRejected(id, adminId, note)            // pending → rejected
    RefundAffiliateAccount(userId, amount)               // balance += amount，total_withdrawn -= amount
```

## 风控

| 规则 | 实现位置 | 说明 |
|---|---|---|
| 24h 累计上限 | `payOneLevel` | `SumCommissionPaidToday + commission > cap → 跳过` |
| 自邀拒绝 | `PayCommission` | `InviterId == userId && !AllowSelfInvite → 跳过` |
| 提现最低额 | `RequestWithdrawal` | `amount < MinWithdrawalRmb → 错误` |
| 提现金额方法白名单 | `RequestWithdrawal` | 4 种合法 method，其它拒 |
| 余额不足 | `DebitAffiliateAccountForWithdrawal` | conditional UPDATE `WHERE balance >= amount`，0 row affected = 不够 |
| 重复审核 | `MarkWithdrawalApproved/Rejected` | `WHERE status = 'pending'`，已处理的拒掉 |

**待补**（Next turn 或 Phase 3）：
- IP / 设备指纹去重（注册时检测同 IP 多账号）
- 邀请人短时间集中充值小额检测（典型刷单模式）
- 退款时联动反冲 commission（topup 退款 → commission_log 写 reversed 行）

## 测试

`service/affiliate_test.go` 覆盖 7 个核心场景：

| 测试 | 验证 |
|---|---|
| `TestRound2` | 浮点取整正确性（避开 1.005 类不可表示边界） |
| `TestPayCommission_Disabled` | 主开关 off 时 no-op |
| `TestPayCommission_NoInviter` | 无邀请人时 no-op |
| `TestPayCommission_L1AndL2HappyPath` | 三层链 grand→parent→child；child 充 ¥100 → parent +¥15、grand +¥5 |
| `TestPayCommission_DailyCap` | 第 2 笔会让 earner 超 cap 时 silent skip |
| `TestRequestWithdrawal_Validation` | 4 种验证错（金额、最低、方法、账户） |
| `TestRequestWithdrawal_HappyPathLocksBalance` | balance 正确锁定 + 拒绝后正确 refund |

```bash
go test ./service/ -run "TestRound2|TestPayCommission|TestRequestWithdrawal" -v -count=1
```

7/7 PASS。

## 上线 checklist

- [ ] env 设 `AFFILIATE_ENABLED=true`、`AFFILIATE_L1_PCT=15`、`AFFILIATE_L2_PCT=5`
- [ ] env 设 `AFFILIATE_DAILY_CAP_RMB=1000`、`AFFILIATE_MIN_WITHDRAWAL_RMB=100`
- [ ] **法律**：`templates/legal/tos.md` 加邀请返佣条款，`templates/legal/refund.md` 说明 commission 不退
- [ ] **运营文案**：在 brand-copy.md 推广素材里补"邀请赚 15% 佣金"卖点
- [ ] **API**：等待 next turn 完成 `/api/affiliate/{overview,log,withdrawal}` + admin endpoints
- [ ] **前端**：等待 next turn 完成"我的推广"页 + admin 审核页
- [ ] **wire 其它充值通道**：等待 next turn 完成 Stripe/Creem/Waffo 4 个 webhook
- [ ] 灰度先开 0.5% L1 + 0% L2 跑一周，观察 balance 增长曲线再放阈值
- [ ] Grafana 加 panel：`sum(rate(commission_logs.commission_rmb[1h]))` 监控总分润趋势
- [ ] 设 Lark 告警：单日新增 withdrawal request > 阈值时通知 admin

## 相关文件

| 文件 | 用途 |
|---|---|
| [`../../model/affiliate_account.go`](../../model/affiliate_account.go) | 账户摘要表 |
| [`../../model/commission_log.go`](../../model/commission_log.go) | 事件 ledger |
| [`../../model/withdrawal.go`](../../model/withdrawal.go) | 提现申请 |
| [`../../setting/operation_setting/affiliate_setting.go`](../../setting/operation_setting/affiliate_setting.go) | 配置 |
| [`../../service/affiliate.go`](../../service/affiliate.go) | 业务核心 |
| [`../../service/affiliate_test.go`](../../service/affiliate_test.go) | 单元测试 |
| [`../../controller/topup.go`](../../controller/topup.go) | epay callsite (其余 next turn) |

## 后续迭代

按优先级：

1. **wire 剩余充值通道**（Stripe / Creem / Waffo / Waffo-Pancake）
2. **API 端点**：用户 `/api/affiliate/{overview,log,withdrawal}` + admin `/api/admin/withdrawal`
3. **前端页**：用户中心 → 我的推广（含邀请链接生成 + 二维码 + 排行榜可选）
4. **admin 审核页**：pending 队列 + 一键 approve/reject + 备注
5. **退款联动**：topup refund → 自动写 commission `reversed` 行 + 反扣 balance
6. **IP/设备指纹**：注册时收集 + 同指纹邀请关系标 frozen 状态
7. **多级排行**：邀请数 top 10 / 累计赚 top 10 leaderboard（运营推广用）
