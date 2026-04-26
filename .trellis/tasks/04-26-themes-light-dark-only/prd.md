# PRD — Converge themes to claude-light + linear-dark

> **Trellis 任务**：`04-26-themes-light-dark-only`
> **分支**：`UI`
> **作者**：sikjmyhre
> **日期**：2026-04-26

---

## 1. Motivation（为什么做）

当前主题膨胀到 **14 个 css 文件**（academic / art-deco / base / bauhaus / classical-chinese / japanese-bw / minimalist-bw / neo-minimal / pop-anime / pop-art / pop-art-dark / qi-baishi / rococo / swiss / workspace），且：

- Header 顶部铺满 14 色块切换器，违反 [`design-principles.md` §1.3 / §5.2 / §6.1`](../../spec/ui-design/design-principles.md)（首屏暴露原则、Less is More、首屏布局）。
- `base.css:57–80` 有 CSS 嵌套 bug，`body { ... .skip-link { ... } }` 编译后 `.skip-link` 不在顶层。
- `academic.css` 主色用了冷调蓝 `oklch(0.42 0.12 255)`，违反 [`claude-light.md` §7 don'ts](../../spec/ui-design/references/claude-light.md)「不能引入冷蓝灰」。
- 14 主题中只有少数对得上视觉锚点（Claude 浅色 / Linear 暗色），其它在两套体系之间「折中」，违反 [`design-principles.md` §0.2`](../../spec/ui-design/design-principles.md)。

**结论**：塌成「**1 浅 + 1 暗**」。浅色基底 = `claude-light`（继承 [`claude-light.md`](../../spec/ui-design/references/claude-light.md)），暗色基底 = `linear-dark`（继承 [`linear-dark.md`](../../spec/ui-design/references/linear-dark.md)）。

---

## 2. Goals（必须达成）

- **G1**：`web/src/themes/` 只保留 `claude-light.css` 和 `linear-dark.css`，外加 `tokens.css`（替代 `base.css`，纯 token bridge + global reset，无嵌套 bug）。
- **G2**：浅色实现严格对齐 `claude-light.md` §2 §3 §6（Parchment `#f5f4ed`、Anthropic Serif fallback、Terracotta `#c96442` CTA、ring shadow `0 0 0 1px`）。
- **G3**：暗色实现严格对齐 `linear-dark.md` §2 §3 §6（`#08090a` 背景、Inter Variable + `cv01/ss03`、weight 510、Indigo `#5e6ad2` CTA、半透明白边 `rgba(255,255,255,0.05~0.08)`）。
- **G4**：Header 移除 14 色块切换器（`Header.tsx:84–108`），只保留语言切换 + Dark Mode Toggle。
- **G5**：`useTheme` hook 只暴露 `colorScheme`（`light` | `dark` | `system`），不再有「主题列表」概念；旧 `theme` 字段从 `appStore` 中移除。
- **G6**：浅色 ↔ 暗色切换由 OS `prefers-color-scheme` + 用户 override（在 SettingsDrawer 中）双重驱动，写入 `data-color-scheme` 而不是 `data-theme`。
- **G7**：测试更新通过，`vitest run --reporter=basic` 0 失败；浏览器验收清单全过（§7）。

## 3. Non-Goals（不做）

- 不做品牌色二次设计（terracotta + indigo 直接用 spec 的颜色，不调）。
- 不做新组件，只重构现有的 Header / SettingsDrawer / theme 注册表。
- 不做 i18n 主题名翻译（因为没有「主题列表」了）。
- 不引入新字体加载器，沿用现有 Google Fonts `<link>`（仅替换字体清单）。
- 不写 E2E（已在前一 commit 退役 golden 套件）；只补 unit + 浏览器手动验收。

---

## 4. Architecture / 数据架构

### 4.1 Token 流向

```
spec/ui-design/references/claude-light.md   spec/ui-design/references/linear-dark.md
              │                                              │
              ▼                                              ▼
     web/src/themes/claude-light.css             web/src/themes/linear-dark.css
              │                                              │
              └─────────── @theme tokens via tokens.css ─────┘
                                       │
                                       ▼
                             Tailwind utility classes
                                       │
                                       ▼
                     React components (Header, GeneratePanel, ...)
```

