package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
	agentstate "github.com/paperbanana/paperbanana/internal/infrastructure/agentstate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestPersistentSnapshotStore_RoundTrip(t *testing.T) {
	t.Parallel()

	db := setupTestDBForSnapshot(t)
	sessionRepo := NewSessionRepository(db)
	store := NewPersistentSnapshotStore(sessionRepo)

	// Create test session and agent state
	input := domainagent.AgentInput{
		SessionID: "test-session-001",
		RequestID: "test-request-001",
		Content:   "Generate a test figure",
		VisualIntent: domainagent.VisualIntent{
			Mode:  domainagent.VisualModeDiagram,
			Goal:  "Test persistence",
			Style: "academic",
		},
		Prompt: domainagent.PromptMetadata{
			SystemInstruction: "Test instruction",
			Version:           "test-v1",
		},
	}

	state := domainagent.AgentState{
		Stage:  domainagent.StageRetriever,
		Status: domainagent.StatusCompleted,
		Timing: domainagent.Timing{
			StartedAt:   time.Date(2026, time.March, 17, 9, 0, 0, 0, time.UTC),
			CompletedAt: time.Date(2026, time.March, 17, 9, 0, 1, 0, time.UTC),
			Duration:    time.Second,
		},
		Input:  input,
		Output: domainagent.AgentOutput{Stage: domainagent.StageRetriever, Content: "retriever output"},
	}

	session := domainagent.SessionState{
		SchemaVersion: "test-schema-v1",
		SessionID:     input.SessionID,
		RequestID:     input.RequestID,
		Status:        domainagent.StatusRunning,
		CurrentStage:  domainagent.StageRetriever,
		Pipeline:      []domainagent.StageName{domainagent.StageRetriever, domainagent.StagePlanner},
		InitialInput:  input,
		StageStates:   []domainagent.AgentState{state},
		StartedAt:     time.Date(2026, time.March, 17, 9, 0, 0, 0, time.UTC),
		UpdatedAt:     time.Date(2026, time.March, 17, 9, 0, 1, 0, time.UTC),
		Metadata:      map[string]string{"test_key": "test_value"},
	}

	// Save the snapshot
	err := store.Save(session, state)
	require.NoError(t, err, "Save should succeed")

	// Restore the snapshot
	snapshot, err := store.Restore(input.SessionID, domainagent.StageRetriever)
	require.NoError(t, err, "Restore should succeed")

	// Verify restored session state
	assert.Equal(t, session.SchemaVersion, snapshot.Session.SchemaVersion)
	assert.Equal(t, session.SessionID, snapshot.Session.SessionID)
	assert.Equal(t, session.RequestID, snapshot.Session.RequestID)
	assert.Equal(t, session.Status, snapshot.Session.Status)
	assert.Equal(t, session.CurrentStage, snapshot.Session.CurrentStage)
	assert.Equal(t, session.Pipeline, snapshot.Session.Pipeline)
	assert.Equal(t, session.Metadata["test_key"], snapshot.Session.Metadata["test_key"])

	// Verify restored stage state
	assert.Equal(t, state.Stage, snapshot.Stage.Stage)
	assert.Equal(t, state.Status, snapshot.Stage.Status)
	assert.Equal(t, state.Output.Content, snapshot.Stage.Output.Content)
	assert.Equal(t, state.Timing.StartedAt, snapshot.Stage.Timing.StartedAt)
	assert.Equal(t, state.Timing.CompletedAt, snapshot.Stage.Timing.CompletedAt)
}

func TestPersistentSnapshotStore_RestoreNotFound(t *testing.T) {
	t.Parallel()

	db := setupTestDBForSnapshot(t)
	sessionRepo := NewSessionRepository(db)
	store := NewPersistentSnapshotStore(sessionRepo)

	// Try to restore non-existent snapshot
	_, err := store.Restore("non-existent-session", domainagent.StageRetriever)
	require.Error(t, err, "Restore should fail for non-existent session")
	assert.ErrorIs(t, err, agentstate.ErrSnapshotNotFound)
}

