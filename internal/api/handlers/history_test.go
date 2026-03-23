package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
	domainworkspace "github.com/paperbanana/paperbanana/internal/domain/workspace"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// mockHistoryService implements HistoryService for testing.
type mockHistoryService struct {
	versions       map[string]*domainworkspace.VisualizationVersion
	sessions       map[string]*domainworkspace.SessionRecord
	latestSession  *domainworkspace.SessionRecord
	err            error
}

func (m *mockHistoryService) ListHistory(ctx context.Context, projectID, visualizationID string, limit int) ([]*domainworkspace.VisualizationVersion, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []*domainworkspace.VisualizationVersion
	for _, v := range m.versions {
		if v.ProjectID == projectID && v.VisualizationID == visualizationID {
			result = append(result, v)
		}
	}
	return result, nil
}

func (m *mockHistoryService) GetVersion(ctx context.Context, projectID, versionID string) (*domainworkspace.VisualizationVersion, error) {
	if m.err != nil {
		return nil, m.err
	}
	if v, ok := m.versions[versionID]; ok && v.ProjectID == projectID {
		return v, nil
	}
	return nil, errors.New("version not found")
}

func (m *mockHistoryService) GetLatestSession(ctx context.Context, projectID, visualizationID string) (*domainworkspace.SessionRecord, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.latestSession != nil {
		return m.latestSession, nil
	}
	return nil, errors.New("session not found")
}

func (m *mockHistoryService) GetSessionByID(ctx context.Context, sessionID string) (*domainworkspace.SessionRecord, error) {
	if m.err != nil {
		return nil, m.err
	}
	if s, ok := m.sessions[sessionID]; ok {
		return s, nil
	}
	return nil, errors.New("session not found")
}

func setupHistoryTest(t *testing.T, mock *mockHistoryService) (*gin.Engine, *HistoryHandler) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	handler := NewHistoryHandler(mock, logger)
	router := gin.New()
	return router, handler
}

func TestHistoryHandler_ListHistory(t *testing.T) {
	projectID := uuid.NewString()
	vizID := uuid.NewString()
	versionID := uuid.NewString()

	mock := &mockHistoryService{
		versions: map[string]*domainworkspace.VisualizationVersion{
			versionID: {
				ID:              versionID,
				VisualizationID: vizID,
				ProjectID:       projectID,
				VersionNumber:   1,
				Summary:         "First version",
				CreatedAt:       time.Now().UTC(),
			},
		},
	}

	router, handler := setupHistoryTest(t, mock)
	router.GET("/history", handler.ListHistory)

	t.Run("returns history for valid request", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/history?project_id="+projectID+"&visualization_id="+vizID, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response ListHistoryResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, projectID, response.ProjectID)
		assert.Len(t, response.Versions, 1)
		assert.Equal(t, 1, response.Versions[0].VersionNumber)
	})

	t.Run("returns error for missing project_id", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/history?visualization_id="+vizID, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns empty list when visualization_id is missing", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/history?project_id="+projectID, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response ListHistoryResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, projectID, response.ProjectID)
		assert.Len(t, response.Versions, 0)
	})
}

func TestHistoryHandler_GetVersion(t *testing.T) {
	projectID := uuid.NewString()
	versionID := uuid.NewString()
	vizID := uuid.NewString()

	mock := &mockHistoryService{
		versions: map[string]*domainworkspace.VisualizationVersion{
			versionID: {
				ID:              versionID,
				VisualizationID: vizID,
				ProjectID:       projectID,
				VersionNumber:   1,
				Summary:         "First version",
				CreatedAt:       time.Now().UTC(),
			},
		},
	}

	router, handler := setupHistoryTest(t, mock)
	router.GET("/history/:project_id/:version_id", handler.GetVersion)

	t.Run("returns version for valid request", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/history/"+projectID+"/"+versionID, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response GetVersionResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, versionID, response.ID)
		assert.Equal(t, 1, response.VersionNumber)
		assert.Equal(t, "First version", response.Summary)
	})

	t.Run("returns not found for missing version", func(t *testing.T) {
		missingVersionID := uuid.NewString()
		req := httptest.NewRequest("GET", "/history/"+projectID+"/"+missingVersionID, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestHistoryHandler_GetLatestSession(t *testing.T) {
	projectID := uuid.NewString()
	vizID := uuid.NewString()
	sessionID := uuid.NewString()

	now := time.Now().UTC()
	mock := &mockHistoryService{
		latestSession: &domainworkspace.SessionRecord{
			ID:              sessionID,
			ProjectID:       projectID,
			VisualizationID: &vizID,
			Status:          string(domainagent.StatusCompleted),
			CurrentStage:    string(domainagent.StageCritic),
			SchemaVersion:   "1.0.0",
			Snapshot: &domainagent.SessionState{
				SessionID:     sessionID,
				SchemaVersion: "1.0.0",
				Status:        domainagent.StatusCompleted,
				CurrentStage:  domainagent.StageCritic,
				StartedAt:     now,
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	router, handler := setupHistoryTest(t, mock)
	router.GET("/session/latest", handler.GetLatestSession)

	t.Run("returns session for valid request", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/session/latest?project_id="+projectID+"&visualization_id="+vizID, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response SessionResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, sessionID, response.ID)
		assert.Equal(t, string(domainagent.StatusCompleted), response.Status)
		assert.Equal(t, string(domainagent.StageCritic), response.CurrentStage)
	})

	t.Run("returns not found when no session exists", func(t *testing.T) {
		mock.err = errors.New("session not found")
		req := httptest.NewRequest("GET", "/session/latest?project_id="+projectID+"&visualization_id="+vizID, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("returns error for missing project_id", func(t *testing.T) {
		mock.err = nil
		req := httptest.NewRequest("GET", "/session/latest?visualization_id="+vizID, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestHistoryHandler_GetSession(t *testing.T) {
	projectID := uuid.NewString()
	vizID := uuid.NewString()
	sessionID := uuid.NewString()

	now := time.Now().UTC()
	mock := &mockHistoryService{
		sessions: map[string]*domainworkspace.SessionRecord{
			sessionID: {
				ID:              sessionID,
				ProjectID:       projectID,
				VisualizationID: &vizID,
				Status:          string(domainagent.StatusFailed),
				CurrentStage:    string(domainagent.StagePlanner),
				SchemaVersion:   "1.0.0",
				Snapshot: &domainagent.SessionState{
					SessionID:     sessionID,
					SchemaVersion: "1.0.0",
					Status:        domainagent.StatusFailed,
					CurrentStage:  domainagent.StagePlanner,
					StartedAt:     now,
				},
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
	}

	router, handler := setupHistoryTest(t, mock)
	router.GET("/session/:session_id", handler.GetSession)

	t.Run("returns session for valid request", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/session/"+sessionID, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response SessionResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, sessionID, response.ID)
		assert.Equal(t, string(domainagent.StatusFailed), response.Status)
		assert.Equal(t, string(domainagent.StagePlanner), response.CurrentStage)
	})

	t.Run("returns not found for missing session", func(t *testing.T) {
		missingSessionID := uuid.NewString()
		req := httptest.NewRequest("GET", "/session/"+missingSessionID, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
