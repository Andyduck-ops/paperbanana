package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// mockAssetService implements AssetService for testing.
type mockAssetService struct {
	assets    map[string]*mockAsset
	listError error
	getError  error
}

type mockAsset struct {
	info  *AssetInfo
	bytes []byte
}

func (m *mockAssetService) ListAssets(ctx context.Context, projectID, visualizationID string) ([]*AssetInfo, error) {
	if m.listError != nil {
		return nil, m.listError
	}
	var result []*AssetInfo
	for _, a := range m.assets {
		if a.info.ProjectID == projectID && a.info.VisualizationID == visualizationID {
			result = append(result, a.info)
		}
	}
	return result, nil
}

func (m *mockAssetService) GetAsset(ctx context.Context, projectID, assetID string) (*AssetInfo, []byte, error) {
	if m.getError != nil {
		return nil, nil, m.getError
	}
	if a, ok := m.assets[assetID]; ok && a.info.ProjectID == projectID {
		return a.info, a.bytes, nil
	}
	return nil, nil, errors.New("asset not found")
}

func (m *mockAssetService) ListAssetsByVersion(ctx context.Context, projectID, versionID string) ([]*AssetInfo, error) {
	if m.listError != nil {
		return nil, m.listError
	}
	var result []*AssetInfo
	for _, a := range m.assets {
		if a.info.ProjectID == projectID && a.info.VersionID != nil && *a.info.VersionID == versionID {
			result = append(result, a.info)
		}
	}
	return result, nil
}

func setupAssetTest(t *testing.T, mock *mockAssetService) (*gin.Engine, *AssetHandler) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	handler := NewAssetHandler(mock, logger)
	router := gin.New()
	return router, handler
}

func TestAssetHandler_ListAssets(t *testing.T) {
	projectID := uuid.NewString()
	vizID := uuid.NewString()
	assetID := uuid.NewString()
	now := time.Now().UTC()

	mock := &mockAssetService{
		assets: map[string]*mockAsset{
			assetID: {
				info: &AssetInfo{
					ID:              assetID,
					ProjectID:       projectID,
					VisualizationID: vizID,
					MIMEType:        "image/png",
					ByteSize:        1024,
					ChecksumSHA256:  "abc123",
					CreatedAt:       now,
				},
				bytes: []byte("test image data"),
			},
		},
	}

	router, handler := setupAssetTest(t, mock)
	router.GET("/assets", handler.ListAssets)

	t.Run("returns assets for valid request", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/assets?project_id="+projectID+"&visualization_id="+vizID, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response ListAssetsResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, projectID, response.ProjectID)
		assert.Len(t, response.Assets, 1)
		assert.Equal(t, assetID, response.Assets[0].ID)
		assert.Equal(t, "image/png", response.Assets[0].MIMEType)
		assert.Equal(t, int64(1024), response.Assets[0].ByteSize)
		assert.Equal(t, "abc123", response.Assets[0].ChecksumSHA256)
		// Storage key should NOT be exposed
		assert.Empty(t, response.Assets[0].StorageKey)
	})

	t.Run("returns error for missing project_id", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/assets?visualization_id="+vizID, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns error for missing visualization_id", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/assets?project_id="+projectID, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns empty list when no assets", func(t *testing.T) {
		emptyMock := &mockAssetService{assets: map[string]*mockAsset{}}
		router2, handler2 := setupAssetTest(t, emptyMock)
		router2.GET("/assets", handler2.ListAssets)

		req := httptest.NewRequest("GET", "/assets?project_id="+projectID+"&visualization_id="+vizID, nil)
		w := httptest.NewRecorder()
		router2.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response ListAssetsResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Empty(t, response.Assets)
	})
}

