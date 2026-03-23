package workspace

import (
	"context"
)

// ProjectRepository manages project aggregates.
// All methods enforce project-scoped access; ID-based lookups must match the project context.
type ProjectRepository interface {
	Create(ctx context.Context, project *Project) error
	GetByID(ctx context.Context, id string) (*Project, error)
	List(ctx context.Context) ([]*Project, error)
	Update(ctx context.Context, project *Project) error
	Delete(ctx context.Context, id string) error // Soft delete to trash
}

// FolderRepository manages folder hierarchies within projects.
// All queries are project-scoped; parent-child relationships are validated.
type FolderRepository interface {
	Create(ctx context.Context, folder *Folder) error
	GetByID(ctx context.Context, projectID, id string) (*Folder, error)
	ListByParent(ctx context.Context, projectID string, parentID *string) ([]*Folder, error)
	ListByProject(ctx context.Context, projectID string) ([]*Folder, error)
	Update(ctx context.Context, folder *Folder) error
	Delete(ctx context.Context, projectID, id string) error // Soft delete to trash
	// GetDescendantIDs returns all folder IDs in the subtree rooted at the given folder.
	// Uses recursive CTE for efficient tree traversal.
	GetDescendantIDs(ctx context.Context, projectID, folderID string) ([]string, error)
	// Restore restores a soft-deleted folder.
	Restore(ctx context.Context, projectID, id string) error
}

// VisualizationRepository manages visualization entries.
// All queries are project-scoped; folder assignment is optional (root level).
type VisualizationRepository interface {
	Create(ctx context.Context, viz *Visualization) error
	GetByID(ctx context.Context, projectID, id string) (*Visualization, error)
	ListByFolder(ctx context.Context, projectID string, folderID *string) ([]*Visualization, error)
	ListByProject(ctx context.Context, projectID string) ([]*Visualization, error)
	Update(ctx context.Context, viz *Visualization) error
	Delete(ctx context.Context, projectID, id string) error // Soft delete to trash
	Restore(ctx context.Context, projectID, id string) error
	SetCurrentVersion(ctx context.Context, projectID, visualizationID, versionID string) error
}

// VersionRepository manages immutable visualization versions.
// Versions are append-only; updates are not permitted after creation.
type VersionRepository interface {
	Create(ctx context.Context, version *VisualizationVersion) error
	GetByID(ctx context.Context, projectID, id string) (*VisualizationVersion, error)
	ListByVisualization(ctx context.Context, projectID, visualizationID string, limit int) ([]*VisualizationVersion, error)
	GetLatestByVisualization(ctx context.Context, projectID, visualizationID string) (*VisualizationVersion, error)
}

// SessionRepository persists session state for restore and audit.
// Sessions store the full SessionState as JSON for complete restore capability.
type SessionRepository interface {
	Create(ctx context.Context, session *SessionRecord) error
	GetByID(ctx context.Context, id string) (*SessionRecord, error)
	GetByProject(ctx context.Context, projectID string, limit int) ([]*SessionRecord, error)
	GetByVisualization(ctx context.Context, projectID, visualizationID string, limit int) ([]*SessionRecord, error)
	Update(ctx context.Context, session *SessionRecord) error
}

// AssetRepository manages asset metadata.
// The actual file bytes are stored externally; this repository tracks ownership and location.
type AssetRepository interface {
	Create(ctx context.Context, asset *Asset) error
	GetByID(ctx context.Context, projectID, id string) (*Asset, error)
	GetByStorageKey(ctx context.Context, storageKey string) (*Asset, error)
	ListByVisualization(ctx context.Context, projectID, visualizationID string) ([]*Asset, error)
	ListByVersion(ctx context.Context, projectID, versionID string) ([]*Asset, error)
	Delete(ctx context.Context, projectID, id string) error // Soft delete to trash
}

// ProjectScopedQuery marks queries that must be filtered by project_id.
// This interface is used for compile-time verification of project isolation.
type ProjectScopedQuery interface {
	GetProjectID() string
}

// FolderQuery retrieves folder contents with mixed items.
type FolderQuery struct {
	ProjectID string
	FolderID  *string // nil for root level
}

func (q *FolderQuery) GetProjectID() string { return q.ProjectID }

// VisualizationQuery retrieves visualizations filtered by location.
type VisualizationQuery struct {
	ProjectID string
	FolderID  *string // nil for root level or all in project
}

func (q *VisualizationQuery) GetProjectID() string { return q.ProjectID }

// HistoryQuery retrieves version history for a visualization.
type HistoryQuery struct {
	ProjectID       string
	VisualizationID string
	Limit           int
}

func (q *HistoryQuery) GetProjectID() string { return q.ProjectID }
