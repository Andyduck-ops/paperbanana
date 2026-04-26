# Web Quality Audit — PaperBanana Frontend

**Date**: 2026-04-25  
**Scope**: `paperbanana-clean/web/` — React + Vite + Tailwind CSS 前端  
**Standards**: Google Lighthouse (Performance, Accessibility, SEO, Best Practices)

---

## 严重问题 (4 处)

### [CRITICAL] Security: 生产环境暴露源码映射 (Source Maps)
- **File**: `web/vite.config.ts:43` — `sourcemap: true`
- **Impact**: 任何人可查看未压缩的原始源代码，泄露 API 密钥、业务逻辑和安全漏洞。
- **Fix**: 生产构建设置 `sourcemap: false`，或条件启用：`sourcemap: process.env.NODE_ENV !== 'production'`

### [CRITICAL] Security: Nginx 缺少安全响应头
- **File**: `nginx.conf` — 全局缺失
- **Impact**: 无 XSS 防护 (CSP)、无点击劫持防护 (X-Frame-Options)、无 MIME 嗅探防护 (X-Content-Type-Options)、无 HTTPS 强制 (HSTS)。
- **Fix**: 添加以下响应头：
```nginx
add_header Content-Security-Policy "default-src 'self'; script-src 'self'; style-src 'self' https://fonts.googleapis.com; font-src https://fonts.gstatic.com; img-src 'self' data: https:; connect-src 'self'";
add_header X-Content-Type-Options "nosniff";
add_header X-Frame-Options "DENY";
add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
add_header Referrer-Policy "strict-origin-when-cross-origin";
```

### [CRITICAL] Security: Nginx 仅监听 HTTP (80 端口)
- **File**: `nginx.conf:2` — `listen 80;`
- **Impact**: 所有流量明文传输，API 密钥等敏感数据可能被拦截。
- **Fix**: 添加 443 端口监听并配置 SSL 证书，将所有 HTTP 重定向至 HTTPS。

### [CRITICAL] Accessibility: 缺少 "跳转到主内容" 链接
- **File**: 全局缺失 (应在 `index.html` 或 `App.tsx` 中添加)
- **Impact**: 键盘用户必须逐项遍历所有导航元素才能到达主工作区。
- **Fix**: 在 `<body>` 头部 `<Header>` 之前添加：
```tsx
<a href="#main-content" className="skip-link">Skip to main content</a>
```

---

## 高优先级 (3 处)

### [HIGH] Performance: Google Fonts 未使用 `font-display: swap`
- **File**: `web/index.html:14`
- **Impact**: 字体加载完成前文字不可见 (FOIT)，移动端/慢速连接下 LCP 可能超过 2.5s。
- **Fix**: 在 Google Fonts URL 中添加 `&display=swap`；建议精简到 1-2 个必需字体族。

### [HIGH] Accessibility: `Input` 和 `Select` 组件 `<label>` 未关联
- **File**: `web/src/components/atoms/Input.tsx:45`, `web/src/components/atoms/Select.tsx:49`
- **Impact**: 屏幕阅读器无法将标签与表单控件关联。
- **Fix**: 添加 `htmlFor` 与 input/select 的 `id` 匹配。

### [HIGH] SEO: 缺少 `<meta name="description">`
- **File**: `web/index.html`
- **Impact**: 搜索引擎无描述文字可展示。
- **Fix**: 添加 `<meta name="description" content="PaperBanana — AI-powered academic illustration generation." />`

---

## 中优先级 (5 处)

### [MEDIUM] SEO: 标题标签过短
- **File**: `web/index.html:7` — `<title>PaperBanana</title>`
- **Fix**: 改为 `PaperBanana — AI Academic Illustration Generator`

### [MEDIUM] SEO: 缺少 Open Graph / Twitter Card 标签
- **File**: `web/index.html`
- **Fix**: 添加 `og:title`, `og:description`, `og:image`, `twitter:card`

### [MEDIUM] Accessibility: WelcomeWizard 模态框缺少 ARIA 属性
- **File**: `web/src/components/WelcomeWizard.tsx:93`
- **Fix**: 添加 `role="dialog" aria-modal="true"`，并使用已有的 `useFocusTrap` hook。

### [MEDIUM] Performance: 静态资源缺少缓存控制头
- **File**: `nginx.conf`
- **Fix**: 为带哈希资源添加 `Cache-Control: public, immutable`。

### [MEDIUM] Performance: 5个字体族全部同步加载
- **File**: `web/index.html:14`
- **Fix**: 精简到项目中实际使用的 1-2 个字体族。

---

## 低优先级 (4 处)

### [LOW] Accessibility: 错误信息未用 `aria-describedby` 关联输入框
- **File**: `web/src/components/atoms/Input.tsx:68`
- **Fix**: 在 `<input>` 上添加 `aria-describedby={errorId}`, 在错误 `<p>` 上添加 `id={errorId}`。

### [LOW] Accessibility: 语言切换后 `lang` 属性未更新
- **File**: `index.html:2` — 固定 `lang="en"`
- **Fix**: 在 `useLanguage` hook 中动态设置 `document.documentElement.lang`。

### [LOW] Accessibility: ImageUpload 上传按钮缺少 `aria-label`
- **File**: `web/src/components/ImageUpload.tsx:106`
- **Fix**: 在按钮元素上添加 `aria-label={t('refine.dropImage')}`。

### [LOW] SEO: 缺少 canonical URL 和 JSON-LD 结构化数据
- **Fix**: 按需添加（如果是营销落地页则更关键）。

---

## 总结

| 类别 | 总数 | 严重 | 高 | 中 | 低 |
|------|------|------|-----|-----|-----|
| Performance | 5 | 0 | 1 | 3 | 1 |
| Accessibility | 5 | 1 | 1 | 1 | 2 |
| SEO | 4 | 0 | 1 | 2 | 1 |
| Best Practices/Security | 3 | 3 | 0 | 0 | 0 |

### 修复优先级

1. **立即修复** — 安全响应头 + 禁用生产环境 source maps + 添加 HTTPS
2. **发布前修复** — 字体 `font-display: swap` + label 关联 + skip link
3. **本次迭代内修复** — meta description, OG 标签, 缓存头, 模态框 focus trap
4. **方便时修复** — `aria-describedby`, `lang` 属性, `aria-label`