### 4.2 文件级变更

| 操作 | 文件 | 说明 |
|------|------|------|
| **新增** | `web/src/themes/claude-light.css` | 浅色实现，定义 `:root, [data-color-scheme="light"]` 下所有 `--theme-*` token |
| **新增** | `web/src/themes/linear-dark.css` | 暗色实现，定义 `[data-color-scheme="dark"]` 下所有 `--theme-*` token |
| **重写** | `web/src/themes/tokens.css` | 替代旧 `base.css`，纯 `@theme` token bridge + global reset；**修复嵌套 bug** |
| **删除** | `web/src/themes/{academic,art-deco,bauhaus,classical-chinese,japanese-bw,minimalist-bw,neo-minimal,pop-anime,pop-art,pop-art-dark,qi-baishi,rococo,swiss,workspace}.css` + 旧 `base.css` | 共 15 个文件 |
| **修改** | `web/src/main.tsx` 或入口 css | `@import` 改为只引入新 3 个文件 |
| **修改** | `web/src/hooks/useTheme.ts` | 删 `themes` 列表 / `theme` / `setTheme`；只保留 `colorScheme` + `setColorScheme` |
| **删除** | `web/src/hooks/useTheme.test.ts` 中所有 14 主题断言；改为 `light/dark/system` 三态测试 |
| **修改** | `web/src/hooks/useDynamicTheme.ts` | 评估是否还需要；如果只是 `data-color-scheme` 切换，可合并进 `useTheme` |
| **修改** | `web/src/stores/appStore.ts` | 移除 `theme` 字段；`colorScheme` 保留；migration 函数把旧 localStorage 的 `theme: "qi-baishi"` 等清掉 |
| **修改** | `web/src/components/layout/Header.tsx` | 删 84–108 行的 14 色块块；保留语言 + 加 DarkModeToggle |
| **修改** | `web/src/components/theme/ThemeSelector.tsx` | 删除（已无主题列表）→ 文件移除 |
| **修改** | `web/src/components/theme/DarkModeToggle.tsx` | 改为 light/dark/system 三态 segmented control |
| **修改** | `web/src/components/settings/SettingsDrawer.tsx` | 把 `DarkModeToggle` 放进抽屉内的「Appearance」分组 |
| **修改** | `web/index.html` | bootstrap script 删 `theme` 读取；`<link>` 替换字体为 Inter Variable + Source Serif 4 |
| **修改** | `web/src/i18n/locales/{en,zh}.json` | 删 `theme.options.*` 的 14 项；保留 `theme.title` / `theme.colorScheme.{light,dark,system}` |

### 4.3 Token 命名（保留向后兼容的 `--theme-*`）

为了不动几十个组件用 `var(--theme-primary)` 的地方，**保留 token 名**，只换值：

| token | claude-light 值 | linear-dark 值 |
|-------|----------------|----------------|
| `--theme-background` | `oklch(0.96 0.01 95)` ≈ `#f5f4ed` Parchment | `oklch(0.10 0.005 264)` ≈ `#08090a` Marketing Black |
| `--theme-card` | `oklch(0.98 0.005 95)` ≈ `#faf9f5` Ivory | `oklch(0.13 0.005 264)` ≈ `#0f1011` Panel |
| `--theme-foreground` | `oklch(0.20 0.005 80)` ≈ `#141413` Near Black | `oklch(0.97 0.005 264)` ≈ `#f7f8f8` Primary Text |
| `--theme-muted-foreground` | `oklch(0.46 0.01 80)` ≈ `#5e5d59` Olive Gray | `oklch(0.62 0.01 260)` ≈ `#8a8f98` Tertiary |
| `--theme-primary` | `oklch(0.56 0.16 40)` ≈ `#c96442` Terracotta | `oklch(0.58 0.16 285)` ≈ `#5e6ad2` Brand Indigo |
| `--theme-primary-foreground` | `oklch(0.98 0.005 95)` Ivory | `#ffffff` |
| `--theme-border` | `oklch(0.93 0.01 90)` ≈ `#f0eee6` Border Cream | `rgba(255,255,255,0.08)` |
| `--theme-border-subtle` | `oklch(0.95 0.01 90)` ≈ `#e8e6dc` | `rgba(255,255,255,0.05)` |
| `--theme-ring` | `oklch(0.85 0.01 80)` ≈ `#d1cfc5` Ring Warm | `rgba(0,0,0,0.2)` |
| `--theme-shadow-whisper` | `0 4px 24px rgba(0,0,0,0.05)` | `0 0 0 1px rgba(0,0,0,0.2)` |
| `--theme-font-heading` | `"Source Serif 4", Georgia, serif` | `"Inter Variable", system-ui, sans-serif` |
| `--theme-font-body` | `"Inter Variable", system-ui, sans-serif` | `"Inter Variable", system-ui, sans-serif` |
| `--theme-font-mono` | `"JetBrains Mono", ui-monospace, monospace` | `"JetBrains Mono", ui-monospace, monospace` |

