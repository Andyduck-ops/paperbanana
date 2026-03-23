package sqlite

import (
	"context"
	"fmt"
	"time"

	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
	"github.com/paperbanana/paperbanana/internal/domain/workspace"
	agentstate "github.com/paperbanana/paperbanana/internal/infrastructure/agentstate"
)

// PersistentSnapshotStore implements the orchestrator SnapshotStore interface
// using the SQLite-backed SessionRepository as the persistence backend.
// It replaces the Phase 2 filesystem snapshot store with formal persistence.
type PersistentSnapshotStore struct {
	sessionRepo workspace.SessionRepository
}

// NewPersistentSnapshotStore creates a new PersistentSnapshotStore.
func NewPersistentSnapshotStore(sessionRepo workspace.SessionRepository) *PersistentSnapshotStore {
	return &PersistentSnapshotStore{
		sessionRepo: sessionRepo,
	}
}

// Save persists the session state and agent state to the SQLite database.
// It creates or updates a SessionRecord with the full session snapshot.
func (s *PersistentSnapshotStore) Save(session domainagent.SessionState, state domainagent.AgentState) error {
	ctx := context.Background()
	return s.SaveWithProject(ctx, session, state, "", nil)
}

// SaveWithProject persists the session with project and visualization ownership.
// This is the primary method used by the generate flow to associate sessions with projects.
func (s *PersistentSnapshotStore) SaveWithProject(ctx context.Context, session domainagent.SessionState, state domainagent.AgentState, projectID string, visualizationID *string) error {
	// Build the session record
	record := &workspace.SessionRecord{
		ID:            session.SessionID,
		ProjectID:     projectID,
		VisualizationID: visualizationID,
		Status:        string(session.Status),
		CurrentStage:  string(session.CurrentStage),
		SchemaVersion: session.SchemaVersion,
		Snapshot:      &session,
		CreatedAt:     session.StartedAt,
		UpdatedAt:     session.UpdatedAt,
	}

	if session.CompletedAt != (time.Time{}) {
		record.CompletedAt = &session.CompletedAt
	}

	// Check if session exists
	existing, err := s.sessionRepo.GetByID(ctx, session.SessionID)
	if err != nil {
		// Session doesn't exist, create it
		if projectID == "" {
			// For sessions without project scope (legacy/Phase 2 compatibility),
			// use a default project ID
			record.ProjectID = "default"
		}
		return s.sessionRepo.Create(ctx, record)
	}

	// Update existing session, preserving project ownership
	record.ProjectID = existing.ProjectID
	record.VisualizationID = existing.VisualizationID
	record.CreatedAt = existing.CreatedAt

	return s.sessionRepo.Update(ctx, record)
}

// Restore retrieves a session snapshot by session ID and stage.
// It reconstructs the agentstate.Snapshot shape expected by the runner.
func (s *PersistentSnapshotStore) Restore(sessionID string, stage domainagent.StageName) (agentstate.Snapshot, error) {
	ctx := context.Background()

	record, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return agentstate.Snapshot{}, fmt.Errorf("%w: %s/%s: %v", agentstate.ErrSnapshotNotFound, sessionID, stage, err)
	}

	if record.Snapshot == nil {
		return agentstate.Snapshot{}, fmt.Errorf("%w: session %s has no snapshot data", agentstate.ErrInvalidSnapshot, sessionID)
	}

	// Find the stage state in the snapshot
	var stageState *domainagent.AgentState
	for i, ss := range record.Snapshot.StageStates {
		if ss.Stage == stage {
			stageState = &record.Snapshot.StageStates[i]
			break
		}
	}

	if stageState == nil {
		return agentstate.Snapshot{}, fmt.Errorf("%w: stage %s not found in session %s", agentstate.ErrSnapshotNotFound, stage, sessionID)
	}

	// Build the snapshot expected by the runner
	return agentstate.Snapshot{
		SchemaVersion: agentstate.SnapshotSchemaVersion,
		Session: agentstate.SessionSnapshot{
			SchemaVersion: record.Snapshot.SchemaVersion,
			SessionID:     record.Snapshot.SessionID,
			RequestID:     record.Snapshot.RequestID,
			Status:        record.Snapshot.Status,
			CurrentStage:  record.Snapshot.CurrentStage,
			Pipeline:      record.Snapshot.Pipeline,
			InitialInput:  record.Snapshot.InitialInput,
			StageStates:   record.Snapshot.StageStates,
			FinalOutput:   record.Snapshot.FinalOutput,
			Error:         record.Snapshot.Error,
			Restore:       record.Snapshot.Restore,
			Metadata:      record.Snapshot.Metadata,
			StartedAt:     record.Snapshot.StartedAt,
			UpdatedAt:     record.Snapshot.UpdatedAt,
			CompletedAt:   record.Snapshot.CompletedAt,
		},
		Stage: agentstate.StageSnapshot{
			AgentType: string(stageState.Stage),
			Stage:     stageState.Stage,
			Status:    stageState.Status,
			Timing:    stageState.Timing,
			Input:     stageState.Input,
			Output:    stageState.Output,
			Error:     stageState.Error,
			Restore:   stageState.Restore,
		},
	}, nil
}

// GetSessionRecord retrieves the full session record including project ownership.
// This is used by the generate handler to check project scope.
func (s *PersistentSnapshotStore) GetSessionRecord(ctx context.Context, sessionID string) (*workspace.SessionRecord, error) {
	return s.sessionRepo.GetByID(ctx, sessionID)
}

// GetSessionByVisualization retrieves the most recent session for a visualization.
// Used for resume operations within a project scope.
func (s *PersistentSnapshotStore) GetSessionByVisualization(ctx context.Context, projectID, visualizationID string) (*workspace.SessionRecord, error) {
	sessions, err := s.sessionRepo.GetByVisualization(ctx, projectID, visualizationID, 1)
	if err != nil {
		return nil, err
	}
	if len(sessions) == 0 {
		return nil, fmt.Errorf("no session found for visualization %s in project %s", visualizationID, projectID)
	}
	return sessions[0], nil
}

// Ensure PersistentSnapshotStore implements the orchestrator SnapshotStore interface
var _ interface {
	Save(session domainagent.SessionState, state domainagent.AgentState) error
	Restore(sessionID string, stage domainagent.StageName) (agentstate.Snapshot, error)
} = (*PersistentSnapshotStore)(nil)
