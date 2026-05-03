# OmniRouter 品牌追加配置指南

## ⚠️ Rule 5 合规声明

按 [CLAUDE.md Rule 5](../../CLAUDE.md)，本项目对以下信息**严格保护、不得删除或替换**：

- **nеw-аρi**（项目名/标识）
- **QuаntumΝоuѕ**（组织/作者标识）

包括：README、license header、HTML title、meta tags、footer text、Go module 路径、Docker 镜像名、注释等。

> **本指南只追加 OmniRouter 品牌信息，不修改任何 new-api / QuantumNous 原标识。**
>
> 具体策略：
> - HTML `<title>New API</title>` **保留不动**（位于 `web/default/index.html`）
> - `web/default/public/logo.png` **保留不动**（项目原 logo 资源）
> - Logo URL 通过运行时配置指向 OmniRouter 自有 logo 文件（不覆盖 `logo.png`）
> - Footer 通过运行时配置展示 OmniRouter 品牌 + "Powered by [new-api](https://github.com/QuantumNous/new-api) by QuantumNous" 双品牌

## 配置项一览

| OptionMap key | 值 | 来源 |
|---|---|---|
| `SystemName` | `OmniRouter` | DB option / admin 后台 |
| `Logo` | `/static/omnirouter-logo.png`（自备） | DB option / admin 后台 |
| `Footer` | 见 [footer.html](#footer-html-template) 段 | DB option / admin 后台 |
| `HomePageContent` | OmniRouter 首页文案（可选） | DB option / admin 后台 |
| `About` | 关于页（可选） | DB option / admin 后台 |
| `ServerAddress` | `https://omnirouter.org` | DB option / admin 后台 |

> 注意：上述配置全部通过 admin 后台或 SQL 设置，**不动 Go 代码**，因此不会影响上游 rebase。

## 配置方法（三选一）

### 方法 A：admin 后台手动配（推荐用于第一次跑通）

1. 启动 new-api（`docker compose up -d`）
2. 浏览器访问 http://localhost:3000，用 root 账号登录
3. 进入 **管理 → 设置 → 系统设置**
4. 逐项填入下表的值，保存
5. 刷新页面，可见 OmniRouter 品牌生效

### 方法 B：SQL 种子脚本（推荐用于一键部署）

执行 [brand-seed.sql](./brand-seed.sql)（按你的数据库选对应段）。这会插入或更新 `options` 表。
适合：CI/CD 脚本、Terraform、Ansible、首次部署的 init container。

### 方法 C：在 Postgres docker init dir 自动跑

把 `brand-seed.sql` 挂载到 `/docker-entrypoint-initdb.d/`：
```yaml
postgres:
  volumes:
    - ./omnirouter-docs/operations/brand-seed.sql:/docker-entrypoint-initdb.d/99-omnirouter-brand.sql:ro
```

注意：Postgres 只在**首次启动**（数据目录为空时）执行 init scripts，已有数据库不会再执行。

## Footer HTML Template

把以下 HTML 粘到 admin 后台的 Footer 字段（或写到 SQL 种子里）：

```html
<div style="text-align:center; padding:8px; font-size:12px; color:#777;">
  <strong>OmniRouter</strong> · <a href="https://omnirouter.org" target="_blank" rel="noopener">omnirouter.org</a>
  &nbsp;·&nbsp;
  Powered by <a href="https://github.com/QuantumNous/new-api" target="_blank" rel="noopener">new-api</a> by QuantumNous
</div>
```

✅ **保留** `new-api` 与 `QuantumNous` 链接（含 GitHub 跳转）→ Rule 5 合规
✅ **追加** `OmniRouter` 品牌 + 主域名

## Logo 资源准备

把 OmniRouter logo（建议 PNG，512×512，透明底）放到以下任一位置：

### 选项 1：宿主机静态目录（推荐）

```bash
# 把 OmniRouter logo 放到 docker-compose 挂载的 ./data/ 目录
cp omnirouter-logo.png ./data/omnirouter-logo.png
```

然后给 nginx / 反代加路径转发，把 `https://omnirouter.org/static/omnirouter-logo.png` 转到 `./data/omnirouter-logo.png`。

### 选项 2：CDN / 对象存储（推荐生产）

上传到你的 OSS / S3 / Cloudflare R2，把公开 URL 填到 `Logo` 配置项：
```
Logo = https://cdn.omnirouter.org/brand/logo.png
```

### 不推荐

❌ **不要覆盖** `web/default/public/logo.png` —— 这是 new-api 项目自带资源，覆盖会破坏 Rule 5 合规。

## 浏览器 Tab 标题问题

`web/default/index.html` 的 `<title>New API</title>` 是首屏渲染前的 HTML 标题，**保留不变**（Rule 5）。
React 应用启动后会读取 `SystemName=OmniRouter` 动态 setDocumentTitle，用户实际看到的页面标题会变成 OmniRouter（首屏可能短暂闪现 "New API"，0.1-0.5 秒，可忽略）。

如果你介意首屏闪现，可以在 ingress / 反代层面做 HTML rewrite：
```nginx
sub_filter '<title>New API</title>' '<title>OmniRouter</title>';
sub_filter_once on;
```

> 这种方式 **不动源码** → 不破坏 Rule 5。

## 验证清单

部署 + 配置完成后：
- [ ] 访问 `https://omnirouter.org`，登录页 logo 是 OmniRouter logo
- [ ] 页面 Footer 同时出现 OmniRouter 与 "Powered by new-api by QuantumNous"
- [ ] 浏览器 Tab 标题（React 加载后）显示 OmniRouter
- [ ] Stripe 充值页商品名以 OmniRouter 展示（需要单独在 [Stripe Dashboard](https://dashboard.stripe.com) 改）
- [ ] Lark 告警卡片标题以 OmniRouter 出现（已在 [`service/lark_notify.go`](../../service/lark_notify.go) 默认值 "OmniRouter Notification"）

## 后续品牌资产清单（不在本次 MVP）

- [ ] 邮件模板品牌化（验证邮件、密码重置邮件，新增模板覆盖默认）
- [ ] PWA manifest 配置 OmniRouter 应用名
- [ ] 微信公众号 / 客服二维码集成（runtime 配置）
- [ ] 用户协议 / 隐私政策 OmniRouter 主体（admin 后台 → 法务设置）
- [ ] 关于页（About）OmniRouter 介绍
- [ ] 自定义首页（HomePageContent）OmniRouter 营销文案

## 相关文件

- [brand-seed.sql](./brand-seed.sql) — SQL 种子脚本
- [`controller/setup.go`](../../controller/setup.go) — 首次安装引导
- [`model/option.go`](../../model/option.go) — Option 模型与读写
- [`common/constants.go`](../../common/constants.go) — `SystemName`/`Logo`/`Footer` 默认值