> Anthropic Sans / Anthropic Serif / Anthropic Mono 是商业字体，无法使用 — 按 spec 推荐用 fallback：Source Serif 4（Adobe，开源）替代 Anthropic Serif；Inter Variable 替代 Anthropic Sans；JetBrains Mono 替代 Anthropic Mono。

### 4.4 颜色模式切换语义

- `data-color-scheme` 标签写在 `<html>` 上，三个值：`"light"` / `"dark"` / 不写（= 跟随 OS）。
- 所有 component css 用 `[data-color-scheme="dark"] { ... }` 选择器；浅色作为 `:root` 默认。
- `index.html` bootstrap 脚本只读 `colorScheme`，去掉 `theme`。

---

## 5. Implementation Plan（落地步骤）

### Step 1 — Spec 文档同步（已完成大部分，做收尾）

- [x] `claude-light.md` 已是 canonical
- [x] `linear-dark.md` 已是 canonical
- [ ] 更新 `design-principles.md` §0 §7 §8.3，把 14 主题列表改为 2 主题；把 P0/P1/P2 路线图标记为本任务收敛
- [ ] 更新 `references/README.md` §3 路线图，把 P2 「删除装饰性主题」改为 ✅ 完成

### Step 2 — Theme CSS 实现

1. 写 `web/src/themes/tokens.css`（替代 base.css）
2. 写 `web/src/themes/claude-light.css`（`:root, [data-color-scheme="light"]` 下的 token）
3. 写 `web/src/themes/linear-dark.css`（`[data-color-scheme="dark"]` 下的 token + `font-feature-settings: "cv01", "ss03"`）
4. 删 `web/src/themes/{14 个主题文件 + 旧 base.css}`

### Step 3 — Hook / Store / 组件改造

5. `useTheme` 砍到 `colorScheme` 一项 + `setColorScheme(s)` + 自动 OS 监听
6. `appStore` 的 `theme` 字段移除；migration 写一段把旧 `theme` 抹掉的代码
7. `Header.tsx` 删 14 色块；保留语言 + 把 DarkModeToggle 移到 SettingsDrawer
8. `DarkModeToggle.tsx` 改成 light / dark / system 三态
9. `SettingsDrawer.tsx` 加「Appearance」分组容纳上面那个 toggle

### Step 4 — Bootstrap / 字体

10. `web/index.html`：替换字体 `<link>` 为 Inter Variable + Source Serif 4 + JetBrains Mono；bootstrap 脚本只处理 `colorScheme`
11. `web/src/main.tsx`（或入口 css）只引 `tokens.css` + `claude-light.css` + `linear-dark.css`

### Step 5 — 测试

12. `useTheme.test.ts` 重写：light/dark/system 三态 + OS 监听
13. 跑 `vitest run --reporter=basic`；0 失败（如有 snapshot 残留，先删）
14. 跑 `pnpm tsc --noEmit` 或 `mcp__ide__getDiagnostics` 确认无类型错误

### Step 6 — 浏览器验收

见 §7 清单，逐条 `agent-browser snapshot`。

### Step 7 — Trellis Phase 3

15. `add_session.py` 记一笔
16. 更新 `design-principles.md` 把这次收敛沉淀进去
17. `task.py finish` 收尾

---

## 6. Test Plan

