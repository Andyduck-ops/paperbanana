package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/paperbanana/paperbanana/internal/application/persistence"
	domainworkspace "github.com/paperbanana/paperbanana/internal/domain/workspace"
	"go.uber.org/zap"
)

// WorkspaceHandler handles workspace-related HTTP requests.
type WorkspaceHandler struct {
	workspaceService WorkspaceService
	logger           *zap.Logger
}

// WorkspaceService provides the application logic for workspace operations.
type WorkspaceService interface {
	CreateProject(ctx context.Context, name, description string) (*domainworkspace.Project, error)
	GetProject(ctx context.Context, id string) (*domainworkspace.Project, error)
	ListProjects(ctx context.Context) ([]*domainworkspace.Project, error)
	CreateFolder(ctx context.Context, projectID string, parentID *string, name string) (*domainworkspace.Folder, error)
	CreateVisualization(ctx context.Context, projectID string, folderID *string, name string) (*domainworkspace.Visualization, error)
	MoveFolder(ctx context.Context, projectID, folderID string, newParentID *string) error
	MoveVisualization(ctx context.Context, projectID, vizID string, newFolderID *string) error
	TrashFolder(ctx context.Context, projectID, folderID string) error
	TrashVisualization(ctx context.Context, projectID, vizID string) error
	TrashProject(ctx context.Context, projectID string) error
	RestoreFolder(ctx context.Context, projectID, folderID string) error
	RestoreVisualization(ctx context.Context, projectID, vizID string) error
	ListFolderContents(ctx context.Context, projectID string, folderID *string) (*persistence.FolderContents, error)
}

// FolderContents is an alias for the persistence.FolderContents type.
// Deprecated: Use persistence.FolderContents directly.
type FolderContents = persistence.FolderContents

// NewWorkspaceHandler creates a new WorkspaceHandler.
func NewWorkspaceHandler(workspaceService WorkspaceService, logger *zap.Logger) *WorkspaceHandler {
	return &WorkspaceHandler{
		workspaceService: workspaceService,
		logger:           logger,
	}
}

