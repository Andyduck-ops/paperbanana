# PaperBanana UI 设计原则 Spec

> 本文档是 PaperBanana 的 **UI 决策法典**。所有视觉细节（颜色、字体、阴影、圆角）请优先查阅 [`references/`](./references/) 目录的两份外部参考。本文档负责说明「哪条原则在什么场景下生效」。

---

## 0. 视觉锚点 (Visual Anchors)

PaperBanana 维护**恰好两套基底**：1 浅 + 1 暗，**没有第三个**。每套都直接 1:1 映射到一个外部参考：

| 模式 | 锚点 | 实现文件 | 关键特征 |
|------|------|---------|---------|
| 白天 / 阅读 / 学术 | **Claude (Anthropic)** [`references/claude-light.md`](./references/claude-light.md) | `web/src/themes/claude-light.css` | 暖羊皮纸 `#f5f4ed`、Anthropic Serif weight 500、赤陶 CTA `#c96442`、ring 阴影 `0 0 0 1px` |
| 夜晚 / 工程 / 高密度 | **Linear** [`references/linear-dark.md`](./references/linear-dark.md) | `web/src/themes/linear-dark.css` | 近黑画布 `#08090a`、Inter Variable + cv01/ss03、510 字重、半透明白边 `rgba(255,255,255,0.05~0.08)` |

**原则**：
1. 当 `web/src/themes/*.css` 与 references 冲突时，**永远修改实现**，不要反向修改 reference。
2. **不允许新增第三个主题**。如果未来确实要加，必须先在本文档与 [`references/README.md`](./references/README.md) 完成「锚点声明 + 收敛门槛 + 删除候选」三件事，否则 PR 会被拒。
3. 切换由 `data-color-scheme="light" | "dark"` 写到 `<html>` 上；缺省值跟随 OS `prefers-color-scheme`。**不再使用 `data-theme`**。
4. 用户在 SettingsDrawer → Appearance 里的选择会持久化到 `localStorage` 的 `colorScheme` 字段，覆盖 OS 偏好。

---

## 1. 首屏暴露原则 (Above the Fold Principle)

> 核心功能必须在首屏可见，无需滚动、无需点击、无需思考。

### 1.1 首屏三要素
| 优先级 | 元素 | 说明 |
|--------|------|------|
| P0 | 价值主张 | 一句话说明这个工具是做什么的 |
| P0 | 核心输入区 | 用户必须填写的最少字段 |
| P0 | 主行动按钮 | 最醒目的提交/生成按钮 |

### 1.2 禁止项
- 禁止在首屏放置纯装饰性内容（动画、大图标、示例卡片）
- 禁止核心功能被折叠、隐藏或需要滚动才能看到
- 禁止首屏出现多个同等重要的行动按钮分散注意力

### 1.3 PaperBanana 应用
- 输入框（方法 + 图注）必须在一打开页面时就可见
- "生成"按钮必须是页面上视觉上最重的元素
- EmptyState 不应占用垂直空间，或直接与输入区融合

---

## 2. 信息架构原则 (Information Architecture)

> 按用户任务流程组织信息，最重要的操作最容易找到。

### 2.1 用户任务流
```
打开页面 → 理解工具用途 → 输入需求 → 点击生成 → 查看结果
   ↑           ↑            ↑         ↑         ↑
  首屏      价值主张      核心输入    主CTA     结果区
```

### 2.2 信息优先级分层
| 层级 | 内容 | 位置 |
|------|------|------|
| L1 (核心) | 方法输入、图注输入、生成按钮 | 首屏顶部 |
| L2 (常用) | 参考图上传、模型选择 | 核心下方 |
| L3 (高级) | 批量模式、配置面板、模板 | 折叠/抽屉 |
| L4 (辅助) | 历史记录、设置、主题 | 顶部导航 |

### 2.3 禁止项
- 禁止将高级选项与核心选项平铺展示
- 禁止在核心任务流中插入无关功能

