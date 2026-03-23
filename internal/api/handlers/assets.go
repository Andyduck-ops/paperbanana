package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AssetHandler handles asset-related HTTP requests.
// It provides access to asset metadata and controlled download access
// while hiding storage internals from API consumers.
type AssetHandler struct {
	assetService AssetService
	logger       *zap.Logger
}

// AssetService provides the application logic for asset operations.
type AssetService interface {
	// ListAssets retrieves all assets for a visualization within a project.
	ListAssets(ctx context.Context, projectID, visualizationID string) ([]*AssetInfo, error)

	// GetAsset retrieves an asset's metadata and bytes by ID, scoped to a project.
	GetAsset(ctx context.Context, projectID, assetID string) (*AssetInfo, []byte, error)

	// ListAssetsByVersion retrieves all assets for a specific version.
	ListAssetsByVersion(ctx context.Context, projectID, versionID string) ([]*AssetInfo, error)
}

// AssetInfo represents asset metadata for API responses.
// Does not include storage internals like storage keys.
type AssetInfo struct {
	ID              string    `json:"id"`
	ProjectID       string    `json:"project_id"`
	VisualizationID string    `json:"visualization_id"`
	VersionID       *string   `json:"version_id,omitempty"`
	MIMEType        string    `json:"mime_type"`
	ByteSize        int64     `json:"byte_size"`
	ChecksumSHA256  string    `json:"checksum_sha256"`
	CreatedAt       time.Time `json:"created_at"`
}

// NewAssetHandler creates a new AssetHandler.
func NewAssetHandler(assetService AssetService, logger *zap.Logger) *AssetHandler {
	return &AssetHandler{
		assetService: assetService,
		logger:       logger,
	}
}

// ListAssetsRequest represents the request for listing assets.
type ListAssetsRequest struct {
	ProjectID       string `form:"project_id" binding:"required"`
	VisualizationID string `form:"visualization_id" binding:"required"`
}

// AssetResponse represents a single asset in the response.
// Storage internals are intentionally not exposed.
type AssetResponse struct {
	ID              string    `json:"id"`
	ProjectID       string    `json:"project_id"`
	VisualizationID string    `json:"visualization_id"`
	VersionID       *string   `json:"version_id,omitempty"`
	MIMEType        string    `json:"mime_type"`
	ByteSize        int64     `json:"byte_size"`
	ChecksumSHA256  string    `json:"checksum_sha256"`
	CreatedAt       string    `json:"created_at"`
	// StorageKey is intentionally NOT included to hide storage internals
	StorageKey string `json:"-"` // Always empty in responses
}

// ListAssetsResponse represents the response for listing assets.
type ListAssetsResponse struct {
	ProjectID string          `json:"project_id"`
	Assets    []AssetResponse `json:"assets"`
}

