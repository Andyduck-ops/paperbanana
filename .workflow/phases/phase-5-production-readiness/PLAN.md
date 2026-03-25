# Phase 5: Production Readiness - Plan

**Created:** 2026-03-24
**Status:** Ready for execution

## Goals

1. Enable SQLite WAL mode by default
2. Add graceful shutdown handlers
3. Implement health check endpoints
4. Add observability (metrics, logging, tracing)
5. Optimize performance

## Tasks

### Task 5.1: SQLite WAL Mode

**Priority:** P1
**Files:**
- `internal/infrastructure/persistence/sqlite/connection.go` - Enable WAL mode
- `internal/config/config.go` - Update defaults

**Implementation:**
1. Enable WAL mode by default:
   - Set `PRAGMA journal_mode=WAL`
   - Set `PRAGMA synchronous=NORMAL`
   - Set `PRAGMA cache_size=-64000` (64MB)

2. Add connection pool configuration:
   - Max open connections
   - Max idle connections
   - Connection max lifetime

**Acceptance Criteria:**
- WAL mode enabled on startup
- Database handles concurrent reads/writes
- Configuration documented

### Task 5.2: Graceful Shutdown

**Priority:** P1
**Files:**
- `cmd/server/main.go` - Add shutdown handling

**Implementation:**
1. Handle SIGINT and SIGTERM:
   - Stop accepting new requests
   - Wait for existing requests to complete
   - Close database connections
   - Flush logs

2. Add shutdown timeout:
   - Default: 30 seconds
   - Configurable via environment

**Acceptance Criteria:**
- Server shuts down gracefully on signal
- In-flight requests complete
- Resources cleaned up properly

### Task 5.3: Health Check Endpoints

**Priority:** P1
**Files:**
- `internal/api/router.go` - Enhance health endpoints

**Implementation:**
1. Enhance `/health` endpoint:
   - Check database connectivity
   - Check LLM provider connectivity
   - Return detailed status

2. Enhance `/ready` endpoint:
   - Check if server is ready for traffic
   - Include dependency checks

**Acceptance Criteria:**
- Health endpoint returns detailed status
- Ready endpoint checks all dependencies
- Response time < 100ms

### Task 5.4: Observability

**Priority:** P1
**Files:**
- `internal/infrastructure/metrics/` (new) - Prometheus metrics
- `internal/api/middleware/metrics.go` (new) - Request metrics

**Implementation:**
1. Add Prometheus metrics:
   - Request count by method/path/status
   - Request duration histogram
   - Active requests gauge
   - Database query duration

2. Add structured logging:
   - Request ID in all logs
   - JSON format for production
   - Log level configuration

3. Add request tracing:
   - Trace ID generation
   - Span creation for major operations

**Acceptance Criteria:**
- Metrics endpoint available at `/metrics`
- All requests logged with trace context
- Performance impact < 5%

### Task 5.5: Performance Optimizations

**Priority:** P1
**Files:**
- `internal/infrastructure/cache/` - Add result caching
- `internal/infrastructure/persistence/sqlite/` - Query optimization

**Implementation:**
1. Add result caching:
   - Cache generated figures
   - Cache retrieval results
   - TTL-based invalidation

2. Optimize database queries:
   - Add indexes for common queries
   - Use prepared statements
   - Batch operations where possible

3. Connection pooling:
   - Pool LLM connections
   - Pool database connections

**Acceptance Criteria:**
- Cache hit rate > 80% for repeated requests
- Database queries < 10ms average
- Connection pools properly managed

## Verification

- WAL mode enabled
- Graceful shutdown works
- Health endpoints return correct status
- Metrics available
- Performance benchmarks pass

## Files Changed

| File | Change |
|------|--------|
| `internal/infrastructure/persistence/sqlite/connection.go` | Modify - WAL mode |
| `internal/config/config.go` | Modify - Connection pool config |
| `cmd/server/main.go` | Modify - Graceful shutdown |
| `internal/api/router.go` | Modify - Enhanced health endpoints |
| `internal/infrastructure/metrics/prometheus.go` | New - Prometheus metrics |
| `internal/api/middleware/metrics.go` | New - Request metrics middleware |
