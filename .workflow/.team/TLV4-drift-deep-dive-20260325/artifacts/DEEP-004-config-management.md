# Configuration Management Deep Dive

**Task ID**: DEEP-004
**Analyst**: analyst
**Date**: 2026-03-25

## Executive Summary

The configuration management system uses Viper with a hybrid approach: YAML config file + environment variable overrides. The `.env.example` is **incomplete** - missing 10+ documented environment variables that are used in production code.

**Risk Level**: MEDIUM
- Missing documentation for critical paths (config file, encryption key file)
- No sample `config.yaml` provided in repository
- Redis password not documented in `.env.example`
- Security configuration completely absent from `.env.example`

---

## Configuration Inventory

### Documented in `.env.example`

| Variable | Purpose | Default | Status |
|----------|---------|---------|--------|
| `PAPERBANANA_SERVER_HOST` | Server bind host | `0.0.0.0` (code) | Documented default differs |
| `PAPERBANANA_SERVER_PORT` | Server port | `8080` | Documented |
| `PAPERBANANA_LLM_DEFAULT` | Default LLM provider | `gemini` | Documented |
| `PAPERBANANA_BENCH_ROOT` | Benchmark data root | `data/PaperBananaBench` | Documented |
| `PAPERBANANA_ENCRYPTION_KEY` | AES-256-GCM key | (dev key fallback) | Documented |
| `PAPERBANANA_OUTPUT_DPI` | Output DPI | `300` | Documented |
| `GEMINI_API_KEY` | Gemini API key | - | Documented |
| `OPENAI_API_KEY` | OpenAI API key | - | Documented |
| `ANTHROPIC_API_KEY` | Anthropic API key | - | Documented |
| `OPENROUTER_API_KEY` | OpenRouter API key | - | Documented |

### Missing from `.env.example`

| Variable | Purpose | Source Location | Severity |
|----------|---------|-----------------|----------|
| `PAPERBANANA_CONFIG_FILE` | Explicit config file path | `config.go:201` | HIGH |
| `PAPERBANANA_NODE_CONFIG_FILE` | Custom node YAML config | `main.go:343` | MEDIUM |
| `PAPERBANANA_ENCRYPTION_KEY_FILE` | Dev key storage path | `aesgcm/service.go:118` | LOW |
| `GROK_API_KEY` | Grok/X.AI API key | `config.go:250` | MEDIUM |
| `XAI_API_KEY` | Alternative to GROK_API_KEY | `config.go:250` | MEDIUM |

### Config File Only (YAML, not env)

The following settings are configurable via `config.yaml` but have no direct environment variable mapping:

| Section | Key | Default | Notes |
|---------|-----|---------|-------|
| `cache.redis.enabled` | Redis cache toggle | `false` | No PAPERBANANA_CACHE_REDIS_ENABLED |
| `cache.redis.addr` | Redis address | `localhost:6379` | No PAPERBANANA_CACHE_REDIS_ADDR |
| `cache.redis.password` | Redis password | - | **Sensitive, not documented** |
| `cache.redis.db` | Redis DB number | `0` | No PAPERBANANA_CACHE_REDIS_DB |
| `persistence.database_path` | SQLite path | `.paperbanana/paperbanana.db` | No env override |
| `persistence.enable_foreign_keys` | FK enforcement | `true` | No env override |
| `persistence.busy_timeout_ms` | Lock timeout | `5000` | No env override |
| `persistence.enable_wal` | WAL mode | `true` | No env override |
| `assets.root` | Asset storage root | `.paperbanana/assets` | No env override |
| `assets.max_file_size` | Max upload size | `100MB` | No env override |
| `stage_timeout.*` | Pipeline timeouts | `180s-360s` | No env override |
| `security.auth_enabled` | Auth toggle | `false` | No env override |
| `security.api_keys` | Server API keys | `[]` | No env override |
| `security.rate_limit.*` | Rate limiting | defaults | No env override |
| `security.cors.*` | CORS settings | defaults | No env override |

---

## Configuration Loading Flow

```
1. Viper initialization (config.go:120-126)
   ├── SetEnvPrefix("PAPERBANANA")
   ├── SetEnvKeyReplacer(".", "_")
   └── AutomaticEnv()

2. Load config file (config.go:128-130)
   ├── PAPERBANANA_CONFIG_FILE (explicit)
   └── Auto-discover: configs/config.yaml

3. Environment expansion (config.go:190)
   └── os.ExpandEnv() on YAML content

4. Unmarshal to struct (config.go:133)

5. Provider API key fallback (config.go:240-263)
   └── Apply env vars if config file key empty

6. Validation (config.go:275-327)
```

### Environment Variable Naming Convention

Viper maps nested config keys to env vars by:
1. Prefix: `PAPERBANANA_`
2. Replacer: `.` -> `_`
3. Example: `cache.redis.addr` -> `PAPERBANANA_CACHE_REDIS_ADDR`

