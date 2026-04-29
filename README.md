# PaperBanana

> AI-powered academic figure generation and refinement workspace.

PaperBanana is an intelligent workspace for researchers to generate, refine, and manage publication-ready figures for academic papers. It combines a multi-agent AI pipeline with an elegant, themeable UI to transform paper context and visual intent into high-quality diagrams and plots.

[English](#overview) | [中文](#概述)

---

## Overview

PaperBanana addresses a critical gap in academic workflows: creating publication-quality figures that accurately reflect research content. Unlike generic image generators, PaperBanana understands academic context through a specialized multi-agent pipeline:

1. **Retrieval** — Searches relevant figure examples from a curated benchmark dataset
2. **Planning** — Analyzes paper context and generates a structured visualization plan
3. **Styling** — Applies academic style guidelines (NeurIPS, IEEE, etc.) to the figure
4. **Visualization** — Executes the plan using code-based rendering (Python/matplotlib, Mermaid, etc.)
5. **Critic** — Reviews the output against academic standards and iteratively improves

The result is figures that are not just visually appealing, but academically accurate and publication-ready.

## Key Features

### Multi-Agent Generation Pipeline
- **5-stage orchestrated pipeline**: Retriever → Planner → Stylist → Visualizer → Critic
- **Batch generation**: Generate multiple candidate figures in parallel and compare
- **Iterative refinement**: Feed generated figures back for style and content improvements
- **Session resumption**: Save and resume generation sessions at any stage

### Intelligent Provider Management
- **Multi-provider support**: Gemini, OpenAI, Anthropic, OpenRouter
- **Role-based routing**: Assign different providers to different pipeline stages
- **API key management**: Secure encrypted storage with per-provider configuration
- **Model auto-selection**: Intelligent model picking based on task requirements

### Elegant, Themeable UI
- **Academic design system**: Warm, scholarly aesthetic with careful typography
- **Light/Dark themes**: Claude Light & Linear Dark built-in
- **Responsive layout**: Optimized for desktop research workflows
- **Real-time progress**: Live stage-by-stage generation feedback via SSE
- **Internationalization**: English & Chinese support

### Workspace & Persistence
- **Project management**: Organize figures into projects with folder hierarchies
- **Version history**: Track figure iterations with full rollback capability
- **Asset management**: Store and retrieve generated figures with metadata
- **Local-first**: SQLite database with optional Redis caching

### Desktop App (Tauri v2)
- **Native desktop wrapper**: Cross-platform EXE via Tauri + Go sidecar
- **Auto-updater**: Built-in update mechanism
- **System tray**: Background operation support

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Frontend                              │
│  React 19 + TypeScript + Tailwind CSS + Zustand             │
│  Vite build | Vitest testing | Playwright E2E               │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼ HTTP/SSE
┌─────────────────────────────────────────────────────────────┐
│                        Backend                               │
│  Go 1.25 + Gin + GORM + SQLite                              │
│                                                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │  API Layer  │  │   Agents    │  │   Infrastructure    │  │
│  │  (REST/SSE) │  │  Retriever  │  │  LLM Clients        │  │
│  │  Middleware │  │  Planner    │  │  SQLite/GORM        │  │
│  │  Validation │  │  Stylist    │  │  Redis Cache        │  │
│  └─────────────┘  │  Visualizer │  │  Crypto/AES-GCM     │  │
│                   │  Critic     │  │  Resilience         │  │
│                   └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.25, Gin, GORM, SQLite |
| Frontend | React 19, TypeScript, Tailwind CSS 4, Vite 6 |
| State | Zustand 5, React Query 5 |
| i18n | i18next |
| Testing | Vitest, Playwright, Go test |
| Desktop | Tauri v2 (Rust) |
| Observability | Prometheus metrics, Zap logging |

## Quick Start

### One-Click Start

**Windows:**
```batch
start.bat
```

**Linux/Mac:**
```bash
./start.sh
```

This will check dependencies, install frontend packages if needed, and start both backend (port 8080) and frontend dev server (port 5173).

### Manual Development

**Backend:**
```bash
go run ./cmd/server --config ./configs/config.yaml
```

**Frontend:**
```bash
cd web
npm install
npm run dev
```

**Build for production:**
```bash
# Frontend
cd web && npm run build

# Backend
go build -o server.exe ./cmd/server
```

### Desktop App (Tauri)

```bash
cd src-tauri
cargo tauri dev    # Development
cargo tauri build  # Production build
```

## Configuration

Copy `.env.example` to `.env` and configure your API keys:

```bash
cp .env.example .env
```

Supported providers: Gemini, OpenAI, Anthropic, OpenRouter. Configure via the in-app Settings panel or `configs/config.yaml`.

## Deployment

### Docker

```bash
docker-compose up -d
```

See `Dockerfile` and `docker-compose.yml` for details.

### Benchmark Dataset

Set `PAPERBANANA_BENCH_ROOT` to point to your local PaperBananaBench dataset for retrieval-augmented generation:

```bash
export PAPERBANANA_BENCH_ROOT=/path/to/PaperBananaBench
```

## Project Structure

```
paperbanana/
├── cmd/server/           # Go backend entry point
├── internal/
│   ├── api/              # HTTP handlers, middleware, DTOs
│   ├── application/      # Business logic: agents, orchestrator, persistence
│   ├── domain/           # Domain models and interfaces
│   ├── infrastructure/   # LLM clients, SQLite, crypto, cache
│   └── config/           # Configuration loading
├── web/                  # React frontend
│   ├── src/
│   │   ├── components/   # UI components
│   │   ├── hooks/        # React hooks
│   │   ├── stores/       # Zustand state stores
│   │   ├── themes/       # CSS theme files
│   │   └── lib/          # Utilities
│   └── e2e/              # Playwright tests
├── src-tauri/            # Tauri desktop app
├── configs/              # Configuration templates
├── data/                 # SQLite database (gitignored)
└── docs/                 # Architecture documentation
```

## Development

### Running Tests

```bash
# Go backend
go test ./...

# Frontend unit tests
cd web && npm run test:run

# E2E tests
cd web && npx playwright test
```

### Code Quality

```bash
# Frontend lint
cd web && npm run lint

# Go format
go fmt ./...
```

## License

MIT License — see LICENSE file for details.

---

## 概述

PaperBanana 是一个面向学术研究者的智能图表生成与精修工作空间。它通过专门的多智能体流水线，将论文上下文和视觉意图转化为高质量的学术插图。

### 核心能力

- **五阶段智能体流水线**：检索 → 规划 → 风格化 → 可视化 → 评审
- **批量生成与对比**：并行生成多个候选图表，择优选用
- **迭代精修**：将生成结果反馈进行风格和内容的持续优化
- **多模型路由**：支持 Gemini、OpenAI、Anthropic、OpenRouter，可按阶段分配不同模型
- **项目化管理**：图表按项目组织，支持版本历史和资产追踪
- **桌面端应用**：基于 Tauri v2 的跨平台原生应用

### 技术栈

- **后端**：Go 1.25 + Gin + GORM + SQLite
- **前端**：React 19 + TypeScript + Tailwind CSS 4 + Vite 6
- **状态管理**：Zustand 5 + React Query 5
- **桌面端**：Tauri v2 (Rust)

### 快速开始

```bash
# Windows
start.bat

# Linux/Mac
./start.sh
```

或手动启动：

```bash
# 后端
go run ./cmd/server --config ./configs/config.yaml

# 前端
cd web && npm install && npm run dev
```
