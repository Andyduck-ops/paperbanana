# Quality Guidelines

> Code quality standards, known issues, and risk areas for PaperBanana.

---

## Overview

This document catalogs the actual state of code quality in PaperBanana: known tech debt, bugs, security concerns, performance issues, and test gaps. It is not an aspirational document; it records what exists today so that future contributors can make informed decisions.

---

## Forbidden Patterns

### 1. `errors.Is(err, errors.New(...))` -- Always False

~~Using `errors.New` inline with `errors.Is` creates a new pointer each time, so the comparison always fails. This pattern exists in three files:~~

~~| File | Line |~~
~~|------|------|~~
~~| `internal/api/handlers/history.go` | 750-751 |~~
~~| `internal/api/handlers/workspace.go` | 426 |~~
~~| `internal/api/handlers/assets.go` | 238 |~~

**Status**: ✅ Fixed (2026-04-18). Replaced with `err.Error() == "..."` string comparison. In the future, these should be migrated to package-level sentinel errors for `errors.Is` compatibility.

### 2. Direct GORM Usage Outside Infrastructure

Application services and handlers should not import or use `*gorm.DB` directly. All database access must go through repository interfaces.

**Exception**: `TxManager` and `HealthChecker` in `internal/api/router.go` legitimately use `*gorm.DB` for transaction management and health checks.

### 3. Logging Sensitive Data

Never log API keys, encryption keys, or large payloads (base64 images, full session snapshots).

---

## Required Patterns

### 1. Error Wrapping with Context

All errors must be wrapped with descriptive context:

```go
// Required
return nil, fmt.Errorf("create project: %w", err)

// Forbidden
return nil, err
```

### 2. Structured Logging

All log calls must use Zap structured fields:

```go
// Required
logger.Info("request", zap.String("method", c.Request.Method))

// Forbidden
logger.Info(fmt.Sprintf("method=%s", c.Request.Method))
```

### 3. Compile-Time Interface Checks

All repository implementations must include:

```go
var _ workspace.ProjectRepository = (*ProjectRepository)(nil)
```

### 4. Project-Scoped Queries

All workspace queries must include `project_id` filtering. Use the `ProjectScopedQuery` interface for compile-time verification.

---

## Known Tech Debt

### Generation Persistence Split-Brain

The generate flow can persist artifacts through two paths: (1) via the `AssetServiceAdapter` in the generate handler, and (2) via the `PersistentSnapshotStore` saving session state. These two paths can produce inconsistent state if one fails and the other succeeds.

### Duplicate Asset Store Implementations

Two separate asset store implementations exist:
- `internal/infrastructure/persistence/sqlite/local_asset_store.go` -- Simple, no path traversal protection
- `internal/infrastructure/assets/localstore/store.go` -- Full-featured with path traversal checks and SHA-256 verification

These should be consolidated. New code should use the `localstore` package.

### Duplicate Config Surfaces

Provider configuration can come from:
1. YAML config file (`configs/config.yaml`)
2. Environment variables (`PAPERBANANA_*`, provider-specific like `GEMINI_API_KEY`)
3. Database-stored provider settings (via `configservice.Service`)

These three sources can conflict. The startup sync (`syncStartupProviders`) attempts to reconcile them, but the precedence is not always clear.

---

## Known Bugs

### Refine API Contract Mismatch

The refine endpoint's request/response DTOs may not fully match what the frontend sends. Verify `internal/api/dto/refine.go` against `web/src/lib/refine.ts` before making changes.

### Legacy Popover Endpoints

Some history endpoints (`/api/v1/session/latest`, `/api/v1/sessions/recent`) return data shaped for a legacy popover UI component that may no longer exist in the current frontend.

### Artifact URL Construction

Generated artifacts include `AssetID` and `ProjectID` fields that the frontend uses to construct download URLs. If either field is missing or inconsistent, asset downloads fail silently.

### Workspace Hierarchy Validation

