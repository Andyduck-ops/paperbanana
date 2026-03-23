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

## Configuration

Copy environment variables from `.env.example` and provide the provider API keys you want to enable. The app exposes provider settings in the UI, while the backend still needs to start successfully for the settings API to respond.

## Deployment

See `DEPLOYMENT.md` plus the provided `Dockerfile`, `docker-compose.yml`, and `nginx.conf`.
