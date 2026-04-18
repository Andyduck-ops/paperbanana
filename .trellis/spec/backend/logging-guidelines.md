# Logging Guidelines

> Structured logging conventions for PaperBanana.

---

## Overview

PaperBanana uses **Zap** (`go.uber.org/zap`) for structured logging on the backend. The frontend generally does not log in production code. All logging uses structured fields, never interpolated strings.

---

## Logger Creation

A single `*zap.Logger` is created at server startup and passed down through dependency injection:

```go
// cmd/server/main.go
logger, err := zap.NewProduction()
if err != nil {
    log.Fatalf("failed to create logger: %v", err)
}
defer func() { _ = logger.Sync() }()
```

The production logger is used (JSON output, `Info` level minimum). There is no per-request logger or logger middleware that enriches the logger with request-scoped fields.

---

## Passing Loggers

Entry points and handlers receive `*zap.Logger` as a constructor parameter:

```go
// internal/api/handlers/generate.go
func NewHandler(runner Runner, logger *zap.Logger) *Handler {
    return &Handler{runner: runner, logger: logger}
}

// internal/api/handlers/history.go
func NewHistoryHandler(historyService HistoryService, logger *zap.Logger) *HistoryHandler {
    return &HistoryHandler{historyService: historyService, logger: logger}
}
```

The logger is stored as a struct field and used throughout the handler's methods.

---

## Structured Fields

Always use Zap's typed field constructors instead of string interpolation:

```go
// GOOD - Structured fields
logger.Info("starting server",
    zap.String("address", address),
    zap.String("provider", cfg.LLM.Default),
    zap.String("model", providerConfig.Model),
    zap.String("database", cfg.Persistence.DatabasePath),
)

// BAD - Interpolated strings
logger.Info(fmt.Sprintf("starting server at %s with provider %s", address, cfg.LLM.Default))
```

### Common Field Constructors

| Zap Constructor | Use For |
|----------------|---------|
| `zap.String(key, val)` | Provider names, paths, session IDs, error messages |
| `zap.Error(err)` | Errors (automatically uses "error" key) |
| `zap.Int(key, val)` | Status codes, counts, port numbers |
| `zap.Duration(key, val)` | Timing, timeouts |
| `zap.Bool(key, val)` | Feature flags |
| `zap.Any(key, val)` | Complex objects (use sparingly) |

### Field Naming Conventions

- Use `snake_case` for field keys: `session_id`, `project_id`, `provider`
- Be consistent: always use the same key name for the same concept across files
- Common keys: `provider`, `model`, `session_id`, `request_id`, `project_id`, `address`, `path`

---

## Request Logging Middleware

All HTTP requests are logged via the `Logger` middleware in `internal/api/middleware/logger.go`:

```go
// internal/api/middleware/logger.go
func Logger(logger *zap.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        c.Next()

        logger.Info("request",
            zap.String("method", c.Request.Method),
            zap.String("path", c.Request.URL.Path),
            zap.Int("status", c.Writer.Status()),
            zap.Duration("duration", time.Since(start)),
        )
    }
}
```

This logs every request at `Info` level with method, path, status, and duration.

---

## Log Levels

| Level | When to Use | Examples |
|-------|-------------|----------|
| **Debug** | Verbose diagnostic info, currently not used in production config | Variable values, internal state dumps |
| **Info** | Normal operational events | Server start, request completion, config loaded, provider synced |
| **Warn** | Unexpected but recoverable situations | Failed to sync startup providers, failed to persist artifacts, using dev encryption key |
| **Error** | Operation failures that need attention | Generation failed, database bootstrap failed, server shutdown error |
| **Fatal** | Unrecoverable startup errors | Config load failed, database bootstrap failed, missing default provider |

### Examples from the Codebase

```go
// Info - Normal operation
logger.Info("starting server",
    zap.String("address", address),
    zap.String("provider", cfg.LLM.Default),
)

// Warn - Degraded but functional
logger.Warn("failed to initialize system providers", zap.Error(err))
logger.Warn("failed to sync startup providers into config store", zap.Error(err))

// Error - Operation failure
h.logger.Error("generation failed", zap.Error(err))
logger.Error("server forced to shutdown", zap.Error(err))

// Fatal - Cannot start
logger.Fatal("failed to load config", zap.Error(err))
logger.Fatal("failed to bootstrap persistence", zap.Error(err))
```

---

## What to Log

- Server lifecycle events (start, shutdown, config reload)
- Request outcomes (via middleware)
- Pipeline stage transitions (via orchestrator events)
- Operation failures with context
- Configuration warnings (missing API keys, fallback behavior)
- Asset persistence outcomes

---

## What NOT to Log

- **API keys and secrets**: Never log raw API keys. Use `aesgcm.MaskAPIKey()` or `hashKeyID()` for masked identifiers
- **Full request bodies**: The generate request can contain large content; log only metadata
- **Base64-encoded images**: The refine endpoint receives images; never log image data
- **Session snapshot JSON**: Session snapshots can be very large; log only session ID and status
- **User prompt content**: Log the existence of a prompt, not its content

---

## Known Issues

1. **No request-scoped logging**: The logger is shared across all requests. There is no request ID in log entries (the middleware logs path and status but not a request-scoped correlation ID).

2. **Inconsistent error logging**: Some handlers log errors before returning, while others rely on the caller to log. This can lead to duplicate error log entries.

3. **Frontend logging**: The React frontend generally does not log in production. Development console logs are not standardized.
