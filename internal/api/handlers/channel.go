package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	configservice "github.com/paperbanana/paperbanana/internal/application/config"
)

// ChannelHandler handles channel-related HTTP requests.
// Channels are an alias for Providers to align with the PRD terminology.
type ChannelHandler struct {
	providerHandler *ProviderHandler
}

// NewChannelHandler creates a new channel handler that delegates to provider handler.
func NewChannelHandler(svc *configservice.Service) *ChannelHandler {
	return &ChannelHandler{
		providerHandler: NewProviderHandler(svc),
	}
}

// ListChannels handles GET /api/v1/channels
// Returns all providers in channel format.
func (h *ChannelHandler) ListChannels(c *gin.Context) {
	// Delegate to provider handler
	h.providerHandler.ListProviders(c)
}

// GetChannel handles GET /api/v1/channels/:id
// Returns a single provider in channel format.
func (h *ChannelHandler) GetChannel(c *gin.Context) {
	h.providerHandler.GetProvider(c)
}

// CreateChannel handles POST /api/v1/channels
// Creates a new provider.
func (h *ChannelHandler) CreateChannel(c *gin.Context) {
	h.providerHandler.CreateProvider(c)
}

// UpdateChannel handles PUT /api/v1/channels/:id
// Updates a provider.
func (h *ChannelHandler) UpdateChannel(c *gin.Context) {
	h.providerHandler.UpdateProvider(c)
}

// DeleteChannel handles DELETE /api/v1/channels/:id
// Deletes a provider.
func (h *ChannelHandler) DeleteChannel(c *gin.Context) {
	h.providerHandler.DeleteProvider(c)
}

// FetchChannelModels handles GET /api/v1/channels/:id/models
// Fetches available models from a provider.
func (h *ChannelHandler) FetchChannelModels(c *gin.Context) {
	h.providerHandler.ListModels(c)
}

// TestChannel handles POST /api/v1/channels/test
// Tests a channel connection.
func (h *ChannelHandler) TestChannel(c *gin.Context) {
	h.providerHandler.TestProvider(c)
}

// ListPresets handles GET /api/v1/channels/presets
// Lists all available provider presets.
func (h *ChannelHandler) ListPresets(c *gin.Context) {
	h.providerHandler.ListPresets(c)
}

// ListChannelAPIKeys handles GET /api/v1/channels/:id/keys
// Lists API keys for a channel.
func (h *ChannelHandler) ListChannelAPIKeys(c *gin.Context) {
	h.providerHandler.ListAPIKeys(c)
}

// AddChannelAPIKey handles POST /api/v1/channels/:id/keys
// Adds an API key to a channel.
func (h *ChannelHandler) AddChannelAPIKey(c *gin.Context) {
	h.providerHandler.AddAPIKey(c)
}

// DeleteChannelAPIKey handles DELETE /api/v1/channels/:id/keys/:keyId
// Deletes an API key from a channel.
func (h *ChannelHandler) DeleteChannelAPIKey(c *gin.Context) {
	h.providerHandler.DeleteAPIKey(c)
}

// ChannelRoleRequest is the request body for setting role assignments.
type ChannelRoleRequest struct {
	Role      string `json:"role" binding:"required"`       // "image_generation" or "retrieval_reasoning"
	ChannelID string `json:"channel_id" binding:"required"`
	ModelID   string `json:"model_id" binding:"required"`
}

// SetRoleAssignment handles PUT /api/v1/config/roles
// Sets the model assignment for a workflow role.
func (h *ChannelHandler) SetRoleAssignment(c *gin.Context) {
	var req ChannelRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate role
	if req.Role != "image_generation" && req.Role != "retrieval_reasoning" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role, must be 'image_generation' or 'retrieval_reasoning'"})
		return
	}
	req.ChannelID = strings.TrimSpace(req.ChannelID)
	req.ModelID = strings.TrimSpace(req.ModelID)
	if req.ChannelID == "" || req.ModelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "channel_id and model_id are required"})
		return
	}

	if err := h.clearRoleAcrossProviders(req.Role, req.ChannelID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update the provider with the appropriate model assignment
	// For image_generation -> gen_model
	// For retrieval_reasoning -> query_model
	updateKey := "gen_model"
	if req.Role == "retrieval_reasoning" {
		updateKey = "query_model"
	}

	// Get the provider and update it
	provider, err := h.providerHandler.svc.GetProvider(req.ChannelID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
		return
	}

	// Update the appropriate model field
	if updateKey == "gen_model" {
		provider.GenModel = req.ModelID
	} else {
		provider.QueryModel = req.ModelID
	}

	if err := h.providerHandler.svc.UpdateProvider(provider); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"role":    req.Role,
		"channel_id": req.ChannelID,
		"model_id": req.ModelID,
	})
}

// ClearRoleAssignment handles DELETE /api/v1/config/roles/:role
// Removes the assigned model for a workflow role across all providers.
func (h *ChannelHandler) ClearRoleAssignment(c *gin.Context) {
	role := strings.TrimSpace(c.Param("role"))
	if role != "image_generation" && role != "retrieval_reasoning" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role, must be 'image_generation' or 'retrieval_reasoning'"})
		return
	}

	if err := h.clearRoleAcrossProviders(role, ""); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"role":    role,
	})
}

func (h *ChannelHandler) clearRoleAcrossProviders(role string, keepProviderID string) error {
	providers, err := h.providerHandler.svc.ListProviders()
	if err != nil {
		return err
	}

	for _, provider := range providers {
		if provider == nil || provider.ID == keepProviderID {
			continue
		}

		updated := false
		switch role {
		case "image_generation":
			if provider.GenModel != "" {
				provider.GenModel = ""
				updated = true
			}
		case "retrieval_reasoning":
			if provider.QueryModel != "" {
				provider.QueryModel = ""
				updated = true
			}
		}

		if updated {
			if err := h.providerHandler.svc.UpdateProvider(provider); err != nil {
				return err
			}
		}
	}

	return nil
}
