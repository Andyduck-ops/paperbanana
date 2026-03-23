package orchestrator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"

	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
)

// AgentFactory creates fresh agent instances for each candidate.
type AgentFactory interface {
	CreateRetriever() domainagent.BaseAgent
	CreatePlanner() domainagent.BaseAgent
	CreateStylist() domainagent.BaseAgent
	CreateVisualizer() domainagent.BaseAgent
	CreateCritic() domainagent.BaseAgent
}

// BatchRunner executes multiple candidates in parallel with shared retriever.
type BatchRunner struct {
	agentFactory  AgentFactory
	maxConcurrent int
	eventBuffer   int
	results       map[string]*domainagent.BatchResult
	mu            sync.RWMutex
}

// BatchOption configures the BatchRunner.
type BatchOption func(*BatchRunner)

// NewBatchRunner creates a new batch runner.
func NewBatchRunner(factory AgentFactory, opts ...BatchOption) *BatchRunner {
	runner := &BatchRunner{
		agentFactory:  factory,
		maxConcurrent: 10,
		eventBuffer:   256,
		results:       make(map[string]*domainagent.BatchResult),
	}
	for _, opt := range opts {
		opt(runner)
	}
	return runner
}

// WithBatchMaxConcurrent sets the maximum concurrent candidates.
func WithBatchMaxConcurrent(n int) BatchOption {
	return func(r *BatchRunner) {
		if n > 0 {
			r.maxConcurrent = n
		}
	}
}

// WithBatchEventBuffer sets the event buffer size.
func WithBatchEventBuffer(size int) BatchOption {
	return func(r *BatchRunner) {
		if size > 0 {
			r.eventBuffer = size
		}
	}
}

// GetBatchResult retrieves a stored batch result by ID.
// Results are kept in memory for a limited time after completion.
func (r *BatchRunner) GetBatchResult(batchID string) (*domainagent.BatchResult, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result, ok := r.results[batchID]
	if !ok {
		return nil, fmt.Errorf("batch result not found: %s", batchID)
	}
	return result, nil
}

// storeResult saves a batch result for later retrieval.
func (r *BatchRunner) storeResult(batchID string, result *domainagent.BatchResult) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.results[batchID] = result
}

// BatchHandle provides access to batch execution results.
type BatchHandle struct {
	events <-chan domainagent.BatchEvent
	done   chan struct{}
	result domainagent.BatchResult
	err    error
	once   sync.Once
}

// Events returns the event channel.
func (h *BatchHandle) Events() <-chan domainagent.BatchEvent {
	return h.events
}

// Wait blocks until the batch completes and returns the result.
func (h *BatchHandle) Wait() (domainagent.BatchResult, error) {
	<-h.done
	return h.result, h.err
}

func (h *BatchHandle) setOutcome(result domainagent.BatchResult, err error) {
	h.once.Do(func() {
		h.result = result
		h.err = err
		close(h.done)
	})
}

// StartBatch starts a batch execution with multiple candidates.
func (r *BatchRunner) StartBatch(ctx context.Context, inputs []domainagent.AgentInput) (*BatchHandle, error) {
	if len(inputs) == 0 {
		return nil, fmt.Errorf("batch requires at least one input")
	}

	batchID := uuid.New().String()
	events := make(chan domainagent.BatchEvent, r.eventBuffer)
	handle := &BatchHandle{
		events: events,
		done:   make(chan struct{}),
	}

	go r.executeBatch(ctx, batchID, inputs, events, handle)

	return handle, nil
}

