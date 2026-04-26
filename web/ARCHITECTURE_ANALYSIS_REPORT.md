# PaperBanana 前端代码聚合与复用分析报告

## 项目概况

| 指标 | 数值 |
|------|------|
| TS/TSX 文件总数 | 170 (51 .ts + 119 .tsx) |
| 源代码文件 (不含测试) | 107 |
| 构建输出 | 451KB JS, 98KB CSS |
| 技术栈 | React + TypeScript + Vite + TailwindCSS |

---

## 1. 组件复用分析

### 1.1 组件目录结构

```
src/components/
├── index.ts                    # 统一导出
├── components/                 # 50+ 个组件
│   ├── folder/                # 4个文件管理组件
│   ├── history/               # 4个历史记录组件
│   ├── input/                 # 1个输入组件
│   ├── layout/                # 1个布局组件
│   ├── model/                 # 6个模型配置组件
│   ├── progress/              # 1个进度组件
│   ├── refine/                # 2个精修组件
│   ├── settings/              # 4个设置组件
│   └── workspace/             # 5个工作区组件
└── [40+ 根级组件]
```

### 1.2 发现的具体重复代码位置

#### 🔴 高优先级 - Panel 组件重复模式

**文件位置:**
- `src/components/GeneratePanel.tsx` (210 lines)
- `src/components/RefinePanel.tsx` (161 lines)
- `src/components/ConfigPanel.tsx` (324 lines)

**重复模式分析:**

| 模式 | GeneratePanel | RefinePanel | ConfigPanel |
|------|--------------|-------------|-------------|
| Form 状态管理 | ✅ useState | ✅ useState | ✅ useState |
| 提交按钮模式 | ✅ w-full 按钮 | ✅ w-full 按钮 | N/A |
| disabled 状态传递 | ✅ isGenerating | ✅ isRefining | ✅ disabled |
| useLanguage hook | ✅ | ✅ | ✅ |
| 折叠/展开模式 | N/A | N/A | ✅ ChevronIcon |

**重复代码片段 (ConfigPanel 行 100-102):**
```tsx
const configuredProviders = providers.filter(
  (p) => p.enabled && p.status === 'configured' && p.models && p.models.length > 0
);
```

**与 ModelSelector.tsx 行 30-34 重复:**
```tsx
for (const channel of channels) {
  if (!channel.enabled) continue;
  for (const model of channel.models) {
    if (model.enabled) {
```

#### 🔴 高优先级 - Selector 类型组件重复

**文件位置:**
- `src/components/ProjectSelector.tsx` (218 lines)
- `src/components/TemplateSelector.tsx` (186 lines)
- `src/components/model/ModelSelector.tsx` (146 lines)

**重复模式:**
```tsx
// 所有三个 Selector 都有相同的模式
const [isOpen, setIsOpen] = useState(false);
const [loading, setLoading] = useState(true);
const [error, setError] = useState<string | null>(null);

// 外部点击关闭逻辑 (TemplateSelector 行 31-39)
useEffect(() => {
  function handleClickOutside(event: MouseEvent) {
    if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
      setIsOpen(false);
    }
  }
  // ...
}, []);
```

#### 🟡 中优先级 - History 相关组件类型重复

**文件位置:**
- `src/components/HistoryItem.tsx` (根级)
- `src/components/history/HistoryItem.tsx` (history目录)
- `src/components/history/HistoryPanel.tsx`
- `src/components/HistorySidebar.tsx`

**问题:** 存在 "Legacy" 和 "V2" 两套历史组件，代码重复度高。

**src/components/index.ts 行 37-39:**
```tsx
// Legacy History components (deprecated, use V2 above)
export { HistorySidebar, type HistorySidebarProps } from './HistorySidebar';
export { HistoryItem as HistoryItemLegacy, ...
```

#### 🟡 中优先级 - 本地存储 Hook 重复模式

**在多个 hooks 中发现相同的 localStorage 模式:**

