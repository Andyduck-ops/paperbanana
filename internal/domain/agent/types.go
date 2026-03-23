package agent

import (
	"time"

	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
)

type StageName string

const (
	StageRetriever  StageName = "retriever"
	StagePlanner    StageName = "planner"
	StageStylist    StageName = "stylist"
	StageVisualizer StageName = "visualizer"
	StageCritic     StageName = "critic"
	StagePolish     StageName = "polish"
)

var pipelineOrder = []StageName{
	StageRetriever,
	StagePlanner,
	StageStylist,
	StageVisualizer,
	StageCritic,
}

func CanonicalPipeline() []StageName {
	return append([]StageName(nil), pipelineOrder...)
}

type RunStatus string

const (
	StatusPending   RunStatus = "pending"
	StatusRunning   RunStatus = "running"
	StatusCompleted RunStatus = "completed"
	StatusFailed    RunStatus = "failed"
	StatusCanceled  RunStatus = "canceled"
)

type VisualMode string

const (
	VisualModeDiagram VisualMode = "diagram"
	VisualModePlot    VisualMode = "plot"
)

type ArtifactKind string

const (
	ArtifactKindReferenceBundle ArtifactKind = "reference_bundle"
	ArtifactKindPlan            ArtifactKind = "plan"
	ArtifactKindRenderedFigure  ArtifactKind = "rendered_figure"
	ArtifactKindPromptTrace     ArtifactKind = "prompt_trace"
	ArtifactKindCritique        ArtifactKind = "critique"
)

type VisualIntent struct {
	Mode             VisualMode `json:"mode"`
	Goal             string     `json:"goal"`
	Audience         string     `json:"audience"`
	Style            string     `json:"style"`
	Constraints      []string   `json:"constraints,omitempty"`
	PreferredOutputs []string   `json:"preferred_outputs,omitempty"`
}

type RetrievedReference struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Source      string    `json:"source"`
	URI         string    `json:"uri"`
	Summary     string    `json:"summary"`
	Score       float64   `json:"score"`
	Snippets    []string  `json:"snippets,omitempty"`
	RetrievedAt time.Time `json:"retrieved_at"`
}

type PromptMetadata struct {
	SystemInstruction string            `json:"system_instruction"`
	Version           string            `json:"version"`
	Template          string            `json:"template"`
	Variables         map[string]string `json:"variables,omitempty"`
}

type Artifact struct {
	ID       string            `json:"id"`
	Kind     ArtifactKind      `json:"kind"`
	MIMEType string            `json:"mime_type"`
	URI      string            `json:"uri"`
	Content  string            `json:"content,omitempty"`
	Bytes    []byte            `json:"bytes,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type CritiqueRound struct {
	Round            int       `json:"round"`
	Summary          string    `json:"summary"`
	Accepted         bool      `json:"accepted"`
	RequestedChanges []string  `json:"requested_changes,omitempty"`
	EvaluatedAt      time.Time `json:"evaluated_at"`
}

type RestoreMetadata struct {
	SnapshotVersion string    `json:"snapshot_version"`
	RestoredFrom    StageName `json:"restored_from"`
	RestoredAt      time.Time `json:"restored_at"`
	ResumeToken     string    `json:"resume_token"`
}

type ErrorDetail struct {
	Message   string    `json:"message"`
	Code      string    `json:"code,omitempty"`
	Retryable bool      `json:"retryable"`
	Stage     StageName `json:"stage,omitempty"`
}

type Timing struct {
	StartedAt   time.Time     `json:"started_at"`
	CompletedAt time.Time     `json:"completed_at"`
	Duration    time.Duration `json:"duration"`
}

type AgentInput struct {
	SessionID           string               `json:"session_id"`
	RequestID           string               `json:"request_id"`
	Stage               StageName            `json:"stage"`
	Content             string               `json:"content"`
	Messages            []domainllm.Message  `json:"messages,omitempty"`
	VisualIntent        VisualIntent         `json:"visual_intent"`
	RetrievedReferences []RetrievedReference `json:"retrieved_references,omitempty"`
	Prompt              PromptMetadata       `json:"prompt"`
	GeneratedArtifacts  []Artifact           `json:"generated_artifacts,omitempty"`
	CritiqueRounds      []CritiqueRound      `json:"critique_rounds,omitempty"`
	Restore             RestoreMetadata      `json:"restore"`
	Metadata            map[string]string    `json:"metadata,omitempty"`
}

type AgentOutput struct {
	Stage               StageName            `json:"stage"`
	Content             string               `json:"content,omitempty"`
	Messages            []domainllm.Message  `json:"messages,omitempty"`
	VisualIntent        VisualIntent         `json:"visual_intent"`
	RetrievedReferences []RetrievedReference `json:"retrieved_references,omitempty"`
	Prompt              PromptMetadata       `json:"prompt"`
	GeneratedArtifacts  []Artifact           `json:"generated_artifacts,omitempty"`
	CritiqueRounds      []CritiqueRound      `json:"critique_rounds,omitempty"`
	Error               *ErrorDetail         `json:"error,omitempty"`
	Metadata            map[string]string    `json:"metadata,omitempty"`
}

type AgentState struct {
	Stage   StageName       `json:"stage"`
	Status  RunStatus       `json:"status"`
	Timing  Timing          `json:"timing"`
	Input   AgentInput      `json:"input"`
	Output  AgentOutput     `json:"output"`
	Error   *ErrorDetail    `json:"error,omitempty"`
	Restore RestoreMetadata `json:"restore"`
}

type SessionState struct {
	SchemaVersion string            `json:"schema_version"`
	SessionID     string            `json:"session_id"`
	RequestID     string            `json:"request_id"`
	Status        RunStatus         `json:"status"`
	CurrentStage  StageName         `json:"current_stage"`
	Pipeline      []StageName       `json:"pipeline"`
	InitialInput  AgentInput        `json:"initial_input"`
	StageStates   []AgentState      `json:"stage_states,omitempty"`
	FinalOutput   AgentOutput       `json:"final_output"`
	Error         *ErrorDetail      `json:"error,omitempty"`
	Restore       RestoreMetadata   `json:"restore"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	StartedAt     time.Time         `json:"started_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	CompletedAt   time.Time         `json:"completed_at"`
}

// BatchTiming tracks the timing information for a batch execution.
type BatchTiming struct {
	StartedAt   time.Time     `json:"started_at"`
	CompletedAt time.Time     `json:"completed_at"`
	Duration    time.Duration `json:"duration"`
}

// CandidateResult represents the result of a single candidate execution within a batch.
type CandidateResult struct {
	CandidateID int          `json:"candidate_id"`
	SessionID   string       `json:"session_id"`
	Status      RunStatus    `json:"status"`
	Artifacts   []Artifact   `json:"artifacts,omitempty"`
	Error       *ErrorDetail `json:"error,omitempty"`
}

// BatchResult aggregates the results from all candidates in a batch execution.
type BatchResult struct {
	BatchID    string           `json:"batch_id"`
	Results    []CandidateResult `json:"results"`
	Successful int              `json:"successful"`
	Failed     int              `json:"failed"`
	Timing     BatchTiming      `json:"timing"`
}
