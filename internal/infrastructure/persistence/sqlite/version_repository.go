package sqlite

import (
	"context"
	"errors"
	"fmt"

	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
	"github.com/paperbanana/paperbanana/internal/domain/workspace"
	"gorm.io/gorm"
)

// VersionRepository implements workspace.VersionRepository using SQLite/GORM.
type VersionRepository struct {
	db *gorm.DB
}

// NewVersionRepository creates a new VersionRepository.
func NewVersionRepository(db *gorm.DB) *VersionRepository {
	return &VersionRepository{db: db}
}

// Create persists a new immutable version record.
// The version number must be unique per visualization.
func (r *VersionRepository) Create(ctx context.Context, version *workspace.VisualizationVersion) error {
	if version == nil {
		return errors.New("version cannot be nil")
	}

	model := versionToModel(version)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("create version: %w", err)
	}
	return nil
}

// GetByID retrieves a version by ID within a project scope.
func (r *VersionRepository) GetByID(ctx context.Context, projectID, id string) (*workspace.VisualizationVersion, error) {
	var model VisualizationVersionModel
	if err := r.db.WithContext(ctx).
		Preload("Artifacts").
		Where("id = ? AND project_id = ?", id, projectID).
		First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("version not found: %s", id)
		}
		return nil, fmt.Errorf("get version: %w", err)
	}
	return modelToVersion(&model), nil
}

// ListByVisualization retrieves versions for a visualization, ordered by version number descending.
func (r *VersionRepository) ListByVisualization(ctx context.Context, projectID, visualizationID string, limit int) ([]*workspace.VisualizationVersion, error) {
	query := r.db.WithContext(ctx).
		Preload("Artifacts").
		Where("project_id = ? AND visualization_id = ?", projectID, visualizationID).
		Order("version_number DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	var models []VisualizationVersionModel
	if err := query.Find(&models).Error; err != nil {
		return nil, fmt.Errorf("list versions: %w", err)
	}

	versions := make([]*workspace.VisualizationVersion, len(models))
	for i, model := range models {
		versions[i] = modelToVersion(&model)
	}
	return versions, nil
}

// GetLatestByVisualization retrieves the most recent version for a visualization.
func (r *VersionRepository) GetLatestByVisualization(ctx context.Context, projectID, visualizationID string) (*workspace.VisualizationVersion, error) {
	var model VisualizationVersionModel
	if err := r.db.WithContext(ctx).
		Preload("Artifacts").
		Where("project_id = ? AND visualization_id = ?", projectID, visualizationID).
		Order("version_number DESC").
		First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("no versions found for visualization: %s", visualizationID)
		}
		return nil, fmt.Errorf("get latest version: %w", err)
	}
	return modelToVersion(&model), nil
}

// versionToModel converts a domain VisualizationVersion to GORM models.
func versionToModel(version *workspace.VisualizationVersion) *VisualizationVersionModel {
	model := &VisualizationVersionModel{
		ID:              version.ID,
		VisualizationID: version.VisualizationID,
		ProjectID:       version.ProjectID,
		VersionNumber:   version.VersionNumber,
		SessionID:       version.SessionID,
		Summary:         version.Summary,
		CreatedAt:       version.CreatedAt,
	}

	if version.SessionSnapshot != nil {
		model.SnapshotJSON = SessionSnapshotPayload{
			SchemaVersion: version.SessionSnapshot.SchemaVersion,
			SessionID:     version.SessionSnapshot.SessionID,
			RequestID:     version.SessionSnapshot.RequestID,
			Status:        version.SessionSnapshot.Status,
			CurrentStage:  version.SessionSnapshot.CurrentStage,
			Pipeline:      version.SessionSnapshot.Pipeline,
			InitialInput:  version.SessionSnapshot.InitialInput,
			StageStates:   version.SessionSnapshot.StageStates,
			FinalOutput:   version.SessionSnapshot.FinalOutput,
			Error:         version.SessionSnapshot.Error,
			Restore:       version.SessionSnapshot.Restore,
			Metadata:      version.SessionSnapshot.Metadata,
			StartedAt:     version.SessionSnapshot.StartedAt,
			UpdatedAt:     version.SessionSnapshot.UpdatedAt,
			CompletedAt:   version.SessionSnapshot.CompletedAt,
		}
	}

	// Convert artifacts
	for _, artifact := range version.Artifacts {
		model.Artifacts = append(model.Artifacts, VersionArtifactModel{
			ID:             artifact.ID,
			VersionID:      version.ID,
			AssetID:        artifact.AssetID,
			Kind:           artifact.Kind,
			MIMEType:       artifact.MIMEType,
			StorageKey:     artifact.StorageKey,
			ByteSize:       artifact.ByteSize,
			ChecksumSHA256: artifact.ChecksumSHA256,
		})
	}

	return model
}

// modelToVersion converts a GORM VisualizationVersionModel to domain type.
func modelToVersion(model *VisualizationVersionModel) *workspace.VisualizationVersion {
	version := &workspace.VisualizationVersion{
		ID:              model.ID,
		VisualizationID: model.VisualizationID,
		ProjectID:       model.ProjectID,
		VersionNumber:   model.VersionNumber,
		SessionID:       model.SessionID,
		Summary:         model.Summary,
		CreatedAt:       model.CreatedAt,
	}

	if model.SnapshotJSON.SessionID != "" {
		version.SessionSnapshot = &domainagent.SessionState{
			SchemaVersion: model.SnapshotJSON.SchemaVersion,
			SessionID:     model.SnapshotJSON.SessionID,
			RequestID:     model.SnapshotJSON.RequestID,
			Status:        model.SnapshotJSON.Status,
			CurrentStage:  model.SnapshotJSON.CurrentStage,
			Pipeline:      model.SnapshotJSON.Pipeline,
			InitialInput:  model.SnapshotJSON.InitialInput,
			StageStates:   model.SnapshotJSON.StageStates,
			FinalOutput:   model.SnapshotJSON.FinalOutput,
			Error:         model.SnapshotJSON.Error,
			Restore:       model.SnapshotJSON.Restore,
			Metadata:      model.SnapshotJSON.Metadata,
			StartedAt:     model.SnapshotJSON.StartedAt,
			UpdatedAt:     model.SnapshotJSON.UpdatedAt,
			CompletedAt:   model.SnapshotJSON.CompletedAt,
		}
	}

	// Convert artifacts
	for _, artifactModel := range model.Artifacts {
		version.Artifacts = append(version.Artifacts, workspace.VersionArtifact{
			ID:             artifactModel.ID,
			VersionID:      artifactModel.VersionID,
			AssetID:        artifactModel.AssetID,
			Kind:           artifactModel.Kind,
			MIMEType:       artifactModel.MIMEType,
			StorageKey:     artifactModel.StorageKey,
			ByteSize:       artifactModel.ByteSize,
			ChecksumSHA256: artifactModel.ChecksumSHA256,
		})
	}

	return version
}

// Ensure VersionRepository implements workspace.VersionRepository.
var _ workspace.VersionRepository = (*VersionRepository)(nil)
