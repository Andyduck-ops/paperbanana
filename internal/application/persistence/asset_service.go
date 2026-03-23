package persistence

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/paperbanana/paperbanana/internal/domain/agent"
	"github.com/paperbanana/paperbanana/internal/domain/workspace"
)

// Retained artifact kinds - these are stored as assets
var retainedKinds = map[agent.ArtifactKind]bool{
	agent.ArtifactKindRenderedFigure: true, // Final output images
	agent.ArtifactKindPromptTrace:    true, // Execution traces for audit
	agent.ArtifactKindCritique:       true, // Critique results
}

// AssetService coordinates asset metadata and byte storage.
// It enforces project isolation and manages the lifecycle of retained assets.
type AssetService struct {
	repo  workspace.AssetRepository
	store AssetStore
}

// NewAssetService creates a new AssetService.
func NewAssetService(repo workspace.AssetRepository, store AssetStore) *AssetService {
	return &AssetService{
		repo:  repo,
		store: store,
	}
}

// RegisterRetainedAssets processes artifacts and stores retained ones as assets.
// Only final outputs, exports, and audit-relevant artifacts are retained.
// Returns the created asset metadata records.
func (s *AssetService) RegisterRetainedAssets(
	ctx context.Context,
	projectID string,
	visualizationID string,
	versionID *string,
	artifacts []agent.Artifact,
) ([]*workspace.Asset, error) {
	if len(artifacts) == 0 {
		return []*workspace.Asset{}, nil
	}

	var createdAssets []*workspace.Asset

	for _, artifact := range artifacts {
		// Check if this artifact kind should be retained
		if !retainedKinds[artifact.Kind] {
			continue
		}

		// Get bytes from artifact (Bytes field or Content field)
		data := artifact.Bytes
		if len(data) == 0 && artifact.Content != "" {
			data = []byte(artifact.Content)
		}

		// Skip empty artifacts
		if len(data) == 0 {
			continue
		}

		// Create asset metadata
		assetID := uuid.New().String()
		now := time.Now()
		checksum := sha256.Sum256(data)

		// Store bytes using the AssetStore interface
		// Generate storage key using project/visualization scope
		storageKey := fmt.Sprintf("projects/%s/viz/%s/%s", projectID, visualizationID, assetID)

		if err := s.store.Write(ctx, storageKey, data); err != nil {
			return nil, fmt.Errorf("store asset bytes: %w", err)
		}

		asset := &workspace.Asset{
			ID:              assetID,
			ProjectID:       projectID,
			VisualizationID: visualizationID,
			VersionID:       versionID,
			MIMEType:        artifact.MIMEType,
			StorageKey:      storageKey,
			ByteSize:        int64(len(data)),
			ChecksumSHA256:  hex.EncodeToString(checksum[:]),
			CreatedAt:       now,
		}

		// Persist metadata
		if err := s.repo.Create(ctx, asset); err != nil {
			// Attempt to clean up stored bytes
			s.store.Delete(ctx, storageKey)
			return nil, fmt.Errorf("create asset metadata: %w", err)
		}

		createdAssets = append(createdAssets, asset)
	}

	return createdAssets, nil
}

// GetAsset retrieves an asset's metadata and bytes by ID, scoped to a project.
// Returns ErrCrossProjectAccess if the asset belongs to a different project.
func (s *AssetService) GetAsset(ctx context.Context, projectID, assetID string) (*workspace.Asset, []byte, error) {
	// Get metadata with project scope
	asset, err := s.repo.GetByID(ctx, projectID, assetID)
	if err != nil {
		return nil, nil, fmt.Errorf("get asset metadata: %w", err)
	}

	// Read bytes from store
	data, err := s.store.Read(ctx, asset.StorageKey)
	if err != nil {
		return nil, nil, fmt.Errorf("read asset bytes: %w", err)
	}

	return asset, data, nil
}

// GetAssetByStorageKey retrieves an asset by its storage key.
// This is useful for download endpoints where the key comes from the URL.
func (s *AssetService) GetAssetByStorageKey(ctx context.Context, storageKey string) (*workspace.Asset, []byte, error) {
	// Get metadata by storage key
	asset, err := s.repo.GetByStorageKey(ctx, storageKey)
	if err != nil {
		return nil, nil, fmt.Errorf("get asset by storage key: %w", err)
	}

	// Read bytes from store
	data, err := s.store.Read(ctx, asset.StorageKey)
	if err != nil {
		return nil, nil, fmt.Errorf("read asset bytes: %w", err)
	}

	return asset, data, nil
}

// ListAssets retrieves all assets for a visualization within a project.
func (s *AssetService) ListAssets(ctx context.Context, projectID, visualizationID string) ([]*workspace.Asset, error) {
	return s.repo.ListByVisualization(ctx, projectID, visualizationID)
}

// ListAssetsByVersion retrieves all assets for a specific version.
func (s *AssetService) ListAssetsByVersion(ctx context.Context, projectID, versionID string) ([]*workspace.Asset, error) {
	return s.repo.ListByVersion(ctx, projectID, versionID)
}

// DeleteAsset removes an asset's metadata and bytes, scoped to a project.
func (s *AssetService) DeleteAsset(ctx context.Context, projectID, assetID string) error {
	// Get asset to find storage key (with project scope)
	asset, err := s.repo.GetByID(ctx, projectID, assetID)
	if err != nil {
		return fmt.Errorf("get asset for deletion: %w", err)
	}

	// Delete from store first
	if err := s.store.Delete(ctx, asset.StorageKey); err != nil {
		// Log but continue - metadata delete is more important
		// In production, use proper logging
	}

	// Delete metadata
	if err := s.repo.Delete(ctx, projectID, assetID); err != nil {
		return fmt.Errorf("delete asset metadata: %w", err)
	}

	return nil
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

// ToInfo converts a workspace.Asset to an AssetInfo for API responses.
func ToInfo(asset *workspace.Asset) *AssetInfo {
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

// Errors
var (
	ErrCrossProjectAccess = errors.New("cross-project asset access denied")
	ErrAssetNotFound      = errors.New("asset not found")
)
