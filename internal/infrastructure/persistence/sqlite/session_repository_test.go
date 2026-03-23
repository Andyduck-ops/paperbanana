package sqlite

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
	"github.com/paperbanana/paperbanana/internal/domain/workspace"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionRepository_RoundTripsSessionState(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "test.db")

	result, err := Bootstrap(ctx, BootstrapConfig{
		DatabasePath:      dbPath,
		EnableForeignKeys: true,
		BusyTimeoutMs:     5000,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	defer Close(result.DB)

	// Create a parent project and visualization first
	projectID := uuid.NewString()
	vizID := uuid.NewString()
	sessionID := uuid.NewString()

	require.NoError(t, result.DB.Create(&ProjectModel{
		ID:        projectID,
		Name:      "Test Project",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}).Error)

	require.NoError(t, result.DB.Create(&VisualizationModel{
		ID:        vizID,
		ProjectID: projectID,
		Name:      "Test Visualization",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}).Error)

	repo := NewSessionRepository(result.DB)

	// Create a session with full SessionState payload
	now := time.Now().UTC()
	session := &workspace.SessionRecord{
		ID:            sessionID,
		ProjectID:     projectID,
		VisualizationID: &vizID,
		Status:        string(domainagent.StatusCompleted),
		CurrentStage:  string(domainagent.StageCritic),
		SchemaVersion: "1.0.0",
		Snapshot: &domainagent.SessionState{
			SchemaVersion: "1.0.0",
			SessionID:     sessionID,
			RequestID:     "request-123",
			Status:        domainagent.StatusCompleted,
			CurrentStage:  domainagent.StageCritic,
			Pipeline:      domainagent.CanonicalPipeline(),
			InitialInput: domainagent.AgentInput{
				SessionID: sessionID,
				RequestID: "request-123",
				Content:   "Create a bar chart",
				VisualIntent: domainagent.VisualIntent{
					Mode: domainagent.VisualModePlot,
					Goal: "Create a bar chart",
				},
			},
			StageStates: []domainagent.AgentState{
				{
					Stage:  domainagent.StageRetriever,
					Status: domainagent.StatusCompleted,
				},
			},
			FinalOutput: domainagent.AgentOutput{
				Stage:   domainagent.StageCritic,
				Content: "Final visualization output",
			},
			StartedAt:   now,
			UpdatedAt:   now,
			CompletedAt: now,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Create the session
	err = repo.Create(ctx, session)
	require.NoError(t, err)

	// Retrieve by ID
	loaded, err := repo.GetByID(ctx, sessionID)
	require.NoError(t, err)
	require.NotNil(t, loaded)

	assert.Equal(t, sessionID, loaded.ID)
	assert.Equal(t, projectID, loaded.ProjectID)
	assert.Equal(t, vizID, *loaded.VisualizationID)
	assert.Equal(t, string(domainagent.StatusCompleted), loaded.Status)
	assert.Equal(t, string(domainagent.StageCritic), loaded.CurrentStage)
	assert.Equal(t, "1.0.0", loaded.SchemaVersion)

	// Verify the full snapshot round-tripped
	require.NotNil(t, loaded.Snapshot)
	assert.Equal(t, sessionID, loaded.Snapshot.SessionID)
	assert.Equal(t, "request-123", loaded.Snapshot.RequestID)
	assert.Equal(t, domainagent.StatusCompleted, loaded.Snapshot.Status)
	assert.Equal(t, domainagent.StageCritic, loaded.Snapshot.CurrentStage)
	assert.Equal(t, "Create a bar chart", loaded.Snapshot.InitialInput.Content)
	assert.Len(t, loaded.Snapshot.Pipeline, 5) // retriever, planner, stylist, visualizer, critic
	require.Len(t, loaded.Snapshot.StageStates, 1)
	assert.Equal(t, domainagent.StageRetriever, loaded.Snapshot.StageStates[0].Stage)
	assert.Equal(t, "Final visualization output", loaded.Snapshot.FinalOutput.Content)
}

func TestSessionRepository_GetByVisualization(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "test.db")

	result, err := Bootstrap(ctx, BootstrapConfig{
		DatabasePath:      dbPath,
		EnableForeignKeys: true,
		BusyTimeoutMs:     5000,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	defer Close(result.DB)

	// Create project and visualization
	projectID := uuid.NewString()
	vizID := uuid.NewString()

	require.NoError(t, result.DB.Create(&ProjectModel{
		ID:        projectID,
		Name:      "Test Project",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}).Error)

	require.NoError(t, result.DB.Create(&VisualizationModel{
		ID:        vizID,
		ProjectID: projectID,
		Name:      "Test Visualization",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}).Error)

	repo := NewSessionRepository(result.DB)

	// Create multiple sessions
	now := time.Now().UTC()
	for i := 0; i < 3; i++ {
		sessionID := uuid.NewString()
		session := &workspace.SessionRecord{
			ID:              sessionID,
			ProjectID:       projectID,
			VisualizationID: &vizID,
			Status:          string(domainagent.StatusCompleted),
			CurrentStage:    string(domainagent.StageCritic),
			SchemaVersion:   "1.0.0",
			Snapshot: &domainagent.SessionState{
				SchemaVersion: "1.0.0",
				SessionID:     sessionID,
				Status:        domainagent.StatusCompleted,
				CurrentStage:  domainagent.StageCritic,
				StartedAt:     now.Add(time.Duration(i) * time.Hour),
				UpdatedAt:     now.Add(time.Duration(i) * time.Hour),
			},
			CreatedAt: now.Add(time.Duration(i) * time.Hour),
			UpdatedAt: now.Add(time.Duration(i) * time.Hour),
		}
		require.NoError(t, repo.Create(ctx, session))
	}

	// List sessions by visualization
	sessions, err := repo.GetByVisualization(ctx, projectID, vizID, 10)
	require.NoError(t, err)
	assert.Len(t, sessions, 3)

	// Verify ordering (most recent first)
	assert.True(t, sessions[0].CreatedAt.After(sessions[1].CreatedAt))
	assert.True(t, sessions[1].CreatedAt.After(sessions[2].CreatedAt))
}

func TestSessionRepository_Update(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "test.db")

	result, err := Bootstrap(ctx, BootstrapConfig{
		DatabasePath:      dbPath,
		EnableForeignKeys: true,
		BusyTimeoutMs:     5000,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	defer Close(result.DB)

	projectID := uuid.NewString()
	sessionID := uuid.NewString()

	require.NoError(t, result.DB.Create(&ProjectModel{
		ID:        projectID,
		Name:      "Test Project",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}).Error)

	repo := NewSessionRepository(result.DB)

	now := time.Now().UTC()
	session := &workspace.SessionRecord{
		ID:            sessionID,
		ProjectID:     projectID,
		Status:        string(domainagent.StatusRunning),
		CurrentStage:  string(domainagent.StagePlanner),
		SchemaVersion: "1.0.0",
		Snapshot: &domainagent.SessionState{
			SchemaVersion: "1.0.0",
			SessionID:     sessionID,
			Status:        domainagent.StatusRunning,
			CurrentStage:  domainagent.StagePlanner,
			StartedAt:     now,
			UpdatedAt:     now,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	require.NoError(t, repo.Create(ctx, session))

	// Update the session
	session.Status = string(domainagent.StatusCompleted)
	session.CurrentStage = string(domainagent.StageCritic)
	session.Snapshot.Status = domainagent.StatusCompleted
	session.Snapshot.CurrentStage = domainagent.StageCritic
	session.UpdatedAt = now.Add(time.Hour)

	require.NoError(t, repo.Update(ctx, session))

	// Verify update
	loaded, err := repo.GetByID(ctx, sessionID)
	require.NoError(t, err)
	assert.Equal(t, string(domainagent.StatusCompleted), loaded.Status)
	assert.Equal(t, string(domainagent.StageCritic), loaded.CurrentStage)
}
