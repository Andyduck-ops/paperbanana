package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	configservice "github.com/paperbanana/paperbanana/internal/application/config"
	domainconfig "github.com/paperbanana/paperbanana/internal/domain/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockProviderRepository implements domainconfig.ProviderRepository for testing
type mockProviderRepository struct {
	providers []*domainconfig.Provider
}

func (m *mockProviderRepository) Create(provider *domainconfig.Provider) error {
	m.providers = append(m.providers, provider)
	return nil
}

func (m *mockProviderRepository) GetByID(id string) (*domainconfig.Provider, error) {
	for _, p := range m.providers {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, assert.AnError
}

func (m *mockProviderRepository) GetByName(name string) (*domainconfig.Provider, error) {
	for _, p := range m.providers {
		if p.Name == name {
			return p, nil
		}
	}
	return nil, assert.AnError
}

func (m *mockProviderRepository) List() ([]*domainconfig.Provider, error) {
	return m.providers, nil
}

func (m *mockProviderRepository) ListEnabled() ([]*domainconfig.Provider, error) {
	var enabled []*domainconfig.Provider
	for _, p := range m.providers {
		if p.Enabled {
			enabled = append(enabled, p)
		}
	}
	return enabled, nil
}

func (m *mockProviderRepository) Update(provider *domainconfig.Provider) error {
	for i, p := range m.providers {
		if p.ID == provider.ID {
			m.providers[i] = provider
			return nil
		}
	}
	return assert.AnError
}

func (m *mockProviderRepository) Delete(id string) error {
	for i, p := range m.providers {
		if p.ID == id {
			m.providers = append(m.providers[:i], m.providers[i+1:]...)
			return nil
		}
	}
	return assert.AnError
}

func (m *mockProviderRepository) SetDefault(id string) error {
	for _, p := range m.providers {
		p.IsDefault = p.ID == id
	}
	return nil
}

func (m *mockProviderRepository) GetDefault() (*domainconfig.Provider, error) {
	for _, p := range m.providers {
		if p.IsDefault {
			return p, nil
		}
	}
	return nil, assert.AnError
}

func (m *mockProviderRepository) InitializeSystemProviders() error {
	return nil
}

// mockAPIKeyRepository implements domainconfig.APIKeyRepository for testing
type mockAPIKeyRepository struct {
	keys []*domainconfig.APIKey
}

func (m *mockAPIKeyRepository) Create(ctx interface{}, key *domainconfig.APIKey, plaintext string) error {
	m.keys = append(m.keys, key)
	return nil
}

func (m *mockAPIKeyRepository) GetByID(id string) (*domainconfig.APIKey, error) {
	for _, k := range m.keys {
		if k.ID == id {
			return k, nil
		}
	}
	return nil, assert.AnError
}

func (m *mockAPIKeyRepository) ListByProvider(providerID string) ([]*domainconfig.APIKey, error) {
	var result []*domainconfig.APIKey
	for _, k := range m.keys {
		if k.ProviderID == providerID {
			result = append(result, k)
		}
	}
	return result, nil
}

func (m *mockAPIKeyRepository) GetActiveKeys(providerID string) ([]*domainconfig.APIKey, error) {
	var result []*domainconfig.APIKey
	for _, k := range m.keys {
		if k.ProviderID == providerID && k.IsActive {
			result = append(result, k)
		}
	}
	return result, nil
}

func (m *mockAPIKeyRepository) GetNextKey(ctx interface{}, providerID string) (*domainconfig.APIKey, string, error) {
	keys, _ := m.GetActiveKeys(providerID)
	if len(keys) == 0 {
		return nil, "", assert.AnError
	}
	return keys[0], "test-decrypted-key", nil
}

func (m *mockAPIKeyRepository) GetDecrypted(ctx interface{}, id string) (string, error) {
	return "test-decrypted-key", nil
}

func (m *mockAPIKeyRepository) Update(key *domainconfig.APIKey) error {
	for i, k := range m.keys {
		if k.ID == key.ID {
			m.keys[i] = key
			return nil
		}
	}
	return assert.AnError
}

func (m *mockAPIKeyRepository) Delete(id string) error {
	for i, k := range m.keys {
		if k.ID == id {
			m.keys = append(m.keys[:i], m.keys[i+1:]...)
			return nil
		}
	}
	return assert.AnError
}

func (m *mockAPIKeyRepository) MarkUsed(id string) error {
	return nil
}

func TestResetSystemProviders_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup mock repositories
	providerRepo := &mockProviderRepository{
		providers: []*domainconfig.Provider{
			{ID: "sys-1", Name: "openai", IsSystem: true, Enabled: true},
			{ID: "sys-2", Name: "anthropic", IsSystem: true, Enabled: true},
			{ID: "custom-1", Name: "custom", IsSystem: false, Enabled: true},
		},
	}
	apiKeyRepo := &mockAPIKeyRepository{
		keys: []*domainconfig.APIKey{
			{ID: "key-1", ProviderID: "sys-1", IsActive: true},
			{ID: "key-2", ProviderID: "sys-1", IsActive: true},
			{ID: "key-3", ProviderID: "sys-2", IsActive: true},
			{ID: "key-4", ProviderID: "custom-1", IsActive: true},
		},
	}

	svc := configservice.NewService(providerRepo, apiKeyRepo)
	handler := NewProviderHandler(svc)

	// Create request with confirm="RESET"
	body := `{"confirm":"RESET"}`
	req := httptest.NewRequest(http.MethodPost, "/providers/reset", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute
	handler.ResetSystemProviders(c)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "System providers reset successfully", response["message"])
	assert.Equal(t, float64(3), response["keys_cleared"]) // 3 keys from system providers

	// Verify only custom provider's keys remain
	remainingKeys := 0
	for _, k := range apiKeyRepo.keys {
		remainingKeys++
		assert.Equal(t, "custom-1", k.ProviderID, "Only custom provider keys should remain")
	}
	assert.Equal(t, 1, remainingKeys, "Only custom provider key should remain")
}

