package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelectPreferredGenerationModelPrefersTextToImageModel(t *testing.T) {
	models := []ModelInfo{
		{ID: "grok-4.1-fast", Name: "Grok 4.1 Fast", Enabled: true},
		{ID: "grok-imagine-1.0", Name: "Grok Imagine 1.0", Enabled: true},
		{ID: "grok-imagine-1.0-edit", Name: "Grok Imagine 1.0 Edit", Enabled: true},
	}

	selected := SelectPreferredGenerationModel(ProviderTypeGrok, "grok", "grok-4.1-fast", models)

	assert.Equal(t, "grok-imagine-1.0", selected)
}

func TestSelectPreferredQueryModelSkipsImageOnlyModels(t *testing.T) {
	models := []ModelInfo{
		{ID: "grok-imagine-1.0", Name: "Grok Imagine 1.0", Enabled: true},
		{ID: "grok-4.1-fast", Name: "Grok 4.1 Fast", Enabled: true},
	}

	selected := SelectPreferredQueryModel("grok-imagine-1.0", models)

	assert.Equal(t, "grok-4.1-fast", selected)
}
