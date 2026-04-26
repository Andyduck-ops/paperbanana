# UI 展示逻辑与功能对齐审查

**Date**: 2026-04-25  
**Scope**: `paperbanana-clean/web/src/` — 所有 UI 组件 + i18n 翻译文件 + 设计系统  
**Method**: 交叉比对 i18n keys ↔ 组件引用 ↔ 实际功能语义

---

## 汇总

| 严重级别 | 数量 | 说明 |
|----------|------|------|
| **CRITICAL** | 2 | i18n key 重复/缺失导致 UI 崩溃或全乱码 |
| **HIGH** | 3 | 概念命名不一致导致用户困惑 |
| **MEDIUM** | 6 | 翻译冗余/缺失/术语混乱 |
| **LOW** | 4 | 小范围文案或样式不匹配 |

---

## CRITICAL — 必须立即修复

### [CRITICAL] i18n: 13 个 WorkspaceHero 翻译键完全缺失

- **File**: `web/src/components/workspace/WorkspaceHero.tsx:21-82`
- **Issue**: WorkspaceHero 组件引用了 13 个 i18n key，但 `en.json` 和 `zh.json` 中均不存在：
  - `workspace.heroTitle`
  - `workspace.heroDescription`
  - `workspace.heroCards.contextLabel` / `contextBody`
  - `workspace.heroCards.outputLabel` / `outputBody`
  - `workspace.heroCards.workflowLabel` / `workflowBody`
  - `workspace.quickActionsTitle`
  - `workspace.quickActions.primaryLabel` / `primaryBody`
  - `workspace.quickActions.history`
  - `workspace.quickActions.settings`
- **Impact**: UI 上将直接显示原始 key 路径（如 `workspace.heroTitle`），整屏文字为乱码。
- **Fix**: 在 `en.json` / `zh.json` 的 `workspace` 块中添加这些 key。

### [CRITICAL] i18n: `history` 节点重复，第二段覆盖第一段

- **File**: `web/src/i18n/locales/en.json:170-194` 与 `en.json:539-553`
- **Issue**: 同一个 JSON 对象中存在两个 `history` 顶级键。JSON 解析时后者覆盖前者。第一段 `history.title = "Work Records"` 被第二段 `history.title = "History"` 覆盖。
  第一段独有的 key（`subtitle`, `empty`, `emptyHint`, `untitled`, `version`, `restore`, `restoreFailed`, `restoreUnavailable`, `status.*`）全部丢失。
- **Impact**: 历史面板展示文案与实际功能完全不匹配；"Work Records" 语义被替换为泛化的 "History"。
- **Fix**: 
  - 方案 A：合并两段为一个 `history` 块。
  - 方案 B：将第二段重命名为 `historyPanel` 或 `historyFilter`。

---

## HIGH — 发布前必须修复

### [HIGH] 术语不一致: "Execution Node" vs "Visualizer Node"

- **Files**:
  - `en.json:97`: `generate.visualizerNode = "Execution Node"`
  - `zh.json:97`: `generate.visualizerNode = "执行节点"`
  - `api.ts:105`: `visualizer_node` 参数名
  - 测试 mock: `generate.visualizerNode = "Visualizer Node"`
- **Issue**: 代码层统一使用 `visualizerNode` / `visualizer_node`，但用户看到的英文标签是 "Execution Node"（执行节点）。测试文件使用的又是 "Visualizer Node"。概念从后端到前端不一致。
- **Fix**: 统一为 "Visualizer Node" / "可视化节点"，或如果功能确实是选择执行引擎则改为 "Execution Node" / "执行引擎"。

### [HIGH] 术语不一致: "Channel" vs "Provider"

- **Files**: 
  - `en.json`: `modelConfig.*` 块使用 "Channel/通道"
  - `en.json`: `settings.*` 块使用 "Provider/提供商"
  - `SettingsPage.tsx`: 同时引用 `modelConfig.channels` 和 `settings.providers`