**However**, not all config keys have been tested with environment override. The test suite only covers:
- `PAPERBANANA_SERVER_PORT` (config_test.go:30)
- `PAPERBANANA_CONFIG_FILE` (config_test.go:57)

---

## Sensitive Information Handling

### API Keys (LLM Providers)

| Aspect | Assessment |
|--------|------------|
| Storage | Encrypted in SQLite (`api_keys` table) via AES-256-GCM |
| Env fallback | Raw env vars used as fallback (config.go:246-250) |
| Masking | `MaskAPIKey()` in validation.go:128-133 |
| Validation | Injection pattern detection (validation.go:52-56) |
| Runtime | `RuntimeClient` fetches from encrypted store, not env |

**Risk**: Env vars are read at startup for initial sync but not persisted if stored in DB.

### Encryption Key (`PAPERBANANA_ENCRYPTION_KEY`)

| Aspect | Assessment |
|--------|------------|
| Required | No - falls back to dev key |
| Dev fallback | Generates random key, persists to `.paperbanana/dev-encryption.key` |
| Production | Should be required but not enforced |

**Issue**: Production deployments may accidentally use dev key.

### Redis Password

| Aspect | Assessment |
|--------|------------|
| Storage | Config file or env (via Viper mapping) |
| Documentation | **NOT in `.env.example`** |
| Default | Empty string |

**Risk**: Redis auth not documented, deployments may use unauthenticated Redis.

---

## Missing Configuration Items

### 1. Sample `config.yaml`

**Problem**: No `config.yaml` or `config.yaml.example` in repository.

**Impact**: Users must reconstruct from:
- Test fixtures (config_test.go)
- Go struct definitions (config.go)
- README hints

**Recommendation**: Add `configs/config.yaml.example` with all documented options.

### 2. Security Configuration

**Problem**: `SecurityConfig` struct defined but no env var or sample config documentation.

```go
type SecurityConfig struct {
    AuthEnabled  bool           `mapstructure:"auth_enabled"`
    APIKeys      []string       `mapstructure:"api_keys"`
    RateLimit    RateLimitConfig `mapstructure:"rate_limit"`
    CORS         CORSConfig     `mapstructure:"cors"`
}
```

**Impact**: Production security settings are undocumented.

### 3. Stage Timeouts

**Problem**: Pipeline stage timeouts configurable but not documented.

```go
type StageTimeoutConfig struct {
    Retriever  time.Duration `mapstructure:"retriever"`
    Planner    time.Duration `mapstructure:"planner"`
    Stylist    time.Duration `mapstructure:"stylist"`
    Visualizer time.Duration `mapstructure:"visualizer"`
    Critic     time.Duration `mapstructure:"critic"`
}
```

**Impact**: Cannot tune timeout without reading source code.

### 4. Provider Models

**Problem**: `model` field per-provider not documented in `.env.example`.

**Impact**: Users don't know they can override model IDs.

---

## Default Value Assessment

### Reasonable Defaults

| Setting | Default | Assessment |
|---------|---------|------------|
| Server host | `0.0.0.0` | Good for container, may need `localhost` for dev |
| Server port | `8080` | Standard |
| SQLite path | `.paperbanana/paperbanana.db` | Appropriate |
| WAL mode | `true` | Good for concurrency |
| FK enforcement | `true` | Correct |
| Asset max size | `100MB` | Reasonable |
| Stage timeouts | `180s-360s` | Conservative, good |
| Rate limit | `60/min, burst 10` | Reasonable |

### Potentially Problematic Defaults

| Setting | Default | Issue |
|---------|---------|-------|
| Redis enabled | `false` | Good, but redis in docker-compose suggests intent to use |
| Auth enabled | `false` | Production should require auth |
| CORS origins | `["*"]` | Permissive for production |

---

## Viper AutomaticEnv Coverage

Viper's `AutomaticEnv()` only binds environment variables that have a corresponding config key in the loaded config file. This means:

1. **Empty config file**: Only defaults applied, no env binding for keys without defaults
2. **Full config file**: All env overrides work

**Test Coverage Gap**: No tests verify `PAPERBANANA_CACHE_REDIS_*` env vars work.

---

## Recommendations

### High Priority

1. **Add `configs/config.yaml.example`**
   - Include all configurable sections
   - Document sensitive fields (API keys, passwords)
   - Add comments explaining each option

2. **Update `.env.example`**
   ```bash
   # Config file path (optional)
   PAPERBANANA_CONFIG_FILE=configs/config.yaml

   # Custom node configuration (optional)
   PAPERBANANA_NODE_CONFIG_FILE=configs/custom_nodes.yaml

   # Encryption key file (fallback if ENCRYPTION_KEY not set)
   PAPERBANANA_ENCRYPTION_KEY_FILE=.paperbanana/dev-encryption.key

   # Additional LLM providers
   GROK_API_KEY=your_grok_key_here
   XAI_API_KEY=your_xai_key_here  # Alternative to GROK_API_KEY

   # Redis cache (optional)
   PAPERBANANA_CACHE_REDIS_ENABLED=false
   PAPERBANANA_CACHE_REDIS_ADDR=localhost:6379
   PAPERBANANA_CACHE_REDIS_PASSWORD=
   PAPERBANANA_CACHE_REDIS_DB=0
   ```

