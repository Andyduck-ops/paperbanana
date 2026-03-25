package config

import (
	"context"
	"testing"
	"time"

	domainconfig "github.com/paperbanana/paperbanana/internal/domain/config"
	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncStartupProvidersSyncsModelsAndDefault(t *testing.T) {
	original := listModelsForProviderFn
	listModelsForProviderFn = func(_ context.Context, provider, apiKey, baseURL string) ([]domainllm.ModelInfo, error) {
		assert.Equal(t, "grok", provider)
		assert.Equal(t, "test-key", apiKey)
		assert.Equal(t, "https://lx.example.com/v1", baseURL)
		return []domainllm.ModelInfo{
			{ID: "grok-4.1-fast", Name: "grok-4.1-fast"},
			{ID: "grok-imagine-1.0", Name: "grok-imagine-1.0"},
		}, nil
	}
	defer func() { listModelsForProviderFn = original }()

	providers := &testProviderRepo{
		providers: map[string]*domainconfig.Provider{
			"p1": {
				ID:          "p1",
				Name:        "grok",
				Type:        domainconfig.ProviderTypeGrok,
				DisplayName: "xAI Grok",
				APIHost:     "https://api.x.ai/v1",
				Models: []domainconfig.ModelInfo{
					{ID: "grok-2-image-1212", Name: "old", Enabled: true},
				},
			},
		},
	}
	keys := &testAPIKeyRepo{
		active: map[string][]*domainconfig.APIKey{},
		plain:  map[string]string{},
	}

	svc := NewService(providers, keys)
	err := svc.SyncStartupProviders(context.Background(), []StartupProviderSpec{
		{
			Name:         "grok",
			BaseURL:      "https://lx.example.com/v1",
			APIKey:       "test-key",
			DefaultModel: "grok-2-image-1212",
			Timeout:      120 * time.Second,
			IsDefault:    true,
		},
	})

	require.NoError(t, err)
	provider := providers.providers["p1"]
	require.NotNil(t, provider)
	assert.Equal(t, "https://lx.example.com/v1", provider.APIHost)
	assert.True(t, provider.Enabled)
	assert.Equal(t, "grok-4.1-fast", provider.QueryModel)
	assert.Equal(t, "grok-imagine-1.0", provider.GenModel)
	assert.Equal(t, "p1", providers.defaultID)
	assert.Len(t, provider.Models, 2)
	require.Len(t, keys.active["p1"], 1)
	assert.Equal(t, "test-key", keys.plain[keys.active["p1"][0].ID])
}