| Hook | localStorage Key | 模式 |
|------|-----------------|------|
| `useTheme.ts` | `paperbanana-theme` | load/save |
| `useLanguage.ts` | `paperbanana-language` | load only |
| `usePromptTemplates.ts` | `prompt-templates` | load/save |
| `useLocalWorkRecords.ts` | `paperbanana-work-records` | load/save |
| `ProjectSelector.tsx` | `paperbanana_projects_cache` | cache pattern |
| `ModelConfigContext.tsx` | `paperbanana_model_config_v1` | full state |

**重复代码示例 (出现 6+ 次):**
```tsx
function loadFromStorage<T>(key: string, defaultValue: T): T {
  if (typeof window === 'undefined') return defaultValue;
  try {
    const stored = localStorage.getItem(key);
    return stored ? JSON.parse(stored) : defaultValue;
  } catch {
    return defaultValue;
  }
}
```

---

## 2. Hooks 复用分析

### 2.1 18 个自定义 Hooks 概览

| Hook | 功能 | 复杂度 | 依赖 |
|------|------|--------|------|
| `useBatchGeneration.ts` | 批量生成流 | 高 | stream API |
| `useGenerate.ts` | 单一生成 | 高 | SSE, network |
| `useRefine.ts` | 图像精修 | 中 | api |
| `useHistory.ts` | 历史记录 | 高 | apiClient |
| `useProviders.ts` | 提供商管理 | 中 | configStream |
| `useFolders.ts` | 文件夹管理 | 中 | api |
| `useVersions.ts` | 版本管理 | 低 | api |
| `useTemplates.ts` | 模板管理 | 低 | api |
| `usePromptTemplates.ts` | 提示词模板 | 低 | localStorage |
| `useLocalWorkRecords.ts` | 本地工作记录 | 低 | localStorage |
| `useTheme.ts` | 主题切换 | 低 | localStorage |
| `useLanguage.ts` | 语言切换 | 低 | i18n |
| `useToast.ts` | Toast通知 | 低 | state |
| `useKeyboardShortcuts.ts` | 快捷键 | 低 | window |
| `useNetworkStatus.ts` | 网络状态 | 低 | navigator |
| `useNetworkStatus` | 网络状态检测 | 低 | navigator |

### 2.2 可合并或抽象的 Hooks

#### 🔴 useGenerate.ts & useBatchGeneration.ts 相似度分析

**useGenerate.ts 行 38-47:**
```tsx
interface GenerateOptions {
  visualizerNode?: string;
  content?: string;
  visualIntent?: string;
  config?: Pick<
    GenerateRequest,
    'aspect_ratio' | 'critic_rounds' | 'retrieval_mode' | 'pipeline_mode' | 'query_model' | 'gen_model'
  >;
}
```

**useBatchGeneration.ts 行 7-15:**
```tsx
interface StartBatchOptions {
  visualizerNode?: string;
  content?: string;
  visualIntent?: string;
  config?: Pick<
    GenerateRequest,
    'aspect_ratio' | 'critic_rounds' | 'retrieval_mode' | 'pipeline_mode' | 'query_model' | 'gen_model'
  >;
}
```

**结论:** 两个接口几乎完全相同，只是名称不同。

#### 🟡 localStorage Hooks 抽象机会

**发现以下 hooks 具有相同的 localStorage 模式:**
- `useTheme.ts`
- `usePromptTemplates.ts`
- `useLocalWorkRecords.ts`

**建议提取通用 hook:**
```tsx
// hooks/useLocalStorage.ts
export function useLocalStorage<T>(key: string, defaultValue: T) {
  const [value, setValue] = useState<T>(() => loadFromStorage(key, defaultValue));
  useEffect(() => saveToStorage(key, value), [key, value]);
  return [value, setValue] as const;
}
```

---

## 3. 类型定义聚合分析

### 3.1 types/ 目录现状

```
src/types/
├── api.ts      # 271 lines - 主要API类型
└── batch.ts    # 73 lines - 批量生成类型
```

