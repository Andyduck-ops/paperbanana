# Docker Dependency Deep Audit

**Session**: TLV4-drift-deep-dive-20260325
**Task**: DEEP-001
**Date**: 2025-03-25
**Analyst**: Claude (Analyst Role)

---

## Executive Summary

**Verdict**: Docker is **NOT essential** for this project. The architecture is already local-first friendly, with Redis being completely optional and all core services capable of running natively.

**Key Findings**:
- Redis: **Optional** - disabled by default in config, only used for LLM response caching
- Backend: **Pure Go binary** - no Docker-specific dependencies
- Frontend: **Static build** - can be served by any HTTP server or the Go backend itself
- Nginx: **Replaceable** - Go server can serve static files directly

---

## 1. Docker Compose Analysis

### Current Services (docker-compose.yml)

| Service | Image | Purpose | Essential? |
|---------|-------|---------|------------|
| `server` | Custom build | Go backend (API) | **No** - run `go run ./cmd/server` |
| `web` | `nginx:alpine` | Frontend static file server | **No** - Vite dev server or Go static handler |
| `redis` | `redis:7-alpine` | LLM response cache | **No** - disabled by default |

### Service Dependencies

```
Browser --> Nginx (port 3000) --> Backend (port 8080) --> SQLite
                                        |
                                        +--> Redis (optional, port 6379)
```

**Critical Observation**: No service-to-service hard dependencies. Each service can run independently.

---

## 2. Redis Necessity Analysis

### Code Evidence

**Location**: `cmd/server/main.go:58-65`
```go
if cfg.Cache.Redis.Enabled {
    redisClient := goredis.NewClient(&goredis.Options{
        Addr:     cfg.Cache.Redis.Addr,
        Password: cfg.Cache.Redis.Password,
        DB:       cfg.Cache.Redis.DB,
    })
    options.Cache = rediscache.NewCache(rediscache.NewStore(redisClient))
}
```

### Configuration Default

**Location**: `internal/config/config.go:150`
```go
v.SetDefault("cache.redis.enabled", false)  // <-- DISABLED BY DEFAULT
```

### Redis Usage Summary

| Aspect | Finding |
|--------|---------|
| **Purpose** | LLM response caching (SHA256-based key, 24h TTL) |
| **Default** | **Disabled** |
| **Impact if absent** | No caching - every LLM call hits API (cost + latency) |
| **Critical?** | No - application functions normally without it |
| **Production need** | Optional optimization for high-traffic scenarios |

### When Redis Might Be Useful

1. **High request volume**: Caches duplicate LLM queries
2. **Multi-instance deployment**: Shared cache across backend replicas
3. **Cost optimization**: Reduces API calls to LLM providers

### When Redis Is Unnecessary

1. **Single-user development**: Low request volume
2. **Local development**: Latency is acceptable
3. **Testing**: Mocks already used in tests

---

## 3. Local-Run Feasibility

### Backend (Go Server)

**Command**:
```bash
cd D:\git有趣\学习项目合集\绘图\paperbanana-clean
go run ./cmd/server --config configs/config.yaml
```

**Requirements**:
- Go 1.23+
- SQLite (built into the binary, no external dependency)
- LLM API key (environment variable or config)

**Ports**:
- 8080 (API + can serve frontend)

### Frontend (React/Vite)

**Development Mode**:
```bash
cd web
npm run dev    # Runs on http://localhost:5173
```

**Production Build**:
```bash
npm run build  # Outputs to web/dist/
```

**Serving Options**:
1. **Vite preview**: `npm run preview` (port 4173)
2. **Go backend**: Add static file handler in Go
3. **Any HTTP server**: Python, nginx, Apache, etc.

---

## 4. Docker Complexity vs. Benefits

### Complexity Introduced by Docker

| Aspect | Docker Approach | Local Approach |
|--------|-----------------|----------------|
| **Prerequisites** | Docker Desktop (~2GB) | Go 1.23 (~200MB) + Node.js 18 (~100MB) |
| **Build time** | Image build (~2-5 min first time) | `go run` (~5-10s) |
| **Hot reload** | Volume mounts, restart container | Native file watching |
| **Debugging** | Container logs, exec into container | Direct console output, IDE debugging |
| **Config changes** | Rebuild or mount files | Edit and restart |
| **Port conflicts** | Map ports in compose | Change config port |
| **Resource usage** | VM overhead + 3 containers | Native process (~50-100MB RAM) |

### Benefits Provided by Docker

