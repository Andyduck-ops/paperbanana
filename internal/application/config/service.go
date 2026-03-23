// Package config provides business logic for provider and API key management.
package config

import (
	"context"
	"fmt"

	domainconfig "github.com/paperbanana/paperbanana/internal/domain/config"
	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
	"github.com/paperbanana/paperbanana/internal/infrastructure/llm/models"
)

// Service provides business logic for provider and API key management.
type Service struct {
	providers domainconfig.ProviderRepository
	apiKeys   domainconfig.APIKeyRepository
	watcher   *Watcher
}

// NewService creates a new config service.
func NewService(providers domainconfig.ProviderRepository, apiKeys domainconfig.APIKeyRepository) *Service {
	return &Service{providers: providers, apiKeys: apiKeys}
}

// NewServiceWithWatcher creates a new config service with event watching.
func NewServiceWithWatcher(providers domainconfig.ProviderRepository, apiKeys domainconfig.APIKeyRepository, watcher *Watcher) *Service {
	return &Service{providers: providers, apiKeys: apiKeys, watcher: watcher}
}

// ListProviders lists all providers.
func (s *Service) ListProviders() ([]*domainconfig.Provider, error) {
	return s.providers.List()
}

// GetProvider gets a provider by ID.
func (s *Service) GetProvider(id string) (*domainconfig.Provider, error) {
	return s.providers.GetByID(id)
}

// GetProviderByName gets a provider by name.
func (s *Service) GetProviderByName(name string) (*domainconfig.Provider, error) {
	return s.providers.GetByName(name)
}

// CreateProvider creates a new provider.
func (s *Service) CreateProvider(p *domainconfig.Provider) error {
	if err := s.providers.Create(p); err != nil {
		return err
	}
	if s.watcher != nil {
		s.watcher.Emit(domainconfig.EventProviderCreated, p.ID, "")
	}
	return nil
}

// UpdateProvider updates a provider.
func (s *Service) UpdateProvider(p *domainconfig.Provider) error {
	if err := s.providers.Update(p); err != nil {
		return err
	}
	if s.watcher != nil {
		s.watcher.Emit(domainconfig.EventProviderUpdated, p.ID, "")
	}
	return nil
}

// DeleteProvider deletes a provider.
func (s *Service) DeleteProvider(id string) error {
	if err := s.providers.Delete(id); err != nil {
		return err
	}
	if s.watcher != nil {
		s.watcher.Emit(domainconfig.EventProviderDeleted, id, "")
	}
	return nil
}

// SetDefaultProvider sets the default provider.
func (s *Service) SetDefaultProvider(id string) error {
	provider, err := s.providers.GetByID(id)
	if err != nil {
		return err
	}
	if !provider.Enabled {
		return fmt.Errorf("provider %s must be enabled before it can be the default", provider.Name)
	}

	keys, err := s.apiKeys.GetActiveKeys(provider.ID)
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return fmt.Errorf("provider %s must have at least one active API key before it can be the default", provider.Name)
	}

	return s.providers.SetDefault(id)
}

// GetDefaultProvider gets the default provider.
func (s *Service) GetDefaultProvider() (*domainconfig.Provider, error) {
	return s.providers.GetDefault()
}

// ListAPIKeys lists all API keys for a provider.
func (s *Service) ListAPIKeys(providerID string) ([]*domainconfig.APIKey, error) {
	return s.apiKeys.ListByProvider(providerID)
}

// AddAPIKey adds a new API key for a provider.
func (s *Service) AddAPIKey(ctx context.Context, providerID, plaintext string) (*domainconfig.APIKey, error) {
	key := &domainconfig.APIKey{
		ProviderID: providerID,
		IsActive:   true,
	}
	if err := s.apiKeys.Create(ctx, key, plaintext); err != nil {
		return nil, err
	}
	if s.watcher != nil {
		s.watcher.Emit(domainconfig.EventKeyAdded, providerID, key.ID)
	}
	return key, nil
}

