package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
	domainworkspace "github.com/paperbanana/paperbanana/internal/domain/workspace"
	"go.uber.org/zap"
)

// HistoryHandler handles history and session-related HTTP requests.
type HistoryHandler struct {
	historyService HistoryService
	logger         *zap.Logger
}

// HistoryService provides the application logic for history operations.
type HistoryService interface {
	ListHistory(ctx context.Context, projectID, visualizationID string, limit int) ([]*domainworkspace.VisualizationVersion, error)
	GetVersion(ctx context.Context, projectID, versionID string) (*domainworkspace.VisualizationVersion, error)
	GetLatestSession(ctx context.Context, projectID, visualizationID string) (*domainworkspace.SessionRecord, error)
	GetSessionByID(ctx context.Context, sessionID string) (*domainworkspace.SessionRecord, error)
}

// NewHistoryHandler creates a new HistoryHandler.
func NewHistoryHandler(historyService HistoryService, logger *zap.Logger) *HistoryHandler {
	return &HistoryHandler{
		historyService: historyService,
		logger:         logger,
	}
}

// ListHistoryRequest represents the request for listing version history.
type ListHistoryRequest struct {
	ProjectID       string `form:"project_id"`
	VisualizationID string `form:"visualization_id"`
	Limit           int    `form:"limit"`
}

// VersionResponse represents a single version in the response.
type VersionResponse struct {
	ID              string                      `json:"id"`
	VisualizationID string                      `json:"visualization_id"`
	VersionNumber   int                         `json:"version_number"`
	Summary         string                      `json:"summary"`
	CreatedAt       string                      `json:"created_at"`
	Artifacts       []VersionArtifactResponse   `json:"artifacts,omitempty"`
}

// VersionArtifactResponse represents an artifact attached to a version.
type VersionArtifactResponse struct {
	ID           string `json:"id"`
	Kind         string `json:"kind"`
	MIMEType     string `json:"mime_type"`
	StorageKey   string `json:"storage_key"`
	ByteSize     int64  `json:"byte_size"`
}

// ListHistoryResponse represents the response for listing version history.
type ListHistoryResponse struct {
	ProjectID string             `json:"project_id"`
	Versions  []VersionResponse  `json:"versions"`
}

// ListHistory lists the version history for a visualization.
// GET /api/v1/history?project_id=xxx&visualization_id=yyy
func (h *HistoryHandler) ListHistory(c *gin.Context) {
	var req ListHistoryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	// If project_id or visualization_id is not provided, return empty list (graceful degradation)
	if req.ProjectID == "" || req.VisualizationID == "" {
		c.JSON(http.StatusOK, ListHistoryResponse{
			ProjectID: req.ProjectID,
			Versions:  []VersionResponse{},
		})
		return
	}

	versions, err := h.historyService.ListHistory(c.Request.Context(), req.ProjectID, req.VisualizationID, req.Limit)
	if err != nil {
		h.logger.Error("failed to list history", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list history"})
		return
	}

	response := ListHistoryResponse{
		ProjectID: req.ProjectID,
		Versions:  make([]VersionResponse, len(versions)),
	}

	for i, v := range versions {
		response.Versions[i] = VersionResponse{
			ID:              v.ID,
			VisualizationID: v.VisualizationID,
			VersionNumber:   v.VersionNumber,
			Summary:         v.Summary,
			CreatedAt:       v.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
		for _, a := range v.Artifacts {
			response.Versions[i].Artifacts = append(response.Versions[i].Artifacts, VersionArtifactResponse{
				ID:         a.ID,
				Kind:       a.Kind,
				MIMEType:   a.MIMEType,
				StorageKey: a.StorageKey,
				ByteSize:   a.ByteSize,
			})
		}
	}

	c.JSON(http.StatusOK, response)
}

// GetVersionRequest represents the request for getting a specific version.
type GetVersionRequest struct {
	ProjectID string `uri:"project_id" binding:"required"`
	VersionID string `uri:"version_id" binding:"required"`
}

// GetVersionResponse represents the response for a single version.
type GetVersionResponse struct {
	ID              string                       `json:"id"`
	VisualizationID string                       `json:"visualization_id"`
	ProjectID       string                       `json:"project_id"`
	VersionNumber   int                          `json:"version_number"`
	Summary         string                       `json:"summary"`
	CreatedAt       string                       `json:"created_at"`
	Artifacts       []VersionArtifactResponse    `json:"artifacts,omitempty"`
	SessionSnapshot *domainagent.SessionState    `json:"session_snapshot,omitempty"`
}

// GetVersion retrieves a specific version by ID.
// GET /api/v1/history/:project_id/:version_id
func (h *HistoryHandler) GetVersion(c *gin.Context) {
	projectID := c.Param("project_id")
	versionID := c.Param("version_id")

	if projectID == "" || versionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id and version_id are required"})
		return
	}

	version, err := h.historyService.GetVersion(c.Request.Context(), projectID, versionID)
	if err != nil {
		h.logger.Error("failed to get version", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "version not found"})
		return
	}

	response := GetVersionResponse{
		ID:              version.ID,
		VisualizationID: version.VisualizationID,
		ProjectID:       version.ProjectID,
		VersionNumber:   version.VersionNumber,
		Summary:         version.Summary,
		CreatedAt:       version.CreatedAt.Format("2006-01-02T15:04:05Z"),
		SessionSnapshot: version.SessionSnapshot,
	}

	for _, a := range version.Artifacts {
		response.Artifacts = append(response.Artifacts, VersionArtifactResponse{
			ID:         a.ID,
			Kind:       a.Kind,
			MIMEType:   a.MIMEType,
			StorageKey: a.StorageKey,
			ByteSize:   a.ByteSize,
		})
	}

	c.JSON(http.StatusOK, response)
}