---

## 3. 视觉层次原则 (Visual Hierarchy)

> 通过大小、颜色、位置、留白建立层次，让"正确的选择"成为"显而易见的选择"。

### 3.1 层次建立手段
| 手段 | 应用 |
|------|------|
| 大小 | 主按钮 > 输入框 > 辅助按钮 > 文字 |
| 颜色 | 主按钮使用主色，辅助元素使用 muted 色 |
| 位置 | 核心输入居中偏上，辅助功能靠边 |
| 留白 | 核心区域周围留白最大，分组之间留白次之 |

### 3.2 主行动按钮规范
- 尺寸：至少 48px 高度，全宽或接近全宽
- 颜色：使用主色，与背景形成强烈对比
  - 浅色锚点：Terracotta `#c96442`，Ivory 文字
  - 暗色锚点：Indigo `#5e6ad2`，纯白文字
- 文案：动作导向，如"生成图表"而非"提交"
- 位置：紧跟核心输入之后，视线自然流动到达
- 阴影：`0 0 0 1px` ring（浅色）或 multi-layer luminance step（暗色），**禁用** drop-shadow

### 3.3 禁止项
- 禁止多个按钮使用同等视觉权重
- 禁止使用装饰性动画干扰核心操作
- 禁止在浅色锚点中引入冷色调灰；禁止在暗色锚点中引入暖色调 chrome

---

## 4. 表单设计原则 (Form Design)

> 每个字段都必须有存在的理由，标签必须清晰，完成路径必须一目了然。

### 4.1 字段组织
- 按逻辑分组，组间用留白或分割线区分
- 核心字段（方法、图注）放在最前面
- 高级字段默认折叠，需要时展开

### 4.2 标签规范
- 标签必须持久可见（不要用 placeholder 替代 label）
- 标签文案简洁明确，如"论文上下文"而非"请输入论文上下文"
- 可选字段明确标注

### 4.3 减少摩擦
- 删除所有非必要字段
- 提供合理的默认值
- 实时验证，错误提示具体可操作

### 4.4 输入框样式（来自锚点）
- 浅色：12px radius、Border Cream `#f0eee6`、focus ring 用 Focus Blue `#3898ec`（**唯一允许的冷色**）
- 暗色：6px radius、`rgba(255,255,255,0.08)` 边框、`rgba(255,255,255,0.02)` 背景、focus 用 multi-layer shadow

---

## 5. 极简主义原则 (Less is More)

> 每个元素都必须有明确目的，否则就是视觉噪音。

### 5.1 评估标准
对每个 UI 元素问三个问题：
1. 用户完成核心任务需要它吗？
2. 没有它用户会困惑吗？
3. 它帮助用户做出正确决策吗？

如果三个答案都是"否"，则删除。

### 5.2 PaperBanana 应用清单
| 元素 | 评估 | 决策 |
|------|------|------|
| Header 14 主题色板 | 否/否/否 | **P0** 从 Header 移除，放入 Settings 抽屉 |
| EmptyState 动画 | 否/否/否 | 删除或极简化 |
| 示例卡片 | 否/否/否 | 删除 |
| 多余 tagline + subtitle 双层介绍 | 否/否/否 | 合并为一行 |
| 批量模式 | 是/否/否 | 折叠到高级选项 |
| 配置面板 | 是/否/是 | 折叠到高级选项 |
| 模板选择器 | 是/否/是 | 保留但简化 |

### 5.3 主题数量
- **历史**：曾膨胀到 14 个主题文件（`academic`、`art-deco`、`base`、`bauhaus`、`classical-chinese`、`japanese-bw`、`minimalist-bw`、`neo-minimal`、`pop-anime`、`pop-art`、`pop-art-dark`、`qi-baishi`、`rococo`、`swiss`、`workspace`），违反 §0 视觉锚点纪律。
- **现状**：收敛到 **2 个**——`claude-light.css`（浅色锚点实现）+ `linear-dark.css`（暗色锚点实现）+ `tokens.css`（共享 token bridge & global reset）。
- **不允许新增第三个主题**。详见 §0 与 §7。

