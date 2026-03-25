# UI 架构重构实施计划

## 项目: PaperBanana UI V2.1 — 极简·悬浮·专注

---

## 一、实施阶段

### Phase 1: 核心布局重构 (P0)
**目标**: 建立极简布局骨架，废弃常驻侧边栏

#### Task 1.1: 创建 MinimalLayout 组件
```
文件: web/src/components/layout/MinimalLayout.tsx
要点:
- 无侧边栏布局
- 角落功能区 (左上 History, 右上 Model)
- 底部预留 FloatingInput 空间
- 主画布居中
```

#### Task 1.2: 实现 FloatingInput 组件
```
文件: web/src/components/input/FloatingInput.tsx
要点:
- position: fixed, bottom: 24px, centered
- 毛玻璃效果: backdrop-blur-lg
- 阴影: shadow-[0_8px_32px_rgba(0,0,0,0.12)]
- 圆角: rounded-2xl
- Method/Caption 双输入
- Generate/Stop 按钮
```

#### Task 1.3: 实现 ProgressLine 组件
```
文件: web/src/components/progress/ProgressLine.tsx
要点:
- 单行状态文本
- 细进度条 (h-1)
- 阶段图标动态切换
- 失败时红色高亮
```

### Phase 2: 模型配置系统 (P0)
**目标**: 实现 Popover 式模型配置，支持多商家管理

#### Task 2.1: 实现 ModelIndicator 组件
```
文件: web/src/components/model/ModelIndicator.tsx
要点:
- 右上角悬浮
- 显示启用的 Provider 名称
- 状态指示器 (⚡/⚠)
- 点击触发 Popover
```

#### Task 2.2: 实现 ModelPopover 组件
```
文件: web/src/components/model/ModelPopover.tsx
要点:
- 检索/生图模型选择器
- 商家列表 (启用/禁用)
- 模型添加/移除 (+/-)
- API Key 配置入口
```

#### Task 2.3: 实现 ProviderManager 组件
```
文件: web/src/components/model/ProviderManager.tsx
要点:
- Provider 卡片列表
- 折叠/展开模型列表
- 内联 API Key 输入
- 连接测试按钮
```

### Phase 3: 历史记录重构 (P1)
**目标**: 废弃常驻侧边栏，改为 Popover

#### Task 3.1: 实现 HistoryTrigger 组件
```
文件: web/src/components/history/HistoryTrigger.tsx
要点:
- 左上角悬浮按钮
- 圆形设计
- hover 微光效
- 点击触发 Popover
```

#### Task 3.2: 实现 HistoryPopover 组件
```
文件: web/src/components/history/HistoryPopover.tsx
要点:
- 最近 5 条历史
- 点击外部关闭
- 选择后关闭
- 完整历史入口
```

### Phase 4: 精修迭代系统 (P1)
**目标**: 实现多轮迭代和批量变体

#### Task 4.1: 重构 RefinePanel
```
文件: web/src/components/refine/RefinePanel.tsx
要点:
- 图片上传区
- 精修指令输入
- 迭代轮数配置 (1-5)
- 批量变体开关
```

#### Task 4.2: 实现 IterationTimeline 组件
```
文件: web/src/components/refine/IterationTimeline.tsx
要点:
- 轮次时间线
- Before/After 对比滑块
- AI 改动说明
- 继续迭代按钮
```

#### Task 4.3: 实现 BatchVariantGrid 组件
```
文件: web/src/components/refine/BatchVariantGrid.tsx
要点:
- 2x2 / 3x3 网格
- 风格描述标签
- 选择/下载按钮
- 放大预览
```

### Phase 5: 后端 API 扩展 (P1)
**目标**: 支持精修迭代和模型配置

#### Task 5.1: 扩展精修 API
```
文件: internal/api/handlers/refine.go
要点:
- 支持 iterations 参数
- SSE 返回每轮结果
- 批量变体支持
```

#### Task 5.2: 扩展模型配置 API
```
文件: internal/api/handlers/provider.go
要点:
- GET /providers - 含模型列表和启用状态
- PATCH /providers/:id/models/:modelId - 添加/移除模型
- PATCH /config/models - 设置检索/生图模型
```

---

## 二、文件变更清单

### 新增文件
```
web/src/components/
├── layout/
│   └── MinimalLayout.tsx
├── input/
│   └── FloatingInput.tsx
├── progress/
│   └── ProgressLine.tsx
├── model/
│   ├── ModelIndicator.tsx
│   ├── ModelPopover.tsx
│   └── ProviderManager.tsx
├── history/
│   ├── HistoryTrigger.tsx
│   └── HistoryPopover.tsx
└── refine/
    ├── IterationTimeline.tsx
    └── BatchVariantGrid.tsx
```

### 修改文件
```
web/src/App.tsx                    # 使用新布局
web/src/components/index.ts        # 导出新组件
web/src/hooks/useProviders.ts      # 扩展模型管理
internal/api/handlers/provider.go  # 新 API
internal/api/handlers/refine.go    # 迭代 API
```

### 删除文件
```
web/src/components/Layout.tsx      # 替换为 MinimalLayout
web/src/pages/SettingsPage.tsx     # 废弃全页设置
```

---

## 三、Golden Data 验证

| ID | 验证点 | 测试方法 |
|----|--------|---------|
| GD-UI-006 | 模型配置可见性 | E2E: 点击 ModelIndicator，验证 Popover 内容 |
| GD-UI-007 | 精修迭代进度 | E2E: 执行多轮精修，验证时间线显示 |
| GD-UI-008 | 批量变体展示 | E2E: 生成批量变体，验证网格布局 |
| GD-UI-009 | 极简状态展示 | E2E: 生成中，验证 ProgressLine 样式 |
| GD-UI-010 | 悬浮输入区 | E2E: 验证 FloatingInput 位置和样式 |

---

## 四、时间线

| 阶段 | 任务数 | 预估 |
|------|--------|------|
| Phase 1 | 3 | 1 天 |
| Phase 2 | 3 | 1 天 |
| Phase 3 | 2 | 0.5 天 |
| Phase 4 | 3 | 1 天 |
| Phase 5 | 2 | 0.5 天 |
| **总计** | **13** | **4 天** |

---

## 五、风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|---------|
| 现有测试用例失败 | 高 | 渐进式重构，保持向后兼容 |
| 动画性能问题 | 中 | 使用 CSS transitions 而非 JS |
| API 兼容性 | 中 | 版本化 API，新旧共存期 |

---

## 六、启动命令

```bash
# 开始 Phase 1
maestro-coordinate "execute Phase 1: 核心布局重构" -y
```
