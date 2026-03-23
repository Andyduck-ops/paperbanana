// Package dto provides Data Transfer Objects for API responses.
package dto

import domainconfig "github.com/paperbanana/paperbanana/internal/domain/config"

// ProviderResponse is the API response for a provider configuration.
// API keys are always masked for security.
type ProviderResponse struct {
	ID          string                   `json:"id"`
	Type        string                   `json:"type"`
	Name        string                   `json:"name"`
	DisplayName string                   `json:"display_name"`
	QueryModel  string                   `json:"query_model"`
	GenModel    string                   `json:"gen_model"`
	BaseURL     string                   `json:"base_url,omitempty"`
	Timeout     string                   `json:"timeout,omitempty"`
	Status      string                   `json:"status"` // "configured", "no_keys", "invalid"
	Enabled     bool                     `json:"enabled"`
	IsSystem    bool                     `json:"is_system"`
	IsDefault   bool                     `json:"is_default"`
	Models      []domainconfig.ModelInfo `json:"models,omitempty"`
}

// ProviderListResponse is the API response for listing providers.
type ProviderListResponse struct {
	Providers []ProviderResponse `json:"providers"`
	Default   string             `json:"default"`
}

// ProviderPresetResponse is the API response for a provider preset.
type ProviderPresetResponse struct {
	Type           string                   `json:"type"`
	DisplayName    string                   `json:"display_name"`
	APIHost        string                   `json:"api_host"`
	DefaultModels  []domainconfig.ModelInfo `json:"default_models"`
	SupportsVision bool                     `json:"supports_vision"`
	DocsURL        string                   `json:"docs_url"`
	APIKeyURL      string                   `json:"api_key_url"`
}