// GetLatestSessionRequest represents the request for getting the latest session.
type GetLatestSessionRequest struct {
	ProjectID       string `form:"project_id" binding:"required"`
	VisualizationID string `form:"visualization_id" binding:"required"`
}

// SessionResponse represents a session in the response.
type SessionResponse struct {
	ID              string                    `json:"id"`
	ProjectID       string                    `json:"project_id"`
	VisualizationID *string                   `json:"visualization_id,omitempty"`
	Status          string                    `json:"status"`
	CurrentStage    string                    `json:"current_stage"`
	SchemaVersion   string                    `json:"schema_version"`
	CreatedAt       string                    `json:"created_at"`
	UpdatedAt       string                    `json:"updated_at"`
	Snapshot        *domainagent.SessionState `json:"snapshot,omitempty"`
}

// GetLatestSession retrieves the latest session for a visualization.
// GET /api/v1/session/latest?project_id=xxx&visualization_id=yyy
func (h *HistoryHandler) GetLatestSession(c *gin.Context) {
	var req GetLatestSessionRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session, err := h.historyService.GetLatestSession(c.Request.Context(), req.ProjectID, req.VisualizationID)
	if err != nil {
		h.logger.Error("failed to get latest session", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	response := SessionResponse{
		ID:            session.ID,
		ProjectID:     session.ProjectID,
		Status:        session.Status,
		CurrentStage:  session.CurrentStage,
		SchemaVersion: session.SchemaVersion,
		CreatedAt:     session.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     session.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		Snapshot:      session.Snapshot,
	}

	if session.VisualizationID != nil {
		response.VisualizationID = session.VisualizationID
	}

	c.JSON(http.StatusOK, response)
}

// GetSessionRequest represents the request for getting a session by ID.
type GetSessionRequest struct {
	SessionID string `uri:"session_id" binding:"required"`
}

// GetSession retrieves a session by ID.
// GET /api/v1/session/:session_id
func (h *HistoryHandler) GetSession(c *gin.Context) {
	sessionID := c.Param("session_id")

	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})
		return
	}

	session, err := h.historyService.GetSessionByID(c.Request.Context(), sessionID)
	if err != nil {
		h.logger.Error("failed to get session", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	response := SessionResponse{
		ID:            session.ID,
		ProjectID:     session.ProjectID,
		Status:        session.Status,
		CurrentStage:  session.CurrentStage,
		SchemaVersion: session.SchemaVersion,
		CreatedAt:     session.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     session.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		Snapshot:      session.Snapshot,
	}

	if session.VisualizationID != nil {
		response.VisualizationID = session.VisualizationID
	}

	c.JSON(http.StatusOK, response)
}

// isNotFoundError checks if an error indicates a not found condition.
func isNotFoundError(err error) bool {
	return errors.Is(err, errors.New("not found")) ||
		errors.Is(err, errors.New("no resumable session found")) ||
		err.Error() == "not found" ||
		err.Error() == "no resumable session found"
}