func (r *BatchRunner) executeBatch(
	ctx context.Context,
	batchID string,
	inputs []domainagent.AgentInput,
	events chan<- domainagent.BatchEvent,
	handle *BatchHandle,
) {
	defer close(events)

	startTime := time.Now().UTC()
	publisher := newBatchEventPublisher(batchID, events)

	// Emit batch start
	publisher.emit(domainagent.EventBatchStarted, 0, domainagent.StatusRunning, startTime, time.Time{}, nil)

	// Phase 1: Run retriever once and share results
	sharedRefs, err := r.runSharedRetriever(ctx, inputs[0])
	if err != nil {
		completedAt := time.Now().UTC()
		publisher.emit(domainagent.EventBatchCompleted, 0, domainagent.StatusFailed, startTime, completedAt, err)
		handle.setOutcome(domainagent.BatchResult{
			BatchID:    batchID,
			Failed:     len(inputs),
			Successful: 0,
			Timing: domainagent.BatchTiming{
				StartedAt:   startTime,
				CompletedAt: completedAt,
			},
		}, err)
		return
	}

	// Phase 2: Run candidates in parallel
	results := make([]domainagent.CandidateResult, len(inputs))
	var mu sync.Mutex
	successful, failed := 0, 0

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(r.maxConcurrent)

	for i := range inputs {
		i := i // Capture loop variable

		g.Go(func() error {
			// Clone input with shared references
			input := cloneAgentInput(inputs[i])
			input.RetrievedReferences = sharedRefs

			// Preserve candidate ID in metadata
			if input.Metadata == nil {
				input.Metadata = make(map[string]string)
			}
			input.Metadata["candidate_id"] = fmt.Sprintf("%d", i)

			publisher.emit(domainagent.EventCandidateStart, i, domainagent.StatusRunning, time.Now().UTC(), time.Time{}, nil)

			// Create fresh runner for this candidate (skip retriever - we already have shared refs)
			agents := map[domainagent.StageName]domainagent.BaseAgent{
				domainagent.StagePlanner:    r.agentFactory.CreatePlanner(),
				domainagent.StageVisualizer: r.agentFactory.CreateVisualizer(),
				domainagent.StageCritic:     r.agentFactory.CreateCritic(),
			}
			if stylist := r.agentFactory.CreateStylist(); stylist != nil {
				agents[domainagent.StageStylist] = stylist
			}
			runner := NewRunner(agents)

			runHandle, err := runner.Start(gctx, input)
			if err != nil {
				mu.Lock()
				results[i] = domainagent.CandidateResult{
					CandidateID: i,
					Status:      domainagent.StatusFailed,
					Error:       &domainagent.ErrorDetail{Message: err.Error()},
				}
				failed++
				mu.Unlock()
				publisher.emit(domainagent.EventCandidateComplete, i, domainagent.StatusFailed, time.Now().UTC(), time.Time{}, err)
				return nil // Don't fail the whole batch
			}

			// Drain events (optional: could forward to batch events)
			for range runHandle.Events() {
			}

			result, err := runHandle.Wait()
			if err != nil {
				mu.Lock()
				results[i] = domainagent.CandidateResult{
					CandidateID: i,
					Status:      domainagent.StatusFailed,
					Error:       &domainagent.ErrorDetail{Message: err.Error()},
				}
				failed++
				mu.Unlock()
				publisher.emit(domainagent.EventCandidateComplete, i, domainagent.StatusFailed, time.Now().UTC(), time.Time{}, err)
				return nil
			}

			mu.Lock()
			results[i] = domainagent.CandidateResult{
				CandidateID: i,
				SessionID:   result.Session.SessionID,
				Status:      domainagent.StatusCompleted,
				Artifacts:   result.Session.FinalOutput.GeneratedArtifacts,
			}
			successful++
			mu.Unlock()

			publisher.emit(domainagent.EventCandidateComplete, i, domainagent.StatusCompleted, time.Now().UTC(), time.Time{}, nil)
			return nil
		})
	}

	_ = g.Wait() // Wait for all candidates

	completedAt := time.Now().UTC()
	batchResult := domainagent.BatchResult{
		BatchID:    batchID,
		Results:    results,
		Successful: successful,
		Failed:     failed,
		Timing: domainagent.BatchTiming{
			StartedAt:   startTime,
			CompletedAt: completedAt,
			Duration:    completedAt.Sub(startTime),
		},
	}

	publisher.emit(domainagent.EventBatchCompleted, 0, domainagent.StatusCompleted, startTime, completedAt, nil)
	r.storeResult(batchID, &batchResult)
	handle.setOutcome(batchResult, nil)
}

func (r *BatchRunner) runSharedRetriever(
	ctx context.Context,
	input domainagent.AgentInput,
) ([]domainagent.RetrievedReference, error) {
	retriever := r.agentFactory.CreateRetriever()
	if err := retriever.Initialize(ctx); err != nil {
		return nil, err
	}

	output, err := retriever.Execute(ctx, input)
	if err != nil {
		_ = retriever.Cleanup(ctx)
		return nil, err
	}

	_ = retriever.Cleanup(ctx)
	return output.RetrievedReferences, nil
}

type batchEventPublisher struct {
	batchID  string
	sequence int64
	out      chan<- domainagent.BatchEvent
}

func newBatchEventPublisher(batchID string, out chan<- domainagent.BatchEvent) *batchEventPublisher {
	return &batchEventPublisher{
		batchID: batchID,
		out:     out,
	}
}

func (p *batchEventPublisher) emit(
	eventType domainagent.EventType,
	candidateID int,
	status domainagent.RunStatus,
	startedAt, completedAt time.Time,
	err error,
) {
	p.sequence++
	var errorDetail *domainagent.ErrorDetail
	if err != nil {
		errorDetail = &domainagent.ErrorDetail{Message: err.Error()}
	}
	var duration time.Duration
	if !startedAt.IsZero() && !completedAt.IsZero() {
		duration = completedAt.Sub(startedAt)
	}
	p.out <- domainagent.BatchEvent{
		Sequence:    p.sequence,
		BatchID:     p.batchID,
		CandidateID: candidateID,
		Type:        eventType,
		Status:      status,
		OccurredAt:  time.Now().UTC(),
		Timing: domainagent.Timing{
			StartedAt:   startedAt,
			CompletedAt: completedAt,
			Duration:    duration,
		},
		Error: errorDetail,
	}
}
