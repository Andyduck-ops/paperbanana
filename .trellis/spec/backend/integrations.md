# Integrations

> External service integrations and configuration for PaperBanana.

**[PRE-TAURI]** This document describes the current web-app integration surface. The integration model will change when the project migrates to Tauri + Go sidecar (HTTP becomes IPC).

---

## LLM Providers

PaperBanana supports multiple LLM providers through a provider abstraction layer. All providers implement the `domainllm.LLMClient` interface defined in `internal/domain/llm/client.go`.

### Supported Providers

| Provider | Package | Client Type | Notes |
|----------|---------|-------------|-------|
| OpenAI | `infrastructure/llm/openai/` | Direct API | GPT-4o, etc. |
| Google Gemini | `infrastructure/llm/gemini/` | Google AI SDK | Gemini 2.0 Flash, etc. |
| Anthropic | `infrastructure/llm/anthropic/` | Direct API | Claude Sonnet, etc. |
| OpenRouter | `infrastructure/llm/openrouter/` | Direct API | Multi-provider routing |
| OpenAI-Compatible | `infrastructure/llm/openai/` (reused) | Provider alias client | DeepSeek, Zhipu, Moonshot, Qwen, Doubao, Baichuan, Minimax, Yi, Hunyuan, Stepfun, Silicon, Ollama, Grok/XAI |

### Provider Preset Catalog

System-defined provider presets live in `internal/domain/config/provider.go`. Each preset specifies:

