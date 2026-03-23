package agent

import "time"

type EventType string

const (
	EventRunStarted     EventType = "run_started"
	EventStageStarted   EventType = "stage_started"
	EventStageCompleted EventType = "stage_completed"
	EventStageFailed    EventType = "stage_failed"
	EventRunCompleted   EventType = "run_completed"
	EventRunFailed      EventType = "run_failed"
	EventRunCanceled    EventType = "run_canceled"

	// Batch event types for multi-candidate generation
	EventBatchStarted     EventType = "batch_start"
	EventCandidateStart   EventType = "candidate_start"
	EventCandidateComplete EventType = "candidate_complete"
	EventBatchCompleted   EventType = "batch_complete"
)

type Event struct {
	Sequence   int64             `json:"sequence"`
	SessionID  string            `json:"session_id"`
	RequestID  string            `json:"request_id"`
	Type       EventType         `json:"type"`
	Stage      StageName         `json:"stage,omitempty"`
	Status     RunStatus         `json:"status"`
	OccurredAt time.Time         `json:"occurred_at"`
	Timing     Timing            `json:"timing"`
	Error      *ErrorDetail      `json:"error,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

func (e Event) Terminal() bool {
	return e.Type == EventRunCompleted || e.Type == EventRunFailed || e.Type == EventRunCanceled
}

// BatchEvent represents an event during batch execution for SSE streaming.
type BatchEvent struct {
	Sequence    int64             `json:"sequence"`
	BatchID     string            `json:"batch_id"`
	CandidateID int               `json:"candidate_id,omitempty"`
	Type        EventType         `json:"type"`
	Stage       StageName         `json:"stage,omitempty"`
	Status      RunStatus         `json:"status"`
	OccurredAt  time.Time         `json:"occurred_at"`
	Timing      Timing            `json:"timing,omitempty"`
	Error       *ErrorDetail      `json:"error,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// Terminal returns true if this is a terminal batch event.
func (e BatchEvent) Terminal() bool {
	return e.Type == EventBatchCompleted
}
