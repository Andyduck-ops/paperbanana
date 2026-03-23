package config

import (
	"context"

	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
	"github.com/paperbanana/paperbanana/internal/infrastructure/llm/models"
)

// ValidationResult contains the result of a provider validation.
type ValidationResult struct {
	Valid           bool                   `json:"valid"`
	Errors          []ValidationError      `json:"errors,omitempty"`
	Message         string                 `json:"message,omitempty"`
	ModelsAvailable int                    `json:"models_available,omitempty"`
	Models          []domainllm.ModelInfo  `json:"models,omitempty"`
}

// ValidationError represents a single validation error.
type ValidationError struct {
	Field      string `json:"field"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
}

// Validator validates provider configurations.
type Validator struct{}

// NewValidator creates a new validator.
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateConnection tests a provider connection.
func (v *Validator) ValidateConnection(ctx context.Context, provider, apiKey, baseURL string) *ValidationResult {
	// Try to list models as a connection test
	lister, err := models.GetModelLister(provider, apiKey, baseURL)
	if err != nil {
		// Anthropic doesn't support listing, use different test
		if provider == "anthropic" {
			return v.validateAnthropicConnection(apiKey)
		}
		return &ValidationResult{
			Valid: false,
			Errors: []ValidationError{{
				Field:   "provider",
				Message: err.Error(),
			}},
		}
	}

	modelList, err := lister.ListModels(ctx)
	if err != nil {
		return &ValidationResult{
			Valid: false,
			Errors: []ValidationError{{
				Field:      "api_key",
				Message:    "Failed to connect to provider API: " + err.Error(),
				Suggestion: "Check that your API key is correct and has not expired",
			}},
		}
	}

	return &ValidationResult{
		Valid:           true,
		Message:         "Connection successful",
		ModelsAvailable: len(modelList),
		Models:          modelList,
	}
}

func (v *Validator) validateAnthropicConnection(apiKey string) *ValidationResult {
	// Anthropic doesn't have a models endpoint, so we validate key format
	if len(apiKey) < 20 {
		return &ValidationResult{
			Valid: false,
			Errors: []ValidationError{{
				Field:      "api_key",
				Message:    "API key appears to be too short",
				Suggestion: "Anthropic API keys are typically 100+ characters starting with 'sk-ant-'",
			}},
		}
	}

	// Return success - actual API test would require a messages call
	return &ValidationResult{
		Valid:   true,
		Message: "API key format is valid (connection test limited for Anthropic)",
	}
}
