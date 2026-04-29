# PaperBanana

> AI 驱动的学术图表生成与精修工作空间

PaperBanana 是一个面向学术研究者的智能工作空间，用于生成、精修和管理学术论文中的出版级图表。它通过多智能体 AI 流水线与优雅的可主题化 UI，将论文上下文和视觉意图转化为高质量的示意图和 plots。

[English README](README.md)

---

## 概述

PaperBanana 解决了学术工作流中的一个关键痛点：创建能够准确反映研究内容的出版级图表。与通用图像生成器不同，PaperBanana 通过专门的多智能体流水线理解学术语境：

1. **检索（Retriever）** — 从精选的基准数据集中搜索相关图表示例
2. **规划（Planner）** — 分析论文上下文并生成结构化的可视化方案
3. **风格化（Stylist）** — 将学术风格指南（NeurIPS、IEEE 等）应用于图表
4. **可视化（Visualizer）** — 使用代码渲染（Python/matplotlib、Mermaid 等）执行方案
5. **评审（Critic）** — 对照学术标准审查输出并迭代改进

最终生成的图表不仅视觉上美观，更具有学术准确性和出版就绪品质。

## 核心功能

### 多智能体生成流水线
- **五阶段编排流水线**：检索 → 规划 → 风格化 → 可视化 → 评审
- **批量生成**：并行生成多个候选图表，对比择优
- **迭代精修**：将生成结果反馈进行风格和内容的持续优化
- **会话恢复**：保存并在任意阶段恢复生成会话

### 智能提供商管理
- **多提供商支持**：Gemini、OpenAI、Anthropic、OpenRouter
- **基于角色的路由**：为不同流水线阶段分配不同的提供商
- **API 密钥管理**：安全加密存储，支持按提供商配置
- **模型自动选择**：基于任务需求的智能模型选择

### 优雅的可主题化 UI
- **学术设计系统**：温暖的学术美学，精心设计的排版
- **明暗主题**：内置 Claude Light 和 Linear Dark
- **响应式布局**：针对桌面研究工作流优化
- **实时进度**：通过 SSE 提供逐阶段生成反馈
- **国际化**：支持英文和中文

### 工作空间与持久化
- **项目管理**：将图表组织到具有文件夹层次结构的项目中
- **版本历史**：追踪图表迭代，支持完整回滚
- **资产管理**：存储和检索带元数据的生成图表
- **本地优先**：SQLite 数据库，可选 Redis 缓存

### 桌面应用（Tauri v2）
- **原生桌面封装**：通过 Tauri + Go sidecar 实现跨平台 EXE
- **自动更新**：内置更新机制
- **系统托盘**：支持后台运行

## 架构

```
┌─────────────────────────────────────────────────────────────┐
│                        前端                                   │
│  React 19 + TypeScript + Tailwind CSS + Zustand             │
│  Vite 构建 | Vitest 测试 | Playwright E2E                   │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼ HTTP/SSE
┌─────────────────────────────────────────────────────────────┐
│                        后端                                   │
│  Go 1.25 + Gin + GORM + SQLite                              │
│                                                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │  API 层     │  │   智能体    │  │     基础设施        │  │
│  │  (REST/SSE) │  │  检索器     │  │  LLM 客户端         │  │
│  │  中间件     │  │  规划器     │  │  SQLite/GORM        │  │
│  │  验证       │  │  风格化器   │  │  Redis 缓存         │  │
│  └─────────────┘  │  可视化器   │  │  加密/AES-GCM       │  │
│                   │  评审器     │  │  弹性机制           │  │
│                   └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端 | Go 1.25, Gin, GORM, SQLite |
| 前端 | React 19, TypeScript, Tailwind CSS 4, Vite 6 |
| 状态管理 | Zustand 5, React Query 5 |
| 国际化 | i18next |
| 测试 | Vitest, Playwright, Go test |
| 桌面端 | Tauri v2 (Rust) |
| 可观测性 | Prometheus 指标, Zap 日志 |

## 快速开始

### 一键启动

**Windows:**
```batch
start.bat
```

**Linux/Mac:**
```bash
./start.sh
```

脚本将检查依赖、安装前端包（如需要），并同时启动后端（端口 8080）和前端开发服务器（端口 5173）。

### 手动开发

**后端：**
```bash
go run ./cmd/server --config ./configs/config.yaml
```

**前端：**
```bash
cd web
npm install
npm run dev
```

**生产构建：**
```bash
# 前端
cd web && npm run build

# 后端
go build -o server.exe ./cmd/server
```

### 桌面应用（Tauri）

```bash
cd src-tauri
cargo tauri dev    # 开发模式
cargo tauri build  # 生产构建
```

## 配置

复制 `.env.example` 到 `.env` 并配置 API 密钥：

```bash
cp .env.example .env
```

支持的提供商：Gemini、OpenAI、Anthropic、OpenRouter。可通过应用内设置面板或 `configs/config.yaml` 配置。

## 部署

### Docker

```bash
docker-compose up -d
```

详见 `Dockerfile` 和 `docker-compose.yml`。

### 基准数据集

设置 `PAPERBANANA_BENCH_ROOT` 指向本地 PaperBananaBench 数据集，以启用检索增强生成：

```bash
export PAPERBANANA_BENCH_ROOT=/path/to/PaperBananaBench
```

## 项目结构

```
paperbanana/
├── cmd/server/           # Go 后端入口
├── internal/
│   ├── api/              # HTTP 处理器、中间件、DTO
│   ├── application/      # 业务逻辑：智能体、编排器、持久化
│   ├── domain/           # 领域模型和接口
│   ├── infrastructure/   # LLM 客户端、SQLite、加密、缓存
│   └── config/           # 配置加载
├── web/                  # React 前端
│   ├── src/
│   │   ├── components/   # UI 组件
│   │   ├── hooks/        # React Hooks
│   │   ├── stores/       # Zustand 状态存储
│   │   ├── themes/       # CSS 主题文件
│   │   └── lib/          # 工具函数
│   └── e2e/              # Playwright 测试
├── src-tauri/            # Tauri 桌面应用
├── configs/              # 配置模板
├── data/                 # SQLite 数据库（gitignored）
└── docs/                 # 架构文档
```

## 开发

### 运行测试

```bash
# Go 后端
go test ./...

# 前端单元测试
cd web && npm run test:run

# E2E 测试
cd web && npx playwright test
```

### 代码质量

```bash
# 前端代码检查
cd web && npm run lint

# Go 格式化
go fmt ./...
```

## 许可证

MIT 许可证 — 详见 LICENSE 文件。
