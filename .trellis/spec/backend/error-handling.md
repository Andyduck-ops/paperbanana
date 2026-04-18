# Error Handling

> How errors are created, propagated, and mapped in PaperBanana.

---

## Overview

PaperBanana follows a layered error strategy: domain code creates and wraps errors with context, application code classifies them, and HTTP handlers map them to status codes at the edge. There are no error types or custom error structs in the domain layer; instead, the project uses `fmt.Errorf` wrapping, sentinel errors, and a structured `ErrorDetail` type for pipeline events.

---

## Error Wrapping Convention

All errors are wrapped with contextual information using `fmt.Errorf` and the `%w` verb. This preserves the error chain for `errors.Is` and `errors.As` unwrapping.

```go
// Bootstrap wraps underlying errors with context
if err := ensureDatabaseDir(cfg.DatabasePath); err != nil {
    return nil, fmt.Errorf("ensure database directory: %w", err)
}

// Repository methods add entity context
return nil, fmt.Errorf("project %s: %w", id, workspace.ErrNotFound)
```

**Pattern**: `fmt.Errorf("<what was happening>: %w", err)`

The context prefix describes the operation, not the error itself. This makes error messages traceable when they surface in logs.

---

## Sentinel Errors

Named sentinel errors are defined at package level using `errors.New`. They are used for comparison with `errors.Is` in callers.

```go
// internal/application/orchestrator/runner.go
var (
    ErrResumeRequiresSession = errors.New("orchestrator: resume requires session id")
    ErrResumeSnapshotMissing = errors.New("orchestrator: resume snapshot not found")
    ErrResumeStoreMissing    = errors.New("orchestrator: resume requires snapshot store")
    ErrResumeSessionNotValid = errors.New("orchestrator: resume session state is not valid for resumption")
    ErrResumePipelineEmpty   = errors.New("orchestrator: resume session has empty pipeline")
)

// internal/infrastructure/agentstate/store.go
var (
    ErrSnapshotNotFound = errors.New("snapshot not found")
    ErrInvalidSnapshot  = errors.New("invalid snapshot data")
)
```

These are compared in handlers:

```go
// internal/api/handlers/generate.go
func (h *Handler) respondRunError(c *gin.Context, err error) {
    status := http.StatusInternalServerError
    if errors.Is(err, orchestrator.ErrResumeRequiresSession) ||
        errors.Is(err, orchestrator.ErrResumeSnapshotMissing) ||
        errors.Is(err, orchestrator.ErrResumeStoreMissing) {
        status = http.StatusBadRequest
    }
    c.JSON(status, gin.H{"error": err.Error()})
}
```

---

## ErrorDetail (Pipeline Errors)

The `domainagent.ErrorDetail` struct is the standardized error type for pipeline stage failures. It is emitted as part of `stage_failed` events and serialized to clients via SSE.

```go
// internal/domain/agent/errors.go
type ErrorDetail struct {
    Message    string `json:"message"`
    Code       string `json:"code"`
    Retryable  bool   `json:"retryable"`
    Stage      string `json:"stage,omitempty"`
    Category   string `json:"category"`
    Suggestion string `json:"suggestion"`
}
```

### Error Codes and Categories

| Category | Codes | Retryable |
|----------|-------|-----------|
| Transient | `llm_timeout`, `rate_limit`, `service_unavailable`, `network_error`, `stage_timeout` | Yes |
| Permanent | `invalid_input`, `unsupported_type`, `resource_not_found`, `cancelled` | No |
| Configuration | `invalid_config`, `missing_api_key`, `invalid_model` | No (user action needed) |
| Internal | `execution_failed`, `internal_error`, `unknown` | No |

### Error Classification

The `ClassifyError` function inspects error messages to classify errors into standard codes:

```go
// internal/domain/agent/errors.go
func ClassifyError(err error) ErrorCode {
    errStr := err.Error()
    switch {
    case containsAny(errStr, []string{"timeout", "deadline exceeded"}):
        return ErrCodeLLMTimeout
    case containsAny(errStr, []string{"rate limit", "429"}):
        return ErrCodeRateLimit
    // ...
    }
}
```