// ListAssets lists all assets for a visualization within a project.
// GET /api/v1/assets?project_id=xxx&visualization_id=yyy
func (h *AssetHandler) ListAssets(c *gin.Context) {
	var req ListAssetsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id and visualization_id are required"})
		return
	}

	assets, err := h.assetService.ListAssets(c.Request.Context(), req.ProjectID, req.VisualizationID)
	if err != nil {
		h.logger.Error("failed to list assets",
			zap.String("project_id", req.ProjectID),
			zap.String("visualization_id", req.VisualizationID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list assets"})
		return
	}

	response := ListAssetsResponse{
		ProjectID: req.ProjectID,
		Assets:    make([]AssetResponse, len(assets)),
	}

	for i, a := range assets {
		response.Assets[i] = AssetResponse{
			ID:              a.ID,
			ProjectID:       a.ProjectID,
			VisualizationID: a.VisualizationID,
			VersionID:       a.VersionID,
			MIMEType:        a.MIMEType,
			ByteSize:        a.ByteSize,
			ChecksumSHA256:  a.ChecksumSHA256,
			CreatedAt:       a.CreatedAt.Format(time.RFC3339),
		}
	}

	c.JSON(http.StatusOK, response)
}

// GetAsset retrieves an asset's metadata by ID, scoped to a project.
// GET /api/v1/assets/:project_id/:asset_id
func (h *AssetHandler) GetAsset(c *gin.Context) {
	projectID := c.Param("project_id")
	assetID := c.Param("asset_id")

	if projectID == "" || assetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id and asset_id are required"})
		return
	}

	asset, _, err := h.assetService.GetAsset(c.Request.Context(), projectID, assetID)
	if err != nil {
		h.logger.Error("failed to get asset",
			zap.String("project_id", projectID),
			zap.String("asset_id", assetID),
			zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "asset not found"})
		return
	}

	response := AssetResponse{
		ID:              asset.ID,
		ProjectID:       asset.ProjectID,
		VisualizationID: asset.VisualizationID,
		VersionID:       asset.VersionID,
		MIMEType:        asset.MIMEType,
		ByteSize:        asset.ByteSize,
		ChecksumSHA256:  asset.ChecksumSHA256,
		CreatedAt:       asset.CreatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}

// DownloadAsset retrieves an asset's bytes for download, scoped to a project.
// The actual file bytes flow through the service boundary, not directly from the filesystem.
// GET /api/v1/assets/:project_id/:asset_id/download
func (h *AssetHandler) DownloadAsset(c *gin.Context) {
	projectID := c.Param("project_id")
	assetID := c.Param("asset_id")

	if projectID == "" || assetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id and asset_id are required"})
		return
	}

	asset, data, err := h.assetService.GetAsset(c.Request.Context(), projectID, assetID)
	if err != nil {
		h.logger.Error("failed to download asset",
			zap.String("project_id", projectID),
			zap.String("asset_id", assetID),
			zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "asset not found"})
		return
	}

	// Set response headers for download
	c.Header("Content-Type", asset.MIMEType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", assetID))
	c.Header("Content-Length", fmt.Sprintf("%d", len(data)))

	// Write bytes directly
	c.Data(http.StatusOK, asset.MIMEType, data)
}

// ListAssetsByVersionRequest represents the request for listing assets by version.
type ListAssetsByVersionRequest struct {
	ProjectID string `uri:"project_id" binding:"required"`
	VersionID string `uri:"version_id" binding:"required"`
}

// ListAssetsByVersion lists all assets for a specific version within a project.
// GET /api/v1/assets/version/:project_id/:version_id
func (h *AssetHandler) ListAssetsByVersion(c *gin.Context) {
	projectID := c.Param("project_id")
	versionID := c.Param("version_id")

	if projectID == "" || versionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id and version_id are required"})
		return
	}

	assets, err := h.assetService.ListAssetsByVersion(c.Request.Context(), projectID, versionID)
	if err != nil {
		h.logger.Error("failed to list assets by version",
			zap.String("project_id", projectID),
			zap.String("version_id", versionID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list assets"})
		return
	}

	response := ListAssetsResponse{
		ProjectID: projectID,
		Assets:    make([]AssetResponse, len(assets)),
	}

	for i, a := range assets {
		response.Assets[i] = AssetResponse{
			ID:              a.ID,
			ProjectID:       a.ProjectID,
			VisualizationID: a.VisualizationID,
			VersionID:       a.VersionID,
			MIMEType:        a.MIMEType,
			ByteSize:        a.ByteSize,
			ChecksumSHA256:  a.ChecksumSHA256,
			CreatedAt:       a.CreatedAt.Format(time.RFC3339),
		}
	}

	c.JSON(http.StatusOK, response)
}

// isAssetNotFoundError checks if an error indicates a not found condition.
func isAssetNotFoundError(err error) bool {
	return errors.Is(err, errors.New("asset not found")) ||
		err.Error() == "asset not found"
}