- **Issue**: Settings 页面中，"Channels" 标题列表显示的是 Providers 数据。两个术语指同一事物。`modelConfig.channels` 和 `settings.providers` 在相同 UI 中混用，且二者的管理功能（添加/删除/配置）完全相同。
- **Fix**: 全项目统一为一个术语。建议用 "Provider"（提供商），因为后端和设计文档使用此术语。

### [HIGH] Empty States 双份定义且内容不同

- **Files**:
  - `en.json:357-382`: 顶层 `emptyState.*` (EmptyState 组件使用)
  - `en.json:44-70`: `workspace.emptyState.*` (未被任何组件引用)
- **Issue**: 
  - `emptyState.generateTitle` = "Create Scientific Visualizations"
  - `workspace.emptyState.generateTitle` = "Start from paper context, then shape the figure"
  - 两套完全不同的文案，后者是面向论文场景的精准文案但从未被使用。
- **Fix**: 删除 `workspace.emptyState.*`（死代码），或将 EmptyState 组件改为使用 `workspace.emptyState.*`（更好的文案）。

---

## MEDIUM — 本迭代内修复

### [MEDIUM] i18n: 3 个引用键缺失（有 fallback 兜底）

| 引用位置 | 缺失 key | fallback 值 |
|----------|----------|-------------|
| `SettingsPage.tsx:194` | `settings.appearance` | `"Appearance"` |
| `SettingsPage.tsx:197` | `settings.appearanceDescription` | `"Customize the look..."` |
| `ProviderEditPage.tsx:127` | `settings.addProvider` | `undefined` (显示空白) |
| `ModelConfigPage.tsx:46` | `modelConfig.subtitle` | `"Manage channels..."` |

- **Impact**: `settings.addProvider` 缺失会导致新建 Provider 页面标题为空。其余有硬编码 fallback，但中英文切换时 fallback 始终是英文。
- **Fix**: 在 `en.json` / `zh.json` 中补全这些 key。

### [MEDIUM] GeneratePanel 标题区冗余

- **File**: `GeneratePanel.tsx:104-109`
- **Issue**: 
  ```tsx
  <h1>{t('app.tagline')}</h1>          // "Academic paper figure generation and refinement"
  <p>{t('generate.subtitle')}</p>       // "Enter paper context and target description..."
  ```
  两个描述性文字上下排列且含义高度重叠。`app.tagline` 是全局标语，`generate.subtitle` 是功能说明。二者同时出现显得啰嗦。
- **Fix**: 保留 `generate.subtitle` 作为功能说明；将 `app.tagline` 移到 `<title>` 标签或 meta description。

### [MEDIUM] `generate.descriptionLabel` 仅存在于测试中

- **Files**:
  - `en.json:95`: `generate.descriptionLabel = "Describe your task"`
  - `GeneratePanel.test.tsx:10`: mock 中引用
  - **实际组件**: DualInputPanel 使用两个独立 label（"Paper Context & References" + "Target Figure Brief"），而非单一 description。
- **Issue**: 翻译键定义了但从未被实际组件使用。如果未来有人改用 `descriptionLabel`，会与当前双栏输入 UI 不匹配。
- **Fix**: 删除 `generate.descriptionLabel` / `descriptionPlaceholder`，或明确保留为未来单栏模式使用。

### [MEDIUM] `generate.modes.manual` = "Manual" vs 实际功能无手动模式

- **File**: `en.json:144`
- **Issue**: Retrieval mode 有 `manual` 选项翻译为 "Manual" / "手动"。但查看后端逻辑 (`agents/retriever/`)，不存在真正的手动检索操作。用户选择 "Manual" 后发生了什么？未见明确文档或代码路径。
- **Fix**: 如果 "Manual" 实际上等同于 "None"（跳过检索），应明确标注；如果是仅用 paper context 作为检索依据，应改为 "Paper Only" / "仅论文上下文"。

### [MEDIUM] `generate.pipelines.vanilla` = "Vanilla (Direct)"

- **File**: `en.json:151`
- **Issue**: "Vanilla" 已意为原始/简单，追加 "(Direct)" 显得冗长冗余。
- **Fix**: 改为 "Direct (No Enrichment)" / "直接生成（无增强）"。

