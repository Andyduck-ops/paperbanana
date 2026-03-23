package sqlite

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/paperbanana/paperbanana/internal/domain/workspace"
	"gorm.io/gorm"
)

// ProjectRepository implements workspace.ProjectRepository using GORM.
type ProjectRepository struct {
	db *gorm.DB
}

// NewProjectRepository creates a new ProjectRepository.
func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

// Create inserts a new project.
func (r *ProjectRepository) Create(ctx context.Context, project *workspace.Project) error {
	model := &ProjectModel{
		ID:          project.ID,
		Name:        project.Name,
		Description: project.Description,
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(model).Error
}

// GetByID retrieves a project by ID.
func (r *ProjectRepository) GetByID(ctx context.Context, id string) (*workspace.Project, error) {
	var model ProjectModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("project not found: %s", id)
		}
		return nil, err
	}
	return projectModelToDomain(&model), nil
}

// List retrieves all projects.
func (r *ProjectRepository) List(ctx context.Context) ([]*workspace.Project, error) {
	var models []ProjectModel
	err := r.db.WithContext(ctx).Find(&models).Error
	if err != nil {
		return nil, err
	}
	projects := make([]*workspace.Project, len(models))
	for i, m := range models {
		projects[i] = projectModelToDomain(&m)
	}
	return projects, nil
}

// Update modifies an existing project.
func (r *ProjectRepository) Update(ctx context.Context, project *workspace.Project) error {
	return r.db.WithContext(ctx).Model(&ProjectModel{}).
		Where("id = ?", project.ID).
		Updates(map[string]interface{}{
			"name":        project.Name,
			"description": project.Description,
			"updated_at":  time.Now().UTC(),
		}).Error
}

// Delete soft-deletes a project.
func (r *ProjectRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&ProjectModel{}).Error
}

func projectModelToDomain(m *ProjectModel) *workspace.Project {
	return &workspace.Project{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// FolderRepository implements workspace.FolderRepository using GORM.
type FolderRepository struct {
	db *gorm.DB
}

// NewFolderRepository creates a new FolderRepository.
func NewFolderRepository(db *gorm.DB) *FolderRepository {
	return &FolderRepository{db: db}
}

// Create inserts a new folder.
func (r *FolderRepository) Create(ctx context.Context, folder *workspace.Folder) error {
	model := &FolderModel{
		ID:        folder.ID,
		ProjectID: folder.ProjectID,
		ParentID:  folder.ParentID,
		Name:      folder.Name,
		CreatedAt: folder.CreatedAt,
		UpdatedAt: folder.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(model).Error
}

// GetByID retrieves a folder by ID within a project.
func (r *FolderRepository) GetByID(ctx context.Context, projectID, id string) (*workspace.Folder, error) {
	var model FolderModel
	err := r.db.WithContext(ctx).
		Where("id = ? AND project_id = ?", id, projectID).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("folder not found: %s", id)
		}
		return nil, err
	}
	return folderModelToDomain(&model), nil
}

// ListByParent retrieves folders with a specific parent (nil for root).
func (r *FolderRepository) ListByParent(ctx context.Context, projectID string, parentID *string) ([]*workspace.Folder, error) {
	var models []FolderModel
	query := r.db.WithContext(ctx).Where("project_id = ?", projectID)
	if parentID == nil {
		query = query.Where("parent_id IS NULL")
	} else {
		query = query.Where("parent_id = ?", *parentID)
	}
	err := query.Order("name ASC").Find(&models).Error
	if err != nil {
		return nil, err
	}
	folders := make([]*workspace.Folder, len(models))
	for i, m := range models {
		folders[i] = folderModelToDomain(&m)
	}
	return folders, nil
}

// ListByProject retrieves all folders in a project.
func (r *FolderRepository) ListByProject(ctx context.Context, projectID string) ([]*workspace.Folder, error) {
	var models []FolderModel
	err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("name ASC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	folders := make([]*workspace.Folder, len(models))
	for i, m := range models {
		folders[i] = folderModelToDomain(&m)
	}
	return folders, nil
}

// Update modifies an existing folder.
func (r *FolderRepository) Update(ctx context.Context, folder *workspace.Folder) error {
	return r.db.WithContext(ctx).Model(&FolderModel{}).
		Where("id = ? AND project_id = ?", folder.ID, folder.ProjectID).
		Updates(map[string]interface{}{
			"name":       folder.Name,
			"parent_id":  folder.ParentID,
			"updated_at": time.Now().UTC(),
		}).Error
}

// Delete soft-deletes a folder.
func (r *FolderRepository) Delete(ctx context.Context, projectID, id string) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND project_id = ?", id, projectID).
		Delete(&FolderModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("folder not found: %s", id)
	}
	return nil
}

// GetDescendantIDs returns all folder IDs in the subtree rooted at the given folder.
// Uses recursive CTE for efficient tree traversal.
func (r *FolderRepository) GetDescendantIDs(ctx context.Context, projectID, folderID string) ([]string, error) {
	// SQLite recursive CTE to get all descendants
	query := `
		WITH RECURSIVE subtree AS (
			SELECT id, parent_id
			FROM folders
			WHERE id = ? AND project_id = ? AND deleted_at IS NULL
			UNION ALL
			SELECT f.id, f.parent_id
			FROM folders f
			INNER JOIN subtree s ON f.parent_id = s.id
			WHERE f.project_id = ? AND f.deleted_at IS NULL
		)
		SELECT id FROM subtree
	`

	var ids []string
	err := r.db.WithContext(ctx).Raw(query, folderID, projectID, projectID).Scan(&ids).Error
	if err != nil {
		return nil, err
	}
	return ids, nil
}

// Restore restores a soft-deleted folder.
func (r *FolderRepository) Restore(ctx context.Context, projectID, id string) error {
	result := r.db.WithContext(ctx).Unscoped().
		Model(&FolderModel{}).
		Where("id = ? AND project_id = ?", id, projectID).
		Update("deleted_at", nil)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("folder not found: %s", id)
	}
	return nil
}

