package agent

import (
	"sync/atomic"
	"time"

	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
)

// SharedBytes provides reference-counted byte slices for efficient sharing
// of large binary data (e.g., image artifacts) without deep copying.
// This is thread-safe for concurrent read access.
type SharedBytes struct {
	data []byte
	ref  *int32 // reference count, accessed atomically
}

// NewSharedBytes creates a new SharedBytes with initial reference count of 1.
func NewSharedBytes(data []byte) *SharedBytes {
	if len(data) == 0 {
		return nil
	}
	ref := int32(1)
	return &SharedBytes{data: data, ref: &ref}
}

// Data returns the underlying byte slice. Callers MUST NOT modify the returned slice.
func (sb *SharedBytes) Data() []byte {
	if sb == nil {
		return nil
	}
	return sb.data
}

// Retain increments the reference count and returns the same SharedBytes pointer.
// Returns nil if sb is nil.
func (sb *SharedBytes) Retain() *SharedBytes {
	if sb == nil {
		return nil
	}
	atomic.AddInt32(sb.ref, 1)
	return sb
}

// Release decrements the reference count. When count reaches zero, data is cleared.
// Returns true if the data was released (count reached zero).
func (sb *SharedBytes) Release() bool {
	if sb == nil {
		return false
	}
	if atomic.AddInt32(sb.ref, -1) == 0 {
		sb.data = nil
		return true
	}
	return false
}

// RefCount returns the current reference count (for testing/debugging).
func (sb *SharedBytes) RefCount() int32 {
	if sb == nil || sb.ref == nil {
		return 0
	}
	return atomic.LoadInt32(sb.ref)
}

// Clone returns a shallow copy that shares the same underlying data.
// This is an alias for Retain() to match common naming conventions.
func (sb *SharedBytes) Clone() *SharedBytes {
	return sb.Retain()
}

// Len returns the length of the underlying data.
func (sb *SharedBytes) Len() int {
	if sb == nil {
		return 0
	}
	return len(sb.data)
}

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

// PipelineMode constants define the valid pipeline execution modes.
const (
	PipelineModeFull          = "full"
	PipelineModePlannerCritic = "planner-critic"
	PipelineModeVanilla       = "vanilla"
)

// ValidPipelineModes returns the list of valid pipeline mode values.
func ValidPipelineModes() []string {
	return []string{PipelineModeFull, PipelineModePlannerCritic, PipelineModeVanilla}
}

// IsValidPipelineMode checks if the given mode is a valid pipeline mode.
func IsValidPipelineMode(mode string) bool {
	for _, valid := range ValidPipelineModes() {
		if mode == valid {
			return true
		}
	}
	return false
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
	ArtifactKindPolishedImage   ArtifactKind = "polished_image"
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
	ID        string            `json:"id"`
	Kind      ArtifactKind      `json:"kind"`
	MIMEType  string            `json:"mime_type"`
	URI       string            `json:"uri"`
	Content   string            `json:"content,omitempty"`
	Bytes     []byte            `json:"data,omitempty"`               // Deprecated: Use SharedBytes for new code
	Shared    *SharedBytes      `json:"-"`                            // Reference-counted bytes, not serialized
	AssetID   string            `json:"assetId,omitempty"`            // camelCase for frontend compatibility
	ProjectID string            `json:"projectId,omitempty"`          // camelCase for frontend URL construction
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// GetBytes returns the artifact's binary data, preferring SharedBytes if available.
// This provides a unified access method for backwards compatibility.
func (a *Artifact) GetBytes() []byte {
	if a.Shared != nil {
		return a.Shared.Data()
	}
	return a.Bytes
}

// SetBytes sets the artifact's binary data using SharedBytes for efficient sharing.
// This also updates the legacy Bytes field for JSON serialization compatibility.
func (a *Artifact) SetBytes(data []byte) {
	a.Shared = NewSharedBytes(data)
	a.Bytes = data // Keep in sync for JSON serialization
}

// Clone creates a shallow copy of the Artifact that shares the same SharedBytes.
// This avoids deep copying large binary data.
func (a Artifact) Clone() Artifact {
	cloned := a
	if a.Shared != nil {
		cloned.Shared = a.Shared.Retain()
	}
	cloned.Metadata = cloneStringMap(a.Metadata)
	return cloned
}

// cloneStringMap is a helper for Clone (defined here to avoid import cycles).
func cloneStringMap(m map[string]string) map[string]string {
	if len(m) == 0 {
		return nil
	}
	result := make(map[string]string, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
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
	Message    string    `json:"message"`
	Code       string    `json:"code,omitempty"`
	Category   string    `json:"category,omitempty"`
	Retryable  bool      `json:"retryable"`
	Suggestion string    `json:"suggestion,omitempty"`
	Stage      StageName `json:"stage,omitempty"`
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
	FailedStage   StageName         `json:"failed_stage,omitempty"`
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
	UpdatedAt   time.Time     `json:"updated_at,omitempty"` // 用于进度更新
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
	BatchID    string            `json:"batch_id"`
	Results    []CandidateResult `json:"results"`
	Successful int               `json:"successful"`
	Failed     int               `json:"failed"`
	Timing     BatchTiming       `json:"timing"`
}
