package agentstate

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
)

func TestSnapshotStoreWritesVersionedStageSnapshot(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	store := NewStore(rootDir)
	session, state := fixtureSessionAndState()

	require.NoError(t, store.Save(session, state))

	snapshotPath := filepath.Join(
		rootDir,
		".paperbanana",
		"sessions",
		session.SessionID,
		"agent_states",
		string(state.Stage)+".json",
	)
	require.FileExists(t, snapshotPath)

	payload, err := os.ReadFile(snapshotPath)
	require.NoError(t, err)

	var snapshot Snapshot
	require.NoError(t, json.Unmarshal(payload, &snapshot))

	require.Equal(t, SnapshotSchemaVersion, snapshot.SchemaVersion)
	require.Equal(t, session.SessionID, snapshot.Session.SessionID)
	require.Equal(t, session.RequestID, snapshot.Session.RequestID)
	require.Equal(t, session.CurrentStage, snapshot.Session.CurrentStage)
	require.Equal(t, session.InitialInput.Content, snapshot.Session.InitialInput.Content)
	require.Len(t, snapshot.Session.StageStates, len(session.StageStates))
	require.Equal(t, session.StageStates[0].Stage, snapshot.Session.StageStates[0].Stage)
	require.Equal(t, session.FinalOutput.Content, snapshot.Session.FinalOutput.Content)
	require.Equal(t, session.Error, snapshot.Session.Error)
	require.Equal(t, state.Stage, snapshot.Stage.Stage)
	require.Equal(t, string(state.Stage), snapshot.Stage.AgentType)
	require.Equal(t, state.Status, snapshot.Stage.Status)
	require.Equal(t, state.Input.Content, snapshot.Stage.Input.Content)
	require.Equal(t, state.Output.Content, snapshot.Stage.Output.Content)
	require.Equal(t, state.Timing.StartedAt, snapshot.Stage.Timing.StartedAt)
	require.Equal(t, state.Timing.CompletedAt, snapshot.Stage.Timing.CompletedAt)
	require.Equal(t, state.Timing.Duration, snapshot.Stage.Timing.Duration)
	require.Equal(t, state.Error, snapshot.Stage.Error)

	restored, err := store.Restore(session.SessionID, state.Stage)
	require.NoError(t, err)
	require.Equal(t, snapshot, restored)
}

func TestSnapshotStoreUsesAtomicRename(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	store := NewStore(rootDir)
	session, state := fixtureSessionAndState()

	var renameFrom string
	var renameTo string
	store.rename = func(from, to string) error {
		renameFrom = from
		renameTo = to
		return os.Rename(from, to)
	}

	require.NoError(t, store.Save(session, state))

	require.NotEmpty(t, renameFrom)
	require.NotEmpty(t, renameTo)
	require.Contains(t, filepath.Base(renameFrom), string(state.Stage))
	require.Equal(
		t,
		filepath.Join(
			rootDir,
			".paperbanana",
			"sessions",
			session.SessionID,
			"agent_states",
			string(state.Stage)+".json",
		),
		renameTo,
	)

	_, err := os.Stat(renameFrom)
	require.Error(t, err)
	require.True(t, errors.Is(err, os.ErrNotExist))

	tempFiles, err := filepath.Glob(filepath.Join(filepath.Dir(renameTo), "*.tmp"))
	require.NoError(t, err)
	require.Empty(t, tempFiles)
}

func TestSnapshotStoreRejectsInvalidRestore(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	store := NewStore(rootDir)
	session, state := fixtureSessionAndState()

	dir := filepath.Join(rootDir, ".paperbanana", "sessions", session.SessionID, "agent_states")
	require.NoError(t, os.MkdirAll(dir, 0o755))

	invalidSchema := BuildSnapshot(session, state)
	invalidSchema.SchemaVersion = "agent-state-store/v999"
	invalidSchemaPath := filepath.Join(dir, string(state.Stage)+".json")
	payload, err := json.MarshalIndent(invalidSchema, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(invalidSchemaPath, payload, 0o644))

	_, err = store.Restore(session.SessionID, state.Stage)
	require.ErrorIs(t, err, ErrInvalidSnapshot)
	require.ErrorContains(t, err, "schema version")

	stageMismatch := BuildSnapshot(session, state)
	stageMismatch.Stage.Stage = domainagent.StageCritic
	stageMismatch.Stage.AgentType = string(domainagent.StageCritic)
	payload, err = json.MarshalIndent(stageMismatch, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(invalidSchemaPath, payload, 0o644))

	_, err = store.Restore(session.SessionID, state.Stage)
	require.ErrorIs(t, err, ErrInvalidSnapshot)
	require.ErrorContains(t, err, "stage mismatch")
}

func fixtureSessionAndState() (domainagent.SessionState, domainagent.AgentState) {
	startedAt := time.Date(2026, time.March, 16, 12, 0, 0, 0, time.UTC)
	completedAt := startedAt.Add(2 * time.Second)

	state := domainagent.AgentState{
		Stage:  domainagent.StagePlanner,
		Status: domainagent.StatusFailed,
		Timing: domainagent.Timing{
			StartedAt:   startedAt,
			CompletedAt: completedAt,
			Duration:    2 * time.Second,
		},
		Input: domainagent.AgentInput{
			SessionID: "session-123",
			RequestID: "request-456",
			Stage:     domainagent.StagePlanner,
			Content:   "draft the plotting plan",
			Prompt: domainagent.PromptMetadata{
				SystemInstruction: "planner-system",
				Version:           "planner/v2",
			},
			Metadata: map[string]string{
				"locale": "zh-CN",
			},
		},
		Output: domainagent.AgentOutput{
			Stage:   domainagent.StagePlanner,
			Content: "plan output",
			Prompt: domainagent.PromptMetadata{
				SystemInstruction: "planner-system",
				Version:           "planner/v2",
			},
			GeneratedArtifacts: []domainagent.Artifact{
				{
					ID:       "artifact-1",
					Kind:     domainagent.ArtifactKindPlan,
					MIMEType: "application/json",
					URI:      "memory://plan",
					Content:  "{\"chart\":\"scatter\"}",
				},
			},
		},
		Error: &domainagent.ErrorDetail{
			Message:   "planner validation failed",
			Code:      "planner_validation",
			Retryable: false,
			Stage:     domainagent.StagePlanner,
		},
		Restore: domainagent.RestoreMetadata{
			SnapshotVersion: "agent-stage-store/v1",
			RestoredFrom:    domainagent.StageRetriever,
			RestoredAt:      startedAt.Add(-1 * time.Minute),
			ResumeToken:     "resume-token-123",
		},
	}

	session := domainagent.SessionState{
		SchemaVersion: "agent-session/v1",
		SessionID:     state.Input.SessionID,
		RequestID:     state.Input.RequestID,
		Status:        domainagent.StatusFailed,
		CurrentStage:  state.Stage,
		Pipeline:      domainagent.CanonicalPipeline(),
		InitialInput:  state.Input,
		StageStates:   []domainagent.AgentState{state},
		FinalOutput:   state.Output,
		Error:         state.Error,
		Restore:       state.Restore,
		Metadata: map[string]string{
			"tenant": "paperbanana-lab",
		},
		StartedAt:   startedAt,
		UpdatedAt:   completedAt,
		CompletedAt: completedAt,
	}

	return session, state
}
