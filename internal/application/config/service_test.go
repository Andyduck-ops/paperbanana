package config

import (
	"testing"

	domainconfig "github.com/paperbanana/paperbanana/internal/domain/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testProviderRepo struct {
	providers map[string]*domainconfig.Provider
	defaultID string
}

func (r *testProviderRepo) Create(provider *domainconfig.Provider) error {
	r.providers[provider.ID] = provider
	return nil
}

func (r *testProviderRepo) GetByID(id string) (*domainconfig.Provider, error) {
	provider, ok := r.providers[id]
	if !ok {
		return nil, assert.AnError
	}
	return provider, nil
}

func (r *testProviderRepo) GetByName(name string) (*domainconfig.Provider, error) {
	for _, provider := range r.providers {
		if provider.Name == name {
			return provider, nil
		}
	}
	return nil, assert.AnError
}

func (r *testProviderRepo) List() ([]*domainconfig.Provider, error) {
	result := make([]*domainconfig.Provider, 0, len(r.providers))
	for _, provider := range r.providers {
		result = append(result, provider)
	}
	return result, nil
}

func (r *testProviderRepo) ListEnabled() ([]*domainconfig.Provider, error) {
	var result []*domainconfig.Provider
	for _, provider := range r.providers {
		if provider.Enabled {
			result = append(result, provider)
		}
	}
	return result, nil
}

func (r *testProviderRepo) Update(provider *domainconfig.Provider) error {
	r.providers[provider.ID] = provider
	return nil
}

func (r *testProviderRepo) Delete(id string) error {
	delete(r.providers, id)
	return nil
}

func (r *testProviderRepo) SetDefault(id string) error {
	r.defaultID = id
	return nil
}

func (r *testProviderRepo) GetDefault() (*domainconfig.Provider, error) {
	return r.GetByID(r.defaultID)
}

func (r *testProviderRepo) InitializeSystemProviders() error {
	return nil
}

type testAPIKeyRepo struct {
	active map[string][]*domainconfig.APIKey
}

func (r *testAPIKeyRepo) Create(ctx interface{}, key *domainconfig.APIKey, plaintext string) error {
	return nil
}

func (r *testAPIKeyRepo) GetByID(id string) (*domainconfig.APIKey, error) {
	return nil, assert.AnError
}

func (r *testAPIKeyRepo) GetDecrypted(ctx interface{}, id string) (string, error) {
	return "", nil
}

func (r *testAPIKeyRepo) ListByProvider(providerID string) ([]*domainconfig.APIKey, error) {
	return r.active[providerID], nil
}

func (r *testAPIKeyRepo) GetActiveKeys(providerID string) ([]*domainconfig.APIKey, error) {
	return r.active[providerID], nil
}

func (r *testAPIKeyRepo) GetNextKey(ctx interface{}, providerID string) (*domainconfig.APIKey, string, error) {
	keys := r.active[providerID]
	if len(keys) == 0 {
		return nil, "", assert.AnError
	}
	return keys[0], "test-key", nil
}

func (r *testAPIKeyRepo) Update(key *domainconfig.APIKey) error {
	return nil
}

func (r *testAPIKeyRepo) Delete(id string) error {
	return nil
}

func (r *testAPIKeyRepo) MarkUsed(id string) error {
	return nil
}

func TestSetDefaultProviderRequiresEnabledProvider(t *testing.T) {
	providers := &testProviderRepo{
		providers: map[string]*domainconfig.Provider{
			"p1": {ID: "p1", Name: "anthropic", Enabled: false},
		},
	}
	keys := &testAPIKeyRepo{
		active: map[string][]*domainconfig.APIKey{
			"p1": {{ID: "k1", ProviderID: "p1", IsActive: true}},
		},
	}

	svc := NewService(providers, keys)

	err := svc.SetDefaultProvider("p1")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be enabled")
	assert.Empty(t, providers.defaultID)
}

func TestSetDefaultProviderRequiresActiveKey(t *testing.T) {
	providers := &testProviderRepo{
		providers: map[string]*domainconfig.Provider{
			"p1": {ID: "p1", Name: "anthropic", Enabled: true},
		},
	}
	keys := &testAPIKeyRepo{
		active: map[string][]*domainconfig.APIKey{},
	}

	svc := NewService(providers, keys)

	err := svc.SetDefaultProvider("p1")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "active API key")
	assert.Empty(t, providers.defaultID)
}

func TestSetDefaultProviderSucceedsForEnabledConfiguredProvider(t *testing.T) {
	providers := &testProviderRepo{
		providers: map[string]*domainconfig.Provider{
			"p1": {ID: "p1", Name: "anthropic", Enabled: true},
		},
	}
	keys := &testAPIKeyRepo{
		active: map[string][]*domainconfig.APIKey{
			"p1": {{ID: "k1", ProviderID: "p1", IsActive: true}},
		},
	}

	svc := NewService(providers, keys)

	err := svc.SetDefaultProvider("p1")

	require.NoError(t, err)
	assert.Equal(t, "p1", providers.defaultID)
}