func TestResetSystemProviders_InvalidConfirmation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	providerRepo := &mockProviderRepository{
		providers: []*domainconfig.Provider{
			{ID: "sys-1", Name: "openai", IsSystem: true, Enabled: true},
		},
	}
	apiKeyRepo := &mockAPIKeyRepository{
		keys: []*domainconfig.APIKey{
			{ID: "key-1", ProviderID: "sys-1", IsActive: true},
		},
	}

	svc := configservice.NewService(providerRepo, apiKeyRepo)
	handler := NewProviderHandler(svc)

	tests := []struct {
		name       string
		body       string
		expectCode int
		expectMsg  string
	}{
		{
			name:       "wrong confirmation text",
			body:       `{"confirm":"reset"}`,
			expectCode: http.StatusBadRequest,
			expectMsg:  "confirmation must be exactly 'RESET' (uppercase)",
		},
		{
			name:       "empty confirmation",
			body:       `{"confirm":""}`,
			expectCode: http.StatusBadRequest,
			expectMsg:  "confirmation required", // binding validation fails first
		},
		{
			name:       "missing confirm field",
			body:       `{}`,
			expectCode: http.StatusBadRequest,
			expectMsg:  "confirmation required",
		},
		{
			name:       "random confirmation text",
			body:       `{"confirm":"yes"}`,
			expectCode: http.StatusBadRequest,
			expectMsg:  "confirmation must be exactly 'RESET' (uppercase)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/providers/reset", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.ResetSystemProviders(c)

			assert.Equal(t, tt.expectCode, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Contains(t, response["error"], tt.expectMsg)
		})
	}
}

func TestResetSystemProviders_OnlySystemProvidersAffected(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup with multiple providers (system and custom)
	providerRepo := &mockProviderRepository{
		providers: []*domainconfig.Provider{
			{ID: "sys-1", Name: "openai", IsSystem: true, Enabled: true},
			{ID: "custom-1", Name: "my-custom", IsSystem: false, Enabled: true},
		},
	}
	apiKeyRepo := &mockAPIKeyRepository{
		keys: []*domainconfig.APIKey{
			{ID: "key-1", ProviderID: "sys-1", IsActive: true},
			{ID: "key-2", ProviderID: "custom-1", IsActive: true},
		},
	}

	svc := configservice.NewService(providerRepo, apiKeyRepo)
	handler := NewProviderHandler(svc)

	// Execute reset
	body := `{"confirm":"RESET"}`
	req := httptest.NewRequest(http.MethodPost, "/providers/reset", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.ResetSystemProviders(c)

	// Verify
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Only 1 key cleared (from system provider)
	assert.Equal(t, float64(1), response["keys_cleared"])

	// Custom provider's key should still exist
	customKeys, err := apiKeyRepo.ListByProvider("custom-1")
	require.NoError(t, err)
	assert.Len(t, customKeys, 1, "Custom provider keys should be preserved")
}

func TestResetSystemProviders_ResponseIncludesCount(t *testing.T) {
	gin.SetMode(gin.TestMode)

	providerRepo := &mockProviderRepository{
		providers: []*domainconfig.Provider{
			{ID: "sys-1", Name: "openai", IsSystem: true, Enabled: true},
		},
	}
	apiKeyRepo := &mockAPIKeyRepository{
		keys: []*domainconfig.APIKey{
			{ID: "key-1", ProviderID: "sys-1", IsActive: true},
			{ID: "key-2", ProviderID: "sys-1", IsActive: true},
			{ID: "key-3", ProviderID: "sys-1", IsActive: true},
		},
	}

	svc := configservice.NewService(providerRepo, apiKeyRepo)
	handler := NewProviderHandler(svc)

	body := `{"confirm":"RESET"}`
	req := httptest.NewRequest(http.MethodPost, "/providers/reset", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.ResetSystemProviders(c)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify keys_cleared is present and accurate
	assert.Contains(t, response, "keys_cleared")
	assert.Equal(t, float64(3), response["keys_cleared"])
}
