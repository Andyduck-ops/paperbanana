package persistence

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/paperbanana/paperbanana/internal/domain/agent"
	"github.com/paperbanana/paperbanana/internal/domain/workspace"
)

// MockAssetStore implements AssetStore for testing
type MockAssetStore struct {
	WriteFunc  func(ctx context.Context, storageKey string, data []byte) error
	ReadFunc   func(ctx context.Context, storageKey string) ([]byte, error)
	DeleteFunc func(ctx context.Context, storageKey string) error
	ExistsFunc func(ctx context.Context, storageKey string) (bool, error)

	// Call tracking
	WriteCalls  []struct{ Key string; Data []byte }
	ReadCalls   []string
	DeleteCalls []string
}

func (m *MockAssetStore) Write(ctx context.Context, storageKey string, data []byte) error {
	m.WriteCalls = append(m.WriteCalls, struct{ Key string; Data []byte }{Key: storageKey, Data: data})
	if m.WriteFunc != nil {
		return m.WriteFunc(ctx, storageKey, data)
	}
	return nil
}

func (m *MockAssetStore) Read(ctx context.Context, storageKey string) ([]byte, error) {
	m.ReadCalls = append(m.ReadCalls, storageKey)
	if m.ReadFunc != nil {
		return m.ReadFunc(ctx, storageKey)
	}
	return nil, nil
}

func (m *MockAssetStore) Delete(ctx context.Context, storageKey string) error {
	m.DeleteCalls = append(m.DeleteCalls, storageKey)
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, storageKey)
	}
	return nil
}

func (m *MockAssetStore) Exists(ctx context.Context, storageKey string) (bool, error) {
	if m.ExistsFunc != nil {
		return m.ExistsFunc(ctx, storageKey)
	}
	return false, nil
}

// MockAssetRepository implements workspace.AssetRepository for testing
type MockAssetRepository struct {
	CreateFunc              func(ctx context.Context, asset *workspace.Asset) error
	GetByIDFunc             func(ctx context.Context, projectID, id string) (*workspace.Asset, error)
	GetByStorageKeyFunc     func(ctx context.Context, storageKey string) (*workspace.Asset, error)
	ListByVisualizationFunc func(ctx context.Context, projectID, visualizationID string) ([]*workspace.Asset, error)
	ListByVersionFunc       func(ctx context.Context, projectID, versionID string) ([]*workspace.Asset, error)
	DeleteFunc              func(ctx context.Context, projectID, id string) error

	CreateCalls  []*workspace.Asset
	GetByIDCalls []struct{ ProjectID, ID string }
}

func (m *MockAssetRepository) Create(ctx context.Context, asset *workspace.Asset) error {
	m.CreateCalls = append(m.CreateCalls, asset)
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, asset)
	}
	return nil
}

func (m *MockAssetRepository) GetByID(ctx context.Context, projectID, id string) (*workspace.Asset, error) {
	m.GetByIDCalls = append(m.GetByIDCalls, struct{ ProjectID, ID string }{ProjectID: projectID, ID: id})
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, projectID, id)
	}
	return nil, nil
}

func (m *MockAssetRepository) GetByStorageKey(ctx context.Context, storageKey string) (*workspace.Asset, error) {
	if m.GetByStorageKeyFunc != nil {
		return m.GetByStorageKeyFunc(ctx, storageKey)
	}
	return nil, nil
}

func (m *MockAssetRepository) ListByVisualization(ctx context.Context, projectID, visualizationID string) ([]*workspace.Asset, error) {
	if m.ListByVisualizationFunc != nil {
		return m.ListByVisualizationFunc(ctx, projectID, visualizationID)
	}
	return nil, nil
}

func (m *MockAssetRepository) ListByVersion(ctx context.Context, projectID, versionID string) ([]*workspace.Asset, error) {
	if m.ListByVersionFunc != nil {
		return m.ListByVersionFunc(ctx, projectID, versionID)
	}
	return nil, nil
}

func (m *MockAssetRepository) Delete(ctx context.Context, projectID, id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, projectID, id)
	}
	return nil
}

