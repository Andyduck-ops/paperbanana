package llm

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	pbconfig "github.com/paperbanana/paperbanana/internal/config"
	domainconfig "github.com/paperbanana/paperbanana/internal/domain/config"
	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
	"github.com/paperbanana/paperbanana/internal/infrastructure/resilience"
	openaisdk "github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/require"
)

type runtimeTestProviderRepo struct {
	provider *domainconfig.Provider
}

func (r runtimeTestProviderRepo) Create(provider *domainconfig.Provider) error {
	r.provider = provider
	return nil
}

func (r runtimeTestProviderRepo) GetByID(id string) (*domainconfig.Provider, error) {
	if r.provider != nil && r.provider.ID == id {
		return r.provider, nil
	}
	return nil, errors.New("provider not found")
}

func (r runtimeTestProviderRepo) GetByName(name string) (*domainconfig.Provider, error) {
	if r.provider != nil && r.provider.Name == name {
		return r.provider, nil
	}
	return nil, errors.New("provider not found")
}

func (r runtimeTestProviderRepo) List() ([]*domainconfig.Provider, error) {
	return nil, nil
}

func (r runtimeTestProviderRepo) ListEnabled() ([]*domainconfig.Provider, error) {
	return nil, nil
}

func (r runtimeTestProviderRepo) Update(provider *domainconfig.Provider) error {
	r.provider = provider
	return nil
}

func (r runtimeTestProviderRepo) Delete(string) error {
	return nil
}

func (r runtimeTestProviderRepo) SetDefault(string) error {
	return nil
}

func (r runtimeTestProviderRepo) GetDefault() (*domainconfig.Provider, error) {
	if r.provider == nil {
		return nil, errors.New("provider not found")
	}
	return r.provider, nil
}

func (r runtimeTestProviderRepo) InitializeSystemProviders() error {
	return nil
}

type runtimeTestAPIKeyRepo struct{}

func (runtimeTestAPIKeyRepo) Create(ctx interface{}, key *domainconfig.APIKey, plaintext string) error {
	return nil
}

func (runtimeTestAPIKeyRepo) GetByID(id string) (*domainconfig.APIKey, error) {
	return nil, nil
}

func (runtimeTestAPIKeyRepo) GetDecrypted(ctx interface{}, id string) (string, error) {
	return "", nil
}

func (runtimeTestAPIKeyRepo) ListByProvider(providerID string) ([]*domainconfig.APIKey, error) {
	return nil, nil
}

func (runtimeTestAPIKeyRepo) GetActiveKeys(providerID string) ([]*domainconfig.APIKey, error) {
	return nil, nil
}

func (runtimeTestAPIKeyRepo) GetNextKey(ctx interface{}, providerID string) (*domainconfig.APIKey, string, error) {
	return &domainconfig.APIKey{ID: "key-1", ProviderID: providerID, IsActive: true}, "test-key", nil
}

func (runtimeTestAPIKeyRepo) Update(key *domainconfig.APIKey) error {
	return nil
}

func (runtimeTestAPIKeyRepo) Delete(id string) error {
	return nil
}

func (runtimeTestAPIKeyRepo) MarkUsed(id string) error {
	return nil
}

func TestRuntimeClientResolveClient_UsesRequestModelOverrideForDefaultProvider(t *testing.T) {
	t.Parallel()

	provider := &domainconfig.Provider{
		ID:          "provider-1",
		Type:        domainconfig.ProviderTypeGrok,
		Name:        "grok",
		DisplayName: "xAI Grok",
		APIHost:     "https://lx.lxsummer.cloud/v1",
		QueryModel:  "grok-4.1-fast",
		GenModel:    "grok-imagine-1.0",
	}

	client := NewRuntimeClient(
		RuntimePurposeQuery,
		"grok",
		pbconfig.ProviderConfig{
			APIKey:  "startup-key",
			BaseURL: "https://lx.lxsummer.cloud/v1",
			Model:   "grok-4.1-fast",
		},
		ClientOptions{},
		runtimeTestProviderRepo{provider: provider},
		runtimeTestAPIKeyRepo{},
	)

	resolvedClient, resolvedReq, err := client.resolveClient(context.Background(), domainllm.GenerateRequest{
		Model: "grok-4.1-thinking",
	})
	require.NoError(t, err)
	require.NotNil(t, resolvedClient)
	require.Equal(t, "grok-4.1-thinking", resolvedReq.Model)
}

