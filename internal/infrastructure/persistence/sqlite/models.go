package sqlite

import (
	"time"

	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
	"gorm.io/gorm"
)

// ProjectModel is the GORM model for projects.
type ProjectModel struct {
	ID          string         `gorm:"primaryKey;type:varchar(36)"`
	Name        string         `gorm:"type:varchar(255);not null"`
	Description string         `gorm:"type:text"`
	CreatedAt   time.Time      `gorm:"not null"`
	UpdatedAt   time.Time      `gorm:"not null"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// TableName returns the table name for ProjectModel.
func (ProjectModel) TableName() string {
	return "projects"
}

// FolderModel is the GORM model for folders.
type FolderModel struct {
	ID        string         `gorm:"primaryKey;type:varchar(36)"`
	ProjectID string         `gorm:"type:varchar(36);not null;index;constraint:OnDelete:CASCADE"`
	ParentID  *string        `gorm:"type:varchar(36);index"`
	Name      string         `gorm:"type:varchar(255);not null"`
	CreatedAt time.Time      `gorm:"not null"`
	UpdatedAt time.Time      `gorm:"not null"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// TableName returns the table name for FolderModel.
func (FolderModel) TableName() string {
	return "folders"
}

// VisualizationModel is the GORM model for visualizations.
type VisualizationModel struct {
	ID               string         `gorm:"primaryKey;type:varchar(36)"`
	ProjectID        string         `gorm:"type:varchar(36);not null;index;constraint:OnDelete:CASCADE"`
	FolderID         *string        `gorm:"type:varchar(36);index"`
	Name             string         `gorm:"type:varchar(255);not null"`
	CurrentVersionID *string        `gorm:"type:varchar(36)"`
	CreatedAt        time.Time      `gorm:"not null"`
	UpdatedAt        time.Time      `gorm:"not null"`
	DeletedAt        gorm.DeletedAt `gorm:"index"`
}

// TableName returns the table name for VisualizationModel.
func (VisualizationModel) TableName() string {
	return "visualizations"
}

// VisualizationVersionModel is the GORM model for visualization versions.
// Versions are immutable after creation.
type VisualizationVersionModel struct {
	ID              string                    `gorm:"primaryKey;type:varchar(36)"`
	VisualizationID string                    `gorm:"type:varchar(36);not null;index;uniqueIndex:idx_version_visualization_number;constraint:OnDelete:CASCADE"`
	ProjectID       string                    `gorm:"type:varchar(36);not null;index;constraint:OnDelete:CASCADE"`
	VersionNumber   int                       `gorm:"not null;uniqueIndex:idx_version_visualization_number"`
	SessionID       string                    `gorm:"type:varchar(36);not null;index"`
	Summary         string                    `gorm:"type:text"`
	Artifacts       []VersionArtifactModel    `gorm:"foreignKey:VersionID;constraint:OnDelete:CASCADE"`
	CreatedAt       time.Time                 `gorm:"not null"`
	SnapshotJSON    SessionSnapshotPayload    `gorm:"type:json;serializer:json"`
}

// TableName returns the table name for VisualizationVersionModel.
func (VisualizationVersionModel) TableName() string {
	return "visualization_versions"
}

// VersionArtifactModel captures asset metadata attached to a version.
type VersionArtifactModel struct {
	ID             string `gorm:"primaryKey;type:varchar(36)"`
	VersionID      string `gorm:"type:varchar(36);not null;index"`
	AssetID        string `gorm:"type:varchar(36);not null;index"`
	Kind           string `gorm:"type:varchar(50);not null"`
	MIMEType       string `gorm:"type:varchar(100);not null"`
	StorageKey     string `gorm:"type:varchar(255);not null"`
	ByteSize       int64  `gorm:"not null"`
	ChecksumSHA256 string `gorm:"type:varchar(64);not null"`
}

// TableName returns the table name for VersionArtifactModel.
func (VersionArtifactModel) TableName() string {
	return "version_artifacts"
}

// SessionModel is the GORM model for session records.
// The SnapshotJSON field stores the complete SessionState as opaque JSON.
type SessionModel struct {
	ID              string                 `gorm:"primaryKey;type:varchar(36)"`
	ProjectID       string                 `gorm:"type:varchar(36);not null;index;constraint:OnDelete:CASCADE"`
	VisualizationID *string                `gorm:"type:varchar(36);index;constraint:OnDelete:SET NULL"`
	Status          string                 `gorm:"type:varchar(50);not null;index"`
	CurrentStage    string                 `gorm:"type:varchar(50);not null;index"`
	SchemaVersion   string                 `gorm:"type:varchar(100);not null"`
	SnapshotJSON    SessionSnapshotPayload `gorm:"type:json;serializer:json;not null"`
	CreatedAt       time.Time              `gorm:"not null"`
	UpdatedAt       time.Time              `gorm:"not null"`
	CompletedAt     *time.Time
}

// TableName returns the table name for SessionModel.
func (SessionModel) TableName() string {
	return "sessions"
}

// SessionSnapshotPayload wraps domainagent.SessionState for JSON serialization.
type SessionSnapshotPayload struct {
	SchemaVersion string                 `json:"schema_version"`
	SessionID     string                 `json:"session_id"`
	RequestID     string                 `json:"request_id"`
	Status        domainagent.RunStatus  `json:"status"`
	CurrentStage  domainagent.StageName  `json:"current_stage"`
	Pipeline      []domainagent.StageName `json:"pipeline,omitempty"`
	InitialInput  domainagent.AgentInput `json:"initial_input"`
	StageStates   []domainagent.AgentState `json:"stage_states,omitempty"`
	FinalOutput   domainagent.AgentOutput `json:"final_output"`
	Error         *domainagent.ErrorDetail `json:"error,omitempty"`
	Restore       domainagent.RestoreMetadata `json:"restore"`
	Metadata      map[string]string      `json:"metadata,omitempty"`
	StartedAt     time.Time              `json:"started_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	CompletedAt   time.Time              `json:"completed_at,omitempty"`
}

// AssetModel is the GORM model for asset metadata.
type AssetModel struct {
	ID              string         `gorm:"primaryKey;type:varchar(36)"`
	ProjectID       string         `gorm:"type:varchar(36);not null;index;constraint:OnDelete:CASCADE"`
	VisualizationID string         `gorm:"type:varchar(36);not null;index;constraint:OnDelete:CASCADE"`
	VersionID       *string        `gorm:"type:varchar(36);index;constraint:OnDelete:SET NULL"`
	MIMEType        string         `gorm:"type:varchar(100);not null"`
	StorageKey      string         `gorm:"type:varchar(255);not null;uniqueIndex"`
	ByteSize        int64          `gorm:"not null"`
	ChecksumSHA256  string         `gorm:"type:varchar(64);not null"`
	CreatedAt       time.Time      `gorm:"not null"`
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}

// TableName returns the table name for AssetModel.
func (AssetModel) TableName() string {
	return "assets"
}

// AllModels returns all GORM models for AutoMigrate.
func AllModels() []interface{} {
	return []interface{}{
		&ProjectModel{},
		&FolderModel{},
		&VisualizationModel{},
		&VisualizationVersionModel{},
		&VersionArtifactModel{},
		&SessionModel{},
		&AssetModel{},
		&ProviderModel{},
		&APIKeyModel{},
	}
}