func TestPersistentSnapshotStore_SaveUpdatesExistingSession(t *testing.T) {
	t.Parallel()

	db := setupTestDBForSnapshot(t)
	sessionRepo := NewSessionRepository(db)
	store := NewPersistentSnapshotStore(sessionRepo)

	input := domainagent.AgentInput{
		SessionID: "update-session-001",
		RequestID: "update-request-001",
		Content:   "Initial content",
		VisualIntent: domainagent.VisualIntent{
			Mode:  domainagent.VisualModeDiagram,
			Goal:  "Test update",
			Style: "academic",
		},
	}

	// Create initial session with retriever completed
	retrieverState := domainagent.AgentState{
		Stage:  domainagent.StageRetriever,
		Status: domainagent.StatusCompleted,
		Timing: domainagent.Timing{
			StartedAt:   time.Date(2026, time.March, 17, 9, 0, 0, 0, time.UTC),
			CompletedAt: time.Date(2026, time.March, 17, 9, 0, 1, 0, time.UTC),
		},
		Input: input,
	}

	session := domainagent.SessionState{
		SchemaVersion: "test-schema-v1",
		SessionID:     input.SessionID,
		RequestID:     input.RequestID,
		Status:        domainagent.StatusRunning,
		CurrentStage:  domainagent.StageRetriever,
		Pipeline:      []domainagent.StageName{domainagent.StageRetriever, domainagent.StagePlanner},
		InitialInput:  input,
		StageStates:   []domainagent.AgentState{retrieverState},
		StartedAt:     time.Date(2026, time.March, 17, 9, 0, 0, 0, time.UTC),
		UpdatedAt:     time.Date(2026, time.March, 17, 9, 0, 1, 0, time.UTC),
	}

	// Save initial state
	err := store.Save(session, retrieverState)
	require.NoError(t, err)

	// Update session with planner state
	plannerState := domainagent.AgentState{
		Stage:  domainagent.StagePlanner,
		Status: domainagent.StatusCompleted,
		Timing: domainagent.Timing{
			StartedAt:   time.Date(2026, time.March, 17, 9, 0, 2, 0, time.UTC),
			CompletedAt: time.Date(2026, time.March, 17, 9, 0, 3, 0, time.UTC),
		},
		Input:  input,
		Output: domainagent.AgentOutput{Stage: domainagent.StagePlanner, Content: "plan output"},
	}

	// Update session with accumulated stage states
	session.CurrentStage = domainagent.StagePlanner
	session.Status = domainagent.StatusRunning
	session.UpdatedAt = time.Date(2026, time.March, 17, 9, 0, 2, 0, time.UTC)
	session.StageStates = []domainagent.AgentState{retrieverState, plannerState}

	err = store.Save(session, plannerState)
	require.NoError(t, err)

	// Restore planner snapshot
	snapshot, err := store.Restore(input.SessionID, domainagent.StagePlanner)
	require.NoError(t, err)

	assert.Equal(t, domainagent.StagePlanner, snapshot.Session.CurrentStage)
	assert.Equal(t, domainagent.StagePlanner, snapshot.Stage.Stage)
	assert.Equal(t, "plan output", snapshot.Stage.Output.Content)

	// Verify we can still restore retriever stage
	retrieverSnapshot, err := store.Restore(input.SessionID, domainagent.StageRetriever)
	require.NoError(t, err)
	assert.Equal(t, domainagent.StageRetriever, retrieverSnapshot.Stage.Stage)
}

