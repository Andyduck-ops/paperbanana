package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/paperbanana/paperbanana/internal/api/dto"
	configservice "github.com/paperbanana/paperbanana/internal/application/config"
	domainconfig "github.com/paperbanana/paperbanana/internal/domain/config"
)

// ProviderHandler handles provider-related HTTP requests.
type ProviderHandler struct {
	svc *configservice.Service
}

// NewProviderHandler creates a new provider handler.
func NewProviderHandler(svc *configservice.Service) *ProviderHandler {
	return &ProviderHandler{svc: svc}
}

// ListProviders handles GET /api/v1/providers
func (h *ProviderHandler) ListProviders(c *gin.Context) {
	providers, err := h.svc.ListProviders()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]dto.ProviderResponse, len(providers))
	for i, p := range providers {
		status := "configured"
		keys, _ := h.svc.ListAPIKeys(p.ID)
		if len(keys) == 0 {
			status = "no_keys"
		}
		response[i] = dto.ProviderResponse{
			ID:          p.ID,
			Type:        string(p.Type),
			Name:        p.Name,
			DisplayName: p.DisplayName,
			QueryModel:  p.QueryModel,
			GenModel:    p.GenModel,
			BaseURL:     p.APIHost,
			Timeout:     formatTimeout(p.TimeoutMs),
			Status:      status,
			Enabled:     p.Enabled,
			IsSystem:    p.IsSystem,
			IsDefault:   p.IsDefault,
			Models:      p.Models,
		}
	}

	c.JSON(http.StatusOK, gin.H{"providers": response})
}

// ListPresets handles GET /api/v1/providers/presets
func (h *ProviderHandler) ListPresets(c *gin.Context) {
	presets := domainconfig.BuiltInPresets()
	c.JSON(http.StatusOK, gin.H{"presets": presets})
}

// GetProvider handles GET /api/v1/providers/:id
func (h *ProviderHandler) GetProvider(c *gin.Context) {
	id := c.Param("id")

	// Try by ID first, then by name
	provider, err := h.svc.GetProvider(id)
	if err != nil {
		provider, err = h.svc.GetProviderByName(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "provider not found"})
			return
		}
	}

	status := "configured"
	keys, _ := h.svc.ListAPIKeys(provider.ID)
	if len(keys) == 0 {
		status = "no_keys"
	}

	c.JSON(http.StatusOK, gin.H{
		"provider": dto.ProviderResponse{
			ID:          provider.ID,
			Type:        string(provider.Type),
			Name:        provider.Name,
			DisplayName: provider.DisplayName,
			QueryModel:  provider.QueryModel,
			GenModel:    provider.GenModel,
			BaseURL:     provider.APIHost,
			Timeout:     formatTimeout(provider.TimeoutMs),
			Status:      status,
			Enabled:     provider.Enabled,
			IsSystem:    provider.IsSystem,
			IsDefault:   provider.IsDefault,
			Models:      provider.Models,
		},
	})
}

// CreateProviderRequest is the request body for creating a provider.
type CreateProviderRequest struct {
	Type        string                   `json:"type"`         // Provider type (openai, gemini, deepseek, custom, etc.)
	Name        string                   `json:"name"`         // Unique identifier (auto-generated if empty)
	DisplayName string                   `json:"display_name"` // Human-readable name
	APIHost     string                   `json:"api_host"`     // API base URL
	QueryModel  string                   `json:"query_model"`  // Model for retrieval/planning/critique
	GenModel    string                   `json:"gen_model"`    // Model for visualization generation
	TimeoutMs   int                      `json:"timeout_ms"`
	APIKey      string                   `json:"api_key"` // Initial API key
	Enabled     bool                     `json:"enabled"` // Whether to enable the provider
	Models      []domainconfig.ModelInfo `json:"models"`
}