### [MEDIUM] ConfigPanel 中 model dropdown 使用 `settings.default` 而非 `generate.*`

- **File**: `ConfigPanel.tsx:247, 298`
- **Issue**: `<option value="">{t('settings.default')}</option>` — 模型选择器的默认选项引用了 `settings.default`（"Default"），但上下文是生成配置面板。这应该使用 `generate.defaultModel` 或单独的 "System Default" label。
- **Fix**: 添加 `generate.defaultModel` key，或使用 "Use System Default" / "使用系统默认"。

---

## LOW — 方便时修复

### [LOW] Header logo 的 `alt=""` 正确但缺品牌名称

- **File**: `Header.tsx:25`
- **Issue**: `<img src="/images/logo.png" alt="" />` — logo 旁有文字 `PaperBanana`，所以 `alt=""`（装饰性图片）是正确的。但 `title` 属性未设置，hover 无提示。
- **Fix**: 添加 `title="PaperBanana"`。

### [LOW] ImageUpload 按钮无 `aria-label` 的 i18n key 对齐

- **File**: `ImageUpload.tsx:106`
- **Issue**: 拖放上传按钮没有 `aria-label`，仅通过 `<span>{t('refine.dropImage')}</span>` 传递文本。
- **Fix**: 添加 `aria-label={t('refine.dropImage')}` 属性。

### [LOW] GeneratePanel 的 svg icon 无语义 ARIA 标签

- **File**: `GeneratePanel.tsx:138`
- **Issue**: Submit 按钮中的闪电 SVG 图标缺少 `aria-hidden="true"`。
- **Fix**: 添加 `aria-hidden="true"`。

### [LOW] 中英文 DualInputPanel textarea 行数差异

- **File**: `DualInputPanel.tsx:57,73`
- **Issue**: method textarea 的 `rows={5}` 和 caption textarea 的 `rows={3}` 是固定值。中文输入时字符占位与英文差异很大，rows 可能需要根据语言动态调整或使用 `autoResize`。
- **Fix**: 添加 `autoResize` 或根据 `language` 动态调整 rows（中文约需英文 1.2-1.3x 行数）。

---

## 设计系统审查

### Tailwind Theme Token 使用一致性

- **状态**: PASS。所有组件一致使用 `--theme-*` CSS 变量和 Tailwind token（`bg-primary`, `text-foreground`, `border-border` 等）。无常量硬编码。
- 未发现使用原生 HEX/RGB 颜色替代 theme token 的情况。

### 圆角 (border-radius) 一致性

- **状态**: PARTIAL。组件间使用不同的 rounded 值：
  - Header: `rounded-[1.6rem]`
  - Buttons: `rounded-lg` / `rounded-xl`
  - Cards: `rounded-2xl`
  - Workspace hero: `rounded-[2rem]`
  - 建议统一为一套 scale（sm: 8px, md: 12px, lg: 16px, xl: 24px），或用 `rounded-2xl` 作为卡片/面板的默认值。

### 间距 (padding) 一致性

- **状态**: PASS。组件内 padding 基本一致（py-3 px-4 或 py-4 px-6）。

---

## 建议修复排序

1. **CRITICAL**: 补全 WorkspaceHero 的 13 个 i18n key → 消除全乱码
2. **CRITICAL**: 移除/合并重复的 `history` 节点 → 恢复文案正确性
3. **HIGH**: 统一 "Execution Node" → "Visualizer Node"
4. **HIGH**: 统一 "Channel" → "Provider" 全项目
5. **HIGH**: 删除 `workspace.emptyState.*` 死代码
6. **MEDIUM**: 补全 4 个缺失的 i18n key
7. **MEDIUM**: 调整 GeneratePanel 标题区避免冗余
8. **MEDIUM**: 清理未使用的 i18n key
9. **MEDIUM**: 修复 ConfigPanel model dropdown 引用错误 i18n key
10. **LOW**: 补充缺失的 ARIA 属性
