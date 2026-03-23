package httpnode

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	pbconfig "github.com/paperbanana/paperbanana/internal/config"
	"github.com/paperbanana/paperbanana/internal/infrastructure/resilience"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfiguredNodeExecuteExtractsConfiguredJSONPaths(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "token-123", r.Header.Get("X-API-Key"))

		var payload map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		assert.Equal(t, "summarize this dataset", payload["prompt"])
		assert.EqualValues(t, 3, payload["max_results"])

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"artifact":{"image":{"base64":"ZmFrZS1pbWFnZQ==","mime_type":"image/svg+xml"},"summary":"rendered with configured selectors","plot_code":"plt.bar([1,2],[3,4])"}}`))
	}))
	defer server.Close()

	adapter := NewAdapter(resilience.NewResilientClient("node-success", time.Second))
	node := pbconfig.NodeDefinition{
		Name:   "external_analyzer",
		URL:    server.URL,
		Method: http.MethodPost,
		Headers: map[string]string{
			"X-API-Key": "token-123",
		},
		RequestTemplate: map[string]any{
			"prompt":      "summarize this dataset",
			"max_results": 3,
		},
		ResponseParser: "json_path",
		ResponseSelectors: pbconfig.NodeResponseSelectors{
			ImageBase64: "$.artifact.image.base64",
			MIMEType:    "$.artifact.image.mime_type",
			Summary:     "$.artifact.summary",
			PlotCode:    "$.artifact.plot_code",
		},
	}

	result, err := adapter.Execute(context.Background(), node)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, result.StatusCode)
	assert.Equal(t, "ZmFrZS1pbWFnZQ==", result.Body["image_base64"])
	assert.Equal(t, "image/svg+xml", result.Body["mime_type"])
	assert.Equal(t, "rendered with configured selectors", result.Body["summary"])
	assert.Equal(t, "plt.bar([1,2],[3,4])", result.Body["plot_code"])
}

func TestConfiguredNodeExecuteExpandsEnvBackedHeaders(t *testing.T) {
	t.Setenv("NODE_TOKEN", "expanded-token")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer expanded-token", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"artifact":{"image_base64":"ZmFrZS1pbWFnZQ==","mime_type":"image/png","summary":"configured node render"}}`))
	}))
	defer server.Close()

	configPath := filepath.Join(t.TempDir(), "custom_nodes.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(`
custom_nodes:
  - name: external_analyzer
    url: `+server.URL+`
    method: POST
    headers:
      Authorization: Bearer ${NODE_TOKEN}
    request_template:
      prompt: summarize this dataset
    response_parser: json_path
    response_selectors:
      image_base64: $.artifact.image_base64
      mime_type: $.artifact.mime_type
      summary: $.artifact.summary
`), 0o644))

	catalog, err := pbconfig.LoadNodeConfig(configPath)
	require.NoError(t, err)
	require.Len(t, catalog.CustomNodes, 1)

	adapter := NewAdapter(resilience.NewResilientClient("node-env", time.Second))
	result, err := adapter.Execute(context.Background(), catalog.CustomNodes[0])
	require.NoError(t, err)
	assert.Equal(t, "configured node render", result.Body["summary"])
}

func TestConfiguredNodeExecutePropagatesHTTPFailures(t *testing.T) {
	t.Run("transport failure", func(t *testing.T) {
		adapter := NewAdapter(resilience.NewResilientClient("node-transport", 50*time.Millisecond))
		node := pbconfig.NodeDefinition{
			Name:   "external_analyzer",
			URL:    "http://127.0.0.1:1",
			Method: http.MethodPost,
			RequestTemplate: map[string]any{
				"prompt": "summarize this dataset",
			},
			ResponseParser: "json_path",
			ResponseSelectors: pbconfig.NodeResponseSelectors{
				ImageBase64: "$.artifact.image_base64",
				MIMEType:    "$.artifact.mime_type",
				Summary:     "$.artifact.summary",
			},
		}

		_, err := adapter.Execute(context.Background(), node)
		require.Error(t, err)

		var execErr *ExecutionError
		require.ErrorAs(t, err, &execErr)
		assert.Equal(t, ErrorKindTransport, execErr.Kind)
	})

	t.Run("response parse failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`not-json`))
		}))
		defer server.Close()

		adapter := NewAdapter(resilience.NewResilientClient("node-parse", time.Second))
		node := pbconfig.NodeDefinition{
			Name:   "external_analyzer",
			URL:    server.URL,
			Method: http.MethodPost,
			RequestTemplate: map[string]any{
				"prompt": "summarize this dataset",
			},
			ResponseParser: "json_path",
			ResponseSelectors: pbconfig.NodeResponseSelectors{
				ImageBase64: "$.artifact.image_base64",
				MIMEType:    "$.artifact.mime_type",
				Summary:     "$.artifact.summary",
			},
		}

		_, err := adapter.Execute(context.Background(), node)
		require.Error(t, err)

		var execErr *ExecutionError
		require.ErrorAs(t, err, &execErr)
		assert.Equal(t, ErrorKindParseResponse, execErr.Kind)
	})
}