| Benefit | Actual Value | Alternative |
|---------|--------------|-------------|
| **Consistent environment** | High | `go.mod` + `package-lock.json` already ensure this |
| **Isolation** | Medium | Not needed for single-developer project |
| **One-command startup** | `docker-compose up` | `go run ./cmd/server` (equally simple) |
| **Redis included** | Convenience | Not needed (disabled by default) |

### Net Assessment

**Docker adds complexity with minimal benefit** for this project because:

1. **Go is already portable**: Static binary, no runtime dependencies
2. **SQLite is embedded**: No database server to manage
3. **Redis is optional**: Not required for core functionality
4. **Frontend is static**: Build once, serve anywhere

---

## 5. Recommended Removal Strategy

### Phase 1: Immediate Removal (Low Risk)

**Delete Files**:
```
docker-compose.yml
Dockerfile
nginx.conf
```

**Update DEPLOYMENT.md**:
Remove Docker references, emphasize local development.

### Phase 2: Simplify Frontend Serving

**Option A: Serve from Go Backend** (Recommended)

Add a static file server in Go:

```go
// In cmd/server/main.go or internal/api/router.go
router.Static("/assets", "./web/dist/assets")
router.NoRoute(func(c *gin.Context) {
    c.File("./web/dist/index.html")
})
```

**Option B: Separate Vite Dev Server** (Development Only)

Run frontend and backend separately:
```bash
# Terminal 1: Backend
go run ./cmd/server

# Terminal 2: Frontend
cd web && npm run dev
```

**Option C: Vite Preview** (Production Simulation)

```bash
cd web && npm run build && npm run preview
```

### Phase 3: Redis Decision

**Keep Redis Dependency** (Code Cleanup):
- Remove from `docker-compose.yml`
- Keep code path for optional Redis usage
- Document when to enable Redis

**Remove Redis Completely**:
- Delete `internal/infrastructure/cache/redis/` directory
- Remove `github.com/redis/go-redis/v9` from `go.mod`
- Remove Redis config from `internal/config/config.go`

**Recommendation**: Keep Redis code but remove from Docker. Users can install Redis locally if needed.

---

## 6. Migration Checklist

### For Developers

- [ ] Install Go 1.23+
- [ ] Install Node.js 18+
- [ ] Set `GEMINI_API_KEY` environment variable (or other LLM provider)
- [ ] Run `go mod download`
- [ ] Run `cd web && npm install`

### For Running Locally

**Backend only**:
```bash
go run ./cmd/server
# API at http://localhost:8080
```

**Full stack (development)**:
```bash
# Terminal 1
go run ./cmd/server

# Terminal 2
cd web && npm run dev
# Frontend at http://localhost:5173
```

**Full stack (production-like)**:
```bash
# Build frontend
cd web && npm run build

# Run backend (configure to serve ./web/dist)
go run ./cmd/server
```

### For CI/CD

Update pipelines to:
1. Use `go test ./...` instead of Docker-based tests
2. Use `npm run build` for frontend
3. Remove Docker build/push steps

---

## 7. Risk Assessment

| Risk | Severity | Mitigation |
|------|----------|------------|
| Different behavior without Docker | Low | Go binaries are deterministic; SQLite is consistent |
| Redis needed later | Low | Keep code, install locally if needed |
| Nginx features lost | Low | SSE already handled by Go middleware |
| Team member environment drift | Medium | Document setup clearly, use `go.mod` + `package-lock.json` |

---

## 8. Conclusion

**Docker is overkill for this project.** The architecture is inherently local-first:

- **SQLite** replaces traditional database servers
- **Go** produces static binaries with no runtime dependencies
- **Redis** is explicitly optional (disabled by default)
- **Nginx** is only a static file server (easily replaced)

**Recommendation**: Remove Docker. Update documentation to guide developers toward native execution. This simplifies the development experience without losing any functionality.

---

## Appendix A: File Locations

| File | Purpose | Action |
|------|---------|--------|
| `docker-compose.yml` | Container orchestration | **Delete** |
| `Dockerfile` | Backend container image | **Delete** |
| `nginx.conf` | Frontend proxy config | **Delete** |
| `DEPLOYMENT.md` | Deployment guide | **Update** |
| `internal/infrastructure/cache/redis/` | Redis cache impl | **Keep** (optional feature) |

## Appendix B: Current Architecture vs. Proposed

```
CURRENT (Docker):
Browser --> Nginx container --> Go container --> SQLite
                                   |
                                   +--> Redis container (unused)

PROPOSED (Native):
Browser --> Go process --> SQLite
                |
                +--> Redis (optional, manual install)
```

---

**End of Report**
