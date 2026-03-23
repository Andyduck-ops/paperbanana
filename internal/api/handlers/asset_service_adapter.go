package handlers

import (
	"context"

	"github.com/paperbanana/paperbanana/internal/domain/workspace"
)

// AssetServiceAdapter adapts persistence.AssetService to handlers.AssetService interface.
type AssetServiceAdapter struct {
	svc AssetPersistenceService
}

// AssetPersistenceService is the interface satisfied by persistence.AssetService.
type AssetPersistenceService interface {
	ListAssets(ctx context.Context, projectID, visualizationID string) ([]*workspace.Asset, error)
	GetAsset(ctx context.Context, projectID, assetID string) (*workspace.Asset, []byte, error)
	ListAssetsByVersion(ctx context.Context, projectID, versionID string) ([]*workspace.Asset, error)
}

// NewAssetServiceAdapter creates a new AssetServiceAdapter.
func NewAssetServiceAdapter(svc AssetPersistenceService) *AssetServiceAdapter {
	return &AssetServiceAdapter{svc: svc}
}

// ListAssets retrieves all assets for a visualization within a project.
func (a *AssetServiceAdapter) ListAssets(ctx context.Context, projectID, visualizationID string) ([]*AssetInfo, error) {
	assets, err := a.svc.ListAssets(ctx, projectID, visualizationID)
	if err != nil {
		return nil, err
	}

	result := make([]*AssetInfo, len(assets))
	for i, asset := range assets {
		result[i] = assetToAssetInfo(asset)
	}
	return result, nil
}

// GetAsset retrieves an asset's metadata and bytes by ID, scoped to a project.
func (a *AssetServiceAdapter) GetAsset(ctx context.Context, projectID, assetID string) (*AssetInfo, []byte, error) {
	asset, data, err := a.svc.GetAsset(ctx, projectID, assetID)
	if err != nil {
		return nil, nil, err
	}
	return assetToAssetInfo(asset), data, nil
}

// ListAssetsByVersion retrieves all assets for a specific version.
func (a *AssetServiceAdapter) ListAssetsByVersion(ctx context.Context, projectID, versionID string) ([]*AssetInfo, error) {
	assets, err := a.svc.ListAssetsByVersion(ctx, projectID, versionID)
	if err != nil {
		return nil, err
	}

	result := make([]*AssetInfo, len(assets))
	for i, asset := range assets {
		result[i] = assetToAssetInfo(asset)
	}
	return result, nil
}

// assetToAssetInfo converts workspace.Asset to handlers.AssetInfo.
func assetToAssetInfo(asset *workspace.Asset) *AssetInfo {
	return &AssetInfo{
		ID:              asset.ID,
		ProjectID:       asset.ProjectID,
		VisualizationID: asset.VisualizationID,
		VersionID:       asset.VersionID,
		MIMEType:        asset.MIMEType,
		ByteSize:        asset.ByteSize,
		ChecksumSHA256:  asset.ChecksumSHA256,
		CreatedAt:       asset.CreatedAt,
	}
}

// Ensure AssetServiceAdapter implements AssetService interface.
var _ AssetService = (*AssetServiceAdapter)(nil)

// AssetInfoFromPersistence creates a handlers.AssetInfo from persistence.AssetInfo.
func AssetInfoFromPersistence(asset *workspace.Asset) *AssetInfo {
	return &AssetInfo{
		ID:              asset.ID,
		ProjectID:       asset.ProjectID,
		VisualizationID: asset.VisualizationID,
		VersionID:       asset.VersionID,
		MIMEType:        asset.MIMEType,
		ByteSize:        asset.ByteSize,
		ChecksumSHA256:  asset.ChecksumSHA256,
		CreatedAt:       asset.CreatedAt,
	}
}