This is a heuristic classifier, not a type system. It relies on error message string matching.

### Wrapping Agent Errors

The `WrapAgentError` helper creates a standardized `ErrorDetail` from a raw error:

```go
detail := domainagent.WrapAgentError(err, stage, errorCode)
```

---

## errors.Join for Cleanup Paths

When multiple operations can fail during cleanup (e.g., agent cleanup after execution error), `errors.Join` combines them without losing either:

```go
// internal/application/orchestrator/runner.go
output, err := stageAgent.Execute(stageCtx, stageInput)
if err != nil {
    if cleanupErr := stageAgent.Cleanup(ctx); cleanupErr != nil {
        err = errors.Join(err, cleanupErr)
    }
    // ...
}
```

Also used when persisting snapshots after a failed stage:

```go
if persistErr := r.persistSnapshot(tracker, stageState); persistErr != nil {
    err = errors.Join(err, persistErr)
}
```

---

## HTTP Error Mapping

Handlers map domain errors to HTTP codes close to the edge (in the handler layer, not in application code). The pattern is:

1. Handler calls application service
2. If error, check with `errors.Is` against known sentinels
3. Map to appropriate HTTP status code
4. Return JSON error response

```go
// Standard pattern in handlers
func (h *Handler) respondRunError(c *gin.Context, err error) {
    h.logger.Error("generation failed", zap.Error(err))
    status := http.StatusInternalServerError
    if errors.Is(err, orchestrator.ErrResumeRequiresSession) {
        status = http.StatusBadRequest
    }
    c.JSON(status, gin.H{"error": err.Error()})
}
```

---

## Anti-Pattern: Inline Sentinel Errors

**Do NOT** use `errors.Is(err, errors.New("..."))` for error comparison. This pattern will never match because `errors.New` creates a new pointer each time.

This anti-pattern currently exists in three handler files:

```go
// BAD - internal/api/handlers/history.go:750-751
return errors.Is(err, errors.New("not found")) ||
    errors.Is(err, errors.New("no resumable session found"))

// BAD - internal/api/handlers/workspace.go:426
return errors.Is(err, errors.New("not found"))

// BAD - internal/api/handlers/assets.go:238
return errors.Is(err, errors.New("asset not found"))
```

These comparisons will always return `false`. The correct approach is to define package-level sentinel errors and compare against those:

```go
// GOOD - Define in domain or infrastructure package
var ErrNotFound = errors.New("not found")

// GOOD - Compare in handler
return errors.Is(err, workspace.ErrNotFound)
```

Note: `internal/infrastructure/assets/localstore/store.go` correctly defines `ErrAssetNotFound` as a package-level sentinel, but `handlers/assets.go` does not reference it.

---

## Validation Errors

Input validation errors are returned directly as `400 Bad Request` with descriptive messages. They are not wrapped in `ErrorDetail`:

```go
// internal/api/handlers/generate.go
func validateGenerateRequest(req GenerateRequest) error {
    if len(req.Prompt) > MaxPromptLength {
        return fmt.Errorf("prompt exceeds maximum length of %d characters", MaxPromptLength)
    }
    if strings.TrimSpace(req.Prompt) == "" && strings.TrimSpace(req.Content) == "" {
        return errors.New("prompt, content, or visual_intent is required")
    }
    // ...
}
```

---

## Common Mistakes

1. **`errors.Is(err, errors.New(...))`**: Always false. Use package-level sentinel errors.

2. **Swapping `%w` for `%v`**: Using `%v` in `fmt.Errorf` breaks the error chain. Always use `%w` when you need the caller to unwrap the error.

3. **Returning raw errors without context**: A bare `return err` loses the operation context. Wrap it: `return fmt.Errorf("create project: %w", err)`.

4. **Logging and returning the same error**: If you log an error and also return it, the caller may log it again, creating duplicate log entries. Log at the boundary where the error is handled, not at every level.

5. **Not classifying pipeline errors**: Raw errors from LLM calls should be classified via `ClassifyError` before being emitted as `stage_failed` events, so the frontend can show appropriate retry suggestions.
