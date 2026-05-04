# OmniRouter Logo 设计稿

> 4 个不同方向的 SVG 概念稿。**直接 open 任何 .svg 在浏览器里**，看实际渲染。
>
> 我自己写的 vector，能用，但不是专业设计师水平。挑一个方向后建议：
> - 给 [Logo.com](https://logo.com) AI 生成器或 [Looka](https://looka.com) 喂"OmniRouter, cyberpunk, kawaii router robot"等关键词，迭代
> - 或扔到 Fiverr / 米画师 / 杉果 找设计师精修（¥200-1500）
> - 或直接当 v0 上线，等有营收再升级

---

## 品牌定位（设计师 brief）

| 维度 | 内容 |
|---|---|
| **品牌名** | OmniRouter |
| **slogan 候选** | "One API for 28+ AI Models" / "Pay-as-you-go in CNY, no VPN" |
| **tone** | 友好但专业（开发者向）· 有趣但不轻佻 · 科技但不冷漠 |
| **风格** | 卡通赛博朋克（Cute Cyberpunk）· Edgerunners + Pixar · Astro Boy + 2077 |
| **目标用户** | 中文开发者圈（V2EX/Linux DO/Twitter）+ 国内 AI Code 用户 |
| **必避免** | 严肃企业风（IBM 蓝 / Oracle 红）· 烂大街的 AI 大脑 / 神经网络云图 · 单纯的"API"字母组合 |
| **必传达** | "全模型聚合（Omni）"+"智能路由（Router）"+"中国友好" |

---

## 颜色规范

OmniRouter 主色板（**v0 提议，可调整**）：

| 名字 | 用途 | Hex | RGB | OKLCH |
|---|---|---|---|---|
| **Cyber Cyan** | 主色 / 链接 / 焦点 | `#06B6D4` | `6 182 212` | `oklch(0.71 0.15 209)` |
| **Neon Magenta** | 辅色 / 高亮 / CTA | `#EC4899` | `236 72 153` | `oklch(0.65 0.25 5)` |
| **Hyper Violet** | 渐变中段 | `#A855F7` | `168 85 247` | `oklch(0.62 0.27 304)` |
| **Plasma Yellow** | LED / 装饰点 | `#FBBF24` | `251 191 36` | `oklch(0.84 0.16 87)` |
| **Slate Black** | 深色背景 | `#0F172A` | `15 23 42` | `oklch(0.20 0.04 265)` |
| **Cloud White** | 浅色背景 / 字 | `#F8FAFC` | `248 250 252` | `oklch(0.98 0.00 0)` |

**调色板逻辑**：
- 主辅是经典 Cyberpunk 双色（Cyan + Magenta，Edgerunners 即视感）
- 紫做渐变桥
- 黄做点缀（亮起的 LED 灯）— 让画面有"通电感"
- 黑白做底

---

## 字体规范

OmniRouter 推荐字体（全部免费、Google Fonts 可加载）：

| 用途 | 字体 | 风格特点 |
|---|---|---|
| **品牌名 / Logo wordmark** | [Audiowide](https://fonts.google.com/specimen/Audiowide) | 圆润 + 复古赛博朋克，像 80 年代街机 |
| **Logo wordmark 备选** | [Rajdhani](https://fonts.google.com/specimen/Rajdhani) | 几何 sans，正经一点的科技感 |
| **Logo wordmark 备选 2** | [Orbitron](https://fonts.google.com/specimen/Orbitron) | 标准未来主义 sans |
| **正文** | [Inter](https://fonts.google.com/specimen/Inter) | 项目已用，保持一致 |
| **代码 / 终端文案** | [JetBrains Mono](https://fonts.google.com/specimen/JetBrains+Mono) | 程序员友好 |

我做的 SVG 用 system-safe 字体（monospace / sans-serif），保证在没装专门字体的地方也能渲染。设计师精修时可换成 Audiowide / Rajdhani。

---

## 4 个 Logo 概念稿

> 每个文件都是 200×200（icon mark）。打开方式：
> ```bash
> open omnirouter-docs/operations/templates/brand/concept-a-hexcore.svg
> ```
> 或拖进浏览器，或在线 SVG viewer。

### 🅰️ Concept A · "HexCore" — 几何六边形核心

**关键词**：抽象 / 几何 / 模块化 / 6 路分发

**叙事**：六边形 = 网络拓扑/分子结构 + 6 个外向箭头 = "Omni"全方位分发 + 中心 O 字 = "OmniRouter"。

**优点**：
- 极度可扩展（favicon 16px 也能识别）
- 永远不过时（极简几何）
- 跨语境通用（toC / toB 都不违和）

**缺点**：
- 不够"卡通"，缺记忆点角色
- 跟 Hexagon Labs / 各种 hex logo 撞型可能性

**文件**：[`concept-a-hexcore.svg`](./concept-a-hexcore.svg)

---

### 🅱️ Concept B · "OmniBot" — 卡通赛博机器人头 ⭐推荐

**关键词**：卡通 / 角色 / 赛博朋克可爱 / 记忆点最强

**叙事**：一只小机器人头，单条 LED 视觉条 + 萌萌大眼 + 双天线（router antenna）+ 侧面 4 个 LED 端口（路由器接口暗示）+ 嘴角微笑 + 胸前小 OR 标。

**优点**：
- **最有记忆点** — 角色比抽象 logo 在用户脑海留存 5-10x
- 卡通 + 赛博朋克的最佳载体
- 衍生周边友好（贴纸 / 表情包 / TG 频道吉祥物）
- 国内开发者圈（特别是 V2EX/Linux DO 文化）天然爱角色 mascot

**缺点**:
- 16px 缩小后细节损失
- 太可爱可能伤"专业感"（toB 大客户可能皱眉）— 解决：保留 OmniBot 做 mascot，正式合同/商务用 wordmark

**文件**：[`concept-b-omnibot.svg`](./concept-b-omnibot.svg)

---

### 🅲️ Concept C · "OR Monogram" — 字母组合

**关键词**：字母 / 简洁 / 商务

**叙事**：字母 O 和 R 的组合，渐变填充。最干净的方案。

**优点**:
- 最稳，永远不会"看着违和"
- 最小空间（favicon、应用图标）效果最好
- toB 友好

**缺点**:
- 最不有趣
- 太多公司用字母 logo，记忆点低

**文件**：[`concept-c-monogram.svg`](./concept-c-monogram.svg)

---

### 🅳️ Concept D · "Neon Loop" — 霓虹无限环

**关键词**：符号 / 抽象 / 大胆 / 视觉冲击

**叙事**：∞ 无限环 = "Omni"无限多模型 + 中间交叉点 = "Router"路由切换 + 霓虹双色 = 赛博朋克。

**优点**:
- 视觉冲击最强
- 概念准确（infinity = "Omni"，crossing = "Router"）
- 适合做动画 logo（loop 转动）

**缺点**:
- 抽象，需要解释才能 get
- 跟很多带无限符号的 AI 产品撞型

**文件**：[`concept-d-neonloop.svg`](./concept-d-neonloop.svg)

---

## Wordmark（字标）

横向品牌字 — header / 名片 / 文档头都要用：

[`wordmark.svg`](./wordmark.svg) — "OmniRouter" + 终端光标点缀

[`lockup-horizontal.svg`](./lockup-horizontal.svg) — Concept B 卡通头 + Wordmark 横排组合

---

## 我的推荐

**主推：Concept B (OmniBot)** + Wordmark 组合。

理由：
1. 你的目标用户（V2EX/Twitter/Linux DO 开发者）天然爱角色 mascot
2. PackyAPI 没有角色 logo（它就是一个文字 + 简单图标），你做角色直接形成视觉差异
3. 角色 logo 可以衍生：表情包、TG sticker、404 页吉祥物、产品文档插画 — 一个 logo 多重用
4. 万一 toB 客户嫌可爱，wordmark 单独用即可（双轨制）

**备选：Concept A (HexCore)** — 如果你想稳一点不冒险。

---

## 怎么进一步精修

### 路径 1（最便宜，¥0-300）
1. 把 Concept B 的 SVG 上传到 [Vectornator](https://www.linearity.io) / [Figma](https://figma.com) 免费版
2. 自己微调（颜色、比例、表情）
3. 导出 PNG / ICO 各尺寸

### 路径 2（中等，¥200-1500）
1. 在 [Fiverr](https://fiverr.com) 搜 "cyberpunk mascot logo design"
2. 给设计师看本 README + Concept B 草图
3. 让他出 3-5 个变体
4. 选一个 + 1-2 轮迭代

### 路径 3（专业，¥3k-15k）
1. 找国内独立设计师（米画师、Behance、即刻 #设计 标签）
2. 完整 brand identity（logo + 字体 + 颜色 + UI 规范）
3. 适合长期 SaaS 品牌投入

### 路径 4（AI 生成器辅助）
喂以下 prompt 给 [Logo.com](https://logo.com) / [Looka](https://looka.com) / Midjourney / DALL-E：

```
Logo for "OmniRouter", an AI API gateway service.
Style: kawaii cyberpunk mascot, like a friendly cyberpunk router robot
Mood: tech but cute, futuristic but approachable
Colors: cyber cyan #06B6D4 + neon magenta #EC4899 + dark slate #0F172A
Elements: rounded robot head, single wide LED visor band with cute eyes,
  two antenna ears with glowing tips, side LED ports suggesting network ports
Output: vector logo on transparent background, both icon-only and with wordmark
Avoid: stereotypical AI brain/neural-network imagery, generic cloud icons,
  serious corporate style, beige/brown/grey palettes
```

---

## 使用指南

### 文件用法
| 场合 | 用哪个 |
|---|---|
| Favicon (16x16, 32x32) | Concept A 或 C（最简洁） |
| App icon (512x512) | Concept B（角色，最 striking）|
| Header / 网站头部 | `lockup-horizontal.svg`（图+字组合）|
| 名片 / PPT | Wordmark only |
| TG 频道头像 | Concept B |
| 文档侧边栏 | Concept A 或 C |
| Loading 动画 | Concept D（无限环天然适合 loop 旋转）|
| 404 / Empty state 插画 | Concept B 表情变体 |

### 间距规范
- Logo 周围至少留 0.25× 高度的 padding
- 永远不要把 logo 拉伸（保持 aspect ratio）
- 永远不要换色（除非有官方深色 / 浅色变体）

### 不要做的事
- ❌ 给 logo 加 drop shadow / 浮雕 / 各种 PS 滤镜
- ❌ 把 logo 放在 busy 背景上（一定要纯色或低对比底）
- ❌ 缩小到 16px 以下（找替代版）
- ❌ 在 Logo 旁加营销文案（slogan 应单独排版）

---

## License / 来源声明

这 5 个 SVG 是 AI（Claude Sonnet）为你直接生成的代码。
**版权归你**，可商用、可修改、可商标注册。
请注意 `Audiowide` / `Rajdhani` 等 Google Fonts 是 SIL 开源协议，免费商用。

如果你真的去注册商标（强烈建议在中国国家知识产权局商标局注册一下，¥600 + 1 年），先去 [中国商标网](http://sbj.cnipa.gov.cn) 搜 "OmniRouter"、"全路由"、"奥米路由" 等，确保没冲突。
