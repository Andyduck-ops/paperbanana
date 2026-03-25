# UI Reference Research Report

## 调研目标
分析 Cherry Studio 和 NewAPI 的模型管理交互设计，为 PaperBanana UI 重构提供参考。

---

## 1. Cherry Studio 分析

### 项目概述
- **技术栈**: Electron + React
- **定位**: 多 LLM Provider 桌面客户端
- **平台**: Windows, Mac, Linux

### 核心设计发现

#### 1.1 Provider 支持策略
```
支持类型:
├── 云服务商: OpenAI, Gemini, Anthropic
├── 本地模型: Ollama, LM Studio
└── MCP 协议: Model Context Protocol
```

#### 1.2 模型选择器设计 (推断)
基于 Electron + React 架构，典型的模型选择器设计模式：
- 下拉式 Provider 选择
- 级联 Model 选择
- 内联 API Key 配置
- 状态指示器 (已配置/未配置)

#### 1.3 UI 风格特点
- 支持亮/暗主题
- 透明窗口效果
- 完整 Markdown 渲染
- 预配置 300+ AI Assistants

### 可借鉴点
1. **简洁的 Provider 切换** - 单一入口，不占用常驻空间
2. **预配置助手库** - 快速开始，降低配置门槛
3. **透明窗口** - 现代感视觉效果

---

## 2. NewAPI / One-API 分析

### 项目概述
- **技术栈**: Go 后端 + React 前端
- **定位**: LLM API 管理平台
- **特点**: 多渠道负载均衡

### 核心设计发现

#### 2.1 渠道管理架构
```
渠道特性:
├── 负载均衡 - 多渠道请求分发
├── 渠道分组 - 不同倍率设置
├── 模型映射 - 请求重定向
├── 自动重试 - 故障转移
└── 渠道测试 - 定期可用性检查
```

#### 2.2 模型配置逻辑
- 每个渠道可配置模型列表
- 模型映射支持请求重构
- 支持自定义系统名称和 Logo

#### 2.3 UI 组件结构
```
web/
├── default/     # 默认主题
├── berry/       # Berry 主题
└── air/         # Air 主题
```

### 可借鉴点
1. **渠道分组** - 按用途/倍率分类管理
2. **模型映射** - 灵活的模型重定向
3. **渠道测试** - 一键验证连接状态
4. **多主题支持** - 可定制的外观

---

## 3. 现代 AI 工具 UI 趋势分析

### 3.1 模型选择器模式对比

| 工具 | 选择器位置 | 展示方式 | 配置入口 |
|------|-----------|---------|---------|
| ChatGPT | 顶部左上 | 下拉菜单 | 设置页 |
| Claude | 顶部居中 | 标签切换 | 无 (固定) |
| Cursor | 底部状态栏 | Popover | 快捷配置 |
| Cherry Studio | 顶部工具栏 | 下拉面板 | 内联 |
| Linear | 顶部右侧 | 命令面板 | 弹窗 |

### 3.2 最佳实践总结

#### 模型指示器设计
```
推荐方案: 右上角 Tag + Popover

显示内容:
├── 简短名称: "Gemini⚡" 或 "GPT-4o"
├── 状态指示: ⚡ (可用) / ⚠ (未配置)
└── 多 Provider: "Gemini + OpenAI"

点击展开:
├── 当前模型配置
├── 快速切换选项
└── 完整设置入口
```

#### 配置流程
```
最小化步骤:
1. 点击指示器 → 展开配置面板
2. 选择 Provider → 自动加载模型列表
3. 输入 API Key → 验证连接
4. 选择模型 → 完成

避免:
× 跳转独立设置页面
× 多层嵌套配置
× 无引导的空白输入
```

---

## 4. PaperBanana UI 设计建议

### 4.1 模型配置 (GD-UI-006)

```
右上角指示器:
┌─────────────────┐
│ Gemini⚡ + OpenAI│  ← 显示启用的 Provider
└─────────────────┘
        │ 点击
        ▼
┌───────────────────────────────┐
│ 检索模型: Gemini 2.0 Flash    │
│ 生图模型: GPT-4o              │
├───────────────────────────────┤
│ 商家:                         │
│ ○ Gemini   [✓] [🔑]          │
│   └ gemini-2.0-flash [-]     │
│   └ gemini-1.5-pro    [+]    │
│ ○ OpenAI   [✓] [🔑]          │
│   └ gpt-4o            [-]    │
│   └ gpt-4o-mini       [+]    │
└───────────────────────────────┘
```

### 4.2 状态展示 (GD-UI-009)

参考 Vercel 部署进度：
```
┌─────────────────────────────────────────────┐
│ 🔍 Retrieving references... ━━━━━━░░░░ 2/5 │
└─────────────────────────────────────────────┘

特点:
- 单行，极简
- 图标 + 文字 + 细进度条
- 失败时变红
```

### 4.3 精修迭代 (GD-UI-007)

```
迭代时间线:
┌─────┐   ┌─────┐   ┌─────┐
│原图 │ → │ R1  │ → │ R2  │ → ...
└─────┘   └─────┘   └─────┘
            ✓        运行中

点击任意轮次展开对比:
┌────────┐    ┌────────┐
│ Before │ →  │ After  │
└────────┘    └────────┘
改动: 增强了色彩对比度
```

---

## 5. 实施建议

### 5.1 技术选型
- **UI 框架**: React 19 + Tailwind CSS v4
- **状态管理**: Zustand (轻量)
- **动画**: Framer Motion (平滑过渡)
- **图标**: Lucide React (现代图标库)

### 5.2 组件优先级
1. P0: FloatingInput + ProgressLine
2. P0: ModelPopover + ModelIndicator
3. P1: HistoryPopover
4. P1: RefineIteration
5. P2: BatchVariantGrid

### 5.3 设计系统
```css
/* 核心变量 */
--radius-popover: 12px;
--radius-input: 24px;  /* 大圆角现代感 */
--shadow-floating: 0 8px 32px rgba(0,0,0,0.12);
--blur-glass: blur(12px);
--transition-smooth: 300ms cubic-bezier(0.4, 0, 0.2, 1);
```

---

## 6. 参考资源

- Cherry Studio: https://github.com/kangfenmao/cherry-studio
- One-API: https://github.com/songquanpeng/one-api
- Vercel 部署进度样式
- Linear 任务状态设计