### 3.2 类型重复问题

#### 🔴 Artifact 类型分散

**src/types/api.ts 行 145-161:**
```tsx
export interface Artifact {
  id: string;
  kind: ArtifactKind;
  mime_type: string;
  // ... snake_case fields
}
```

**src/types/batch.ts 行 34-43:**
```tsx
export interface UIArtifact {
  id: string;
  kind: string;
  mimeType: string;  // camelCase
  // ... camelCase fields
}
```

**src/hooks/useGenerate.ts 行 26-36:**
```tsx
export interface GenerateResult {
  artifacts: Array<{
    kind: string;
    mimeType: string;  // camelCase
    // ...
  }>;
}
```

**问题:** 同一概念存在 snake_case (API) 和 camelCase (UI) 两种版本。

#### 🔴 Provider/Channel 类型冗余

**src/context/ModelConfigContext.tsx 行 34-50:**
```tsx
export interface Provider {
  id: string;
  name: string;
  type: 'openai' | 'anthropic' | ...;
  // ...
}
```

**src/hooks/useProviders.ts 行 22-36:**
```tsx
export interface Provider {
  id: string;
  type: string;  // 无具体枚举
  name: string;
  // ... 字段不完全一致
}
```

#### 🟡 Stage 状态类型重复

**src/components/StageCard.tsx 行 3:**
```tsx
export type StageStatus = 'pending' | 'running' | 'complete' | 'error' | 'not_run';
```

**src/hooks/useGenerate.ts 行 8-19:**
```tsx
export interface StageState {
  stage: string;
  agent: string;
  status: StageStatus;  // 从 StageCard 导入
  // ...
}
```

**src/components/ProgressPanel.tsx 行 5-12:**
```tsx
export interface StageState {  // 重新定义！
  stage: string;
  agent: string;
  status: StageStatus;
  // ... 不完全一致
}
```

### 3.3 建议的统一类型导出策略

```typescript
// types/index.ts - 建议创建
export type { Artifact, ArtifactKind } from './api';
export type { BatchProgress, UIArtifact } from './batch';
export type { StageStatus, StageState } from './stage';  // 新建统一文件

// 统一命名规范
export type ApiArtifact = import('./api').Artifact;      // snake_case
export type UiArtifact = import('./batch').UIArtifact;   // camelCase
```

---

## 4. 状态管理评估

### 4.1 App.tsx 中的 useState 分析 (16 个状态)

| 状态 | 类型 | 用途 | 建议 |
|------|------|------|------|
| `currentPage` | Page | 页面路由 | 可用 URL state |
| `editingProvider` | string | 编辑的提供商 | 可用 URL state |
| `mainTab` | MainTab | 主标签 | 可用 URL state |
| `currentPath` | string | URL 路径 | 已有 |
| `selectedSessionId` | string | 选中的会话 | 可合并 |
| `selectedBatchCandidateId` | string \| null | 批处理候选 | 可合并 |
| `refineSeedImageData` | string \| null | 精修种子图 | 可合并 |
| `exportArtifact` | Artifact | 导出目标 | 可合并 |
| `showExport` | boolean | 显示导出弹窗 | 可合并为对象 |
| `isHistoryPanelOpen` | boolean | 历史面板 | 可合并为对象 |
| `isSettingsOpen` | boolean | 设置抽屉 | 可合并为对象 |
| `showShortcutsHelp` | boolean | 快捷键帮助 | 可合并为对象 |
| `showWelcomeWizard` | boolean | 欢迎向导 | 可合并为对象 |
| `examplePrompt` | string \| null | 示例提示词 | 可合并 |
| `pendingHistoryContext` | object \| null | 历史上下文 | 可合并 |

**布尔标志建议合并:**
```tsx
// 当前 (7 个独立状态)
const [showExport, setShowExport] = useState(false);
const [isHistoryPanelOpen, setIsHistoryPanelOpen] = useState(false);
const [isSettingsOpen, setIsSettingsOpen] = useState(false);
// ...

// 建议
const [uiState, setUiState] = useState({
  showExport: false,
  historyPanelOpen: false,
  settingsOpen: false,
  shortcutsHelpOpen: false,
  welcomeWizardOpen: false,
});
```