func TestAssetService_RegisterRetainedAssets(t *testing.T) {
	ctx := context.Background()

	mockStore := &MockAssetStore{}
	mockRepo := &MockAssetRepository{}
	service := NewAssetService(mockRepo, mockStore)

	projectID := "project-1"
	visualizationID := "viz-1"
	versionID := "version-1"

	// Create test artifacts (retained final outputs)
	artifacts := []agent.Artifact{
		{
			ID:       "artifact-1",
			Kind:     agent.ArtifactKindRenderedFigure,
			MIMEType: "image/png",
			Bytes:    []byte("png-image-data"),
		},
		{
			ID:       "artifact-2",
			Kind:     agent.ArtifactKindPromptTrace,
			MIMEType: "application/json",
			Content:  `{"prompt": "test"}`,
		},
	}

	// Register retained assets
	result, err := service.RegisterRetainedAssets(ctx, projectID, visualizationID, &versionID, artifacts)
	if err != nil {
		t.Fatalf("failed to register assets: %v", err)
	}

	// Verify assets were created
	if len(result) != len(artifacts) {
		t.Errorf("expected %d assets, got %d", len(artifacts), len(result))
	}

	// Verify repository Create was called for each artifact
	if len(mockRepo.CreateCalls) != len(artifacts) {
		t.Errorf("expected %d Create calls, got %d", len(artifacts), len(mockRepo.CreateCalls))
	}

	// Verify store Write was called for each artifact with bytes
	// Note: artifacts with Content should also be stored
	if len(mockStore.WriteCalls) != len(artifacts) {
		t.Errorf("expected %d Write calls, got %d", len(artifacts), len(mockStore.WriteCalls))
	}

	// Verify project isolation: each asset should be linked to the project
	for _, asset := range mockRepo.CreateCalls {
		if asset.ProjectID != projectID {
			t.Errorf("asset project_id mismatch: got %s, want %s", asset.ProjectID, projectID)
		}
		if asset.VisualizationID != visualizationID {
			t.Errorf("asset visualization_id mismatch: got %s, want %s", asset.VisualizationID, visualizationID)
		}
		if asset.VersionID == nil || *asset.VersionID != versionID {
			t.Errorf("asset version_id mismatch: expected %s", versionID)
		}
	}
}