Folder reparenting and deletion operations may not fully validate workspace hierarchy constraints (e.g., circular references in folder trees, cross-project moves).

### SSE Event Type Drift

Frontend `web/src/lib/sse.ts` and backend `internal/domain/agent/events.go` had mismatched event type constants. Specifically:
- `result` and `error` were used in frontend union but absent from backend `EventType` constants
- `batch_result` was emitted by `internal/api/handlers/batch.go` but absent from frontend `SSEEventType`

**Status**: ✅ Fixed (2026-04-18). Added `EventResult`, `EventError`, and `EventBatchResult` to backend constants; added `'batch_result'` to frontend union. Both sides now include explicit `batch_result` handling in dispatch logic.

---

## Security Concerns

### Auth Middleware Not Mounted

`internal/api/middleware/auth.go` implements API key authentication with constant-time comparison, but the `Auth` middleware is **not mounted** in any router configuration in `internal/api/router.go`. All endpoints are currently unauthenticated.

### History/Version Endpoints Expose Internal Details

The history and version endpoints return internal schema details (e.g., `schema_version`, full session snapshot data) that could reveal implementation details to potential attackers.

### Refine Accepts Unbounded Base64

The refine endpoint accepts base64-encoded images in the request body with no size limit enforcement beyond the HTTP server's `ReadTimeout`. Large payloads could cause memory exhaustion.

### Asset Store Path Traversal Risk

`internal/infrastructure/persistence/sqlite/local_asset_store.go` uses `filepath.Join(root, storageKey)` without validating that the resulting path stays within the root directory. A crafted `storageKey` like `../../etc/passwd` could read arbitrary files.

The alternative implementation `internal/infrastructure/assets/localstore/store.go` does check for path traversal. See "Duplicate Asset Store Implementations" above.

---

## Performance Concerns

### Batch Results In Memory Without Eviction

`BatchResultModel` stores batch execution results as JSON blobs in the database, but there is no eviction or TTL mechanism. Over time, batch results accumulate indefinitely.

### Large Assets Fully Materialized

Asset bytes are fully loaded into memory via `os.ReadFile` before being written to the HTTP response. For very large assets, this can cause significant memory pressure.

### SSE with Weak Back-Pressure

The SSE streaming implementation uses buffered channels with a fixed size (default 32). If the client cannot consume events fast enough, the channel fills up and events are silently dropped when `groupCtx.Done()` is selected:

```go
// internal/application/orchestrator/runner.go
select {
case publicEvents <- event:
case <-groupCtx.Done():
    // Event silently dropped
}
```

---

## Test Gaps

### Security Middleware Untested

`internal/api/middleware/auth.go` has no test coverage. The `Auth`, `OptionalAuth`, and `BearerAuth` middleware functions are untested despite implementing security-critical logic.

### Go Test Suite Red in Config/Critic

Some Go tests in the config and critic agent packages are currently failing. Before adding new tests, run the existing suite to identify the current state:

```bash
go test ./internal/application/agents/critic/...
go test ./internal/application/config/...
```

### Frontend Tests Not Runnable from Clean Checkout

The frontend test suite (`web/`) requires specific setup that may not be documented. Running `npm test` from a clean checkout may fail due to missing dependencies or configuration.

---

## Code Review Checklist

When reviewing backend code changes, verify:

- [ ] Errors are wrapped with context (`fmt.Errorf("context: %w", err)`)
- [ ] No `errors.Is(err, errors.New(...))` patterns
- [ ] Structured logging (no interpolated strings in log calls)
- [ ] No sensitive data in logs (API keys, encryption keys, large payloads)
- [ ] New repository methods include project-scoped filtering
- [ ] Compile-time interface check present (`var _ Interface = (*Impl)(nil)`)
- [ ] New endpoints have input validation
- [ ] Asset paths are validated against traversal
- [ ] New GORM models registered in `AllModels()`