### 4.2 ModelConfigContext 评估

**位置:** `src/context/ModelConfigContext.tsx` (687 lines)

**优点:**
- ✅ 使用 useReducer 管理复杂状态
- ✅ 向后兼容 (Channel/Provider 别名)
- ✅ localStorage 持久化
- ✅ 完整的 CRUD 操作

**问题:**
- ⚠️ 单一 Context 过大 (687 lines)
- ⚠️ 包含 API 调用逻辑 (与 hooks 重复)
- ⚠️ useProviders hook 和 Context 功能重叠

**代码重叠示例:**

**useProviders.ts 行 51-67:**
```tsx
const fetchProviders = useCallback(() => {
  fetch('/api/v1/providers')
    .then(res => res.json())
    .then((data: ProvidersResponse) => setProviders(data.providers || []))
    .catch(err => setError(err.message))
    .finally(() => setLoading(false));
}, []);
```

**ModelConfigContext.tsx 行 293-335:**
```tsx
const refreshProviders = useCallback(async () => {
  const response = await fetch('/api/v1/providers');
  const data = await response.json();
  // ... 几乎相同的逻辑
}, []);
```

### 4.3 是否需要 Redux/Zustand?

**当前状态:**
- 使用 React Context + useReducer (ModelConfigContext)
- 使用自定义 hooks 管理局部状态

**评估结论:**

| 维度 | 现状 | Redux | Zustand |
|------|------|-------|---------|
| 复杂度 | 中等 | 过高 | 合适 |
| 学习成本 | 低 | 高 | 低 |
| 性能优化 | 手动优化 | 内置 | 内置 |
| 中间件支持 | 无 | 丰富 | 部分 |
| 推荐度 | - | ⭐⭐ | ⭐⭐⭐⭐ |

**建议:** 引入 Zustand 替代 ModelConfigContext，可简化代码并提升性能。

---

## 5. 重构优先级建议

### 🔴 P0 - 立即重构 (技术债务)

| 优先级 | 项目 | 文件 | 预计节省代码 |
|--------|------|------|-------------|
| P0-1 | 提取 localStorage 通用 Hook | `hooks/useLocalStorage.ts` | ~80 lines |
| P0-2 | 合并 Generate/Batch Options 类型 | `types/generate.ts` | ~40 lines |
| P0-3 | 统一 StageState 类型定义 | `types/stage.ts` | ~30 lines |

### 🟡 P1 - 短期重构 (1-2 周)

| 优先级 | 项目 | 影响文件 | 预计节省代码 |
|--------|------|----------|-------------|
| P1-1 | 创建通用 Selector 组件 | `components/ui/Selector.tsx` | ~200 lines |
| P1-2 | 提取 Panel 通用布局 | `components/ui/Panel.tsx` | ~100 lines |
| P1-3 | 统一 Artifact 类型转换 | `types/converters.ts` | ~50 lines |
| P1-4 | 合并 UI 布尔状态 | `App.tsx` | ~30 lines |

### 🟢 P2 - 中期重构 (1 个月)

| 优先级 | 项目 | 影响文件 | 说明 |
|--------|------|----------|------|
| P2-1 | 引入 Zustand 状态管理 | `stores/` | 替代 ModelConfigContext |
| P2-2 | 创建统一 API Hook | `hooks/useApi.ts` | 封装 fetch 逻辑 |
| P2-3 | 删除 Legacy History 组件 | `HistorySidebar.tsx` | 已标记为 deprecated |
| P2-4 | 组件目录重组 | `components/` | 按功能域划分 |

### 🔵 P3 - 长期优化