// CreateProjectRequest represents the request for creating a project.
type CreateProjectRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// ProjectResponse represents a project in the response.
type ProjectResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// CreateProject creates a new project.
// POST /api/v1/projects
func (h *WorkspaceHandler) CreateProject(c *gin.Context) {
	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project, err := h.workspaceService.CreateProject(c.Request.Context(), req.Name, req.Description)
	if err != nil {
		h.logger.Error("failed to create project", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create project"})
		return
	}

	c.JSON(http.StatusCreated, ProjectResponse{
		ID:          project.ID,
		Name:        project.Name,
		Description: project.Description,
		CreatedAt:   project.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   project.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// ListProjects lists all projects.
// GET /api/v1/projects
func (h *WorkspaceHandler) ListProjects(c *gin.Context) {
	projects, err := h.workspaceService.ListProjects(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to list projects", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list projects"})
		return
	}

	response := make([]ProjectResponse, len(projects))
	for i, p := range projects {
		response[i] = ProjectResponse{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			CreatedAt:   p.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	c.JSON(http.StatusOK, gin.H{"projects": response})
}

// GetProject retrieves a project by ID.
// GET /api/v1/projects/:project_id
func (h *WorkspaceHandler) GetProject(c *gin.Context) {
	projectID := c.Param("project_id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id is required"})
		return
	}

	project, err := h.workspaceService.GetProject(c.Request.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get project", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	c.JSON(http.StatusOK, ProjectResponse{
		ID:          project.ID,
		Name:        project.Name,
		Description: project.Description,
		CreatedAt:   project.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   project.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// CreateFolderRequest represents the request for creating a folder.
type CreateFolderRequest struct {
	ProjectID string  `json:"project_id" binding:"required"`
	ParentID  *string `json:"parent_id"`
	Name      string  `json:"name" binding:"required"`
}

// FolderResponse represents a folder in the response.
type FolderResponse struct {
	ID        string  `json:"id"`
	ProjectID string  `json:"project_id"`
	ParentID  *string `json:"parent_id,omitempty"`
	Name      string  `json:"name"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

// CreateFolder creates a new folder within a project.
// POST /api/v1/folders
func (h *WorkspaceHandler) CreateFolder(c *gin.Context) {
	var req CreateFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	folder, err := h.workspaceService.CreateFolder(c.Request.Context(), req.ProjectID, req.ParentID, req.Name)
	if err != nil {
		h.logger.Error("failed to create folder", zap.Error(err), zap.String("project_id", req.ProjectID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create folder"})
		return
	}

	c.JSON(http.StatusCreated, FolderResponse{
		ID:        folder.ID,
		ProjectID: folder.ProjectID,
		ParentID:  folder.ParentID,
		Name:      folder.Name,
		CreatedAt: folder.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: folder.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// CreateVisualizationRequest represents the request for creating a visualization.
type CreateVisualizationRequest struct {
	ProjectID string  `json:"project_id" binding:"required"`
	FolderID  *string `json:"folder_id"`
	Name      string  `json:"name" binding:"required"`
}

// VisualizationResponse represents a visualization in the response.
type VisualizationResponse struct {
	ID               string  `json:"id"`
	ProjectID        string  `json:"project_id"`
	FolderID         *string `json:"folder_id,omitempty"`
	Name             string  `json:"name"`
	CurrentVersionID *string `json:"current_version_id,omitempty"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

// CreateVisualization creates a new visualization within a project.
// POST /api/v1/visualizations
func (h *WorkspaceHandler) CreateVisualization(c *gin.Context) {
	var req CreateVisualizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	viz, err := h.workspaceService.CreateVisualization(c.Request.Context(), req.ProjectID, req.FolderID, req.Name)
	if err != nil {
		h.logger.Error("failed to create visualization", zap.Error(err), zap.String("project_id", req.ProjectID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create visualization"})
		return
	}

	c.JSON(http.StatusCreated, VisualizationResponse{
		ID:        viz.ID,
		ProjectID: viz.ProjectID,
		FolderID:  viz.FolderID,
		Name:      viz.Name,
		CreatedAt: viz.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: viz.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// ListFolderContentsRequest represents the request for listing folder contents.
type ListFolderContentsRequest struct {
	ProjectID string  `form:"project_id" binding:"required"`
	FolderID  *string `form:"folder_id"`
}

// ListItemType represents the type of item in a folder listing.
type ListItemType string

const (
	ItemTypeFolder        ListItemType = "folder"
	ItemTypeVisualization ListItemType = "visualization"
)

// ListItemResponse represents an item in a folder listing.
type ListItemResponse struct {
	ID        string       `json:"id"`
	Type      ListItemType `json:"type"`
	Name      string       `json:"name"`
	CreatedAt string       `json:"created_at"`
	UpdatedAt string       `json:"updated_at"`
}

// ListFolderContentsResponse represents the response for folder contents.
type ListFolderContentsResponse struct {
	ProjectID string             `json:"project_id"`
	FolderID  *string            `json:"folder_id,omitempty"`
	Items     []ListItemResponse `json:"items"`
}

// ListFolderContents lists the contents of a folder (mixed folders and visualizations).
// GET /api/v1/folders/contents
func (h *WorkspaceHandler) ListFolderContents(c *gin.Context) {
	var req ListFolderContentsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	contents, err := h.workspaceService.ListFolderContents(c.Request.Context(), req.ProjectID, req.FolderID)
	if err != nil {
		h.logger.Error("failed to list folder contents", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list folder contents"})
		return
	}

	// Combine folders and visualizations into a mixed list
	items := make([]ListItemResponse, 0, len(contents.Folders)+len(contents.Visualizations))

	for _, f := range contents.Folders {
		items = append(items, ListItemResponse{
			ID:        f.ID,
			Type:      ItemTypeFolder,
			Name:      f.Name,
			CreatedAt: f.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: f.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	for _, v := range contents.Visualizations {
		items = append(items, ListItemResponse{
			ID:        v.ID,
			Type:      ItemTypeVisualization,
			Name:      v.Name,
			CreatedAt: v.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: v.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	c.JSON(http.StatusOK, ListFolderContentsResponse{
		ProjectID: req.ProjectID,
		FolderID:  req.FolderID,
		Items:     items,
	})
}

// MoveItemRequest represents the request for moving an item.
type MoveItemRequest struct {
	ProjectID   string  `json:"project_id" binding:"required"`
	ItemType    string  `json:"item_type" binding:"required,oneof=folder visualization"`
	ItemID      string  `json:"item_id" binding:"required"`
	NewParentID *string `json:"new_parent_id"` // For folders, this is new parent folder; for visualizations, this is new folder
}

// MoveItem moves a folder or visualization to a new parent within the same project.
// POST /api/v1/workspace/move
func (h *WorkspaceHandler) MoveItem(c *gin.Context) {
	var req MoveItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var err error
	switch req.ItemType {
	case "folder":
		err = h.workspaceService.MoveFolder(c.Request.Context(), req.ProjectID, req.ItemID, req.NewParentID)
	case "visualization":
		err = h.workspaceService.MoveVisualization(c.Request.Context(), req.ProjectID, req.ItemID, req.NewParentID)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item_type"})
		return
	}

	if err != nil {
		h.logger.Error("failed to move item",
			zap.Error(err),
			zap.String("project_id", req.ProjectID),
			zap.String("item_type", req.ItemType),
			zap.String("item_id", req.ItemID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to move item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// TrashItemRequest represents the request for trashing an item.
type TrashItemRequest struct {
	ProjectID string `json:"project_id" binding:"required"`
	ItemType  string `json:"item_type" binding:"required,oneof=project folder visualization"`
	ItemID    string `json:"item_id" binding:"required"`
}

// TrashItem soft-deletes a project, folder, or visualization.
// POST /api/v1/workspace/trash
func (h *WorkspaceHandler) TrashItem(c *gin.Context) {
	var req TrashItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var err error
	switch req.ItemType {
	case "project":
		err = h.workspaceService.TrashProject(c.Request.Context(), req.ItemID)
	case "folder":
		err = h.workspaceService.TrashFolder(c.Request.Context(), req.ProjectID, req.ItemID)
	case "visualization":
		err = h.workspaceService.TrashVisualization(c.Request.Context(), req.ProjectID, req.ItemID)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item_type"})
		return
	}

	if err != nil {
		h.logger.Error("failed to trash item",
			zap.Error(err),
			zap.String("project_id", req.ProjectID),
			zap.String("item_type", req.ItemType),
			zap.String("item_id", req.ItemID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to trash item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// RestoreItemRequest represents the request for restoring an item.
type RestoreItemRequest struct {
	ProjectID string `json:"project_id" binding:"required"`
	ItemType  string `json:"item_type" binding:"required,oneof=folder visualization"`
	ItemID    string `json:"item_id" binding:"required"`
}

// RestoreItem restores a soft-deleted folder or visualization.
// POST /api/v1/workspace/restore
func (h *WorkspaceHandler) RestoreItem(c *gin.Context) {
	var req RestoreItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var err error
	switch req.ItemType {
	case "folder":
		err = h.workspaceService.RestoreFolder(c.Request.Context(), req.ProjectID, req.ItemID)
	case "visualization":
		err = h.workspaceService.RestoreVisualization(c.Request.Context(), req.ProjectID, req.ItemID)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item_type"})
		return
	}

	if err != nil {
		h.logger.Error("failed to restore item",
			zap.Error(err),
			zap.String("project_id", req.ProjectID),
			zap.String("item_type", req.ItemType),
			zap.String("item_id", req.ItemID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to restore item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// isNotFoundError checks if an error indicates a not found condition.
func isWorkspaceNotFoundError(err error) bool {
	return errors.Is(err, errors.New("not found")) ||
		err.Error() == "project not found" ||
		err.Error() == "folder not found" ||
		err.Error() == "visualization not found"
}
