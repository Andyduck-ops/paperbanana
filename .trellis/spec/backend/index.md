# Backend Development Guidelines

> Best practices and architecture documentation for backend development in PaperBanana.

**[PRE-TAURI]** Several spec files carry a `[PRE-TAURI]` marker, indicating they describe the current web-app architecture (Go HTTP server + React SPA). These documents will need revision when the project migrates to Tauri + Go sidecar.

---

## Overview

This directory contains guidelines for backend development. Each file documents the project's **actual conventions and patterns** (not ideals), with code examples and file references from the real codebase.

---

## Guidelines Index

| Guide | Description | Status |
|-------|-------------|--------|
| [Architecture](./architecture.md) | Layered monolith architecture, data flow, key abstractions, entry points | Filled |
| [Conventions](./conventions.md) | Go naming, code style, import aliases, function design, module patterns | Filled |
| [Directory Structure](./directory-structure.md) | Module organization, ownership boundaries, where to add new code | Filled |
| [Database Guidelines](./database-guidelines.md) | SQLite/GORM patterns, repository pattern, transactions, asset storage | Filled |
| [Error Handling](./error-handling.md) | Error wrapping, sentinel errors, ErrorDetail, HTTP mapping, anti-patterns | Filled |
| [Integrations](./integrations.md) | LLM providers, data storage, auth, monitoring, environment config | Filled |
| [Logging Guidelines](./logging-guidelines.md) | Zap structured logging, log levels, what to log and what not to log | Filled |
| [Quality Guidelines](./quality-guidelines.md) | Tech debt, known bugs, security concerns, performance, test gaps | Filled |

---

## How to Use These Guidelines

1. **Before starting a task**: Read the relevant spec file(s) to understand existing patterns
2. **During implementation**: Follow the conventions and patterns documented here
3. **After implementation**: Check against the [Quality Guidelines](./quality-guidelines.md) review checklist
4. **When in doubt**: Prefer consistency with existing code over introducing new patterns

---

## Pre-Tauri Markers

Files with the `[PRE-TAURI]` marker at the top describe architecture that will change when the project migrates from a Go HTTP server + React SPA to Tauri + Go sidecar. Key changes expected:

- HTTP API will become Tauri IPC commands
- SSE streaming will become Tauri event emission
- Router middleware (auth, CORS, rate-limit) will be replaced or removed
- Frontend API transport layer (`web/src/lib/api.ts`) will be replaced with Tauri invoke calls
- Server entrypoint (`cmd/server/main.go`) will become a sidecar entrypoint

---

**Language**: All documentation is written in **English**.
