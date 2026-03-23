package agentstate

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
)

const SnapshotSchemaVersion = "agent-stage-store/v1"

var (
	ErrSnapshotNotFound = errors.New("agentstate: snapshot not found")
	ErrInvalidSnapshot  = errors.New("agentstate: invalid snapshot")
)

type Store struct {
	rootDir string
	rename  func(oldPath, newPath string) error
}

type Snapshot struct {
	SchemaVersion string          `json:"schema_version"`
	Session       SessionSnapshot `json:"session"`
	Stage         StageSnapshot   `json:"stage"`
}

type SessionSnapshot struct {
	SchemaVersion string                      `json:"schema_version"`
	SessionID     string                      `json:"session_id"`
	RequestID     string                      `json:"request_id"`
	Status        domainagent.RunStatus       `json:"status"`
	CurrentStage  domainagent.StageName       `json:"current_stage"`
	Pipeline      []domainagent.StageName     `json:"pipeline,omitempty"`
	InitialInput  domainagent.AgentInput      `json:"initial_input"`
	StageStates   []domainagent.AgentState    `json:"stage_states,omitempty"`
	FinalOutput   domainagent.AgentOutput     `json:"final_output"`
	Error         *domainagent.ErrorDetail    `json:"error,omitempty"`
	Restore       domainagent.RestoreMetadata `json:"restore"`
	Metadata      map[string]string           `json:"metadata,omitempty"`
	StartedAt     time.Time                   `json:"started_at"`
	UpdatedAt     time.Time                   `json:"updated_at"`
	CompletedAt   time.Time                   `json:"completed_at,omitempty"`
}

type StageSnapshot struct {
	AgentType string                      `json:"agent_type"`
	Stage     domainagent.StageName       `json:"stage"`
	Status    domainagent.RunStatus       `json:"status"`
	Timing    domainagent.Timing          `json:"timing"`
	Input     domainagent.AgentInput      `json:"input"`
	Output    domainagent.AgentOutput     `json:"output"`
	Error     *domainagent.ErrorDetail    `json:"error,omitempty"`
	Restore   domainagent.RestoreMetadata `json:"restore"`
}

func NewStore(rootDir string) *Store {
	if rootDir == "" {
		rootDir = "."
	}

	return &Store{
		rootDir: rootDir,
		rename:  os.Rename,
	}
}

func BuildSnapshot(session domainagent.SessionState, state domainagent.AgentState) Snapshot {
	return Snapshot{
		SchemaVersion: SnapshotSchemaVersion,
		Session: SessionSnapshot{
			SchemaVersion: session.SchemaVersion,
			SessionID:     session.SessionID,
			RequestID:     session.RequestID,
			Status:        session.Status,
			CurrentStage:  session.CurrentStage,
			Pipeline:      append([]domainagent.StageName(nil), session.Pipeline...),
			InitialInput:  cloneAgentInput(session.InitialInput),
			StageStates:   cloneAgentStates(session.StageStates),
			FinalOutput:   cloneAgentOutput(session.FinalOutput),
			Error:         cloneErrorDetail(session.Error),
			Restore:       session.Restore,
			Metadata:      cloneStringMap(session.Metadata),
			StartedAt:     session.StartedAt,
			UpdatedAt:     session.UpdatedAt,
			CompletedAt:   session.CompletedAt,
		},
		Stage: StageSnapshot{
			AgentType: string(state.Stage),
			Stage:     state.Stage,
			Status:    state.Status,
			Timing:    state.Timing,
			Input:     cloneAgentInput(state.Input),
			Output:    cloneAgentOutput(state.Output),
			Error:     cloneErrorDetail(state.Error),
			Restore:   state.Restore,
		},
	}
}

func (s *Store) Save(session domainagent.SessionState, state domainagent.AgentState) error {
	snapshot := BuildSnapshot(session, state)
	snapshotDir := s.snapshotDir(session.SessionID)
	if err := os.MkdirAll(snapshotDir, 0o755); err != nil {
		return fmt.Errorf("create snapshot dir: %w", err)
	}

	payload, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal snapshot: %w", err)
	}
	payload = append(payload, '\n')

	tempFile, err := os.CreateTemp(snapshotDir, fmt.Sprintf("%s-*.tmp", state.Stage))
	if err != nil {
		return fmt.Errorf("create temp snapshot: %w", err)
	}

	tempPath := tempFile.Name()
	cleanup := func() {
		_ = tempFile.Close()
		_ = os.Remove(tempPath)
	}

	if _, err := tempFile.Write(payload); err != nil {
		cleanup()
		return fmt.Errorf("write snapshot temp file: %w", err)
	}
	if err := tempFile.Sync(); err != nil {
		cleanup()
		return fmt.Errorf("sync snapshot temp file: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("close snapshot temp file: %w", err)
	}

	finalPath := s.snapshotPath(session.SessionID, state.Stage)
	if err := s.rename(tempPath, finalPath); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("rename snapshot temp file: %w", err)
	}

	return nil
}

func (s *Store) Restore(sessionID string, stage domainagent.StageName) (Snapshot, error) {
	payload, err := os.ReadFile(s.snapshotPath(sessionID, stage))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Snapshot{}, fmt.Errorf("%w: %s/%s", ErrSnapshotNotFound, sessionID, stage)
		}
		return Snapshot{}, fmt.Errorf("read snapshot: %w", err)
	}

	var snapshot Snapshot
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&snapshot); err != nil {
		return Snapshot{}, fmt.Errorf("%w: decode snapshot: %v", ErrInvalidSnapshot, err)
	}

	if err := validateSnapshot(snapshot, sessionID, stage); err != nil {
		return Snapshot{}, err
	}

	return snapshot, nil
}