func TestAssetService_RegisterRetainedAssets_NoVersion(t *testing.T) {
	ctx := context.Background()

	mockStore := &MockAssetStore{}
	mockRepo := &MockAssetRepository{}
	service := NewAssetService(mockRepo, mockStore)

	projectID := "project-noversion"
	visualizationID := "viz-noversion"

	artifacts := []agent.Artifact{
		{
			ID:       "artifact-noversion",
			Kind:     agent.ArtifactKindRenderedFigure,
			MIMEType: "image/png",
			Bytes:    []byte("data"),
		},
	}

	// Register without version
	result, err := service.RegisterRetainedAssets(ctx, projectID, visualizationID, nil, artifacts)
	if err != nil {
		t.Fatalf("failed to register assets: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("expected 1 asset, got %d", len(result))
	}

	// Verify version_id is nil
	for _, asset := range mockRepo.CreateCalls {
		if asset.VersionID != nil {
			t.Errorf("asset version_id should be nil, got %s", *asset.VersionID)
		}
	}
}

func TestAssetService_GetAsset(t *testing.T) {
	ctx := context.Background()

	expectedAsset := &workspace.Asset{
		ID:              "asset-get",
		ProjectID:       "project-get",
		VisualizationID: "viz-get",
		MIMEType:        "image/png",
		StorageKey:      "projects/project-get/viz/viz-get/uuid",
		ByteSize:        100,
		ChecksumSHA256:  "checksum",
		CreatedAt:       time.Now(),
	}

	mockStore := &MockAssetStore{
		ReadFunc: func(ctx context.Context, storageKey string) ([]byte, error) {
			return []byte("image-data"), nil
		},
	}
	mockRepo := &MockAssetRepository{
		GetByIDFunc: func(ctx context.Context, projectID, id string) (*workspace.Asset, error) {
			if projectID == expectedAsset.ProjectID && id == expectedAsset.ID {
				return expectedAsset, nil
			}
			return nil, errors.New("not found")
		},
	}
	service := NewAssetService(mockRepo, mockStore)

	// Get asset
	asset, data, err := service.GetAsset(ctx, expectedAsset.ProjectID, expectedAsset.ID)
	if err != nil {
		t.Fatalf("failed to get asset: %v", err)
	}

	if asset.ID != expectedAsset.ID {
		t.Errorf("asset ID mismatch: got %s, want %s", asset.ID, expectedAsset.ID)
	}
	if string(data) != "image-data" {
		t.Errorf("data mismatch: got %s", string(data))
	}

	// Verify project isolation was enforced
	if len(mockRepo.GetByIDCalls) != 1 {
		t.Fatal("GetByID should be called once")
	}
	call := mockRepo.GetByIDCalls[0]
	if call.ProjectID != expectedAsset.ProjectID {
		t.Errorf("project_id should be passed to GetByID: got %s, want %s", call.ProjectID, expectedAsset.ProjectID)
	}
}

func TestAssetService_GetAsset_CrossProjectRejected(t *testing.T) {
	ctx := context.Background()

	expectedAsset := &workspace.Asset{
		ID:              "asset-cross",
		ProjectID:       "project-owner",
		VisualizationID: "viz-owner",
		MIMEType:        "image/png",
		StorageKey:      "projects/project-owner/viz/viz-owner/uuid",
		ByteSize:        100,
		ChecksumSHA256:  "checksum",
		CreatedAt:       time.Now(),
	}

	mockStore := &MockAssetStore{}
	mockRepo := &MockAssetRepository{
		GetByIDFunc: func(ctx context.Context, projectID, id string) (*workspace.Asset, error) {
			// Return nil if project doesn't match
			if projectID != expectedAsset.ProjectID {
				return nil, errors.New("asset not found")
			}
			return expectedAsset, nil
		},
	}
	service := NewAssetService(mockRepo, mockStore)

	// Try to get asset from different project
	_, _, err := service.GetAsset(ctx, "project-other", expectedAsset.ID)
	if err == nil {
		t.Error("expected error for cross-project access")
	}
}

func TestAssetService_ListAssets(t *testing.T) {
	ctx := context.Background()

	projectID := "project-list"
	visualizationID := "viz-list"

	expectedAssets := []*workspace.Asset{
		{
			ID:              "asset-1",
			ProjectID:       projectID,
			VisualizationID: visualizationID,
			MIMEType:        "image/png",
			StorageKey:      "key-1",
			ByteSize:        100,
			ChecksumSHA256:  "checksum-1",
			CreatedAt:       time.Now(),
		},
		{
			ID:              "asset-2",
			ProjectID:       projectID,
			VisualizationID: visualizationID,
			MIMEType:        "application/json",
			StorageKey:      "key-2",
			ByteSize:        200,
			ChecksumSHA256:  "checksum-2",
			CreatedAt:       time.Now(),
		},
	}

	mockStore := &MockAssetStore{}
	mockRepo := &MockAssetRepository{
		ListByVisualizationFunc: func(ctx context.Context, projID, vizID string) ([]*workspace.Asset, error) {
			if projID == projectID && vizID == visualizationID {
				return expectedAssets, nil
			}
			return nil, nil
		},
	}
	service := NewAssetService(mockRepo, mockStore)

	// List assets
	assets, err := service.ListAssets(ctx, projectID, visualizationID)
	if err != nil {
		t.Fatalf("failed to list assets: %v", err)
	}

	if len(assets) != len(expectedAssets) {
		t.Errorf("expected %d assets, got %d", len(expectedAssets), len(assets))
	}
}

func TestAssetService_DeleteAsset(t *testing.T) {
	ctx := context.Background()

	projectID := "project-del"
	assetID := "asset-del"
	storageKey := "projects/project-del/viz/viz-del/uuid"

	mockStore := &MockAssetStore{}
	mockRepo := &MockAssetRepository{
		GetByIDFunc: func(ctx context.Context, projectID, id string) (*workspace.Asset, error) {
			return &workspace.Asset{
				ID:         assetID,
				ProjectID:  projectID,
				StorageKey: storageKey,
			}, nil
		},
		DeleteFunc: func(ctx context.Context, projectID, id string) error {
			return nil
		},
	}
	service := NewAssetService(mockRepo, mockStore)

	// Delete asset
	err := service.DeleteAsset(ctx, projectID, assetID)
	if err != nil {
		t.Fatalf("failed to delete asset: %v", err)
	}

	// Verify both store delete and repo delete were called
	if len(mockStore.DeleteCalls) != 1 || mockStore.DeleteCalls[0] != storageKey {
		t.Errorf("store Delete should be called with storage key: %s", storageKey)
	}
}

func TestAssetService_RegisterRetainedAssets_OnlyFinalAndExport(t *testing.T) {
	ctx := context.Background()

	mockStore := &MockAssetStore{}
	mockRepo := &MockAssetRepository{}
	service := NewAssetService(mockRepo, mockStore)

	projectID := "project-filter"
	visualizationID := "viz-filter"

	// Mix of artifact kinds
	artifacts := []agent.Artifact{
		{
			ID:       "artifact-final",
			Kind:     agent.ArtifactKindRenderedFigure, // Retained
			MIMEType: "image/png",
			Bytes:    []byte("final-data"),
		},
		{
			ID:       "artifact-intermediate",
			Kind:     agent.ArtifactKindPromptTrace, // Retained for audit
			MIMEType: "application/json",
			Content:  `{"trace": true}`,
		},
		{
			ID:       "artifact-critique",
			Kind:     agent.ArtifactKindCritique, // Also retained
			MIMEType: "text/plain",
			Content:  "critique content",
		},
		{
			ID:       "artifact-references",
			Kind:     agent.ArtifactKindReferenceBundle, // Not retained
			MIMEType: "application/json",
			Content:  `{"refs": []}`,
		},
		{
			ID:       "artifact-plan",
			Kind:     agent.ArtifactKindPlan, // Not retained as asset
			MIMEType: "text/plain",
			Content:  "plan content",
		},
	}

	result, err := service.RegisterRetainedAssets(ctx, projectID, visualizationID, nil, artifacts)
	if err != nil {
		t.Fatalf("failed to register assets: %v", err)
	}

	// Should only register: rendered_figure, prompt_trace, critique
	expectedCount := 3
	if len(result) != expectedCount {
		t.Errorf("expected %d retained assets, got %d", expectedCount, len(result))
	}
}

func TestAssetService_GetAssetByStorageKey(t *testing.T) {
	ctx := context.Background()

	storageKey := "projects/project-storage/viz/viz-storage/uuid"
	expectedAsset := &workspace.Asset{
		ID:              "asset-storage",
		ProjectID:       "project-storage",
		VisualizationID: "viz-storage",
		MIMEType:        "image/png",
		StorageKey:      storageKey,
		ByteSize:        100,
		ChecksumSHA256:  "checksum",
		CreatedAt:       time.Now(),
	}

	mockStore := &MockAssetStore{
		ReadFunc: func(ctx context.Context, key string) ([]byte, error) {
			return []byte("image-data"), nil
		},
	}
	mockRepo := &MockAssetRepository{
		GetByStorageKeyFunc: func(ctx context.Context, key string) (*workspace.Asset, error) {
			if key == storageKey {
				return expectedAsset, nil
			}
			return nil, errors.New("not found")
		},
	}
	service := NewAssetService(mockRepo, mockStore)

	// Get asset by storage key
	asset, data, err := service.GetAssetByStorageKey(ctx, storageKey)
	if err != nil {
		t.Fatalf("failed to get asset by storage key: %v", err)
	}

	if asset.StorageKey != storageKey {
		t.Errorf("storage key mismatch: got %s, want %s", asset.StorageKey, storageKey)
	}
	if string(data) != "image-data" {
		t.Errorf("data mismatch: got %s", string(data))
	}
}

func TestAssetService_RegisterRetainedAssets_EmptyArtifacts(t *testing.T) {
	ctx := context.Background()

	mockStore := &MockAssetStore{}
	mockRepo := &MockAssetRepository{}
	service := NewAssetService(mockRepo, mockStore)

	// Empty artifacts
	result, err := service.RegisterRetainedAssets(ctx, "project", "viz", nil, []agent.Artifact{})
	if err != nil {
		t.Fatalf("failed to register empty assets: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("expected 0 assets for empty artifacts, got %d", len(result))
	}
}

func TestAssetService_RegisterRetainedAssets_StoreFailure(t *testing.T) {
	ctx := context.Background()

	mockStore := &MockAssetStore{
		WriteFunc: func(ctx context.Context, storageKey string, data []byte) error {
			return errors.New("disk full")
		},
	}
	mockRepo := &MockAssetRepository{}
	service := NewAssetService(mockRepo, mockStore)

	artifacts := []agent.Artifact{
		{
			ID:       "artifact-fail",
			Kind:     agent.ArtifactKindRenderedFigure,
			MIMEType: "image/png",
			Bytes:    []byte("data"),
		},
	}

	_, err := service.RegisterRetainedAssets(ctx, "project", "viz", nil, artifacts)
	if err == nil {
		t.Error("expected error when store fails")
	}
}

func TestAssetService_GetAsset_WithContent(t *testing.T) {
	ctx := context.Background()

	// Test artifact with Content field (string) instead of Bytes
	artifacts := []agent.Artifact{
		{
			ID:       "artifact-content",
			Kind:     agent.ArtifactKindPromptTrace,
			MIMEType: "application/json",
			Content:  `{"prompt": "test content"}`,
		},
	}

	mockStore := &MockAssetStore{
		WriteFunc: func(ctx context.Context, key string, data []byte) error {
			// Verify Content is converted to bytes
			if !bytes.Contains(data, []byte("test content")) {
				t.Error("expected Content to be converted to bytes")
			}
			return nil
		},
	}
	mockRepo := &MockAssetRepository{}
	service := NewAssetService(mockRepo, mockStore)

	result, err := service.RegisterRetainedAssets(ctx, "project", "viz", nil, artifacts)
	if err != nil {
		t.Fatalf("failed to register assets: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("expected 1 asset, got %d", len(result))
	}
}
