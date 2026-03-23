package llm

import (
	"context"
	"testing"

	pbconfig "github.com/paperbanana/paperbanana/internal/config"
	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLLMClient(t *testing.T) {
	tests := []struct {
		provider string
		wantErr  bool
	}{
		{provider: "gemini"},
		{provider: "openai"},
		{provider: "anthropic"},
		{provider: "openrouter"},
		{provider: "unknown", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			client, err := NewLLMClient(tt.provider, "test-key", "test-model")
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, client)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, client)
			assert.Equal(t, tt.provider, client.Provider())
		})
	}
}

type mockProviderClient struct {
	provider string
	calls    int
	response *domainllm.GenerateResponse
}

func (m *mockProviderClient) Generate(_ context.Context, _ domainllm.GenerateRequest) (*domainllm.GenerateResponse, error) {
	m.calls++
	return m.response, nil
}

func (m *mockProviderClient) GenerateStream(_ context.Context, _ domainllm.GenerateRequest) (<-chan domainllm.StreamChunk, <-chan error) {
	chunks := make(chan domainllm.StreamChunk)
	errs := make(chan error)
	close(chunks)
	close(errs)
	return chunks, errs
}

func (m *mockProviderClient) Provider() string {
	return m.provider
}

type mockCache struct {
	values map[string]*domainllm.GenerateResponse
	sets   int
}

func (m *mockCache) Get(_ context.Context, provider, model string, req domainllm.GenerateRequest) (*domainllm.GenerateResponse, bool, error) {
	value, ok := m.values[m.cacheKey(provider, model, req)]
	return value, ok, nil
}

func (m *mockCache) Set(_ context.Context, provider, model string, req domainllm.GenerateRequest, resp *domainllm.GenerateResponse) error {
	if m.values == nil {
		m.values = map[string]*domainllm.GenerateResponse{}
	}

	m.values[m.cacheKey(provider, model, req)] = resp
	m.sets++
	return nil
}

func (m *mockCache) cacheKey(provider, model string, req domainllm.GenerateRequest) string {
	return provider + "|" + model + "|" + req.PromptVersion + "|" + domainllm.CollectText(req.Messages[0].Parts)
}

func TestCachedLLMClientBypassesUpstreamProvider(t *testing.T) {
	upstream := &mockProviderClient{
		provider: "openai",
		response: &domainllm.GenerateResponse{Content: "cached"},
	}
	cache := &mockCache{}
	client := NewCachedClient("openai", "gpt-4o", upstream, cache)

	req := domainllm.GenerateRequest{
		Messages: []domainllm.Message{
			{Role: domainllm.RoleUser, Parts: []domainllm.Part{domainllm.TextPart("build a chart")}},
		},
		PromptVersion: "v1",
	}

	first, err := client.Generate(context.Background(), req)
	require.NoError(t, err)

	second, err := client.Generate(context.Background(), req)
	require.NoError(t, err)

	assert.Equal(t, 1, upstream.calls)
	assert.Equal(t, 1, cache.sets)
	assert.Equal(t, "cached", first.Content)
	assert.Equal(t, "cached", second.Content)
}

func TestFactoryReturnsCachedClientWhenCacheConfigured(t *testing.T) {
	client, err := NewLLMClientWithOptions("openai", pbconfig.ProviderConfig{
		APIKey: "test-key",
		Model:  "gpt-4o",
	}, ClientOptions{
		Cache: &mockCache{},
	})
	require.NoError(t, err)

	cached, ok := client.(*CachedClient)
	require.True(t, ok)
	assert.Equal(t, "openai", cached.Provider())
}
