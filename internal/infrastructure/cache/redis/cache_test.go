package redis

import (
	"context"
	"testing"
	"time"

	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeStore struct {
	values  map[string][]byte
	lastKey string
	lastTTL time.Duration
}

func (s *fakeStore) Get(_ context.Context, key string) ([]byte, error) {
	value, ok := s.values[key]
	if !ok {
		return nil, ErrCacheMiss
	}

	return value, nil
}

func (s *fakeStore) Set(_ context.Context, key string, value []byte, ttl time.Duration) error {
	if s.values == nil {
		s.values = map[string][]byte{}
	}

	s.values[key] = value
	s.lastKey = key
	s.lastTTL = ttl
	return nil
}

func TestLLMCacheApplies24HourTTL(t *testing.T) {
	store := &fakeStore{}
	cache := NewCache(store)

	err := cache.Set(context.Background(), "openai", "gpt-4o", sampleRequest(), &domainllm.GenerateResponse{
		Content:      "cached response",
		TokensUsed:   42,
		FinishReason: "stop",
		Parts:        []domainllm.Part{domainllm.TextPart("cached response")},
	})
	require.NoError(t, err)
	assert.Equal(t, 24*time.Hour, store.lastTTL)
	assert.Equal(t, DefaultTTL, store.lastTTL)
}

func TestLLMCacheUsesCanonicalRequestKey(t *testing.T) {
	base := sampleRequest()

	keyA, err := CacheKey("openai", "gpt-4o", base)
	require.NoError(t, err)

	keyB, err := CacheKey("openai", "gpt-4o", sampleRequest())
	require.NoError(t, err)

	withPromptVersion := sampleRequest()
	withPromptVersion.PromptVersion = "v2"
	keyPromptVersion, err := CacheKey("openai", "gpt-4o", withPromptVersion)
	require.NoError(t, err)

	keyProvider, err := CacheKey("anthropic", "gpt-4o", base)
	require.NoError(t, err)

	keyModel, err := CacheKey("openai", "claude-3-5-sonnet", base)
	require.NoError(t, err)

	assert.Equal(t, keyA, keyB)
	assert.NotEqual(t, keyA, keyPromptVersion)
	assert.NotEqual(t, keyA, keyProvider)
	assert.NotEqual(t, keyA, keyModel)
}

func sampleRequest() domainllm.GenerateRequest {
	return domainllm.GenerateRequest{
		SystemInstruction: "You are a chart planner.",
		Messages: []domainllm.Message{
			{
				Role: domainllm.RoleUser,
				Parts: []domainllm.Part{
					domainllm.TextPart("Build a rainfall chart."),
					domainllm.URLImagePart("image/png", "https://example.com/reference.png"),
				},
			},
		},
		Model:         "gpt-4o",
		Temperature:   0.2,
		MaxTokens:     512,
		PromptVersion: "v1",
	}
}