func TestRuntimeClientResolveClient_UsesDefaultProviderModelWhenNoOverrideProvided(t *testing.T) {
	t.Parallel()

	provider := &domainconfig.Provider{
		ID:          "provider-1",
		Type:        domainconfig.ProviderTypeGrok,
		Name:        "grok",
		DisplayName: "xAI Grok",
		APIHost:     "https://lx.lxsummer.cloud/v1",
		QueryModel:  "grok-4.1-thinking",
		GenModel:    "grok-imagine-1.0",
	}

	client := NewRuntimeClient(
		RuntimePurposeQuery,
		"grok",
		pbconfig.ProviderConfig{
			APIKey:  "startup-key",
			BaseURL: "https://lx.lxsummer.cloud/v1",
			Model:   "grok-4.1-fast",
		},
		ClientOptions{},
		runtimeTestProviderRepo{provider: provider},
		runtimeTestAPIKeyRepo{},
	)

	_, resolvedReq, err := client.resolveClient(context.Background(), domainllm.GenerateRequest{})
	require.NoError(t, err)
	require.Equal(t, "grok-4.1-thinking", resolvedReq.Model)
}

func TestRuntimeClientGenerate_UsesProviderScopedHTTPClient(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/chat/completions", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(openaisdk.ChatCompletionResponse{
			Choices: []openaisdk.ChatCompletionChoice{{
				FinishReason: openaisdk.FinishReasonStop,
				Message: openaisdk.ChatCompletionMessage{
					Role:    openaisdk.ChatMessageRoleAssistant,
					Content: "ok",
				},
			}},
			Usage: openaisdk.Usage{TotalTokens: 12},
		}))
	}))
	defer server.Close()

	sharedClient := resilience.NewResilientClient("shared-breaker", 200*time.Millisecond).HTTPClient()
	failingServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer failingServer.Close()

	for i := 0; i < 6; i++ {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, failingServer.URL, nil)
		require.NoError(t, err)
		_, _ = sharedClient.Do(req)
	}

	provider := &domainconfig.Provider{
		ID:          "provider-1",
		Type:        domainconfig.ProviderTypeOpenAICompatible,
		Name:        "tencent-coding",
		DisplayName: "Tencent Coding",
		APIHost:     server.URL,
		QueryModel:  "kimi-k2.5",
	}

	client := NewRuntimeClient(
		RuntimePurposeQuery,
		"grok",
		pbconfig.ProviderConfig{
			APIKey:  "startup-key",
			BaseURL: "https://lx.lxsummer.cloud/v1",
			Model:   "grok-4.1-fast",
			Timeout: time.Second,
		},
		ClientOptions{HTTPClient: sharedClient},
		runtimeTestProviderRepo{provider: provider},
		runtimeTestAPIKeyRepo{},
	)

	resp, err := client.Generate(context.Background(), domainllm.GenerateRequest{
		Model: "tencent-coding:kimi-k2.5",
		Messages: []domainllm.Message{
			{Role: domainllm.RoleUser, Parts: []domainllm.Part{domainllm.TextPart("hello")}},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "ok", resp.Content)
}

func TestRuntimeClientGenerate_RetriesTransientProviderFailures(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/chat/completions", r.URL.Path)
		attempt := attempts.Add(1)
		if attempt < 3 {
			http.Error(w, "model engine error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(openaisdk.ChatCompletionResponse{
			Choices: []openaisdk.ChatCompletionChoice{{
				FinishReason: openaisdk.FinishReasonStop,
				Message: openaisdk.ChatCompletionMessage{
					Role:    openaisdk.ChatMessageRoleAssistant,
					Content: "retried-ok",
				},
			}},
			Usage: openaisdk.Usage{TotalTokens: 18},
		}))
	}))
	defer server.Close()

	provider := &domainconfig.Provider{
		ID:          "provider-1",
		Type:        domainconfig.ProviderTypeOpenAICompatible,
		Name:        "tencent-coding",
		DisplayName: "Tencent Coding",
		APIHost:     server.URL,
		QueryModel:  "minimax-m2.5",
		TimeoutMs:   1000,
	}

	client := NewRuntimeClient(
		RuntimePurposeQuery,
		"grok",
		pbconfig.ProviderConfig{
			APIKey:  "startup-key",
			BaseURL: "https://lx.lxsummer.cloud/v1",
			Model:   "grok-4.1-fast",
			Timeout: time.Second,
		},
		ClientOptions{},
		runtimeTestProviderRepo{provider: provider},
		runtimeTestAPIKeyRepo{},
	)

	resp, err := client.Generate(context.Background(), domainllm.GenerateRequest{
		Model: "tencent-coding:minimax-m2.5",
		Messages: []domainllm.Message{
			{Role: domainllm.RoleUser, Parts: []domainllm.Part{domainllm.TextPart("hello")}},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "retried-ok", resp.Content)
	require.EqualValues(t, 3, attempts.Load())
}