// CreateProvider handles POST /api/v1/providers
func (h *ProviderHandler) CreateProvider(c *gin.Context) {
	var req CreateProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Determine provider type
	providerType := domainconfig.ProviderType(req.Type)
	if providerType == "" {
		providerType = domainconfig.ProviderTypeOpenAICompatible
	}

	// Apply preset defaults if type is a known preset
	preset := domainconfig.GetPresetByType(providerType)
	if preset != nil {
		if req.APIHost == "" {
			req.APIHost = preset.APIHost
		}
		if req.DisplayName == "" {
			req.DisplayName = preset.DisplayName
		}
		if req.QueryModel == "" && len(preset.DefaultModels) > 0 {
			req.QueryModel = preset.DefaultModels[0].ID
		}
		if req.GenModel == "" && len(preset.DefaultModels) > 0 {
			req.GenModel = preset.DefaultModels[0].ID
		}
	}

	// Generate name if not provided
	name := req.Name
	if name == "" {
		name = string(providerType)
		if providerType == domainconfig.ProviderTypeOpenAICompatible {
			name = "custom-" + uuid.New().String()[:8]
		}
	}

	provider := &domainconfig.Provider{
		ID:          uuid.New().String(),
		Type:        providerType,
		Name:        name,
		DisplayName: req.DisplayName,
		APIHost:     req.APIHost,
		QueryModel:  req.QueryModel,
		GenModel:    req.GenModel,
		TimeoutMs:   req.TimeoutMs,
		IsDefault:   false,
		Enabled:     req.Enabled,
		IsSystem:    false,
	}

	if provider.TimeoutMs == 0 {
		provider.TimeoutMs = 60000
	}
	if len(req.Models) > 0 {
		provider.Models = req.Models
	}
	if provider.QueryModel == "" {
		provider.QueryModel = firstEnabledModelID(provider.Models)
	}
	if provider.GenModel == "" {
		provider.GenModel = provider.QueryModel
	}

	// Copy models from preset if available
	if preset != nil && len(provider.Models) == 0 {
		provider.Models = make([]domainconfig.ModelInfo, len(preset.DefaultModels))
		for i, m := range preset.DefaultModels {
			provider.Models[i] = domainconfig.ModelInfo{
				ID:             m.ID,
				Name:           m.Name,
				MaxTokens:      m.MaxTokens,
				SupportsVision: m.SupportsVision,
				Enabled:        m.Enabled,
			}
		}
	}
	if provider.QueryModel == "" {
		provider.QueryModel = "default"
	}
	if provider.GenModel == "" {
		provider.GenModel = provider.QueryModel
	}

	if err := h.svc.CreateProvider(provider); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Add initial API key if provided
	if req.APIKey != "" {
		if _, err := h.svc.AddAPIKey(c.Request.Context(), provider.ID, req.APIKey); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "provider created but failed to add API key: " + err.Error()})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{"provider": provider})
}

// UpdateProviderRequest is the request body for updating a provider.
type UpdateProviderRequest struct {
	DisplayName string                   `json:"display_name"`
	APIHost     string                   `json:"api_host"`
	QueryModel  string                   `json:"query_model"`
	GenModel    string                   `json:"gen_model"`
	TimeoutMs   int                      `json:"timeout_ms"`
	Enabled     *bool                    `json:"enabled"`
	Models      []domainconfig.ModelInfo `json:"models"`
}

// UpdateProvider handles PUT /api/v1/providers/:id
func (h *ProviderHandler) UpdateProvider(c *gin.Context) {
	id := c.Param("id")

	var req UpdateProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	provider, err := h.svc.GetProvider(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "provider not found"})
		return
	}

	if req.DisplayName != "" {
		provider.DisplayName = req.DisplayName
	}
	if req.APIHost != "" {
		provider.APIHost = req.APIHost
	}
	if req.QueryModel != "" {
		provider.QueryModel = req.QueryModel
	}
	if req.GenModel != "" {
		provider.GenModel = req.GenModel
	}
	if req.TimeoutMs > 0 {
		provider.TimeoutMs = req.TimeoutMs
	}
	if req.Enabled != nil {
		provider.Enabled = *req.Enabled
	}
	if len(req.Models) > 0 {
		provider.Models = req.Models
	}
	if req.QueryModel == "" && len(req.Models) > 0 && !modelExists(provider.QueryModel, req.Models) {
		provider.QueryModel = firstEnabledModelID(req.Models)
	}
	if req.GenModel == "" && len(req.Models) > 0 && !modelExists(provider.GenModel, req.Models) {
		provider.GenModel = firstEnabledModelID(req.Models)
		if provider.GenModel == "" {
			provider.GenModel = provider.QueryModel
		}
	}

	if err := h.svc.UpdateProvider(provider); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"provider": provider})
}

