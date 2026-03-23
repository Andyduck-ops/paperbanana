package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/paperbanana/paperbanana/internal/config"
	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
	llminfra "github.com/paperbanana/paperbanana/internal/infrastructure/llm"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestServerStartupFailsOnInvalidNodeConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "custom_nodes.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(`
custom_nodes:
  - name: invalid_node
    url: https://api.example.com/render
    method: POST
    request_template:
      prompt: render this chart
    response_parser: json_path
`), 0o644))

	t.Setenv("PAPERBANANA_NODE_CONFIG_FILE", configPath)

	_, err := loadNodeCatalog(zap.NewNop())
	require.Error(t, err)
	require.ErrorContains(t, err, "response_selectors.image_base64 must be set")
}

func TestBuildStartupLLMClientFallsBackToUnavailableClient(t *testing.T) {
	client, err := buildStartupLLMClient(zap.NewNop(), "gemini", config.ProviderConfig{
		Model:   "gemini-2.0-flash-exp",
		Timeout: 0,
	}, llminfra.ClientOptions{})
	require.NoError(t, err)
	require.NotNil(t, client)
	require.Equal(t, "gemini", client.Provider())

	_, err = client.Generate(t.Context(), domainllm.GenerateRequest{})
	require.Error(t, err)
	require.ErrorContains(t, err, "open Settings to add a key before generating")
}
