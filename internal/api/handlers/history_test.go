package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
	domainworkspace "github.com/paperbanana/paperbanana/internal/domain/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// mockHistoryService implements HistoryService for testing.
type mockHistoryService struct {
	versions       map[string]*domainworkspace.VisualizationVersion
	sessions       map[string]*domainworkspace.SessionRecord
	recentSessions []*domainworkspace.SessionRecord
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

func (m *mockHistoryService) ListRecentSessions(ctx context.Context, projectID string, limit int) ([]*domainworkspace.SessionRecord, error) {
	if m.err != nil {
		return nil, m.err
	}

	sessions := append([]*domainworkspace.SessionRecord(nil), m.recentSessions...)
	if len(sessions) == 0 {
		for _, session := range m.sessions {
			sessions = append(sessions, session)
		}
	}

	filtered := sessions[:0]
	for _, session := range sessions {
		if projectID == "" || session.ProjectID == projectID {
			filtered = append(filtered, session)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})

	if limit > 0 && len(filtered) > limit {
		filtered = filtered[:limit]
	}

	return filtered, nil
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

	t.Run("returns empty list for missing project_id", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/history?visualization_id="+vizID, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response ListHistoryResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "", response.ProjectID)
		assert.Len(t, response.Versions, 0)
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

func TestHistoryHandler_ListRecentSessions(t *testing.T) {
	projectOneID := uuid.NewString()
	projectTwoID := uuid.NewString()
	vizID := uuid.NewString()
	now := time.Now().UTC()

	mock := &mockHistoryService{
		recentSessions: []*domainworkspace.SessionRecord{
			{
				ID:              uuid.NewString(),
				ProjectID:       projectOneID,
				VisualizationID: &vizID,
				Status:          string(domainagent.StatusCompleted),
				CurrentStage:    string(domainagent.StageCritic),
				SchemaVersion:   "1.0.0",
				Snapshot: &domainagent.SessionState{
					SessionID:    uuid.NewString(),
					CurrentStage: domainagent.StageCritic,
					InitialInput: domainagent.AgentInput{
						Content: "Paper Context: transformer compression.",
						VisualIntent: domainagent.VisualIntent{
							Goal: "Draw a narrow paper workflow figure",
						},
						Metadata: map[string]string{"http.visual_intent": "Draw a narrow paper workflow figure"},
					},
					StageStates: []domainagent.AgentState{
						{
							Stage: domainagent.StageVisualizer,
							Output: domainagent.AgentOutput{
								Content:  "Rendered a concise workflow figure.",
								Metadata: map[string]string{"summary": "Rendered a concise workflow figure."},
							},
						},
					},
					FinalOutput: domainagent.AgentOutput{
						Content:  "Final figure ready.",
						Metadata: map[string]string{"summary": "Final figure ready."},
					},
				},
				CreatedAt:   now,
				UpdatedAt:   now,
				CompletedAt: &now,
			},
			{
				ID:            uuid.NewString(),
				ProjectID:     projectTwoID,
				Status:        string(domainagent.StatusRunning),
				CurrentStage:  string(domainagent.StageVisualizer),
				SchemaVersion: "1.0.0",
				Snapshot: &domainagent.SessionState{
					SessionID:    uuid.NewString(),
					CurrentStage: domainagent.StageVisualizer,
					InitialInput: domainagent.AgentInput{
						Metadata: map[string]string{"http.visual_intent": "Plot the ablation comparison"},
					},
				},
				CreatedAt: now.Add(-time.Hour),
				UpdatedAt: now.Add(-time.Hour),
			},
		},
	}

	router, handler := setupHistoryTest(t, mock)
	router.GET("/sessions/recent", handler.ListRecentSessions)

	t.Run("returns recent sessions globally", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/sessions/recent?limit=5", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response ListRecentSessionsResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.Len(t, response.Sessions, 2)
		assert.Equal(t, "", response.ProjectID)
		assert.Equal(t, projectOneID, response.Sessions[0].ProjectID)
		assert.Equal(t, "Draw a narrow paper workflow figure", response.Sessions[0].Prompt)
		assert.Equal(t, "Final figure ready.", response.Sessions[0].Summary)
		require.NotNil(t, response.Sessions[0].CompletedAt)
	})

	t.Run("filters recent sessions by project", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/sessions/recent?project_id="+projectTwoID+"&limit=5", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response ListRecentSessionsResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.Len(t, response.Sessions, 1)
		assert.Equal(t, projectTwoID, response.ProjectID)
		assert.Equal(t, "Plot the ablation comparison", response.Sessions[0].Prompt)
	})

	t.Run("groups batch candidates into one recent history item", func(t *testing.T) {
		batchGroupID := "batch-root-001"
		candidateOneID := batchGroupID + "-candidate-0"
		candidateTwoID := batchGroupID + "-candidate-1"

		mock.recentSessions = []*domainworkspace.SessionRecord{
			{
				ID:            candidateOneID,
				ProjectID:     projectOneID,
				Status:        string(domainagent.StatusCompleted),
				CurrentStage:  string(domainagent.StageCritic),
				SchemaVersion: "1.0.0",
				Snapshot: &domainagent.SessionState{
					SessionID:    candidateOneID,
					CurrentStage: domainagent.StageCritic,
					InitialInput: domainagent.AgentInput{
						Content: "Paper Context: vision transformer ablation.",
						VisualIntent: domainagent.VisualIntent{
							Goal: "Generate three ablation figure variants",
						},
						Metadata: map[string]string{
							"http.visual_intent": "Generate three ablation figure variants",
							"batch.group_id":     batchGroupID,
							"batch.candidate_id": "0",
						},
					},
				},
				CreatedAt: now.Add(-2 * time.Minute),
				UpdatedAt: now.Add(-2 * time.Minute),
			},
			{
				ID:            candidateTwoID,
				ProjectID:     projectOneID,
				Status:        string(domainagent.StatusFailed),
				CurrentStage:  string(domainagent.StageVisualizer),
				SchemaVersion: "1.0.0",
				Snapshot: &domainagent.SessionState{
					SessionID:    candidateTwoID,
					CurrentStage: domainagent.StageVisualizer,
					InitialInput: domainagent.AgentInput{
						Content: "Paper Context: vision transformer ablation.",
						VisualIntent: domainagent.VisualIntent{
							Goal: "Generate three ablation figure variants",
						},
						Metadata: map[string]string{
							"http.visual_intent": "Generate three ablation figure variants",
							"batch.group_id":     batchGroupID,
							"batch.candidate_id": "1",
						},
					},
				},
				CreatedAt: now.Add(-3 * time.Minute),
				UpdatedAt: now.Add(-3 * time.Minute),
			},
		}

		req := httptest.NewRequest("GET", "/sessions/recent?project_id="+projectOneID+"&limit=5", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response ListRecentSessionsResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.Len(t, response.Sessions, 1)
		assert.Equal(t, "batch", response.Sessions[0].Mode)
		assert.Equal(t, batchGroupID, response.Sessions[0].ID)
		assert.Equal(t, batchGroupID, response.Sessions[0].BatchID)
		assert.Equal(t, []string{candidateOneID, candidateTwoID}, response.Sessions[0].CandidateSessionIDs)
		assert.Equal(t, string(domainagent.StatusCompleted), response.Sessions[0].Status)
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
			CreatedAt:   now,
			UpdatedAt:   now,
			CompletedAt: &now,
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
		require.NotNil(t, response.CompletedAt)
		assert.Equal(t, now.Format("2006-01-02T15:04:05Z"), *response.CompletedAt)
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
				CreatedAt:   now,
				UpdatedAt:   now,
				CompletedAt: &now,
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
		require.NotNil(t, response.CompletedAt)
		assert.Equal(t, now.Format("2006-01-02T15:04:05Z"), *response.CompletedAt)
	})

	t.Run("returns not found for missing session", func(t *testing.T) {
		missingSessionID := uuid.NewString()
		req := httptest.NewRequest("GET", "/session/"+missingSessionID, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