func TestAssetHandler_GetAsset(t *testing.T) {
	projectID := uuid.NewString()
	vizID := uuid.NewString()
	assetID := uuid.NewString()
	now := time.Now().UTC()

	imageData := []byte("test image binary data")

	mock := &mockAssetService{
		assets: map[string]*mockAsset{
			assetID: {
				info: &AssetInfo{
					ID:              assetID,
					ProjectID:       projectID,
					VisualizationID: vizID,
					MIMEType:        "image/png",
					ByteSize:        int64(len(imageData)),
					ChecksumSHA256:  "sha256hash",
					CreatedAt:       now,
				},
				bytes: imageData,
			},
		},
	}

	router, handler := setupAssetTest(t, mock)
	router.GET("/assets/:project_id/:asset_id", handler.GetAsset)

	t.Run("returns asset metadata for valid request", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/assets/"+projectID+"/"+assetID, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response AssetResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, assetID, response.ID)
		assert.Equal(t, projectID, response.ProjectID)
		assert.Equal(t, vizID, response.VisualizationID)
		assert.Equal(t, "image/png", response.MIMEType)
		assert.Equal(t, int64(len(imageData)), response.ByteSize)
		assert.Equal(t, "sha256hash", response.ChecksumSHA256)
		// Storage key should NOT be exposed
		assert.Empty(t, response.StorageKey)
	})

	t.Run("returns not found for missing asset", func(t *testing.T) {
		missingAssetID := uuid.NewString()
		req := httptest.NewRequest("GET", "/assets/"+projectID+"/"+missingAssetID, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("returns not found for cross-project access", func(t *testing.T) {
		differentProjectID := uuid.NewString()
		req := httptest.NewRequest("GET", "/assets/"+differentProjectID+"/"+assetID, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestAssetHandler_DownloadAsset(t *testing.T) {
	projectID := uuid.NewString()
	vizID := uuid.NewString()
	assetID := uuid.NewString()
	now := time.Now().UTC()

	imageData := []byte("test image binary data")

	mock := &mockAssetService{
		assets: map[string]*mockAsset{
			assetID: {
				info: &AssetInfo{
					ID:              assetID,
					ProjectID:       projectID,
					VisualizationID: vizID,
					MIMEType:        "image/png",
					ByteSize:        int64(len(imageData)),
					ChecksumSHA256:  "sha256hash",
					CreatedAt:       now,
				},
				bytes: imageData,
			},
		},
	}

	router, handler := setupAssetTest(t, mock)
	router.GET("/assets/:project_id/:asset_id/download", handler.DownloadAsset)

	t.Run("returns asset bytes with correct headers", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/assets/"+projectID+"/"+assetID+"/download", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "image/png", w.Header().Get("Content-Type"))
		assert.Equal(t, "attachment", w.Header().Get("Content-Disposition"))
		assert.Contains(t, w.Header().Get("Content-Disposition"), assetID)
		assert.Equal(t, string(imageData), w.Body.String())
	})

	t.Run("returns not found for missing asset", func(t *testing.T) {
		missingAssetID := uuid.NewString()
		req := httptest.NewRequest("GET", "/assets/"+projectID+"/"+missingAssetID+"/download", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("returns not found for cross-project access", func(t *testing.T) {
		differentProjectID := uuid.NewString()
		req := httptest.NewRequest("GET", "/assets/"+differentProjectID+"/"+assetID+"/download", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestAssetHandler_ListAssetsByVersion(t *testing.T) {
	projectID := uuid.NewString()
	vizID := uuid.NewString()
	versionID := uuid.NewString()
	assetID := uuid.NewString()
	now := time.Now().UTC()

	mock := &mockAssetService{
		assets: map[string]*mockAsset{
			assetID: {
				info: &AssetInfo{
					ID:              assetID,
					ProjectID:       projectID,
					VisualizationID: vizID,
					VersionID:       &versionID,
					MIMEType:        "image/png",
					ByteSize:        2048,
					ChecksumSHA256:  "def456",
					CreatedAt:       now,
				},
				bytes: []byte("versioned asset data"),
			},
		},
	}

	router, handler := setupAssetTest(t, mock)
	router.GET("/assets/version/:project_id/:version_id", handler.ListAssetsByVersion)

	t.Run("returns assets for valid version", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/assets/version/"+projectID+"/"+versionID, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response ListAssetsResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, projectID, response.ProjectID)
		assert.Len(t, response.Assets, 1)
		assert.Equal(t, assetID, response.Assets[0].ID)
		assert.Equal(t, versionID, *response.Assets[0].VersionID)
	})

	t.Run("returns empty list for missing version", func(t *testing.T) {
		missingVersionID := uuid.NewString()
		req := httptest.NewRequest("GET", "/assets/version/"+projectID+"/"+missingVersionID, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response ListAssetsResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Empty(t, response.Assets)
	})
}
