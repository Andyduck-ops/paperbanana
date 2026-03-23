# PaperBanana

PaperBanana is a full-stack image refinement workspace with a Go backend and a Vite + React frontend.

## Repository scope

This repository is intended to contain product code only:

- backend application code
- frontend application code
- tests
- runtime and deployment configuration
- product documentation

It intentionally excludes local planning artifacts, nested repositories, runtime databases, logs, and build outputs.

## Stack

- Go backend
- React + TypeScript frontend
- SQLite for local persistence
- Provider configuration for Gemini, OpenAI, Anthropic, and OpenRouter

## Local development

Backend:

```bash
go run ./cmd/server --config ./configs/config.yaml
```

Frontend:

```bash
cd web
npm install
npm run dev
```

Reference benchmark data:

```bash
# Optional: point to a shared local PaperBananaBench directory
set PAPERBANANA_BENCH_ROOT=D:\datasets\PaperBananaBench
```

If `PAPERBANANA_BENCH_ROOT` is not set, the backend uses `data/PaperBananaBench` relative to the repository root. This dataset is intended for local development and should usually stay out of Git.

## Configuration

Copy environment variables from `.env.example` and provide the provider API keys you want to enable. The app exposes provider settings in the UI, while the backend still needs to start successfully for the settings API to respond.

## Deployment

See `DEPLOYMENT.md` plus the provided `Dockerfile`, `docker-compose.yml`, and `nginx.conf`.
