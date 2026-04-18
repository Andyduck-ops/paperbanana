# Pre-Tauri Architecture Reference

> **PRE-TAURI SNAPSHOT** -- This document captures the architecture state before the Tauri + Go sidecar migration. It consolidates tech debt, known issues, and upstream drift analysis so nothing is lost during the transition.

**Snapshot Date:** 2026-04-17

---

## Product Position

PaperBanana is a focused academic paper figure workspace for generation, comparison, and refinement, built on a productized Go + React pipeline.

---

## M1 Milestone Status

**Status:** PASSED with Technical Debt

| Metric | Value |
|--------|-------|
| Overall Completion | 100% (6/6 phases) |
| Golden Data Tests | 298/298 passing |
| Technical Debt | 54 issues |
| Critical Issues | 17 |
| High Priority | 21 |

---

## Critical Issues (P0)

These must be addressed before or during the Tauri migration:

1. **Refine API contract mismatch** -- Frontend expects `image.data`, backend returns only `content`. The advertised refine/iterate workflow is not trustworthy.
   - Files: `web/src/lib/refine.ts`, `internal/api/handlers/refine.go`, `internal/agents/polish/agent.go`

2. **Generation persistence split-brain** -- `project_id`, `folder_id`, `visualization_id` accepted at HTTP layer but not connected to workspace persistence in the generation pipeline.
   - Files: `cmd/server/main.go`, `internal/api/handlers/generate.go`, `internal/application/orchestrator/runner.go`

3. **No Cancel API endpoint** -- No way to abort a running generation from the frontend.
   - Missing: `DELETE /api/v1/sessions/:id` or equivalent

4. **Missing AbortController** -- `useGenerate` hook doesn't support aborting in-flight SSE streams.
   - File: `web/src/hooks/useGenerate.ts`

5. **Security middleware not mounted** -- Auth, CORS, rate limiting, and metrics middleware exist but are not wired into the router.
   - Files: `internal/api/middleware/auth.go`, `internal/api/router.go`

---

## High Priority Issues (P1)

6. **Duplicate asset store implementations** -- Runtime uses weaker filesystem store; hardened store exists but is only covered by tests.
7. **Duplicate frontend config surfaces** -- Channel/settings flow vs legacy popover flow with separate hooks and endpoint assumptions.
8. **Artifact URLs missing project scope** -- Frontend uses `/api/v1/assets/<assetId>`, backend requires `/api/v1/assets/:project_id/:asset_id`.
9. **Plot RCE vulnerability** -- Plot rendering may execute arbitrary code.
10. **Refine accepts unbounded base64** -- No request size limit on the refine endpoint.
11. **Asset store path traversal risk** -- Runtime store joins root and storageKey directly without traversal checks.

---

## Upstream Drift Analysis

The current branch has drifted from the original PaperBanana upstream in these areas:

### What Improved
- Typed staged pipeline, session state, resumable execution (`internal/domain/agent/`, `internal/application/orchestrator/`)
- Provider/model abstraction and persistence-backed configuration
- Batch orchestration and session/history endpoints
- Productized React frontend with stronger workspace shell

### What Drifted
1. **Product identity** -- UI mixes workspace/tool language, backstage configuration, and generic image wording
2. **Frontstage UX** -- App opens into a tool surface, not a continuation-first desk
3. **Retrieval/Planner/Stylist semantics** -- Pragmatic limits changed capability semantics (candidate pre-shortlisting, reduced few-shot, simplified style guides)
4. **Refine/Polish regression** -- Frontend expects image-oriented behavior, backend returns text-oriented content shape
5. **Persistence wiring** -- Rich persistence primitives exist but generation results bypass workspace ownership

---

## Test Suite Status

### Backend (`go test ./...`)
- **FAILING**: `internal/application/agents/critic` (metadata mismatch), `internal/config` (WAL default change)

### Frontend (`cd web && npm run test:run`)
- **31 failing tests** out of 487 across 54 files
- Root causes: Playwright/Vitest conflict, stale golden test assertions, semantic token class drift, i18n key shape changes

### Configuration Issues
- ESLint declared in `package.json` but not installed
- Vitest doesn't exclude Playwright specs
- Playwright config only matches `test-ui.spec.ts`, not `web/e2e/` specs

---

## Architecture Summary (Current)

```
Browser (React SPA)
    ↓ HTTP/SSE (port 8080)
Go Server (Gin + GORM + SQLite)
    ↓ File I/O
SQLite DB + File Assets
    ↓ HTTP
LLM Providers (OpenAI, Gemini, Anthropic, OpenRouter)
```

### Key Components
- **Backend**: Go 1.23, Gin v1.10, GORM + SQLite, Zap logging, Viper config
- **Frontend**: React 19, TypeScript 5.9, Vite 6, Tailwind 4.2, i18next
- **Pipeline**: 5-stage (Retriever → Planner → Stylist → Visualizer → Critic)
- **Storage**: SQLite (sessions, versions, providers) + Filesystem (assets)

### Files That Will Change Most in Tauri Migration
| File | Change Type |
|------|-------------|
| `cmd/server/main.go` | Add port auto-detection, Tauri environment detection |
| `web/src/lib/api.ts` | Add backend port detection for Tauri |
| `web/vite.config.ts` | Add Tauri dev plugin, update proxy config |
| `web/package.json` | Add @tauri-apps/api, @tauri-apps/cli |
| New: `src-tauri/` | Entire Tauri scaffold (Rust) |

### Files That Will NOT Change
- All `internal/application/agents/*` -- Pipeline agents
- All `internal/domain/*` -- Domain types and interfaces
- All `internal/infrastructure/llm/*` -- LLM provider clients
- All `internal/infrastructure/persistence/*` -- Database code
- All `web/src/components/*` -- React components
- All `web/src/hooks/*` -- React hooks (except API URL detection)
- `web/src/lib/sse.ts` -- SSE client (works over localhost)

---

## Recommended Migration Priority

1. Fix P0 issues first (refine contract, persistence wiring, cancel API)
2. Set up Tauri scaffold with Go sidecar
3. Verify SSE streaming works over localhost in Tauri webview
4. Implement sidecar lifecycle management (spawn, health check, graceful shutdown)
5. Build cross-platform binaries and CI pipeline

---

*This snapshot preserves the pre-Tauri architecture state for reference during migration.*
