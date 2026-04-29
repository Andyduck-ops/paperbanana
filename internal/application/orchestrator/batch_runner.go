package orchestrator

import (
	"context"
	"fmt"
	"log"
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

type snapshotStoreProvider interface {
	SnapshotStore() SnapshotStore
}

// BatchResultStore persists batch results for recovery across restarts.
type BatchResultStore interface {
	Save(result *domainagent.BatchResult) error
	Get(batchID string) (*domainagent.BatchResult, error)
}

// BatchProgress 表示批处理的中间进度状态
type BatchProgress struct {
	BatchID     string                        `json:"batch_id"`
	Status      domainagent.RunStatus         `json:"status"`
	Total       int                           `json:"total"`
	Completed   int                           `json:"completed"`
	Failed      int                           `json:"failed"`
	Results     []domainagent.CandidateResult `json:"results,omitempty"`
	StartedAt   time.Time                     `json:"started_at"`
	UpdatedAt   time.Time                     `json:"updated_at"`
	ErrorDetail *domainagent.ErrorDetail      `json:"error,omitempty"`
}

// BatchRunner executes multiple candidates in parallel with shared retriever.
type BatchRunner struct {
	agentFactory  AgentFactory
	maxConcurrent int
	eventBuffer   int
	results       map[string]*domainagent.BatchResult
	progress      map[string]*BatchProgress // 中间进度跟踪
	resultStore   BatchResultStore
	mu            sync.RWMutex
	resultTTL     time.Duration
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
		progress:      make(map[string]*BatchProgress),
		resultTTL:     30 * time.Minute,
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

// WithBatchResultTTL sets the TTL for in-memory batch results.
func WithBatchResultTTL(ttl time.Duration) BatchOption {
	return func(r *BatchRunner) {
		if ttl > 0 {
			r.resultTTL = ttl
		}
	}
}

// WithBatchResultStore sets the persistent result store.
func WithBatchResultStore(store BatchResultStore) BatchOption {
	return func(r *BatchRunner) {
		r.resultStore = store
	}
}

// GetBatchResult retrieves a stored batch result by ID.
// Results are kept in memory for a limited time after completion.
// Falls back to persistent store if not found in memory.
func (r *BatchRunner) GetBatchResult(batchID string) (*domainagent.BatchResult, error) {
	// First check in-memory cache
	r.mu.RLock()
	result, ok := r.results[batchID]
	r.mu.RUnlock()

	if ok {
		return result, nil
	}

	// Fall back to persistent store
	if r.resultStore != nil {
		return r.resultStore.Get(batchID)
	}

	return nil, fmt.Errorf("batch result not found: %s", batchID)
}

// GetBatchProgress 获取批处理的当前进度（用于状态检查）
func (r *BatchRunner) GetBatchProgress(batchID string) (*BatchProgress, error) {
	r.mu.RLock()
	progress, ok := r.progress[batchID]
	r.mu.RUnlock()

	if ok {
		return progress, nil
	}

	// 如果没有找到进度，尝试从结果中恢复
	result, err := r.GetBatchResult(batchID)
	if err != nil {
		return nil, fmt.Errorf("batch progress not found: %s", batchID)
	}

	// 从结果重建进度
	status := domainagent.StatusCompleted
	if result.Failed > 0 && result.Successful == 0 {
		status = domainagent.StatusFailed
	} else if result.Failed > 0 {
		status = domainagent.StatusRunning // 部分完成视为进行中
	}

	return &BatchProgress{
		BatchID:   batchID,
		Status:    status,
		Total:     len(result.Results),
		Completed: result.Successful,
		Failed:    result.Failed,
		Results:   result.Results,
		StartedAt: result.Timing.StartedAt,
		UpdatedAt: result.Timing.CompletedAt,
	}, nil
}

// storeResult saves a batch result for later retrieval.
// Persists to the database if a result store is configured.
// Old results are evicted if the in-memory cache exceeds a reasonable size.
func (r *BatchRunner) storeResult(batchID string, result *domainagent.BatchResult) {
	r.mu.Lock()
	r.results[batchID] = result
	// 完成后删除进度跟踪
	delete(r.progress, batchID)
	// Evict old completed results to prevent unbounded growth
	r.evictOldResultsLocked()
	r.mu.Unlock()

	// Persist to database if store is configured
	if r.resultStore != nil {
		if err := r.resultStore.Save(result); err != nil {
			// Log but don't fail - memory storage is still available
			log.Printf("Failed to persist batch result %s: %v", batchID, err)
		}
	}
}

// evictOldResultsLocked removes completed results older than resultTTL.
// Must be called with mu held for writing.
func (r *BatchRunner) evictOldResultsLocked() {
	const maxInMemoryResults = 100
	if len(r.results) <= maxInMemoryResults {
		return
	}
	cutoff := time.Now().UTC().Add(-r.resultTTL)
	for id, result := range r.results {
		if result.Timing.CompletedAt.Before(cutoff) {
			delete(r.results, id)
		}
	}
}

// updateProgress 更新批处理中间进度
func (r *BatchRunner) updateProgress(progress *BatchProgress) {
	r.mu.Lock()
	r.progress[progress.BatchID] = progress
	// Evict stale progress entries for batches that haven't updated recently
	const maxInMemoryProgress = 50
	if len(r.progress) > maxInMemoryProgress {
		cutoff := time.Now().UTC().Add(-r.resultTTL)
		for id, p := range r.progress {
			if p.UpdatedAt.Before(cutoff) {
				delete(r.progress, id)
			}
		}
	}
	r.mu.Unlock()

	// 创建临时结果用于持久化
	tempResult := &domainagent.BatchResult{
		BatchID:    progress.BatchID,
		Results:    progress.Results,
		Successful: progress.Completed,
		Failed:     progress.Failed,
		Timing: domainagent.BatchTiming{
			StartedAt: progress.StartedAt,
			UpdatedAt: progress.UpdatedAt,
		},
	}

	// 持久化中间进度
	if r.resultStore != nil {
		if err := r.resultStore.Save(tempResult); err != nil {
			log.Printf("Failed to persist batch progress %s: %v", progress.BatchID, err)
		}
	}
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

	// 初始化进度跟踪
	progress := &BatchProgress{
		BatchID:   batchID,
		Status:    domainagent.StatusRunning,
		Total:     len(inputs),
		Completed: 0,
		Failed:    0,
		Results:   make([]domainagent.CandidateResult, len(inputs)),
		StartedAt: startTime,
		UpdatedAt: startTime,
	}

	// Emit batch start
	publisher.emit(domainagent.EventBatchStarted, 0, domainagent.StatusRunning, startTime, time.Time{}, nil)

	// Phase 1: Run retriever once and share results
	sharedRefs, err := r.runSharedRetriever(ctx, inputs[0])
	if err != nil {
		completedAt := time.Now().UTC()
		publisher.emit(domainagent.EventBatchCompleted, 0, domainagent.StatusFailed, startTime, completedAt, err)

		progress.Status = domainagent.StatusFailed
		progress.UpdatedAt = completedAt
		progress.ErrorDetail = &domainagent.ErrorDetail{Message: err.Error()}
		r.updateProgress(progress)

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
	completedCount := 0

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
			runnerOptions := make([]RunnerOption, 0, 1)
			if provider, ok := r.agentFactory.(snapshotStoreProvider); ok {
				if snapshotStore := provider.SnapshotStore(); snapshotStore != nil {
					runnerOptions = append(runnerOptions, WithSnapshotStore(snapshotStore))
				}
			}
			runner := NewRunner(agents, runnerOptions...)

			runHandle, err := runner.Start(gctx, input)
			if err != nil {
				mu.Lock()
				results[i] = domainagent.CandidateResult{
					CandidateID: i,
					Status:      domainagent.StatusFailed,
					Error:       &domainagent.ErrorDetail{Message: err.Error()},
				}
				progress.Results[i] = results[i]
				failed++
				completedCount++
				progress.Failed = failed
				progress.Completed = completedCount
				progress.UpdatedAt = time.Now().UTC()
				mu.Unlock()

				// 每个候选完成后更新进度
				r.updateProgress(progress)

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
				progress.Results[i] = results[i]
				failed++
				completedCount++
				progress.Failed = failed
				progress.Completed = completedCount
				progress.UpdatedAt = time.Now().UTC()
				mu.Unlock()

				// 每个候选完成后更新进度
				r.updateProgress(progress)

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
			progress.Results[i] = results[i]
			successful++
			completedCount++
			progress.Completed = completedCount
			progress.Failed = failed
			progress.UpdatedAt = time.Now().UTC()
			mu.Unlock()

			// 每个候选完成后更新进度
			r.updateProgress(progress)

			publisher.emit(domainagent.EventCandidateComplete, i, domainagent.StatusCompleted, time.Now().UTC(), time.Time{}, nil)
			return nil
		})
	}

	_ = g.Wait() // Wait for all candidates

	completedAt := time.Now().UTC()

	// 最终进度更新
	progress.Status = domainagent.StatusCompleted
	progress.UpdatedAt = completedAt
	r.updateProgress(progress)

	batchResult := domainagent.BatchResult{
		BatchID:    batchID,
		Results:    results,
		Successful: successful,
		Failed:     failed,
		Timing: domainagent.BatchTiming{
			StartedAt:   startTime,
			CompletedAt: completedAt,
			Duration:    completedAt.Sub(startTime),
			UpdatedAt:   completedAt,
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