func (s *Store) snapshotDir(sessionID string) string {
	return filepath.Join(s.rootDir, ".paperbanana", "sessions", sessionID, "agent_states")
}

func (s *Store) snapshotPath(sessionID string, stage domainagent.StageName) string {
	return filepath.Join(s.snapshotDir(sessionID), string(stage)+".json")
}

func validateSnapshot(snapshot Snapshot, sessionID string, stage domainagent.StageName) error {
	if snapshot.SchemaVersion != SnapshotSchemaVersion {
		return fmt.Errorf(
			"%w: schema version mismatch: got %q want %q",
			ErrInvalidSnapshot,
			snapshot.SchemaVersion,
			SnapshotSchemaVersion,
		)
	}
	if snapshot.Session.SessionID != sessionID {
		return fmt.Errorf(
			"%w: session mismatch: got %q want %q",
			ErrInvalidSnapshot,
			snapshot.Session.SessionID,
			sessionID,
		)
	}
	if snapshot.Stage.Stage != stage {
		return fmt.Errorf(
			"%w: stage mismatch: got %q want %q",
			ErrInvalidSnapshot,
			snapshot.Stage.Stage,
			stage,
		)
	}
	if snapshot.Stage.AgentType == "" {
		return fmt.Errorf("%w: missing agent type", ErrInvalidSnapshot)
	}
	return nil
}

func cloneAgentInput(input domainagent.AgentInput) domainagent.AgentInput {
	cloned := input
	cloned.Messages = cloneMessages(input.Messages)
	cloned.VisualIntent = cloneVisualIntent(input.VisualIntent)
	cloned.RetrievedReferences = cloneReferences(input.RetrievedReferences)
	cloned.Prompt = clonePrompt(input.Prompt)
	cloned.GeneratedArtifacts = cloneArtifacts(input.GeneratedArtifacts)
	cloned.CritiqueRounds = cloneCritiqueRounds(input.CritiqueRounds)
	cloned.Metadata = cloneStringMap(input.Metadata)
	return cloned
}

func cloneAgentOutput(output domainagent.AgentOutput) domainagent.AgentOutput {
	cloned := output
	cloned.Messages = cloneMessages(output.Messages)
	cloned.VisualIntent = cloneVisualIntent(output.VisualIntent)
	cloned.RetrievedReferences = cloneReferences(output.RetrievedReferences)
	cloned.Prompt = clonePrompt(output.Prompt)
	cloned.GeneratedArtifacts = cloneArtifacts(output.GeneratedArtifacts)
	cloned.CritiqueRounds = cloneCritiqueRounds(output.CritiqueRounds)
	cloned.Error = cloneErrorDetail(output.Error)
	cloned.Metadata = cloneStringMap(output.Metadata)
	return cloned
}

func cloneAgentStates(states []domainagent.AgentState) []domainagent.AgentState {
	if len(states) == 0 {
		return nil
	}

	cloned := make([]domainagent.AgentState, len(states))
	for i, state := range states {
		cloned[i] = state
		cloned[i].Input = cloneAgentInput(state.Input)
		cloned[i].Output = cloneAgentOutput(state.Output)
		cloned[i].Error = cloneErrorDetail(state.Error)
	}
	return cloned
}

func cloneMessages(messages []domainllm.Message) []domainllm.Message {
	if len(messages) == 0 {
		return nil
	}

	cloned := make([]domainllm.Message, len(messages))
	for i, message := range messages {
		cloned[i] = message
		cloned[i].Parts = append([]domainllm.Part(nil), message.Parts...)
	}
	return cloned
}

func cloneVisualIntent(intent domainagent.VisualIntent) domainagent.VisualIntent {
	cloned := intent
	cloned.Constraints = append([]string(nil), intent.Constraints...)
	cloned.PreferredOutputs = append([]string(nil), intent.PreferredOutputs...)
	return cloned
}

func cloneReferences(references []domainagent.RetrievedReference) []domainagent.RetrievedReference {
	if len(references) == 0 {
		return nil
	}

	cloned := make([]domainagent.RetrievedReference, len(references))
	for i, reference := range references {
		cloned[i] = reference
		cloned[i].Snippets = append([]string(nil), reference.Snippets...)
	}
	return cloned
}

func clonePrompt(prompt domainagent.PromptMetadata) domainagent.PromptMetadata {
	cloned := prompt
	cloned.Variables = cloneStringMap(prompt.Variables)
	return cloned
}

func cloneArtifacts(artifacts []domainagent.Artifact) []domainagent.Artifact {
	if len(artifacts) == 0 {
		return nil
	}

	cloned := make([]domainagent.Artifact, len(artifacts))
	for i, artifact := range artifacts {
		cloned[i] = artifact
		cloned[i].Bytes = append([]byte(nil), artifact.Bytes...)
		cloned[i].Metadata = cloneStringMap(artifact.Metadata)
	}
	return cloned
}

func cloneCritiqueRounds(rounds []domainagent.CritiqueRound) []domainagent.CritiqueRound {
	if len(rounds) == 0 {
		return nil
	}

	cloned := make([]domainagent.CritiqueRound, len(rounds))
	for i, round := range rounds {
		cloned[i] = round
		cloned[i].RequestedChanges = append([]string(nil), round.RequestedChanges...)
	}
	return cloned
}

func cloneErrorDetail(detail *domainagent.ErrorDetail) *domainagent.ErrorDetail {
	if detail == nil {
		return nil
	}

	cloned := *detail
	return &cloned
}

func cloneStringMap(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}

	cloned := make(map[string]string, len(input))
	for key, value := range input {
		cloned[key] = value
	}
	return cloned
}
