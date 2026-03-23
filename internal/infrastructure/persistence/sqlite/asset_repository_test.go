package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/paperbanana/paperbanana/internal/domain/workspace"
	"gorm.io/gorm"
)

func setupAssetTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	if err := db.AutoMigrate(AllModels()...); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	return db
}

func TestAssetRepository_Create(t *testing.T) {
	db := setupAssetTestDB(t)
	repo := NewAssetRepository(db)
	ctx := context.Background()

	asset := &workspace.Asset{
		ID:              "asset-1",
		ProjectID:       "project-1",
		VisualizationID: "viz-1",
		VersionID:       strPtr("version-1"),
		MIMEType:        "image/png",
		StorageKey:      "projects/project-1/viz/viz-1/uuid-1",
		ByteSize:        1024,
		ChecksumSHA256:  "abc123def456",
		CreatedAt:       time.Now(),
	}

	err := repo.Create(ctx, asset)
	if err != nil {
		t.Fatalf("failed to create asset: %v", err)
	}
}

func TestAssetRepository_GetByID(t *testing.T) {
	db := setupAssetTestDB(t)
	repo := NewAssetRepository(db)
	ctx := context.Background()

	// Create test asset
	asset := &workspace.Asset{
		ID:              "asset-2",
		ProjectID:       "project-2",
		VisualizationID: "viz-2",
		MIMEType:        "image/png",
		StorageKey:      "projects/project-2/viz/viz-2/uuid-2",
		ByteSize:        2048,
		ChecksumSHA256:  "checksum-2",
		CreatedAt:       time.Now(),
	}
	if err := repo.Create(ctx, asset); err != nil {
		t.Fatalf("setup: failed to create asset: %v", err)
	}

	// Retrieve by ID with project scope
	retrieved, err := repo.GetByID(ctx, "project-2", "asset-2")
	if err != nil {
		t.Fatalf("failed to get asset: %v", err)
	}

	if retrieved.ID != asset.ID {
		t.Errorf("ID mismatch: got %s, want %s", retrieved.ID, asset.ID)
	}
	if retrieved.ProjectID != asset.ProjectID {
		t.Errorf("ProjectID mismatch: got %s, want %s", retrieved.ProjectID, asset.ProjectID)
	}
}

func TestAssetRepository_GetByID_WrongProject(t *testing.T) {
	db := setupAssetTestDB(t)
	repo := NewAssetRepository(db)
	ctx := context.Background()

	// Create test asset in project-3
	asset := &workspace.Asset{
		ID:              "asset-3",
		ProjectID:       "project-3",
		VisualizationID: "viz-3",
		MIMEType:        "image/png",
		StorageKey:      "projects/project-3/viz/viz-3/uuid-3",
		ByteSize:        1024,
		ChecksumSHA256:  "checksum-3",
		CreatedAt:       time.Now(),
	}
	if err := repo.Create(ctx, asset); err != nil {
		t.Fatalf("setup: failed to create asset: %v", err)
	}

	// Try to retrieve with wrong project ID (cross-project access)
	_, err := repo.GetByID(ctx, "project-other", "asset-3")
	if err == nil {
		t.Error("expected error for cross-project asset access")
	}
}

func TestAssetRepository_GetByStorageKey(t *testing.T) {
	db := setupAssetTestDB(t)
	repo := NewAssetRepository(db)
	ctx := context.Background()

	storageKey := "projects/project-4/viz/viz-4/uuid-4"
	asset := &workspace.Asset{
		ID:              "asset-4",
		ProjectID:       "project-4",
		VisualizationID: "viz-4",
		MIMEType:        "application/pdf",
		StorageKey:      storageKey,
		ByteSize:        4096,
		ChecksumSHA256:  "checksum-4",
		CreatedAt:       time.Now(),
	}
	if err := repo.Create(ctx, asset); err != nil {
		t.Fatalf("setup: failed to create asset: %v", err)
	}

	// Retrieve by storage key
	retrieved, err := repo.GetByStorageKey(ctx, storageKey)
	if err != nil {
		t.Fatalf("failed to get asset by storage key: %v", err)
	}

	if retrieved.ID != asset.ID {
		t.Errorf("ID mismatch: got %s, want %s", retrieved.ID, asset.ID)
	}
	if retrieved.StorageKey != storageKey {
		t.Errorf("StorageKey mismatch: got %s, want %s", retrieved.StorageKey, storageKey)
	}
}