| 类别 | 范围 | 工具 |
|------|------|------|
| Unit | `useTheme` 三态 + OS prefers 监听 + `data-color-scheme` 切换 | vitest |
| Type check | 全仓 | `tsc --noEmit` / `mcp__ide__getDiagnostics` |
| Build | `web/` 构建无 css 解析错误 | `pnpm build` |
| 视觉 | 浏览器手动验收（§7） | `agent-browser` |

---

## 7. Browser Acceptance Checklist（浏览器验收）

每条都要 `agent-browser open` + `snapshot` + 视觉确认：

### 7.1 浅色（Claude）
- [ ] 页面背景是 Parchment 暖羊皮纸（`#f5f4ed`），不是纯白
- [ ] Hero H1 用 Source Serif 4 weight 500（fallback Georgia 也接受），64px 桌面 / 36px 移动
- [ ] 主 CTA 是 Terracotta `#c96442` + Ivory 文字，圆角 12px，hover 有 ring shadow `0 0 0 1px`
- [ ] 输入框 focus 用 Focus Blue `#3898ec` ring（spec 中唯一允许的冷色）
- [ ] body 文字 `oklch(0.20 0.005 80)`（≈ `#141413`），不是纯黑
- [ ] 没有任何 `box-shadow: 0 18px 60px rgba(...)` 这种重 drop shadow
- [ ] 圆角集中在 8 / 12 / 16 / 32px 四档
- [ ] Header 没有 14 色块切换器
- [ ] Header 不渲染主题切换；仅保留语言 + 设置按钮

### 7.2 暗色（Linear）
- [ ] 页面背景是 `#08090a`，不是纯黑也不是 `#0f0f0f`
- [ ] 文字主色 `#f7f8f8`（不是纯白）
- [ ] Hero H1 用 Inter Variable weight 510 + `font-feature-settings: "cv01", "ss03"` + letter-spacing -1.056px @ 48px
- [ ] 主 CTA 是 Indigo `#5e6ad2`，hover `#828fff`
- [ ] 卡片边用 `rgba(255,255,255,0.08)`，不是 solid 灰
- [ ] 卡片背景用 `rgba(255,255,255,0.02)~0.05`，不是 solid
- [ ] 没有暖色调 chrome（borders/backgrounds 都是冷灰 + 蓝紫 accent）
- [ ] 圆角集中在 4 / 6 / 8 / 12 / 22px

### 7.3 切换
- [ ] OS 切到暗色 → 页面自动切 linear-dark（用户没在 Settings 里 override 时）
- [ ] Settings → Appearance → Light：强制浅色，不跟 OS
- [ ] Settings → Appearance → Dark：强制暗色
- [ ] Settings → Appearance → System：跟 OS
- [ ] 刷新页面后选择持久化（localStorage `colorScheme`）

### 7.4 回归
- [ ] 生成主流程仍可点击 / 提交
- [ ] 历史 / 设置抽屉打开正常
- [ ] 语言切换正常
- [ ] 控制台无 css 加载 404 / no-such-token 警告

---

## 8. Risks & Mitigation

| 风险 | 缓解 |
|------|------|
| 删除 14 主题后某组件硬编码 `data-theme="academic"` 等失效 | 落地前 `grep -r 'data-theme=' web/src` 找出所有引用，统一改成不带值 |
| `appStore` migration 不完美导致旧用户 localStorage 报错 | migration 函数 try/catch + 默认值兜底 |
| Source Serif 4 / Inter Variable 加载失败 | fallback 链至 Georgia / system-ui |
| Tailwind 已编译类没有立刻反映 token 变化 | dev server 重启 + hard reload |
| 14 主题 css 中有 `--theme-*` 之外的 ad-hoc 变量被组件依赖 | grep 全仓 `var(--theme-` 找出未在新 token 列表中的变量，补 / 重命名 |

---

## 9. Definition of Done

1. `web/src/themes/` 只剩 `tokens.css` + `claude-light.css` + `linear-dark.css`
2. `vitest run --reporter=basic` 0 失败
3. `tsc --noEmit` 0 错
4. §7 浏览器清单全部勾选
5. `design-principles.md` 与 `references/README.md` 反映新现实（2 主题）
6. 1 个 atomic commit 在 `UI` 分支上，message 引用本 PRD
7. `task.py finish` 完成本任务
