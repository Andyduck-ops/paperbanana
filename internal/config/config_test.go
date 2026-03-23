package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadYAML(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "gemini-test-key")
	t.Setenv("OPENAI_API_KEY", "openai-test-key")
	t.Setenv("ANTHROPIC_API_KEY", "anthropic-test-key")
	t.Setenv("OPENROUTER_API_KEY", "openrouter-test-key")

	cfg, err := Load()
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "gemini", cfg.LLM.Default)
}

func TestEnvOverride(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "gemini-test-key")
	t.Setenv("OPENAI_API_KEY", "openai-test-key")
	t.Setenv("ANTHROPIC_API_KEY", "anthropic-test-key")
	t.Setenv("OPENROUTER_API_KEY", "openrouter-test-key")
	t.Setenv("PAPERBANANA_SERVER_PORT", "9090")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, 9090, cfg.Server.Port)
}

func TestAPIKeyValidation(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	content := []byte(`
server:
  host: localhost
  port: 8080
llm:
  default: gemini
  providers:
    gemini:
      api_key: ""
      base_url: https://generativelanguage.googleapis.com
      model: gemini-2.0-flash-exp
      timeout: 60s
output:
  dpi: 300
  formats: [png]
`)
	require.NoError(t, os.WriteFile(configPath, content, 0o644))

	t.Setenv("PAPERBANANA_CONFIG_FILE", configPath)

	cfg, err := Load()
	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.ErrorContains(t, err, "provider gemini missing api_key")
}

func TestDefaultModelSelection(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "gemini-test-key")
	t.Setenv("OPENAI_API_KEY", "openai-test-key")
	t.Setenv("ANTHROPIC_API_KEY", "anthropic-test-key")
	t.Setenv("OPENROUTER_API_KEY", "openrouter-test-key")

	cfg, err := Load()
	require.NoError(t, err)
	assert.NotEmpty(t, cfg.LLM.Default)
	assert.Contains(t, cfg.LLM.Providers, cfg.LLM.Default)
	assert.NotEmpty(t, cfg.LLM.Providers[cfg.LLM.Default].Model)
}

func TestOutputParams(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "gemini-test-key")
	t.Setenv("OPENAI_API_KEY", "openai-test-key")
	t.Setenv("ANTHROPIC_API_KEY", "anthropic-test-key")
	t.Setenv("OPENROUTER_API_KEY", "openrouter-test-key")

	cfg, err := Load()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, cfg.Output.DPI, 72)
	assert.NotEmpty(t, cfg.Output.Formats)
}

func TestConfigLoadsRedisCacheSettings(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	content := []byte(`
server:
  host: localhost
  port: 8080
llm:
  default: gemini
  providers:
    gemini:
      api_key: test-key
      base_url: https://generativelanguage.googleapis.com
      model: gemini-2.0-flash-exp
      timeout: 60s
cache:
  redis:
    enabled: true
    addr: localhost:6379
    password: secret
    db: 5
output:
  dpi: 300
  formats: [png]
`)
	require.NoError(t, os.WriteFile(configPath, content, 0o644))

	t.Setenv("PAPERBANANA_CONFIG_FILE", configPath)

	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.True(t, cfg.Cache.Redis.Enabled)
	assert.Equal(t, "localhost:6379", cfg.Cache.Redis.Addr)
	assert.Equal(t, "secret", cfg.Cache.Redis.Password)
	assert.Equal(t, 5, cfg.Cache.Redis.DB)
}

func TestPersistenceDefaults(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	content := []byte(`
server:
  host: localhost
  port: 8080
llm:
  default: gemini
  providers:
    gemini:
      api_key: test-key
      base_url: https://generativelanguage.googleapis.com
      model: gemini-2.0-flash-exp
      timeout: 60s
output:
  dpi: 300
  formats: [png]
`)
	require.NoError(t, os.WriteFile(configPath, content, 0o644))

	t.Setenv("PAPERBANANA_CONFIG_FILE", configPath)

	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify persistence defaults
	assert.Equal(t, ".paperbanana/paperbanana.db", cfg.Persistence.DatabasePath)
	assert.True(t, cfg.Persistence.EnableForeignKeys)
	assert.Equal(t, 5000, cfg.Persistence.BusyTimeoutMs)
	assert.False(t, cfg.Persistence.EnableWAL)
}