func TestAssetRepository_ListByVisualization(t *testing.T) {
	db := setupAssetTestDB(t)
	repo := NewAssetRepository(db)
	ctx := context.Background()

	projectID := "project-5"
	visualizationID := "viz-5"

	// Create multiple assets for the same visualization
	for i := 1; i <= 3; i++ {
		asset := &workspace.Asset{
			ID:              string(rune('a' + i)),
			ProjectID:       projectID,
			VisualizationID: visualizationID,
			MIMEType:        "image/png",
			StorageKey:      "projects/project-5/viz/viz-5/uuid-" + string(rune('0'+i)),
			ByteSize:        int64(i * 1024),
			ChecksumSHA256:  "checksum-" + string(rune('0'+i)),
			CreatedAt:       time.Now(),
		}
		if err := repo.Create(ctx, asset); err != nil {
			t.Fatalf("setup: failed to create asset %d: %v", i, err)
		}
	}

	// Create asset for different visualization
	otherAsset := &workspace.Asset{
		ID:              "other-asset",
		ProjectID:       projectID,
		VisualizationID: "viz-other",
		MIMEType:        "image/png",
		StorageKey:      "projects/project-5/viz/viz-other/uuid-other",
		ByteSize:        100,
		ChecksumSHA256:  "checksum-other",
		CreatedAt:       time.Now(),
	}
	if err := repo.Create(ctx, otherAsset); err != nil {
		t.Fatalf("setup: failed to create other asset: %v", err)
	}

	// List assets for visualization
	assets, err := repo.ListByVisualization(ctx, projectID, visualizationID)
	if err != nil {
		t.Fatalf("failed to list assets: %v", err)
	}

	if len(assets) != 3 {
		t.Errorf("expected 3 assets, got %d", len(assets))
	}

	// Verify all returned assets belong to the correct visualization
	for _, a := range assets {
		if a.VisualizationID != visualizationID {
			t.Errorf("wrong visualization ID in result: %s", a.VisualizationID)
		}
	}
}

func TestAssetRepository_ListByVersion(t *testing.T) {
	db := setupAssetTestDB(t)
	repo := NewAssetRepository(db)
	ctx := context.Background()

	projectID := "project-6"
	versionID := "version-6"

	// Create assets linked to a version
	for i := 1; i <= 2; i++ {
		asset := &workspace.Asset{
			ID:              "asset-v6-" + string(rune('0'+i)),
			ProjectID:       projectID,
			VisualizationID: "viz-6",
			VersionID:       &versionID,
			MIMEType:        "image/png",
			StorageKey:      "projects/project-6/viz/viz-6/uuid-v6-" + string(rune('0'+i)),
			ByteSize:        int64(i * 100),
			ChecksumSHA256:  "checksum-v6-" + string(rune('0'+i)),
			CreatedAt:       time.Now(),
		}
		if err := repo.Create(ctx, asset); err != nil {
			t.Fatalf("setup: failed to create asset %d: %v", i, err)
		}
	}

	// Create asset without version
	noVersionAsset := &workspace.Asset{
		ID:              "asset-noversion",
		ProjectID:       projectID,
		VisualizationID: "viz-6",
		MIMEType:        "image/png",
		StorageKey:      "projects/project-6/viz/viz-6/uuid-noversion",
		ByteSize:        50,
		ChecksumSHA256:  "checksum-noversion",
		CreatedAt:       time.Now(),
	}
	if err := repo.Create(ctx, noVersionAsset); err != nil {
		t.Fatalf("setup: failed to create no-version asset: %v", err)
	}

	// List assets for version
	assets, err := repo.ListByVersion(ctx, projectID, versionID)
	if err != nil {
		t.Fatalf("failed to list assets by version: %v", err)
	}

	if len(assets) != 2 {
		t.Errorf("expected 2 assets for version, got %d", len(assets))
	}

	// Verify all returned assets have the correct version
	for _, a := range assets {
		if a.VersionID == nil || *a.VersionID != versionID {
			t.Errorf("wrong version ID in result: %v", a.VersionID)
		}
	}
}