func TestPersistentSnapshotStore_WithProjectScope(t *testing.T) {
	t.Parallel()

	db := setupTestDBForSnapshot(t)
	sessionRepo := NewSessionRepository(db)
	store := NewPersistentSnapshotStore(sessionRepo)

	ctx := context.Background()

	projectID := "test-project-001"
	visualizationID := "test-viz-001"

	input := domainagent.AgentInput{
		SessionID: "scoped-session-001",
		RequestID: "scoped-request-001",
		Content:   "Scoped content",
		VisualIntent: domainagent.VisualIntent{
			Mode:  domainagent.VisualModePlot,
			Goal:  "Test project scope",
			Style: "academic",
		},
	}

	state := domainagent.AgentState{
		Stage:  domainagent.StageCritic,
		Status: domainagent.StatusCompleted,
		Timing: domainagent.Timing{
			StartedAt:   time.Date(2026, time.March, 17, 9, 0, 4, 0, time.UTC),
			CompletedAt: time.Date(2026, time.March, 17, 9, 0, 5, 0, time.UTC),
		},
		Input:  input,
		Output: domainagent.AgentOutput{Stage: domainagent.StageCritic, Content: "final output"},
	}

	session := domainagent.SessionState{
		SchemaVersion: "test-schema-v1",
		SessionID:     input.SessionID,
		RequestID:     input.RequestID,
		Status:        domainagent.StatusCompleted,
		CurrentStage:  domainagent.StageCritic,
		Pipeline:      domainagent.CanonicalPipeline(),
		InitialInput:  input,
		StageStates:   []domainagent.AgentState{state},
		StartedAt:     time.Date(2026, time.March, 17, 9, 0, 0, 0, time.UTC),
		UpdatedAt:     time.Date(2026, time.March, 17, 9, 0, 5, 0, time.UTC),
		CompletedAt:   time.Date(2026, time.March, 17, 9, 0, 5, 0, time.UTC),
	}

	// Save with project scope
	err := store.SaveWithProject(ctx, session, state, projectID, &visualizationID)
	require.NoError(t, err)

	// Restore the snapshot
	snapshot, err := store.Restore(input.SessionID, domainagent.StageCritic)
	require.NoError(t, err)

	// Verify restored state
	assert.Equal(t, input.SessionID, snapshot.Session.SessionID)
	assert.Equal(t, domainagent.StageCritic, snapshot.Stage.Stage)

	// Verify the session was persisted with project scope
	sessions, err := sessionRepo.GetByVisualization(ctx, projectID, visualizationID, 10)
	require.NoError(t, err)
	require.Len(t, sessions, 1)
	assert.Equal(t, projectID, sessions[0].ProjectID)
	assert.Equal(t, visualizationID, *sessions[0].VisualizationID)
}

func setupTestDBForSnapshot(t *testing.T) *gorm.DB {
	t.Helper()

	// Use separate in-memory databases for each test to avoid locking issues
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Migrate only the SessionModel
	err = db.AutoMigrate(&SessionModel{})
	require.NoError(t, err)

	return db
}

// Ensure PersistentSnapshotStore implements the orchestrator SnapshotStore interface
func TestPersistentSnapshotStore_ImplementsInterface(t *testing.T) {
	t.Parallel()

	db := setupTestDBForSnapshot(t)
	sessionRepo := NewSessionRepository(db)
	store := NewPersistentSnapshotStore(sessionRepo)

	// This is a compile-time check that PersistentSnapshotStore satisfies the SnapshotStore interface
	var _ interface {
		Save(session domainagent.SessionState, state domainagent.AgentState) error
		Restore(sessionID string, stage domainagent.StageName) (agentstate.Snapshot, error)
	} = store

	_ = store // Use store to avoid unused variable warning
}

// Additional test for GetSessionRecord to verify project-scoped session access
func TestPersistentSnapshotStore_GetSessionRecord(t *testing.T) {
	t.Parallel()

	db := setupTestDBForSnapshot(t)
	sessionRepo := NewSessionRepository(db)
	store := NewPersistentSnapshotStore(sessionRepo)

	ctx := context.Background()

	projectID := "project-for-record"
	visualizationID := "viz-for-record"

	input := domainagent.AgentInput{
		SessionID: "record-session-001",
		RequestID: "record-request-001",
		Content:   "Test content",
		VisualIntent: domainagent.VisualIntent{
			Mode:  domainagent.VisualModeDiagram,
			Goal:  "Test record access",
			Style: "academic",
		},
	}

	session := domainagent.SessionState{
		SchemaVersion: "test-schema-v1",
		SessionID:     input.SessionID,
		RequestID:     input.RequestID,
		Status:        domainagent.StatusRunning,
		CurrentStage:  domainagent.StageRetriever,
		Pipeline:      domainagent.CanonicalPipeline(),
		InitialInput:  input,
		StartedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	state := domainagent.AgentState{
		Stage:  domainagent.StageRetriever,
		Status: domainagent.StatusCompleted,
		Input:  input,
	}

	err := store.SaveWithProject(ctx, session, state, projectID, &visualizationID)
	require.NoError(t, err)

	// Get the session record
	record, err := store.GetSessionRecord(ctx, input.SessionID)
	require.NoError(t, err)
	assert.Equal(t, projectID, record.ProjectID)
	assert.Equal(t, visualizationID, *record.VisualizationID)
	assert.Equal(t, input.SessionID, record.ID)
}
