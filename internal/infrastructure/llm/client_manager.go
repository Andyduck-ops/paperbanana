package llm

import (
	"context"
	"sync"

	domainconfig "github.com/paperbanana/paperbanana/internal/domain/config"
	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
)

// ClientManager manages cached LLM clients with hot reload support.
type ClientManager struct {
	mu           sync.RWMutex
	clients      map[string]domainllm.LLMClient
	factory      func(provider, apiKey, model string) (domainllm.LLMClient, error)
	providerRepo domainconfig.ProviderRepository
	apiKeyRepo   domainconfig.APIKeyRepository
}

// NewClientManager creates a new client manager.
func NewClientManager(
	factory func(provider, apiKey, model string) (domainllm.LLMClient, error),
	providerRepo domainconfig.ProviderRepository,
	apiKeyRepo domainconfig.APIKeyRepository,
) *ClientManager {
	return &ClientManager{
		clients:      make(map[string]domainllm.LLMClient),
		factory:      factory,
		providerRepo: providerRepo,
		apiKeyRepo:   apiKeyRepo,
	}
}

// GetClient returns a cached or newly created client for the provider.
func (m *ClientManager) GetClient(ctx context.Context, providerID string) (domainllm.LLMClient, error) {
	return m.GetClientForPurpose(ctx, providerID, "query")
}

// GetClientForPurpose returns a client configured for a specific purpose (query or gen).
func (m *ClientManager) GetClientForPurpose(ctx context.Context, providerID, purpose string) (domainllm.LLMClient, error) {
	cacheKey := providerID + ":" + purpose

	m.mu.RLock()
	if client, ok := m.clients[cacheKey]; ok {
		m.mu.RUnlock()
		return client, nil
	}
	m.mu.RUnlock()

	// Create new client
	provider, err := m.providerRepo.GetByID(providerID)
	if err != nil {
		return nil, err
	}

	key, plaintext, err := m.apiKeyRepo.GetNextKey(ctx, providerID)
	if err != nil {
		return nil, err
	}

	// Select model based on purpose
	model := provider.QueryModel
	if purpose == "gen" && provider.GenModel != "" {
		model = provider.GenModel
	}

	client, err := m.factory(string(provider.Type), plaintext, model)
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	m.clients[cacheKey] = client
	m.mu.Unlock()

	_ = key // Mark as used

	return client, nil
}

// Invalidate removes a cached client.
func (m *ClientManager) Invalidate(providerID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.clients, providerID)
}

// InvalidateAll removes all cached clients.
func (m *ClientManager) InvalidateAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clients = make(map[string]domainllm.LLMClient)
}

// SubscribeToConfigEvents subscribes to config changes and invalidates cache.
func (m *ClientManager) SubscribeToConfigEvents(events <-chan domainconfig.ConfigEvent) {
	go func() {
		for event := range events {
			switch event.Type {
			case domainconfig.EventProviderUpdated,
				domainconfig.EventProviderDeleted,
				domainconfig.EventKeyAdded,
				domainconfig.EventKeyDeleted,
				domainconfig.EventKeyToggled:
				m.Invalidate(event.ProviderID)
			case domainconfig.EventProviderCreated:
				// No action needed - will be created on first use
			}
		}
	}()
}
