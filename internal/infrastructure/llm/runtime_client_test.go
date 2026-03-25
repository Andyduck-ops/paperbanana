package llm

import (
	"context"
	"errors"
	"testing"

	pbconfig "github.com/paperbanana/paperbanana/internal/config"
	domainconfig "github.com/paperbanana/paperbanana/internal/domain/config"
	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
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
		ID:         "provider-1",
		Type:       domainconfig.ProviderTypeGrok,
		Name:       "grok",
		DisplayName: "xAI Grok",
		APIHost:    "https://lx.lxsummer.cloud/v1",
		QueryModel: "grok-4.1-fast",
		GenModel:   "grok-imagine-1.0",
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
		ID:         "provider-1",
		Type:       domainconfig.ProviderTypeGrok,
		Name:       "grok",
		DisplayName: "xAI Grok",
		APIHost:    "https://lx.lxsummer.cloud/v1",
		QueryModel: "grok-4.1-thinking",
		GenModel:   "grok-imagine-1.0",
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