func TestAssetRepository_Delete(t *testing.T) {
	db := setupAssetTestDB(t)
	repo := NewAssetRepository(db)
	ctx := context.Background()

	asset := &workspace.Asset{
		ID:              "asset-del",
		ProjectID:       "project-del",
		VisualizationID: "viz-del",
		MIMEType:        "image/png",
		StorageKey:      "projects/project-del/viz/viz-del/uuid-del",
		ByteSize:        100,
		ChecksumSHA256:  "checksum-del",
		CreatedAt:       time.Now(),
	}
	if err := repo.Create(ctx, asset); err != nil {
		t.Fatalf("setup: failed to create asset: %v", err)
	}

	// Delete the asset
	err := repo.Delete(ctx, "project-del", "asset-del")
	if err != nil {
		t.Fatalf("failed to delete asset: %v", err)
	}

	// Verify it's deleted (soft delete)
	_, err = repo.GetByID(ctx, "project-del", "asset-del")
	if err == nil {
		t.Error("expected error getting deleted asset")
	}
}

func TestAssetRepository_ProjectIsolation(t *testing.T) {
	db := setupAssetTestDB(t)
	repo := NewAssetRepository(db)
	ctx := context.Background()

	// Create assets for different projects
	assetA := &workspace.Asset{
		ID:              "asset-a",
		ProjectID:       "project-a",
		VisualizationID: "viz-a",
		MIMEType:        "image/png",
		StorageKey:      "projects/project-a/viz/viz-a/uuid-a",
		ByteSize:        100,
		ChecksumSHA256:  "checksum-a",
		CreatedAt:       time.Now(),
	}
	assetB := &workspace.Asset{
		ID:              "asset-b",
		ProjectID:       "project-b",
		VisualizationID: "viz-b",
		MIMEType:        "image/png",
		StorageKey:      "projects/project-b/viz/viz-b/uuid-b",
		ByteSize:        200,
		ChecksumSHA256:  "checksum-b",
		CreatedAt:       time.Now(),
	}

	if err := repo.Create(ctx, assetA); err != nil {
		t.Fatalf("setup: failed to create asset A: %v", err)
	}
	if err := repo.Create(ctx, assetB); err != nil {
		t.Fatalf("setup: failed to create asset B: %v", err)
	}

	// Project A should only see asset A
	assetsA, err := repo.ListByVisualization(ctx, "project-a", "viz-a")
	if err != nil {
		t.Fatalf("failed to list project A assets: %v", err)
	}
	if len(assetsA) != 1 || assetsA[0].ID != "asset-a" {
		t.Errorf("project A should only see asset-a, got %d assets", len(assetsA))
	}

	// Project B should only see asset B
	assetsB, err := repo.ListByVisualization(ctx, "project-b", "viz-b")
	if err != nil {
		t.Fatalf("failed to list project B assets: %v", err)
	}
	if len(assetsB) != 1 || assetsB[0].ID != "asset-b" {
		t.Errorf("project B should only see asset-b, got %d assets", len(assetsB))
	}

	// Cross-project GetByID should fail
	_, err = repo.GetByID(ctx, "project-a", "asset-b")
	if err == nil {
		t.Error("cross-project access should fail")
	}

	// Cross-project Delete should fail
	err = repo.Delete(ctx, "project-a", "asset-b")
	if err == nil {
		t.Error("cross-project delete should fail")
	}
}

func TestAssetRepository_UniqueStorageKey(t *testing.T) {
	db := setupAssetTestDB(t)
	repo := NewAssetRepository(db)
	ctx := context.Background()

	storageKey := "projects/project-unique/viz/viz-unique/uuid-unique"

	asset1 := &workspace.Asset{
		ID:              "asset-unique-1",
		ProjectID:       "project-unique",
		VisualizationID: "viz-unique",
		MIMEType:        "image/png",
		StorageKey:      storageKey,
		ByteSize:        100,
		ChecksumSHA256:  "checksum-1",
		CreatedAt:       time.Now(),
	}
	if err := repo.Create(ctx, asset1); err != nil {
		t.Fatalf("setup: failed to create first asset: %v", err)
	}

	// Attempt to create second asset with same storage key
	asset2 := &workspace.Asset{
		ID:              "asset-unique-2",
		ProjectID:       "project-unique",
		VisualizationID: "viz-unique",
		MIMEType:        "image/png",
		StorageKey:      storageKey, // Duplicate key
		ByteSize:        200,
		ChecksumSHA256:  "checksum-2",
		CreatedAt:       time.Now(),
	}

	err := repo.Create(ctx, asset2)
	if err == nil {
		t.Error("expected error for duplicate storage key")
	}
}

// Helper function
func strPtr(s string) *string {
	return &s
}
