package sqlite

import (
	"context"
	"errors"
	"fmt"

	"github.com/paperbanana/paperbanana/internal/domain/workspace"
	"gorm.io/gorm"
)

// AssetRepository implements workspace.AssetRepository using SQLite/GORM.
type AssetRepository struct {
	db *gorm.DB
}

// NewAssetRepository creates a new AssetRepository.
func NewAssetRepository(db *gorm.DB) *AssetRepository {
	return &AssetRepository{db: db}
}

// Create persists a new asset record.
func (r *AssetRepository) Create(ctx context.Context, asset *workspace.Asset) error {
	if asset == nil {
		return errors.New("asset cannot be nil")
	}

	model := assetToModel(asset)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("create asset: %w", err)
	}
	return nil
}

// GetByID retrieves an asset by its ID, scoped to the project.
// Returns an error if the asset does not exist or belongs to a different project.
func (r *AssetRepository) GetByID(ctx context.Context, projectID, id string) (*workspace.Asset, error) {
	var model AssetModel
	if err := r.db.WithContext(ctx).
		Where("id = ? AND project_id = ?", id, projectID).
		First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("asset not found: %s (project: %s)", id, projectID)
		}
		return nil, fmt.Errorf("get asset: %w", err)
	}
	return modelToAsset(&model), nil
}

// GetByStorageKey retrieves an asset by its storage key.
// This is used by the asset service to resolve storage keys for download.
func (r *AssetRepository) GetByStorageKey(ctx context.Context, storageKey string) (*workspace.Asset, error) {
	var model AssetModel
	if err := r.db.WithContext(ctx).
		Where("storage_key = ?", storageKey).
		First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("asset not found for storage key: %s", storageKey)
		}
		return nil, fmt.Errorf("get asset by storage key: %w", err)
	}
	return modelToAsset(&model), nil
}

// ListByVisualization retrieves all assets for a visualization within a project.
func (r *AssetRepository) ListByVisualization(ctx context.Context, projectID, visualizationID string) ([]*workspace.Asset, error) {
	var models []AssetModel
	if err := r.db.WithContext(ctx).
		Where("project_id = ? AND visualization_id = ?", projectID, visualizationID).
		Order("created_at DESC").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("list assets by visualization: %w", err)
	}

	assets := make([]*workspace.Asset, len(models))
	for i, model := range models {
		assets[i] = modelToAsset(&model)
	}
	return assets, nil
}

// ListByVersion retrieves all assets for a specific version within a project.
func (r *AssetRepository) ListByVersion(ctx context.Context, projectID, versionID string) ([]*workspace.Asset, error) {
	var models []AssetModel
	if err := r.db.WithContext(ctx).
		Where("project_id = ? AND version_id = ?", projectID, versionID).
		Order("created_at DESC").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("list assets by version: %w", err)
	}

	assets := make([]*workspace.Asset, len(models))
	for i, model := range models {
		assets[i] = modelToAsset(&model)
	}
	return assets, nil
}

// Delete removes an asset record (soft delete).
// The operation is scoped to the project to prevent cross-project deletion.
func (r *AssetRepository) Delete(ctx context.Context, projectID, id string) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND project_id = ?", id, projectID).
		Delete(&AssetModel{})

	if result.Error != nil {
		return fmt.Errorf("delete asset: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("asset not found: %s (project: %s)", id, projectID)
	}
	return nil
}

// assetToModel converts a domain Asset to a GORM AssetModel.
func assetToModel(asset *workspace.Asset) *AssetModel {
	model := &AssetModel{
		ID:              asset.ID,
		ProjectID:       asset.ProjectID,
		VisualizationID: asset.VisualizationID,
		MIMEType:        asset.MIMEType,
		StorageKey:      asset.StorageKey,
		ByteSize:        asset.ByteSize,
		ChecksumSHA256:  asset.ChecksumSHA256,
		CreatedAt:       asset.CreatedAt,
	}

	if asset.VersionID != nil {
		model.VersionID = asset.VersionID
	}

	if asset.DeletedAt != nil {
		model.DeletedAt = gorm.DeletedAt{Time: *asset.DeletedAt, Valid: true}
	}

	return model
}

// modelToAsset converts a GORM AssetModel to a domain Asset.
func modelToAsset(model *AssetModel) *workspace.Asset {
	asset := &workspace.Asset{
		ID:              model.ID,
		ProjectID:       model.ProjectID,
		VisualizationID: model.VisualizationID,
		MIMEType:        model.MIMEType,
		StorageKey:      model.StorageKey,
		ByteSize:        model.ByteSize,
		ChecksumSHA256:  model.ChecksumSHA256,
		CreatedAt:       model.CreatedAt,
	}

	if model.VersionID != nil && *model.VersionID != "" {
		asset.VersionID = model.VersionID
	}

	if model.DeletedAt.Valid {
		asset.DeletedAt = &model.DeletedAt.Time
	}

	return asset
}

// Ensure AssetRepository implements workspace.AssetRepository.
var _ workspace.AssetRepository = (*AssetRepository)(nil)
