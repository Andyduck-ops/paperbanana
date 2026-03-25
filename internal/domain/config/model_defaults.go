package config

import "strings"

// SupportsImageGeneration reports whether the model can generate an image from
// a text prompt directly, without requiring an input image.
func SupportsImageGeneration(modelID string) bool {
	lowerID := strings.ToLower(strings.TrimSpace(modelID))
	if lowerID == "" {
		return false
	}

	switch lowerID {
	case "grok-imagine-1.0", "gpt-image-1", "chatgpt-image-latest":
		return true
	}

	if strings.Contains(lowerID, "imagine") && !strings.Contains(lowerID, "edit") {
		return true
	}
	if strings.Contains(lowerID, "image") && !strings.Contains(lowerID, "edit") && !strings.Contains(lowerID, "vision") {
		return true
	}

	return false
}

func SelectPreferredQueryModel(configured string, models []ModelInfo) string {
	configured = strings.TrimSpace(configured)
	if configured != "" && hasModel(configured, models) && !SupportsImageGeneration(configured) {
		return configured
	}

	preferred := []string{
		"grok-4.1-fast",
		"grok-4.1-thinking",
		"grok-4.20-beta",
		"gpt-5.4",
		"gpt-5.2",
		"gemini-3-pro",
		"gemini-3-flash",
		"claude-sonnet-4-6",
	}
	for _, candidate := range preferred {
		if hasModel(candidate, models) {
			return candidate
		}
	}

	for _, model := range models {
		lowerID := strings.ToLower(strings.TrimSpace(model.ID))
		if strings.Contains(lowerID, "embedding") || strings.Contains(lowerID, "video") || SupportsImageGeneration(lowerID) {
			continue
		}
		if model.Enabled {
			return model.ID
		}
	}

	for _, model := range models {
		if !SupportsImageGeneration(model.ID) && model.Enabled {
			return model.ID
		}
	}

	if configured != "" && hasModel(configured, models) {
		return configured
	}

	for _, model := range models {
		if model.Enabled {
			return model.ID
		}
	}

	return configured
}

func SelectPreferredGenerationModel(providerType ProviderType, providerName, configured string, models []ModelInfo) string {
	configured = strings.TrimSpace(configured)
	if configured != "" && hasModel(configured, models) && SupportsImageGeneration(configured) {
		return configured
	}

	preferred := []string{
		"grok-imagine-1.0",
		"gpt-image-1",
		"chatgpt-image-latest",
	}
	for _, candidate := range preferred {
		if hasModel(candidate, models) {
			return candidate
		}
	}

	for _, model := range models {
		if model.Enabled && SupportsImageGeneration(model.ID) {
			return model.ID
		}
	}

	if providerType == ProviderTypeGrok || strings.EqualFold(strings.TrimSpace(providerName), "grok") || strings.EqualFold(strings.TrimSpace(providerName), "xai") {
		for _, model := range models {
			if strings.HasPrefix(strings.ToLower(strings.TrimSpace(model.ID)), "grok-imagine") && model.Enabled {
				return model.ID
			}
		}
	}

	if configured != "" && hasModel(configured, models) {
		return configured
	}

	for _, model := range models {
		if model.Enabled {
			return model.ID
		}
	}

	return configured
}

func hasModel(modelID string, models []ModelInfo) bool {
	for _, model := range models {
		if model.ID == modelID {
			return true
		}
	}
	return false
}