// DeleteAPIKey deletes an API key.
func (s *Service) DeleteAPIKey(id string) error {
	key, err := s.apiKeys.GetByID(id)
	if err != nil {
		return err
	}
	if err := s.apiKeys.Delete(id); err != nil {
		return err
	}
	if s.watcher != nil {
		s.watcher.Emit(domainconfig.EventKeyDeleted, key.ProviderID, id)
	}
	return nil
}

// ToggleAPIKey toggles an API key's active status.
func (s *Service) ToggleAPIKey(id string, active bool) error {
	key, err := s.apiKeys.GetByID(id)
	if err != nil {
		return err
	}
	key.IsActive = active
	if err := s.apiKeys.Update(key); err != nil {
		return err
	}
	if s.watcher != nil {
		s.watcher.Emit(domainconfig.EventKeyToggled, key.ProviderID, id)
	}
	return nil
}

// GetDecryptedKey gets and decrypts an API key.
func (s *Service) GetDecryptedKey(ctx context.Context, id string) (string, error) {
	return s.apiKeys.GetDecrypted(ctx, id)
}

// GetNextAPIKey gets the next API key for rotation.
func (s *Service) GetNextAPIKey(ctx context.Context, providerID string) (*domainconfig.APIKey, string, error) {
	return s.apiKeys.GetNextKey(ctx, providerID)
}

// ValidateProvider validates a provider configuration.
func (s *Service) ValidateProvider(ctx context.Context, provider, apiKey, baseURL string) *ValidationResult {
	return NewValidator().ValidateConnection(ctx, provider, apiKey, baseURL)
}

// ListModels lists available models for a provider.
func (s *Service) ListModels(ctx context.Context, providerID string) ([]domainllm.ModelInfo, error) {
	// Try by ID first, then by name
	provider, err := s.providers.GetByID(providerID)
	if err != nil {
		provider, err = s.providers.GetByName(providerID)
		if err != nil {
			return nil, err
		}
	}

	keys, err := s.apiKeys.GetActiveKeys(provider.ID)
	if err != nil || len(keys) == 0 {
		return nil, err
	}

	decrypted, err := s.apiKeys.GetDecrypted(ctx, keys[0].ID)
	if err != nil {
		return nil, err
	}

	// Use provider type for adapter lookup
	providerType := string(provider.Type)
	if providerType == "" || providerType == "custom" {
		providerType = provider.Name
	}
	return ListModelsForProvider(ctx, providerType, decrypted, provider.APIHost)
}

// ListModelsForProvider lists models for a provider using the given credentials.
func ListModelsForProvider(ctx context.Context, provider, apiKey, baseURL string) ([]domainllm.ModelInfo, error) {
	lister, err := models.GetModelLister(provider, apiKey, baseURL)
	if err != nil {
		return nil, err
	}
	return lister.ListModels(ctx)
}

// GetWatcher returns the config watcher.
func (s *Service) GetWatcher() *Watcher {
	return s.watcher
}

// ClearAllAPIKeysForSystemProviders removes all API keys from system providers.
// This preserves the provider records but clears all stored credentials.
// Returns the number of keys cleared.
func (s *Service) ClearAllAPIKeysForSystemProviders(ctx context.Context) (int, error) {
	providers, err := s.providers.List()
	if err != nil {
		return 0, fmt.Errorf("failed to list providers: %w", err)
	}

	cleared := 0
	for _, p := range providers {
		if !p.IsSystem {
			continue
		}

		keys, err := s.apiKeys.ListByProvider(p.ID)
		if err != nil {
			continue // Skip on error, don't fail entire operation
		}

		for _, key := range keys {
			if err := s.apiKeys.Delete(key.ID); err == nil {
				cleared++
				if s.watcher != nil {
					s.watcher.Emit(domainconfig.EventKeyDeleted, p.ID, key.ID)
				}
			}
		}
	}

	return cleared, nil
}