// DeleteProvider handles DELETE /api/v1/providers/:id
func (h *ProviderHandler) DeleteProvider(c *gin.Context) {
	id := c.Param("id")

	if err := h.svc.DeleteProvider(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// SetDefaultProvider handles POST /api/v1/providers/:id/default
func (h *ProviderHandler) SetDefaultProvider(c *gin.Context) {
	id := c.Param("id")

	if err := h.svc.SetDefaultProvider(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// ListAPIKeys handles GET /api/v1/providers/:id/keys
func (h *ProviderHandler) ListAPIKeys(c *gin.Context) {
	providerID := c.Param("id")

	keys, err := h.svc.ListAPIKeys(providerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return masked keys only
	response := make([]gin.H, len(keys))
	for i, k := range keys {
		response[i] = gin.H{
			"id":         k.ID,
			"key_prefix": k.KeyPrefix,
			"key_suffix": k.KeySuffix,
			"masked":     k.MaskedKey(),
			"is_active":  k.IsActive,
			"last_used":  k.LastUsedAt,
			"created_at": k.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{"keys": response})
}

// AddAPIKeyRequest is the request body for adding an API key.
type AddAPIKeyRequest struct {
	APIKey string `json:"api_key" binding:"required"`
}

// AddAPIKey handles POST /api/v1/providers/:id/keys
func (h *ProviderHandler) AddAPIKey(c *gin.Context) {
	providerID := c.Param("id")

	var req AddAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	key, err := h.svc.AddAPIKey(c.Request.Context(), providerID, req.APIKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":        key.ID,
		"masked":    key.MaskedKey(),
		"is_active": key.IsActive,
	})
}

// DeleteAPIKey handles DELETE /api/v1/providers/:id/keys/:keyId
func (h *ProviderHandler) DeleteAPIKey(c *gin.Context) {
	keyID := c.Param("keyId")

	if err := h.svc.DeleteAPIKey(keyID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// ToggleAPIKeyRequest is the request body for toggling an API key.
type ToggleAPIKeyRequest struct {
	IsActive bool `json:"is_active"`
}

// ToggleAPIKey handles PATCH /api/v1/providers/:id/keys/:keyId
func (h *ProviderHandler) ToggleAPIKey(c *gin.Context) {
	keyID := c.Param("keyId")

	var req ToggleAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.ToggleAPIKey(keyID, req.IsActive); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func formatTimeout(ms int) string {
	if ms >= 1000 {
		return strconv.Itoa(ms/1000) + "s"
	}
	return strconv.Itoa(ms) + "ms"
}

func firstEnabledModelID(models []domainconfig.ModelInfo) string {
	for _, model := range models {
		if model.Enabled && model.ID != "" {
			return model.ID
		}
	}
	for _, model := range models {
		if model.ID != "" {
			return model.ID
		}
	}
	return ""
}

func modelExists(modelID string, models []domainconfig.ModelInfo) bool {
	if modelID == "" {
		return false
	}
	for _, model := range models {
		if model.ID == modelID {
			return true
		}
	}
	return false
}

// TestProviderRequest is the request body for testing a provider.
type TestProviderRequest struct {
	Provider string `json:"provider" binding:"required"`
	APIKey   string `json:"api_key" binding:"required"`
	BaseURL  string `json:"base_url"`
}

// TestProvider handles POST /api/v1/providers/test
func (h *ProviderHandler) TestProvider(c *gin.Context) {
	var req TestProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := h.svc.ValidateProvider(c.Request.Context(), req.Provider, req.APIKey, req.BaseURL)

	if result.Valid {
		c.JSON(http.StatusOK, result)
	} else {
		c.JSON(http.StatusBadRequest, result)
	}
}

// TestExistingProvider handles POST /api/v1/providers/:id/test
func (h *ProviderHandler) TestExistingProvider(c *gin.Context) {
	id := c.Param("id")

	provider, err := h.svc.GetProvider(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "provider not found"})
		return
	}

	// Get a decrypted API key
	keys, err := h.svc.ListAPIKeys(provider.ID)
	if err != nil || len(keys) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no API keys configured"})
		return
	}

	decrypted, err := h.svc.GetDecryptedKey(c.Request.Context(), keys[0].ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decrypt API key"})
		return
	}

	result := h.svc.ValidateProvider(c.Request.Context(), provider.Name, decrypted, provider.APIHost)

	if result.Valid {
		c.JSON(http.StatusOK, result)
	} else {
		c.JSON(http.StatusBadRequest, result)
	}
}

// ListModels handles GET /api/v1/providers/:id/models
func (h *ProviderHandler) ListModels(c *gin.Context) {
	providerID := c.Param("id")

	models, err := h.svc.ListModels(c.Request.Context(), providerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"models": models})
}

// ResetSystemProvidersRequest is the request body for resetting system providers.
type ResetSystemProvidersRequest struct {
	Confirm string `json:"confirm" binding:"required"`
}

// ResetSystemProviders handles POST /api/v1/providers/reset
// Clears all API keys for system providers. Requires explicit confirmation.
func (h *ProviderHandler) ResetSystemProviders(c *gin.Context) {
	var req ResetSystemProvidersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "confirmation required"})
		return
	}

	// Require exact "RESET" string for safety
	if req.Confirm != "RESET" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "confirmation must be exactly 'RESET' (uppercase)",
		})
		return
	}

	// Clear all API keys for system providers
	cleared, err := h.svc.ClearAllAPIKeysForSystemProviders(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "System providers reset successfully",
		"keys_cleared": cleared,
	})
}
