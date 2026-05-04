# 法律文书模板使用说明

> ⚠️ **重要免责声明**
>
> 本目录下的所有文档（用户协议 / 隐私政策 / 退款政策 / 公平使用）都是
> **AI 生成的起步草稿**，**不构成法律建议**，**不能直接发布**。
>
> 它们的目的是帮你跟律师对话时有具体起点，而不是替代律师。
>
> ## 必须做的事
>
> 1. **找一位执业律师 review**（特别是涉及跨境数据 / 支付 / AI 服务的律所）
> 2. **针对你的实际主体形态调整**（个人 vs 公司 vs 海外公司，措辞差异极大）
> 3. **针对目标用户所在司法管辖区调整**：
>    - 中国大陆用户 → PIPL（个人信息保护法）+ 《生成式人工智能服务管理暂行办法》
>    - 欧洲用户 → GDPR
>    - 加州用户 → CCPA
>    - 其它 → 各地数据保护法规
> 4. **填上具体信息**（公司名、注册地址、联系方式、生效日期）
> 5. **保留版本号 + 历史记录**（任何修改都给版本号 + 留档）
>
> ## 模板里的占位符
>
> - `{{COMPANY}}` — 公司全称
> - `{{LEGAL_ENTITY}}` — 法律主体（个人 or 公司）
> - `{{DOMAIN}}` — omnirouter.org
> - `{{SUPPORT_EMAIL}}` — hi@omnirouter.org
> - `{{ICP_NUMBER}}` — ICP 备案号（仅国内）
> - `{{EFFECTIVE_DATE}}` — 生效日期 YYYY-MM-DD
> - `{{JURISDICTION}}` — 管辖法院 / 仲裁地
>
> grep 替换即可。

## 文件清单

| 文件 | 用途 | 上线必需 |
|---|---|---|
| [tos.md](./tos.md) | 用户协议 / Terms of Service | **是** |
| [privacy.md](./privacy.md) | 隐私政策 / Privacy Policy | **是** |
| [refund.md](./refund.md) | 退款政策 / Refund Policy | **是** |
| [fair-use.md](./fair-use.md) | 公平使用条款 / Fair Use | **是**（特别是 Codex 包月） |

## 部署到 OmniRouter 站点

new-api 后台支持运行时配置法律文档（admin → 设置 → 法律设置）。流程：

1. 用 [Markdown to HTML 工具](https://markdowntohtml.com/) 把 .md 转 HTML
2. admin 后台粘贴 HTML 到对应字段：
   - 用户协议 → `UserAgreement` 字段
   - 隐私政策 → `PrivacyPolicy` 字段
3. 注册流程会自动展示链接 + 强制勾选同意（参考 `controller/user.go::Register`）
4. 改动后留个 changelog（`omnirouter-docs/operations/templates/legal/CHANGELOG.md`）

## 法务清单（律师 review 时给他看）

- [ ] 主体形态是什么？（影响 TOS 的"提供方"段）
- [ ] 收款主体是哪个？（影响退款政策）
- [ ] 用户数据存在哪？（影响隐私政策的存储位置披露）
- [ ] 上游 LLM 厂商是否会看到用户 prompt？（必须披露！）
- [ ] 公司是否做内容审核？标准是什么？
- [ ] 责任上限怎么定？（API 故障导致用户损失的兜底）
- [ ] 跨境数据流是否合规？（中国主体 → 海外 OpenAI 调用）
- [ ] 是否合规处理未成年人数据？（建议禁止未成年人使用）
- [ ] 仲裁条款 vs 法院诉讼？
- [ ] 知识产权归属（用户输入 prompt + 模型输出归谁）
