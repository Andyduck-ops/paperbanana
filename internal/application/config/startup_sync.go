package config

import (
	"context"
	"strings"
	"time"

	domainconfig "github.com/paperbanana/paperbanana/internal/domain/config"
	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
)

var listModelsForProviderFn = ListModelsForProvider

type StartupProviderSpec struct {
	Name         string
	DisplayName  string
	BaseURL      string
	APIKey       string
	DefaultModel string
	Timeout      time.Duration
	IsDefault    bool
}

func (s *Service) SyncStartupProviders(ctx context.Context, specs []StartupProviderSpec) error {
	for _, spec := range specs {
		if strings.TrimSpace(spec.Name) == "" {
			continue
		}

		provider, err := s.ensureStartupProvider(spec)
		if err != nil {
			return err
		}

		models, err := s.fetchStartupModels(ctx, provider, spec)
		if err != nil {
			models = ensureModelPresent(provider.Models, spec.DefaultModel)
		}
		if len(models) == 0 {
			models = ensureModelPresent(nil, spec.DefaultModel)
		}
		models = normalizeStartupModels(models)
		queryModel := domainconfig.SelectPreferredQueryModel(spec.DefaultModel, models)
		genModel := domainconfig.SelectPreferredGenerationModel(provider.Type, provider.Name, spec.DefaultModel, models)
		provider.APIHost = strings.TrimSpace(spec.BaseURL)
		provider.DisplayName = startupDisplayName(provider.DisplayName, spec)
		provider.Enabled = true
		provider.TimeoutMs = durationToMilliseconds(spec.Timeout)
		provider.Models = models
		provider.QueryModel = queryModel
		provider.GenModel = genModel

		if err := s.UpdateProvider(provider); err != nil {
			return err
		}

		if err := s.ensureAPIKey(ctx, provider.ID, spec.APIKey); err != nil {
			return err
		}

		if spec.IsDefault {
			if err := s.SetDefaultProvider(provider.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Service) ensureStartupProvider(spec StartupProviderSpec) (*domainconfig.Provider, error) {
	provider, err := s.GetProviderByName(spec.Name)
	if err == nil {
		return provider, nil
	}

	preset := domainconfig.GetPresetByName(spec.Name)
	providerType := domainconfig.ProviderTypeOpenAICompatible
	isSystem := false
	displayName := startupDisplayName("", spec)
	if preset != nil && sameBaseURL(spec.BaseURL, preset.APIHost) {
		providerType = preset.Type
		isSystem = true
		if displayName == "" {
			displayName = preset.DisplayName
		}
	}

	provider = &domainconfig.Provider{
		Name:        spec.Name,
		Type:        providerType,
		DisplayName: displayName,
		APIHost:     strings.TrimSpace(spec.BaseURL),
		Enabled:     true,
		IsSystem:    isSystem,
		TimeoutMs:   durationToMilliseconds(spec.Timeout),
	}

	if err := s.CreateProvider(provider); err != nil {
		return nil, err
	}

	return provider, nil
}

func (s *Service) fetchStartupModels(ctx context.Context, provider *domainconfig.Provider, spec StartupProviderSpec) ([]domainconfig.ModelInfo, error) {
	providerName := provider.Name
	if provider.Type != "" && provider.Type != domainconfig.ProviderTypeOpenAICompatible {
		providerName = string(provider.Type)
	}

	fetched, err := listModelsForProviderFn(ctx, providerName, spec.APIKey, spec.BaseURL)
	if err != nil {
		return nil, err
	}

	return toDomainModels(fetched), nil
}

func (s *Service) ensureAPIKey(ctx context.Context, providerID, plaintext string) error {
	if strings.TrimSpace(plaintext) == "" {
		return nil
	}

	keys, err := s.ListAPIKeys(providerID)
	if err == nil {
		for _, key := range keys {
			decrypted, decryptErr := s.GetDecryptedKey(ctx, key.ID)
			if decryptErr == nil && decrypted == plaintext {
				if !key.IsActive {
					key.IsActive = true
					return s.apiKeys.Update(key)
				}
				return nil
			}
		}
	}

	_, err = s.AddAPIKey(ctx, providerID, plaintext)
	return err
}

func toDomainModels(models []domainllm.ModelInfo) []domainconfig.ModelInfo {
	if len(models) == 0 {
		return nil
	}

	result := make([]domainconfig.ModelInfo, 0, len(models))
	seen := make(map[string]struct{}, len(models))
	for _, model := range models {
		modelID := strings.TrimSpace(model.ID)
		if modelID == "" {
			continue
		}
		if _, exists := seen[modelID]; exists {
			continue
		}
		seen[modelID] = struct{}{}

		name := strings.TrimSpace(model.Name)
		if name == "" {
			name = modelID
		}

		result = append(result, domainconfig.ModelInfo{
			ID:             modelID,
			Name:           name,
			MaxTokens:      model.MaxTokens,
			SupportsVision: model.SupportsVision,
			Enabled:        true,
		})
	}

	return normalizeStartupModels(result)
}

func ensureModelPresent(models []domainconfig.ModelInfo, modelID string) []domainconfig.ModelInfo {
	modelID = strings.TrimSpace(modelID)
	if modelID == "" {
		return models
	}

	for _, model := range models {
		if model.ID == modelID {
			return models
		}
	}

	return append(models, domainconfig.ModelInfo{
		ID:      modelID,
		Name:    modelID,
		Enabled: true,
	})
}

func durationToMilliseconds(value time.Duration) int {
	if value <= 0 {
		return 60000
	}
	return int(value / time.Millisecond)
}

func startupDisplayName(current string, spec StartupProviderSpec) string {
	if value := strings.TrimSpace(spec.DisplayName); value != "" {
		return value
	}
	if value := strings.TrimSpace(current); value != "" {
		return value
	}
	if preset := domainconfig.GetPresetByName(spec.Name); preset != nil {
		return preset.DisplayName
	}
	return spec.Name
}

func sameBaseURL(left, right string) bool {
	return strings.TrimRight(strings.ToLower(strings.TrimSpace(left)), "/") ==
		strings.TrimRight(strings.ToLower(strings.TrimSpace(right)), "/")
}

func hasModel(modelID string, models []domainconfig.ModelInfo) bool {
	for _, model := range models {
		if model.ID == modelID {
			return true
		}
	}
	return false
}

func normalizeStartupModels(models []domainconfig.ModelInfo) []domainconfig.ModelInfo {
	if len(models) == 0 {
		return nil
	}

	normalized := make([]domainconfig.ModelInfo, len(models))
	for i, model := range models {
		model.SupportsVision = model.SupportsVision || domainconfig.SupportsImageGeneration(model.ID)
		normalized[i] = model
	}

	return normalized
}