3. **Document security configuration**
   - Add to `config.yaml.example`
   - Document production requirements

### Medium Priority

4. **Add env var tests for Redis and Security**
   - Verify `PAPERBANANA_CACHE_REDIS_*` bindings
   - Verify `PAPERBANANA_SECURITY_*` bindings

5. **Production mode enforcement**
   - Consider `PAPERBANANA_ENV=production` flag
   - Require `PAPERBANANA_ENCRYPTION_KEY` in production
   - Require `security.auth_enabled=true` in production

### Low Priority

6. **Document stage timeout tuning**
   - Add to `config.yaml.example`
   - Explain impact of each timeout

---

## Configuration Schema Summary

```
PAPERBANANA_SERVER_HOST
PAPERBANANA_SERVER_PORT
PAPERBANANA_LLM_DEFAULT
PAPERBANANA_LLM_PROVIDERS_<NAME>_API_KEY     (via config file)
PAPERBANANA_LLM_PROVIDERS_<NAME>_BASE_URL    (via config file)
PAPERBANANA_LLM_PROVIDERS_<NAME>_MODEL       (via config file)
PAPERBANANA_LLM_PROVIDERS_<NAME>_TIMEOUT     (via config file)
PAPERBANANA_CACHE_REDIS_ENABLED
PAPERBANANA_CACHE_REDIS_ADDR
PAPERBANANA_CACHE_REDIS_PASSWORD
PAPERBANANA_CACHE_REDIS_DB
PAPERBANANA_OUTPUT_DPI
PAPERBANANA_OUTPUT_FORMATS
PAPERBANANA_PERSISTENCE_DATABASE_PATH
PAPERBANANA_PERSISTENCE_ENABLE_FOREIGN_KEYS
PAPERBANANA_PERSISTENCE_BUSY_TIMEOUT_MS
PAPERBANANA_PERSISTENCE_ENABLE_WAL
PAPERBANANA_ASSETS_ROOT
PAPERBANANA_ASSETS_MAX_FILE_SIZE
PAPERBANANA_STAGE_TIMEOUT_RETRIEVER
PAPERBANANA_STAGE_TIMEOUT_PLANNER
PAPERBANANA_STAGE_TIMEOUT_STYLIST
PAPERBANANA_STAGE_TIMEOUT_VISUALIZER
PAPERBANANA_STAGE_TIMEOUT_CRITIC
PAPERBANANA_SECURITY_AUTH_ENABLED
PAPERBANANA_SECURITY_API_KEYS
PAPERBANANA_SECURITY_RATE_LIMIT_ENABLED
PAPERBANANA_SECURITY_RATE_LIMIT_REQUESTS_PER_MINUTE
PAPERBANANA_SECURITY_RATE_LIMIT_BURST
PAPERBANANA_SECURITY_CORS_ALLOWED_ORIGINS
PAPERBANANA_SECURITY_CORS_ALLOW_CREDENTIALS
PAPERBANANA_CONFIG_FILE
PAPERBANANA_NODE_CONFIG_FILE
PAPERBANANA_BENCH_ROOT
PAPERBANANA_ENCRYPTION_KEY
PAPERBANANA_ENCRYPTION_KEY_FILE

# External API keys (fallback if not in config file)
GEMINI_API_KEY
OPENAI_API_KEY
ANTHROPIC_API_KEY
OPENROUTER_API_KEY
GROK_API_KEY
XAI_API_KEY
```

**Total**: 37+ configurable environment variables
**Documented in `.env.example`**: 10

---

## Files Analyzed

- `D:\git有趣\学习项目合集\绘图\paperbanana-clean\.env.example`
- `D:\git有趣\学习项目合集\绘图\paperbanana-clean\internal\config\config.go`
- `D:\git有趣\学习项目合集\绘图\paperbanana-clean\internal\config\config_test.go`
- `D:\git有趣\学习项目合集\绘图\paperbanana-clean\internal\config\validation.go`
- `D:\git有趣\学习项目合集\绘图\paperbanana-clean\internal\config\nodes.go`
- `D:\git有趣\学习项目合集\绘图\paperbanana-clean\internal\infrastructure\crypto\aesgcm\service.go`
- `D:\git有趣\学习项目合集\绘图\paperbanana-clean\cmd\server\main.go`
- `D:\git有趣\学习项目合集\绘图\paperbanana-clean\docker-compose.yml`
- `D:\git有趣\学习项目合集\绘图\paperbanana-clean\README.md`
