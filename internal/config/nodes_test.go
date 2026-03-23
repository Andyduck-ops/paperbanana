package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadNodeConfigRejectsUnknownFields(t *testing.T) {
	path := writeNodeConfigFile(t, `
custom_nodes:
  - name: external_analyzer
    url: https://api.example.com/analyze
    methd: POST
    response_parser: json_path
    response_selectors:
      image_base64: $.artifact.image_base64
      mime_type: $.artifact.mime_type
      summary: $.artifact.summary
    request_template:
      prompt: summarize this
`)

	cfg, err := LoadNodeConfig(path)
	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.ErrorContains(t, err, "field methd not found")
}

func TestLoadNodeConfigValidatesRequiredFields(t *testing.T) {
	testCases := []struct {
		name    string
		body    string
		wantErr string
	}{
		{
			name: "missing url",
			body: `
custom_nodes:
  - name: external_analyzer
    method: POST
    response_parser: json_path
    response_selectors:
      image_base64: $.artifact.image_base64
      mime_type: $.artifact.mime_type
      summary: $.artifact.summary
    request_template:
      prompt: summarize this
`,
			wantErr: "custom_nodes[0].url must be set",
		},
		{
			name: "missing method",
			body: `
custom_nodes:
  - name: external_analyzer
    url: https://api.example.com/analyze
    response_parser: json_path
    response_selectors:
      image_base64: $.artifact.image_base64
      mime_type: $.artifact.mime_type
      summary: $.artifact.summary
    request_template:
      prompt: summarize this
`,
			wantErr: "custom_nodes[0].method must be set",
		},
		{
			name: "missing response parser",
			body: `
custom_nodes:
  - name: external_analyzer
    url: https://api.example.com/analyze
    method: POST
    request_template:
      prompt: summarize this
`,
			wantErr: "custom_nodes[0].response_parser must be set",
		},
		{
			name: "missing request template",
			body: `
custom_nodes:
  - name: external_analyzer
    url: https://api.example.com/analyze
    method: POST
    response_parser: json_path
    response_selectors:
      image_base64: $.artifact.image_base64
      mime_type: $.artifact.mime_type
      summary: $.artifact.summary
`,
			wantErr: "custom_nodes[0].request_template must be set",
		},
		{
			name: "missing response selectors",
			body: `
custom_nodes:
  - name: external_analyzer
    url: https://api.example.com/analyze
    method: POST
    response_parser: json_path
    request_template:
      prompt: summarize this
`,
			wantErr: "custom_nodes[0].response_selectors.image_base64 must be set",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			path := writeNodeConfigFile(t, tc.body)

			cfg, err := LoadNodeConfig(path)
			require.Error(t, err)
			assert.Nil(t, cfg)
			assert.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestLoadNodeConfigExpandsEnvPlaceholders(t *testing.T) {
	t.Setenv("NODE_URL", "https://api.example.com/analyze")
	t.Setenv("NODE_API_KEY", "secret-token")
	t.Setenv("NODE_PROMPT", "render the configured chart")

	path := writeNodeConfigFile(t, `
custom_nodes:
  - name: external_analyzer
    url: ${NODE_URL}
    method: post
    headers:
      Authorization: Bearer ${NODE_API_KEY}
    request_template:
      prompt: ${NODE_PROMPT}
      nested:
        value: prefix-${NODE_API_KEY}
    response_parser: json_path
    response_selectors:
      image_base64: $.artifact.image_base64
      mime_type: $.artifact.mime_type
      summary: $.artifact.summary
      plot_code: $.artifact.plot_code
`)

	cfg, err := LoadNodeConfig(path)
	require.NoError(t, err)
	require.Len(t, cfg.CustomNodes, 1)

	node := cfg.CustomNodes[0]
	assert.Equal(t, "https://api.example.com/analyze", node.URL)
	assert.Equal(t, "POST", node.Method)
	assert.Equal(t, "Bearer secret-token", node.Headers["Authorization"])
	assert.Equal(t, "render the configured chart", node.RequestTemplate["prompt"])

	nested, ok := node.RequestTemplate["nested"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "prefix-secret-token", nested["value"])
}

func TestLoadNodeConfigValidatesResponseSelectors(t *testing.T) {
	path := writeNodeConfigFile(t, `
custom_nodes:
  - name: external_analyzer
    url: https://api.example.com/analyze
    method: POST
    headers:
      Authorization: Bearer ${NODE_API_KEY}
    request_template:
      prompt: summarize this
    response_parser: json_path
    response_selectors:
      image_base64: $.artifact.image_base64
      mime_type: $.artifact.mime_type
      summary: $.artifact.summary
      plot_code: $.artifact.plot_code
`)

	cfg, err := LoadNodeConfig(path)
	require.NoError(t, err)
	require.Len(t, cfg.CustomNodes, 1)

	selectors := cfg.CustomNodes[0].ResponseSelectors
	assert.Equal(t, "$.artifact.image_base64", selectors.ImageBase64)
	assert.Equal(t, "$.artifact.mime_type", selectors.MIMEType)
	assert.Equal(t, "$.artifact.summary", selectors.Summary)
	assert.Equal(t, "$.artifact.plot_code", selectors.PlotCode)
}

func TestLoadNodeConfigRejectsUnsupportedParserContracts(t *testing.T) {
	testCases := []struct {
		name    string
		body    string
		wantErr string
	}{
		{
			name: "unsupported parser",
			body: `
custom_nodes:
  - name: external_analyzer
    url: https://api.example.com/analyze
    method: POST
    request_template:
      prompt: summarize this
    response_parser: body_map
    response_selectors:
      image_base64: $.artifact.image_base64
      mime_type: $.artifact.mime_type
      summary: $.artifact.summary
`,
			wantErr: `custom_nodes[0].response_parser "body_map" is not supported`,
		},
		{
			name: "missing summary selector",
			body: `
custom_nodes:
  - name: external_analyzer
    url: https://api.example.com/analyze
    method: POST
    request_template:
      prompt: summarize this
    response_parser: json_path
    response_selectors:
      image_base64: $.artifact.image_base64
      mime_type: $.artifact.mime_type
`,
			wantErr: "custom_nodes[0].response_selectors.summary must be set",
		},
		{
			name: "invalid selector path",
			body: `
custom_nodes:
  - name: external_analyzer
    url: https://api.example.com/analyze
    method: POST
    request_template:
      prompt: summarize this
    response_parser: json_path
    response_selectors:
      image_base64: artifact.image_base64
      mime_type: $.artifact.mime_type
      summary: $.artifact.summary
`,
			wantErr: "custom_nodes[0].response_selectors.image_base64 must start with $.",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := LoadNodeConfig(writeNodeConfigFile(t, tc.body))
			require.Error(t, err)
			assert.Nil(t, cfg)
			assert.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func writeNodeConfigFile(t *testing.T, body string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "custom_nodes.yaml")
	require.NoError(t, os.WriteFile(path, []byte(body), 0o644))

	return path
}
