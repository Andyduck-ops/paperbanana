package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server      ServerConfig      `mapstructure:"server"`
	LLM         LLMConfig         `mapstructure:"llm"`
	Cache       CacheConfig       `mapstructure:"cache"`
	Output      OutputConfig      `mapstructure:"output"`
	Persistence PersistenceConfig `mapstructure:"persistence"`
	Assets      AssetsConfig      `mapstructure:"assets"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type LLMConfig struct {
	Providers map[string]ProviderConfig `mapstructure:"providers"`
	Default   string                    `mapstructure:"default"`
}

type ProviderConfig struct {
	APIKey  string        `mapstructure:"api_key"`
	BaseURL string        `mapstructure:"base_url"`
	Model   string        `mapstructure:"model"`
	Timeout time.Duration `mapstructure:"timeout"`
}

type CacheConfig struct {
	Redis RedisConfig `mapstructure:"redis"`
}

type RedisConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type OutputConfig struct {
	DPI     int      `mapstructure:"dpi"`
	Formats []string `mapstructure:"formats"`
}

// PersistenceConfig defines database connection settings.
type PersistenceConfig struct {
	// Database path for SQLite. Defaults to .paperbanana/paperbanana.db
	DatabasePath string `mapstructure:"database_path"`

	// Enable foreign key enforcement (recommended: true)
	EnableForeignKeys bool `mapstructure:"enable_foreign_keys"`

	// Busy timeout in milliseconds for SQLite locks
	BusyTimeoutMs int `mapstructure:"busy_timeout_ms"`

	// Enable WAL mode (deferred until SQLite version verification)
	EnableWAL bool `mapstructure:"enable_wal"`
}

// AssetsConfig defines the asset storage backend.
type AssetsConfig struct {
	// Root directory for asset storage. Defaults to .paperbanana/assets
	Root string `mapstructure:"root"`

	// Max file size in bytes for uploads (default: 100MB)
	MaxFileSize int64 `mapstructure:"max_file_size"`
}

func Load() (*Config, error) {
	v := viper.New()
	setDefaults(v)

	v.SetEnvPrefix("PAPERBANANA")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := readConfigFile(v); err != nil {
		return nil, err
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	applyProviderEnvFallbacks(&cfg)

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
func setDefaults(v *viper.Viper) {
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)

	v.SetDefault("llm.default", "gemini")
	v.SetDefault("cache.redis.enabled", false)
	v.SetDefault("cache.redis.addr", "localhost:6379")
	v.SetDefault("cache.redis.db", 0)
	v.SetDefault("output.dpi", 300)
	v.SetDefault("output.formats", []string{"png", "svg", "pdf"})
	// Persistence defaults
	v.SetDefault("persistence.database_path", ".paperbanana/paperbanana.db")
	v.SetDefault("persistence.enable_foreign_keys", true)
	v.SetDefault("persistence.busy_timeout_ms", 5000)
	v.SetDefault("persistence.enable_wal", false)
	// Assets defaults
	v.SetDefault("assets.root", ".paperbanana/assets")
	v.SetDefault("assets.max_file_size", int64(100*1024*1024)) // 100MB
}

func readConfigFile(v *viper.Viper) error {
	configPath, found := findConfigFile()
	if !found {
		return nil
	}

	raw, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	expanded := os.ExpandEnv(string(raw))
	v.SetConfigFile(configPath)

	if err := v.ReadConfig(bytes.NewBufferString(expanded)); err != nil {
		return fmt.Errorf("failed to parse config file %s: %w", configPath, err)
	}

	return nil
}

func findConfigFile() (string, bool) {
	if explicit := os.Getenv("PAPERBANANA_CONFIG_FILE"); explicit != "" {
		if fileExists(explicit) {
			return explicit, true
		}
		return "", false
	}

	candidates := []string{
		"configs/config.yaml",
		"./configs/config.yaml",
		"../configs/config.yaml",
		"../../configs/config.yaml",
	}

	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates,
			filepath.Join(wd, "configs", "config.yaml"),
			filepath.Join(wd, "..", "configs", "config.yaml"),
			filepath.Join(wd, "..", "..", "configs", "config.yaml"),
		)
	}

	for _, candidate := range candidates {
		if fileExists(candidate) {
			return candidate, true
		}
	}

	return "", false
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func applyProviderEnvFallbacks(cfg *Config) {
	if cfg.LLM.Providers == nil {
		cfg.LLM.Providers = map[string]ProviderConfig{}
	}

	fallbacks := map[string]string{
		"gemini":     os.Getenv("GEMINI_API_KEY"),
		"openai":     os.Getenv("OPENAI_API_KEY"),
		"anthropic":  os.Getenv("ANTHROPIC_API_KEY"),
		"openrouter": os.Getenv("OPENROUTER_API_KEY"),
	}

	for provider, apiKey := range fallbacks {
		entry, ok := cfg.LLM.Providers[provider]
		if !ok {
			continue
		}
		if entry.APIKey == "" && apiKey != "" {
			entry.APIKey = apiKey
			cfg.LLM.Providers[provider] = entry
		}
	}
}

func validate(cfg *Config) error {
	if cfg.LLM.Default == "" {
		return errors.New("llm.default provider must be set")
	}

	if len(cfg.LLM.Providers) == 0 {
		return errors.New("llm.providers must contain at least one provider")
	}

	defaultProvider, exists := cfg.LLM.Providers[cfg.LLM.Default]
	if !exists {
		return fmt.Errorf("default provider %s not found in providers", cfg.LLM.Default)
	}

	if defaultProvider.Model == "" {
		return fmt.Errorf("default provider %s missing model", cfg.LLM.Default)
	}

	for name, provider := range cfg.LLM.Providers {
		if provider.Model == "" {
			return fmt.Errorf("provider %s missing model", name)
		}
		if provider.Timeout <= 0 {
			return fmt.Errorf("provider %s timeout must be > 0", name)
		}
	}

	if cfg.Output.DPI < 72 {
		return fmt.Errorf("output.dpi must be >= 72, got %d", cfg.Output.DPI)
	}

	if len(cfg.Output.Formats) == 0 {
		return errors.New("output.formats must contain at least one format")
	}

	// Persistence validation
	if cfg.Persistence.DatabasePath == "" {
		return errors.New("persistence.database_path must be set")
	}
	if cfg.Persistence.BusyTimeoutMs < 0 {
		return fmt.Errorf("persistence.busy_timeout_ms must be >= 0, got %d", cfg.Persistence.BusyTimeoutMs)
	}

	// Assets validation
	if cfg.Assets.Root == "" {
		return errors.New("assets.root must be set")
	}
	if cfg.Assets.MaxFileSize <= 0 {
		return fmt.Errorf("assets.max_file_size must be > 0, got %d", cfg.Assets.MaxFileSize)
	}

	return nil
}