func folderModelToDomain(m *FolderModel) *workspace.Folder {
	var deletedAt *time.Time
	if m.DeletedAt.Valid {
		deletedAt = &m.DeletedAt.Time
	}
	return &workspace.Folder{
		ID:        m.ID,
		ProjectID: m.ProjectID,
		ParentID:  m.ParentID,
		Name:      m.Name,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		DeletedAt: deletedAt,
	}
}

// VisualizationRepository implements workspace.VisualizationRepository using GORM.
type VisualizationRepository struct {
	db *gorm.DB
}

// NewVisualizationRepository creates a new VisualizationRepository.
func NewVisualizationRepository(db *gorm.DB) *VisualizationRepository {
	return &VisualizationRepository{db: db}
}

// Create inserts a new visualization.
func (r *VisualizationRepository) Create(ctx context.Context, viz *workspace.Visualization) error {
	model := &VisualizationModel{
		ID:        viz.ID,
		ProjectID: viz.ProjectID,
		FolderID:  viz.FolderID,
		Name:      viz.Name,
		CreatedAt: viz.CreatedAt,
		UpdatedAt: viz.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(model).Error
}

// GetByID retrieves a visualization by ID within a project.
func (r *VisualizationRepository) GetByID(ctx context.Context, projectID, id string) (*workspace.Visualization, error) {
	var model VisualizationModel
	err := r.db.WithContext(ctx).
		Where("id = ? AND project_id = ?", id, projectID).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("visualization not found: %s", id)
		}
		return nil, err
	}
	return vizModelToDomain(&model), nil
}

// ListByFolder retrieves visualizations in a specific folder (nil for root).
func (r *VisualizationRepository) ListByFolder(ctx context.Context, projectID string, folderID *string) ([]*workspace.Visualization, error) {
	var models []VisualizationModel
	query := r.db.WithContext(ctx).Where("project_id = ?", projectID)
	if folderID == nil {
		query = query.Where("folder_id IS NULL")
	} else {
		query = query.Where("folder_id = ?", *folderID)
	}
	err := query.Order("name ASC").Find(&models).Error
	if err != nil {
		return nil, err
	}
	vizs := make([]*workspace.Visualization, len(models))
	for i, m := range models {
		vizs[i] = vizModelToDomain(&m)
	}
	return vizs, nil
}

// ListByProject retrieves all visualizations in a project.
func (r *VisualizationRepository) ListByProject(ctx context.Context, projectID string) ([]*workspace.Visualization, error) {
	var models []VisualizationModel
	err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("name ASC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	vizs := make([]*workspace.Visualization, len(models))
	for i, m := range models {
		vizs[i] = vizModelToDomain(&m)
	}
	return vizs, nil
}

// Update modifies an existing visualization.
func (r *VisualizationRepository) Update(ctx context.Context, viz *workspace.Visualization) error {
	return r.db.WithContext(ctx).Model(&VisualizationModel{}).
		Where("id = ? AND project_id = ?", viz.ID, viz.ProjectID).
		Updates(map[string]interface{}{
			"name":       viz.Name,
			"folder_id":  viz.FolderID,
			"updated_at": time.Now().UTC(),
		}).Error
}

// Delete soft-deletes a visualization.
func (r *VisualizationRepository) Delete(ctx context.Context, projectID, id string) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND project_id = ?", id, projectID).
		Delete(&VisualizationModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("visualization not found: %s", id)
	}
	return nil
}

// SetCurrentVersion updates the current version pointer for a visualization.
func (r *VisualizationRepository) SetCurrentVersion(ctx context.Context, projectID, visualizationID, versionID string) error {
	return r.db.WithContext(ctx).Model(&VisualizationModel{}).
		Where("id = ? AND project_id = ?", visualizationID, projectID).
		Update("current_version_id", versionID).Error
}

// Restore restores a soft-deleted visualization.
func (r *VisualizationRepository) Restore(ctx context.Context, projectID, id string) error {
	result := r.db.WithContext(ctx).Unscoped().
		Model(&VisualizationModel{}).
		Where("id = ? AND project_id = ?", id, projectID).
		Update("deleted_at", nil)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("visualization not found: %s", id)
	}
	return nil
}

func vizModelToDomain(m *VisualizationModel) *workspace.Visualization {
	var deletedAt *time.Time
	if m.DeletedAt.Valid {
		deletedAt = &m.DeletedAt.Time
	}
	return &workspace.Visualization{
		ID:               m.ID,
		ProjectID:        m.ProjectID,
		FolderID:         m.FolderID,
		Name:             m.Name,
		CurrentVersionID: m.CurrentVersionID,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
		DeletedAt:        deletedAt,
	}
}