```go
type SystemProviderPreset struct {
    Type           ProviderType `json:"type"`
    Name           string       `json:"name"`
    DisplayName    string       `json:"display_name"`
    APIHost        string       `json:"api_host"`
    DocsURL        string       `json:"docs_url"`
    APIKeyURL      string       `api_key_url"`
    DefaultModels  []ModelInfo  `json:"default_models"`
    SupportsVision bool         `json:"supports_vision"`
}
```

Presets are initialized into the database on startup via `providerRepo.InitializeSystemProviders()`.

### Client Creation

The factory (`internal/infrastructure/llm/factory.go`) selects and creates clients:

```go
func newRawLLMClient(provider string, cfg pbconfig.ProviderConfig, httpClient *http.Client) (domainllm.LLMClient, error) {
    switch provider {
    case "gemini":
        return gemini.NewClientWithHTTPClient(cfg.APIKey, cfg.Model, httpClient)
    case "openai":
        return openaiclient.NewClientWithConfig(cfg.APIKey, cfg.BaseURL, cfg.Model, cfg.Timeout, httpClient)
    case "anthropic":
        return anthropic.NewClientWithConfig(cfg.APIKey, cfg.BaseURL, cfg.Model, cfg.Timeout, httpClient)
    case "openrouter":
        return openrouter.NewClientWithConfig(cfg.APIKey, cfg.BaseURL, cfg.Model, cfg.Timeout, httpClient)
    default:
        // OpenAI-compatible providers reuse the OpenAI client
        if openAICompatibleProviders[provider] || cfg.BaseURL != "" {
            client, _ := openaiclient.NewClientWithConfig(...)
            return newProviderAliasClient(provider, client), nil
        }
        return nil, fmt.Errorf("unknown provider: %s", provider)
    }
}
```

### RuntimeClient

The `RuntimeClient` in `internal/infrastructure/llm/client_manager.go` wraps LLM clients with:
- Provider resolution from the database (not just config)
- API key decryption before use
- Purpose separation (query vs. generation)

Two instances are created at startup:
- `queryClient` -- Used for Retriever, Planner, Stylist, Critic
- `genClient` -- Used for Visualizer (code generation)

### Custom Nodes

Custom node definitions in `configs/custom_nodes.yaml` allow extending the Visualizer with external processing endpoints:

```go
// cmd/server/main.go
nodeCatalog, err := loadNodeCatalog(logger)
```

Node catalog is loaded from `PAPERBANANA_NODE_CONFIG_FILE` env var or `configs/custom_nodes.yaml`.

---

## Data Storage

### SQLite + GORM

Primary data store. See [Database Guidelines](./database-guidelines.md) for full details.

- **Database file**: Default `.paperbanana/paperbanana.db`, configurable via `persistence.database_path`
- **ORM**: GORM with `github.com/glebarez/sqlite` (pure Go driver)
- **Migration**: AutoMigrate on startup

### Local Filesystem Assets

Asset bytes are stored on the local filesystem under a configurable root directory:

- **Default root**: `.paperbanana/assets`
- **Configurable via**: `assets.root`
- **Max upload size**: 100MB (configurable via `assets.max_file_size`)

Two implementations exist (see Database Guidelines for details).

### Optional Redis Cache

LLM response caching via Redis is optional. When enabled, the `CachedClient` wraps the raw LLM client:

```go
// cmd/server/main.go
if cfg.Cache.Redis.Enabled {
    redisClient := goredis.NewClient(&goredis.Options{
        Addr:     cfg.Cache.Redis.Addr,
        Password: cfg.Cache.Redis.Password,
        DB:       cfg.Cache.Redis.DB,
    })
    options.Cache = rediscache.NewCache(rediscache.NewStore(redisClient))
}
```

Configuration via `cache.redis.*` settings:
- `cache.redis.enabled` (default: `false`)
- `cache.redis.addr` (default: `localhost:6379`)
- `cache.redis.password`
- `cache.redis.db` (default: `0`)

---

## Authentication

### Static API Key Auth

`internal/api/middleware/auth.go` implements static API key authentication using constant-time comparison:

- Header: `X-API-Key` (configurable)
- Multiple keys supported
- Constant-time comparison via `crypto/subtle`
- `Auth` middleware (strict) and `OptionalAuth` middleware (permissive)
- `BearerAuth` middleware for `Authorization: Bearer <token>` format

**Current status**: Auth middleware is **not mounted** in any router. All endpoints are open.

### Encrypted Provider Credentials

API keys stored in the database are encrypted at rest using AES-256-GCM with Argon2id key derivation:

```go
// internal/infrastructure/crypto/aesgcm/service.go
type Service struct {
    gcm cipher.AEAD
    kdf *keyderivation.Argon2idKDF
}
```

- Encryption key source: `PAPERBANANA_ENCRYPTION_KEY` env var
- Fallback: Auto-generated dev key in `.paperbanana/dev-encryption.key`
- Dev key warning: Printed to stderr on startup if using fallback

---

## Monitoring

### Zap Logging

See [Logging Guidelines](./logging-guidelines.md). Production logger with JSON output.

### Prometheus Metrics

Prometheus client is included (`github.com/prometheus/client_golang`) and a metrics endpoint is registered:

```go
// internal/api/router.go
router.GET("/metrics", gin.WrapH(promhttp.Handler()))
```

A `Metrics` middleware (`internal/api/middleware/metrics.go`) records request counts and durations.

**Note**: The metrics endpoint is registered but the middleware is only mounted in the `SetupRouterWithPersistenceWithRegistryAndDB` variant. The most commonly used `SetupRouterWithConfigAndBatchAndDB` does mount it.

### Health Checks

Two health endpoints are provided:

| Endpoint | Purpose | Response |
|----------|---------|----------|
| `GET /health` | Liveness | `{"status": "ok"}` or `{"status": "degraded"}` with database status |
| `GET /ready` | Readiness | `{"status": "ready"}` or `{"status": "not_ready"}` with database status |

Both endpoints check database connectivity when a `*gorm.DB` is provided.

---

## Environment Configuration

### Viper with Environment Expansion

Configuration is loaded via Viper with environment variable expansion:

```go
// internal/config/config.go
func Load() (*Config, error) {
    v := viper.New()
    v.SetEnvPrefix("PAPERBANANA")
    v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
    v.AutomaticEnv()
    // ...
}
```

### Environment Variables

| Variable | Purpose | Default |
|----------|---------|---------|
| `PAPERBANANA_CONFIG_FILE` | Explicit config file path | Auto-discovered |
| `PAPERBANANA_BENCH_ROOT` | Benchmark data root | `data/PaperBananaBench` |
| `PAPERBANANA_NODE_CONFIG_FILE` | Custom node config path | `configs/custom_nodes.yaml` |
| `PAPERBANANA_ENCRYPTION_KEY` | Encryption key for API keys | Dev key auto-generated |
| `PAPERBANANA_ENCRYPTION_KEY_FILE` | Dev key file path | `.paperbanana/dev-encryption.key` |
| `GEMINI_API_KEY` | Gemini API key | -- |
| `OPENAI_API_KEY` | OpenAI API key | -- |
| `ANTHROPIC_API_KEY` | Anthropic API key | -- |
| `OPENROUTER_API_KEY` | OpenRouter API key | -- |
| `GROK_API_KEY` / `XAI_API_KEY` | Grok/XAI API key | -- |

### Config File

YAML config file with env expansion support:

```yaml
# configs/config.example.yaml
server:
  host: "0.0.0.0"
  port: 8080

llm:
  default: "gemini"
  providers:
    gemini:
      api_key: ${GEMINI_API_KEY}
      model: "gemini-2.0-flash-exp"
```

The `${VAR}` syntax in YAML files is expanded via `os.ExpandEnv` before parsing.

### Provider Environment Fallbacks

If a provider is not in the config file but its API key is available as an environment variable, it is automatically created with default settings via `applyProviderEnvFallbacks()`.