func TestAssetsDefaults(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	content := []byte(`
server:
  host: localhost
  port: 8080
llm:
  default: gemini
  providers:
    gemini:
      api_key: test-key
      base_url: https://generativelanguage.googleapis.com
      model: gemini-2.0-flash-exp
      timeout: 60s
output:
  dpi: 300
  formats: [png]
`)
	require.NoError(t, os.WriteFile(configPath, content, 0o644))

	t.Setenv("PAPERBANANA_CONFIG_FILE", configPath)

	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify assets defaults
	assert.Equal(t, ".paperbanana/assets", cfg.Assets.Root)
	assert.Equal(t, int64(100*1024*1024), cfg.Assets.MaxFileSize)
}

func TestPersistenceConfigValidation(t *testing.T) {
	t.Run("rejects empty database path", func(t *testing.T) {
		configPath := filepath.Join(t.TempDir(), "config.yaml")
		content := []byte(`
server:
  host: localhost
  port: 8080
llm:
  default: gemini
  providers:
    gemini:
      api_key: test-key
      model: gemini-2.0-flash-exp
      timeout: 60s
output:
  dpi: 300
  formats: [png]
persistence:
  database_path: ""
`)
		require.NoError(t, os.WriteFile(configPath, content, 0o644))
		t.Setenv("PAPERBANANA_CONFIG_FILE", configPath)

		cfg, err := Load()
		require.Error(t, err)
		assert.Nil(t, cfg)
		assert.ErrorContains(t, err, "persistence.database_path must be set")
	})

	t.Run("rejects negative busy timeout", func(t *testing.T) {
		configPath := filepath.Join(t.TempDir(), "config.yaml")
		content := []byte(`
server:
  host: localhost
  port: 8080
llm:
  default: gemini
  providers:
    gemini:
      api_key: test-key
      model: gemini-2.0-flash-exp
      timeout: 60s
output:
  dpi: 300
  formats: [png]
persistence:
  database_path: test.db
  busy_timeout_ms: -1
`)
		require.NoError(t, os.WriteFile(configPath, content, 0o644))
		t.Setenv("PAPERBANANA_CONFIG_FILE", configPath)

		cfg, err := Load()
		require.Error(t, err)
		assert.Nil(t, cfg)
		assert.ErrorContains(t, err, "persistence.busy_timeout_ms must be >= 0")
	})
}

func TestAssetsConfigValidation(t *testing.T) {
	t.Run("rejects empty assets root", func(t *testing.T) {
		configPath := filepath.Join(t.TempDir(), "config.yaml")
		content := []byte(`
server:
  host: localhost
  port: 8080
llm:
  default: gemini
  providers:
    gemini:
      api_key: test-key
      model: gemini-2.0-flash-exp
      timeout: 60s
output:
  dpi: 300
  formats: [png]
assets:
  root: ""
`)
		require.NoError(t, os.WriteFile(configPath, content, 0o644))
		t.Setenv("PAPERBANANA_CONFIG_FILE", configPath)

		cfg, err := Load()
		require.Error(t, err)
		assert.Nil(t, cfg)
		assert.ErrorContains(t, err, "assets.root must be set")
	})

	t.Run("rejects zero max file size", func(t *testing.T) {
		configPath := filepath.Join(t.TempDir(), "config.yaml")
		content := []byte(`
server:
  host: localhost
  port: 8080
llm:
  default: gemini
  providers:
    gemini:
      api_key: test-key
      model: gemini-2.0-flash-exp
      timeout: 60s
output:
  dpi: 300
  formats: [png]
assets:
  root: /tmp/assets
  max_file_size: 0
`)
		require.NoError(t, os.WriteFile(configPath, content, 0o644))
		t.Setenv("PAPERBANANA_CONFIG_FILE", configPath)

		cfg, err := Load()
		require.Error(t, err)
		assert.Nil(t, cfg)
		assert.ErrorContains(t, err, "assets.max_file_size must be > 0")
	})
}
