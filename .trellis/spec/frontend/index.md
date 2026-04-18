# Frontend Development Guidelines

> Best practices and conventions for the PaperBanana frontend, built with React 19 + TypeScript + Vite + Tailwind CSS.

---

## Overview

This directory contains guidelines for frontend development. Each file documents **actual conventions** observed in the `web/src/` codebase -- not ideals. Follow these when writing new code, reviewing PRs, or onboarding to the project.

> **Pre-Tauri Marker**: These guidelines describe the current browser-based SPA architecture. A migration to Tauri + Go sidecar is planned. Files and patterns marked with `[PRE-TAURI]` will change during that migration. Update this spec when the migration begins.

---

## Guidelines Index

| Guide | Description | Status |
|-------|-------------|--------|
| [Directory Structure](./directory-structure.md) | Module organization, file layout, naming rules | Active |
| [Conventions](./conventions.md) | Naming, code style, imports, error handling, comments | Active |
| [State Management](./state-management.md) | Hook-and-context model, Zustand migration, persistence | Active |
| [Component Patterns](./component-patterns.md) | Component design, props, i18n, co-location, fragile areas | Active |
| [Testing](./testing.md) | Vitest setup, test patterns, mocking, known issues | Active |

---

## Tech Stack

| Technology | Version | Purpose |
|------------|---------|---------|
| React | 19.2.4 | UI framework |
| TypeScript | 5.9.3 | Type system |
| Vite | 6.0.0 | Build tool and dev server |
| Tailwind CSS | 4.2.1 | Utility-first styling |
| Zustand | 5.0.12 | Global state stores (migrating from hooks) |
| @tanstack/react-query | 5.96.2 | Server state management |
| i18next / react-i18next | 25.x / 16.x | Internationalization |
| Vitest | 3.2.4 | Unit testing |
| Playwright | 1.58.x | E2E testing |
| @testing-library/react | 16.3.2 | Component testing utilities |

---

## Architecture Summary

```
Browser SPA (Vite dev server / static build)
  |
  +-- React 19 app (App.tsx shell)
  |     +-- Zustand stores (appStore, providerStore, generationStore)
  |     +-- React Query (server cache)
  |     +-- Custom hooks (useGenerate, useHistory, useRefine, ...)
  |     +-- Feature components (workspace, history, settings, refine)
  |
  +-- Transport layer (web/src/lib/)
  |     +-- api.ts      -> REST client
  |     +-- sse.ts      -> SSE streaming (stage progress)
  |     +-- configStream.ts -> Config SSE subscription
  |     +-- refine.ts   -> Refine REST client
  |
  +-- Go backend (sidecar, :8080)  [PRE-TAURI: currently standalone server]
```

---

## How to Use These Guidelines

1. **Before coding**: Read the relevant guideline file for the area you are changing.
2. **During coding**: Follow naming, import, and error-handling conventions exactly.
3. **During review**: Check changed files against the applicable guideline.
4. **After bugs**: Update guidelines with lessons learned -- capture *why*, not just *what*.

---

**Language**: All documentation is written in **English**.
