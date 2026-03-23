package handlers

import (
	"context"
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/paperbanana/paperbanana/internal/api/dto"
	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
	polishagent "github.com/paperbanana/paperbanana/internal/agents/polish"
	"go.uber.org/zap"
)

// ClientManagerInterface provides LLM client access for the refine handler.
type ClientManagerInterface interface {
	GetClient(ctx context.Context, providerID string) (domainllm.LLMClient, error)
}

// RefineHandler handles image refinement requests.
type RefineHandler struct {
	clientManager ClientManagerInterface
	logger        *zap.Logger
}

// NewRefineHandler creates a new refine handler.
func NewRefineHandler(clientManager ClientManagerInterface, logger *zap.Logger) *RefineHandler {
	return &RefineHandler{
		clientManager: clientManager,
		logger:        logger,
	}
}

// Refine handles POST /api/v1/refine for image enhancement.
func (h *RefineHandler) Refine(c *gin.Context) {
	var req dto.RefineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.RefineResponse{
			Error: err.Error(),
		})
		return
	}

	// Validate request
	if req.ImageData == "" {
		c.JSON(http.StatusBadRequest, dto.RefineResponse{
			Error: "image_data is required",
		})
		return
	}
	if req.Instructions == "" {
		c.JSON(http.StatusBadRequest, dto.RefineResponse{
			Error: "instructions is required",
		})
		return
	}

	// Set default resolution
	resolution := req.Resolution
	if resolution == "" {
		resolution = "2K"
	}
	if resolution != "2K" && resolution != "4K" {
		c.JSON(http.StatusBadRequest, dto.RefineResponse{
			Error: "resolution must be '2K' or '4K'",
		})
		return
	}

	// Decode base64 image data
	imageBytes, err := base64.StdEncoding.DecodeString(req.ImageData)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.RefineResponse{
			Error: "invalid base64 image data: " + err.Error(),
		})
		return
	}

	// Get LLM client
	providerID := req.ProviderID
	if providerID == "" {
		providerID = "default"
	}

	client, err := h.clientManager.GetClient(c.Request.Context(), providerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.RefineResponse{
			Error: "failed to get LLM client: " + err.Error(),
		})
		return
	}

	// Build agent input with image and instructions
	input := domainagent.AgentInput{
		SessionID: req.SessionID,
		Content:   req.Instructions,
		Messages: []domainllm.Message{
			{
				Role: domainllm.RoleUser,
				Parts: []domainllm.Part{
					// Detect MIME type from image bytes (simple detection)
					domainllm.InlineImagePart(detectMIMEType(imageBytes), imageBytes),
				},
			},
		},
	}

	// Create and execute PolishAgent
	config := polishagent.Config{
		Model:      req.Model,
		Resolution: resolution,
	}
	agent := polishagent.NewAgent(client, config, h.logger)

	if err := agent.Initialize(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, dto.RefineResponse{
			Error: "failed to initialize agent: " + err.Error(),
		})
		return
	}

	output, err := agent.Execute(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.RefineResponse{
			Error: "refinement failed: " + err.Error(),
		})
		return
	}

	_ = agent.Cleanup(c.Request.Context())

	c.JSON(http.StatusOK, dto.RefineResponse{
		SessionID: req.SessionID,
		Status:    "completed",
		Content:   output.Content,
	})
}

// detectMIMEType attempts to detect the MIME type from image bytes.
func detectMIMEType(data []byte) string {
	if len(data) < 4 {
		return "image/png" // default
	}

	// PNG: 89 50 4E 47
	if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return "image/png"
	}
	// JPEG: FF D8 FF
	if data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return "image/jpeg"
	}
	// GIF: 47 49 46 38
	if data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x38 {
		return "image/gif"
	}
	// WebP: 52 49 46 46 (RIFF) + later WEBP
	if len(data) >= 12 && data[0] == 0x52 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x46 {
		if data[8] == 0x57 && data[9] == 0x45 && data[10] == 0x42 && data[11] == 0x50 {
			return "image/webp"
		}
	}

	return "image/png" // default to PNG
}