---

## 6. PaperBanana 首屏设计规范

### 6.1 首屏布局（桌面端）
```
+--------------------------------------------------+
| [Logo]  PaperBanana          [Nav] [Dark] [Lang] |  <- Header 已剔除主题色块
+--------------------------------------------------+
|                                                  |
|         创建科学可视化图表                        |  <- 价值主张 (Serif H1, 48–64px)
|                                                  |
|  +--------------------------------------------+  |
|  | 论文上下文与参考                               |  |  <- 核心输入
|  | [                              ]            |  |
|  +--------------------------------------------+  |
|                                                  |
|  +--------------------------------------------+  |
|  | 目标图注描述                                   |  |
|  | [                              ]            |  |
|  +--------------------------------------------+  |
|                                                  |
|  +--------------------------------------------+  |
|  |           [  生成图表  ]                      |  |  <- 主CTA (Terracotta/Indigo)
|  +--------------------------------------------+  |
|                                                  |
|  [高级选项 ▼]  [模板 ▼]                           |  <- 次要操作 (ghost button)
|                                                  |
+--------------------------------------------------+
```

### 6.2 首屏布局（移动端）
- 相同结构，输入框全宽
- 导航按钮收进汉堡菜单或只保留图标
- 高级选项默认折叠

### 6.3 关键尺寸
- 主按钮高度：48px
- 输入框最小高度：120px（多行文本）
- 核心区域最大宽度：720px，居中
- 首屏总高度：不超过 100vh，核心内容在 above the fold
- Hero H1：浅色 64px Serif weight 500 line-height 1.10；暗色 48px Inter Variable weight 510 letter-spacing -1.056px

---

## 7. 主题与视觉宪法

> 14 主题膨胀曾是最大的视觉债务源。任务 [`04-26-themes-light-dark-only`](../../tasks/04-26-themes-light-dark-only/prd.md) 把它压缩到「1 浅 + 1 暗」。本节固化收敛后的纪律。

### 7.1 当前主题清单（恰好两个，不允许新增）
| 文件 | 锚点 | 触发条件 |
|------|------|---------|
| `web/src/themes/claude-light.css` | [`claude-light.md`](./references/claude-light.md) | `:root` / `[data-color-scheme="light"]`，或 OS 偏好浅色 |
| `web/src/themes/linear-dark.css` | [`linear-dark.md`](./references/linear-dark.md) | `[data-color-scheme="dark"]`，或 OS 偏好暗色 |
| `web/src/themes/tokens.css` | — | 共享 `@theme` token bridge + global reset，**没有颜色主张** |

> 已删除的 15 个旧文件：`academic.css`、`art-deco.css`、`bauhaus.css`、`classical-chinese.css`、`japanese-bw.css`、`minimalist-bw.css`、`neo-minimal.css`、`pop-anime.css`、`pop-art.css`、`pop-art-dark.css`、`qi-baishi.css`、`rococo.css`、`swiss.css`、`workspace.css`、旧 `base.css`。

### 7.2 收敛路线图（spec 决策已收尾）

> 状态列反映的是**spec 决策**。代码落地进度请看任务 [`04-26-themes-light-dark-only/prd.md`](../../tasks/04-26-themes-light-dark-only/prd.md) 的 §5 实施步骤与 §9 Definition of Done。

