package models

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenAIModelListerReturnsAllModelsFromCompatibleEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/models", r.URL.Path)
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
			"object": "list",
			"data": []map[string]any{
				{"id": "grok-4.1-fast", "object": "model"},
				{"id": "gemini-3-pro", "object": "model"},
				{"id": "gpt-5.4", "object": "model"},
			},
		}))
	}))
	defer server.Close()

	lister := NewOpenAIModelLister("test-key", server.URL)
	models, err := lister.ListModels(context.Background())

	require.NoError(t, err)
	require.Len(t, models, 3)
	assert.Equal(t, []string{"gemini-3-pro", "gpt-5.4", "grok-4.1-fast"}, []string{
		models[0].ID,
		models[1].ID,
		models[2].ID,
	})
}