| 优先级 | 项目 | 说明 |
|--------|------|------|
| P3-1 | 引入 React Query/TanStack Query | 替代手动 fetch 和缓存逻辑 |
| P3-2 | 代码分割优化 | 按路由懒加载 |
| P3-3 | 测试覆盖率提升 | 当前测试文件较多但覆盖不完整 |

---

## 6. 具体重构方案

### 6.1 方案 1: localStorage Hook 抽象

```typescript
// src/hooks/useLocalStorage.ts
export function useLocalStorage<T>(
  key: string,
  defaultValue: T,
  options?: { migrate?: (old: unknown) => T }
) {
  const [value, setValue] = useState<T>(() => {
    if (typeof window === 'undefined') return defaultValue;
    try {
      const stored = localStorage.getItem(key);
      if (!stored) return defaultValue;
      const parsed = JSON.parse(stored);
      return options?.migrate ? options.migrate(parsed) : parsed;
    } catch {
      return defaultValue;
    }
  });

  useEffect(() => {
    if (typeof window !== 'undefined') {
      localStorage.setItem(key, JSON.stringify(value));
    }
  }, [key, value]);

  return [value, setValue] as const;
}
```

**可简化:**
- `useTheme.ts`: 从 66 lines → ~25 lines
- `usePromptTemplates.ts`: 从 104 lines → ~60 lines
- `useLocalWorkRecords.ts`: 从 92 lines → ~50 lines

### 6.2 方案 2: 通用 Selector 组件

```typescript
// src/components/ui/Selector.tsx
interface SelectorProps<T> {
  items: T[];
  selectedId?: string;
  onSelect: (item: T) => void;
  getId: (item: T) => string;
  getLabel: (item: T) => string;
  placeholder?: string;
  loading?: boolean;
  error?: string | null;
}

export function Selector<T>({ items, ...props }: SelectorProps<T>) {
  // 统一的 Selector 实现
}
```

**可替换:**
- `ProjectSelector.tsx` → 使用通用组件
- `TemplateSelector.tsx` → 使用通用组件
- `ModelSelector.tsx` → 使用通用组件

### 6.3 方案 3: Zustand Store 结构

```typescript
// src/stores/modelConfigStore.ts
import { create } from 'zustand';

interface ModelConfigState {
  providers: Provider[];
  roleAssignments: Record<WorkflowRole, RoleAssignment | null>;
  loading: boolean;
  error: string | null;
  
  // Actions
  fetchProviders: () => Promise<void>;
  addProvider: (provider: Omit<Provider, 'id'>) => Promise<void>;
  updateProvider: (id: string, updates: Partial<Provider>) => Promise<void>;
  deleteProvider: (id: string) => Promise<void>;
  assignRole: (role: WorkflowRole, providerId: string, modelId: string) => Promise<void>;
}

export const useModelConfigStore = create<ModelConfigState>((set, get) => ({
  // ... implementation
}));
```

---

## 7. 总结与建议

### 发现的主要问题

1. **类型分散:** Artifact、Provider 等核心类型在多个文件中重复定义
2. **localStorage 模式重复:** 7+ 处重复的 localStorage 读写逻辑
3. **Selector 组件重复:** 3 个 Selector 具有 70%+ 相似代码
4. **API 调用分散:** fetch 逻辑散布在 hooks 和 context 中
5. **状态管理边界模糊:** Context 和 Hooks 职责不清

### 预估重构收益

| 指标 | 当前 | 重构后 | 收益 |
|------|------|--------|------|
| 代码行数 | ~12,000 | ~10,500 | -12.5% |
| 重复代码块 | 15+ | 3-5 | -70% |
| 组件数量 | 50+ | 45+ | -10% |
| 类型定义分散度 | 高 | 低 | 显著提升 |

### 推荐执行顺序

```
Week 1: P0 任务 (技术债务清理)
Week 2-3: P1 任务 (组件抽象)
Week 4: P2-1 (Zustand 迁移)
Month 2+: P2/P3 任务
```

---

*报告生成时间: 2026-04-06*
*分析范围: paperbanana-clean/web/src/*
