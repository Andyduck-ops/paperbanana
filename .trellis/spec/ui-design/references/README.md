---
ref_id: ui-design-references-index
role: 设计参考索引
captured_on: 2026-04-26
status: canonical
---

# PaperBanana 设计参考索引

PaperBanana 把外部设计语言收敛到**两个权威参考**，浅色与暗色各一：

| 文件 | 角色 | 核心动作 | 适用主题 |
|------|------|---------|---------|
| [`claude-light.md`](./claude-light.md) | 浅色基底 (Light Anchor) | 暖羊皮纸 + Anthropic Serif 500 + 赤陶 CTA + ring 阴影 | `academic` (warm-brown 回归)、`art-deco`、`classical-chinese`、`japanese-bw`、`workspace`、`base` |
| [`linear-dark.md`](./linear-dark.md) | 暗色基底 (Dark Anchor) | 近黑画布 + Inter Variable cv01/ss03 + 510 字重 + 半透明白边 | `pop-art-dark`、未来引入的 `dark` 通用主题；其他主题在 `prefers-color-scheme: dark` 下回退到本系统 |

## 1. 为什么是这两个

- **互补但不冲突**：Claude 是「白天 / 阅读 / 学术」，Linear 是「夜晚 / 工程 / 高密度」。PaperBanana 的核心场景同时覆盖论文阅读（白天 Serif）与图谱编排（晚间 Sans+Mono），两者刚好对应。
- **教条性强**：两个系统都明确说明「不要做什么」。这正是当前 14 个主题需要的纪律——避免再次膨胀。
- **可量化**：颜色、字号、字重、letter-spacing、shadow 都有具体值，可以直接映射到 OKLCH / Tailwind / theme tokens。

## 2. 当 Spec 与实现冲突时

**永远以本目录为准**。如果 `web/src/themes/*.css` 中的实现与本参考不一致：

1. 先在 PR 描述里 tag `ui-design`
2. 修改实现而不是修改 spec
3. 如果发现 spec 真的过时（外部参考更新），先在本目录提 patch + git commit，再动实现

## 3. 主题收敛路线图 (P0/P1/P2)

| 阶段 | 动作 | 受影响主题 |
|------|------|----------|
| **P0** | 修复 `base.css` 的 CSS 嵌套 bug；把 `academic.css` 主色从冷调蓝 (`oklch(0.42 0.12 255)`) 改回暖棕 (`oklch(0.56 0.16 40)`) | `base`、`academic` |
| **P0** | Header 移除 14 个主题色块，改为 Settings 抽屉内的下拉/分组选择 | 全局 |
| **P1** | 收敛圆角到 `radius-sm/md/lg/2xl` 四档；删除 `rounded-[1.6rem]`、`rounded-[2rem]` 等 ad-hoc 值 | 全局 |
| **P1** | 把所有 box-shadow 替换为 `0 0 0 1px` ring + 单层 whisper shadow（浅色）或 luminance step（暗色） | 全局 |
| **P2** | 删除 `pop-art`、`pop-art-dark`、`rococo` 等装饰性强的主题；保留 6–8 个学术/工程向主题 | 主题文件 |
| **P2** | 引入 Anthropic Serif fallback (Source Serif 4) 与 Inter Variable cv01/ss03，统一字体加载 | `index.html`、`base.css` |

## 4. 与 PaperBanana 自身 DESIGN.md 的关系

- 项目根目录的 `DESIGN.md` 是**实现指南**（OKLCH token、组件命名、Tailwind class），需要被持续更新以反映本目录决策。
- 本目录是**视觉宪法**（颜色哲学、字体哲学、阴影哲学、do/don't），变更频率低、权威度高。
- 当根目录 `DESIGN.md` 与本目录冲突时，先修根目录。

## 5. 如何在 PR / commit 中引用

在 commit message 中：

```
feat(ui): unify ring shadow per spec/ui-design/references/claude-light.md §6
```

在 PR 描述中链接具体段落，方便 review 时核对。

## 6. 维护节奏

- 外部参考有重大更新（每季度）：手动 diff，patch 写入本目录
- 项目主题增删：必须先在 `design-principles.md` 标记 motivation，再创建/删除 theme css

---

> 本索引短小是有意为之。详细参数请打开对应的 `*.md`。