| 阶段 | 动作 | spec 决策 |
|------|------|----------|
| **P0** | 修复 `base.css` 第 57–80 行 CSS 嵌套 bug | ✅ 通过删除 `base.css`、新建 `tokens.css` 解决 |
| **P0** | `academic.css` 主色 `oklch(0.42 0.12 255)` → `oklch(0.56 0.16 40)` | ✅ 通过删除 `academic.css` 解决；浅色锚点的 `--theme-primary` 现在统一是 `oklch(0.56 0.16 40)` Terracotta |
| **P0** | Header 14 色块迁出 | ✅ Header 不再渲染主题切换；DarkModeToggle 三态进入 SettingsDrawer → Appearance |
| **P1** | 收敛 box-shadow 到 ring + whisper 两档；删除 ad-hoc 圆角 | ✅ 仅保留 `0 0 0 1px` ring（浅色）+ whisper `0 4px 24px rgba(0,0,0,0.05)`；暗色用 `rgba(0,0,0,0.2) 0 0 0 1px` |
| **P1** | 引入 Inter Variable cv01/ss03 + Source Serif 4 | ✅ `index.html` 字体链改为 Inter Variable + Source Serif 4 + JetBrains Mono |
| **P2** | 删除装饰主题 | ✅ 14 → 2，全部装饰主题已删 |


### 7.3 新主题门槛（事实上禁用）
**当前规则：不允许新增第三个主题。** 如要破例：

1. 必须先在 PR 描述中说明为什么 `claude-light` 与 `linear-dark` 都无法承载这个场景；
2. 必须提名一个删除候选（绝不能净增加）；
3. 必须在 `references/` 下补一份新锚点 spec（与 `claude-light.md` / `linear-dark.md` 同等深度）；
4. 必须通过本文件 §8 的实现检查清单；
5. PRD 必须经过 ui-design 维护者签字。


---

## 8. 实现检查清单

### 8.1 首屏
- [ ] 打开页面后，输入框在首屏可见（无需滚动）
- [ ] "生成"按钮是页面上视觉上最重的元素
- [ ] 没有纯装饰性内容占用首屏空间
- [ ] 高级选项默认折叠
- [ ] 标签持久可见，placeholder 仅用于格式示例
- [ ] 每个元素都通过"三问"测试

### 8.2 视觉锚点合规
- [ ] 浅色页面背景使用 `oklch(0.96 0.01 95)` 暖羊皮纸（≈ `#f5f4ed`），不是纯白
- [ ] 暗色页面背景使用 `oklch(0.10 0.005 264)` 近黑（≈ `#08090a`），不是纯黑
- [ ] 文字色：浅色用 `oklch(0.20 0.005 80)`（warm near-black），暗色用 `oklch(0.97 0.005 264)`（warm off-white）
- [ ] 主 CTA：浅色 Terracotta，暗色 Indigo
- [ ] 阴影只用 `0 0 0 1px` ring（浅色）或 luminance step + `rgba(0,0,0,0.2) 0 0 0 1px`（暗色），**没有任何 drop-shadow**
- [ ] 圆角集中在 sm/md/lg/2xl 四档，没有 ad-hoc 像素值
- [ ] Header 不展示主题色块；主题切换在 Settings 抽屉内
- [ ] Hero H1：浅色 Serif、暗色 Inter Variable + cv01/ss03

### 8.3 主题清洁
- [ ] `web/src/themes/` 只有 `tokens.css` + `claude-light.css` + `linear-dark.css`，**不能多一个文件**
- [ ] 浅色（`claude-light.css`）中没有冷色调灰（`oklch(... 0.0X 200~280)`）；唯一允许的冷色是 input focus 的 Focus Blue `#3898ec`
- [ ] 暗色（`linear-dark.css`）中没有暖色调 chrome（borders/backgrounds）；只允许语义化的暖色 accent（如 warning/error）
- [ ] 没有 `data-theme="..."` 属性残留；只用 `data-color-scheme="light" | "dark"`
- [ ] CSS 嵌套语法正确（旧 `base.css` 嵌套 bug 已通过删除该文件解决）
- [ ] 单个主题文件 < 200 行；超过即代表掺杂了组件级样式，应该迁去 `web/src/components/**/*.module.css` 或 Tailwind utility
