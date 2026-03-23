package workspace

import (
	"time"

	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
)

// Project is the top-level organizational unit for a user's workspace.
// All folders, visualizations, sessions, and assets belong to exactly one project.
type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Folder provides nested organization within a project.
// Folders can contain other folders and visualizations in a mixed listing.
type Folder struct {
	ID        string     `json:"id"`
	ProjectID string     `json:"project_id"`
	ParentID  *string    `json:"parent_id,omitempty"`
	Name      string     `json:"name"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"` // Soft delete for trash semantics
}

// Visualization represents a single chart or diagram entry.
// It points to the current version and maintains project/folder ownership.
type Visualization struct {
	ID               string     `json:"id"`
	ProjectID        string     `json:"project_id"`
	FolderID         *string    `json:"folder_id,omitempty"`
	Name             string     `json:"name"`
	CurrentVersionID *string    `json:"current_version_id,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	DeletedAt        *time.Time `json:"deleted_at,omitempty"` // Soft delete for trash semantics
}

// VisualizationVersion is an immutable snapshot of a visualization's state.
// Each successful generation creates a new version; versions are never modified.
type VisualizationVersion struct {
	ID              string                       `json:"id"`
	VisualizationID string                       `json:"visualization_id"`
	ProjectID       string                       `json:"project_id"`
	VersionNumber   int                          `json:"version_number"`
	SessionID       string                       `json:"session_id"`
	Summary         string                       `json:"summary,omitempty"`
	Artifacts       []VersionArtifact            `json:"artifacts,omitempty"`
	CreatedAt       time.Time                    `json:"created_at"`
	SessionSnapshot *domainagent.SessionState    `json:"session_snapshot,omitempty"`
}

// VersionArtifact captures asset metadata attached to a version.
type VersionArtifact struct {
	ID           string `json:"id"`
	VersionID    string `json:"version_id"`
	AssetID      string `json:"asset_id"`
	Kind         string `json:"kind"` // e.g., "final", "export", "intermediate"
	MIMEType     string `json:"mime_type"`
	StorageKey   string `json:"storage_key"`
	ByteSize     int64  `json:"byte_size"`
	ChecksumSHA256 string `json:"checksum_sha256"`
}

// SessionRecord persists the full session state for restore and audit.
// The SnapshotJSON field stores the complete SessionState as opaque JSON.
type SessionRecord struct {
	ID              string                    `json:"id"`
	ProjectID       string                    `json:"project_id"`
	VisualizationID *string                   `json:"visualization_id,omitempty"`
	Status          string                    `json:"status"`
	CurrentStage    string                    `json:"current_stage"`
	SchemaVersion   string                    `json:"schema_version"`
	Snapshot        *domainagent.SessionState `json:"snapshot,omitempty"`
	CreatedAt       time.Time                 `json:"created_at"`
	UpdatedAt       time.Time                 `json:"updated_at"`
	CompletedAt     *time.Time                `json:"completed_at,omitempty"`
}

// Asset stores metadata for a file belonging to a visualization.
// The actual file bytes live in an external asset store keyed by StorageKey.
type Asset struct {
	ID             string     `json:"id"`
	ProjectID      string     `json:"project_id"`
	VisualizationID string    `json:"visualization_id"`
	VersionID      *string    `json:"version_id,omitempty"`
	MIMEType       string     `json:"mime_type"`
	StorageKey     string     `json:"storage_key"`
	ByteSize       int64      `json:"byte_size"`
	ChecksumSHA256 string     `json:"checksum_sha256"`
	CreatedAt      time.Time  `json:"created_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`
}
