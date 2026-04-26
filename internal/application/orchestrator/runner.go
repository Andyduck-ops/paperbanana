package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
	agentstate "github.com/paperbanana/paperbanana/internal/infrastructure/agentstate"
)

type RunnerOption func(*Runner)

var (
	ErrResumeRequiresSession = errors.New("orchestrator: resume requires session id")
	ErrResumeSnapshotMissing = errors.New("orchestrator: resume snapshot not found")
	ErrResumeStoreMissing    = errors.New("orchestrator: resume requires snapshot store")
	ErrResumeSessionNotValid = errors.New("orchestrator: resume session state is not valid for resumption")
	ErrResumePipelineEmpty   = errors.New("orchestrator: resume session has empty pipeline")
)

type SnapshotStore interface {
	Save(session domainagent.SessionState, state domainagent.AgentState) error
	Restore(sessionID string, stage domainagent.StageName) (agentstate.Snapshot, error)
}

// StageTimeouts maps stage names to their timeout durations.
type StageTimeouts map[domainagent.StageName]time.Duration

// DefaultStageTimeouts returns the default timeout configuration.
func DefaultStageTimeouts() StageTimeouts {
	return StageTimeouts{
		domainagent.StageRetriever:  180 * time.Second,
		domainagent.StagePlanner:    180 * time.Second,
		domainagent.StageStylist:    180 * time.Second,
		domainagent.StageVisualizer: 180 * time.Second,
		domainagent.StageCritic:     360 * time.Second,
	}
}

type Runner struct {
	agents          map[domainagent.StageName]domainagent.BaseAgent
	pipeline        []domainagent.StageName
	eventBuffer     int
	snapshotStore   SnapshotStore
	stageTimeouts   StageTimeouts
	gracefulDegrade bool
}

func NewCanonicalRunner(retriever, planner, stylist, visualizer, critic domainagent.BaseAgent, opts ...RunnerOption) *Runner {
	agents := map[domainagent.StageName]domainagent.BaseAgent{
		domainagent.StageRetriever:  retriever,
		domainagent.StagePlanner:    planner,
		domainagent.StageStylist:    stylist,
		domainagent.StageVisualizer: visualizer,
		domainagent.StageCritic:     critic,
	}
	return NewRunner(agents, opts...)
}

func NewRunner(agents map[domainagent.StageName]domainagent.BaseAgent, opts ...RunnerOption) *Runner {
	runner := &Runner{
		agents:        cloneRegistry(agents),
		pipeline:      orderedPipeline(agents),
		eventBuffer:   32,
		stageTimeouts: DefaultStageTimeouts(),
	}

	for _, opt := range opts {
		opt(runner)
	}

	return runner
}

func WithEventBuffer(size int) RunnerOption {
	return func(r *Runner) {
		if size > 0 {
			r.eventBuffer = size
		}
	}
}

func WithSnapshotStore(store SnapshotStore) RunnerOption {
	return func(r *Runner) {
		r.snapshotStore = store
	}
}

// WithStageTimeouts sets custom timeout durations for each stage.
func WithStageTimeouts(timeouts StageTimeouts) RunnerOption {
	return func(r *Runner) {
		if timeouts == nil {
			return
		}
		// Merge with defaults to ensure all stages have timeouts
		r.stageTimeouts = DefaultStageTimeouts()
		for stage, timeout := range timeouts {
			if timeout > 0 {
				r.stageTimeouts[stage] = timeout
			}
		}
	}
}

// WithGracefulDegradation enables graceful degradation for non-critical stages.
// When enabled, failures in non-critical stages (retriever, stylist, critic) will
// allow the pipeline to continue with fallback behavior.
func WithGracefulDegradation(enabled bool) RunnerOption {
	return func(r *Runner) {
		r.gracefulDegrade = enabled
	}
}

func (r *Runner) Start(ctx context.Context, input domainagent.AgentInput) (*RunHandle, error) {
	pipeline := r.pipelineForInput(input.Metadata)
	tracker := newSessionTracker(input, pipeline)
	return r.startWithTracker(ctx, tracker, pipeline)
}

func (r *Runner) Resume(ctx context.Context, input domainagent.AgentInput) (*RunHandle, error) {
	tracker, remaining, err := r.resumeTracker(input)
	if err != nil {
		return nil, err
	}
	return r.startWithTracker(ctx, tracker, remaining)
}

func (r *Runner) startWithTracker(ctx context.Context, tracker *sessionTracker, stages []domainagent.StageName) (*RunHandle, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	internalEvents := make(chan domainagent.Event, r.eventBuffer)
	publicEvents := make(chan domainagent.Event, r.eventBuffer)
	handle := newRunHandle(publicEvents)

	go func() {
		var (
			result RunResult
			runErr error
		)

		group, groupCtx := errgroup.WithContext(ctx)
		group.Go(func() error {
			defer close(publicEvents)
			for event := range internalEvents {
				select {
				case publicEvents <- event:
				case <-groupCtx.Done():
				}
			}
			return nil
		})
		group.Go(func() error {
			defer close(internalEvents)
			result, runErr = r.execute(groupCtx, tracker, stages, internalEvents)
			return nil
		})

		_ = group.Wait()
		handle.setOutcome(result, runErr)
	}()

	return handle, nil
}

func (r *Runner) execute(ctx context.Context, tracker *sessionTracker, stages []domainagent.StageName, events chan<- domainagent.Event) (RunResult, error) {
	publisher := eventPublisher{
		sessionID: tracker.state.SessionID,
		requestID: tracker.state.RequestID,
		out:       events,
	}
	var (
		lastCompletedState domainagent.AgentState
		hasCompletedStage  bool
		degradedStages     []string
	)

	// GD-UI-004: Emit resume_start event if this is a restored session
	if tracker.state.Restore.RestoredFrom != "" {
		resumeMetadata := map[string]string{
			"resumed_from_stage": string(tracker.state.Restore.RestoredFrom),
			"session_id":         tracker.state.SessionID,
		}
		publisher.emit(domainagent.EventResumeStarted, tracker.state.Restore.RestoredFrom, domainagent.StatusRunning, domainagent.Timing{}, nil, resumeMetadata)
	}

	publisher.emit(domainagent.EventRunStarted, "", domainagent.StatusRunning, domainagent.Timing{}, nil, tracker.state.Metadata)

	for _, stage := range stages {
		stageInput := prepareStageInput(stage, tracker.stageInput(stage), tracker.state.InitialInput)
		if err := ctx.Err(); err != nil {
			return r.finishStageError(ctx, tracker, publisher, stage, stageInput, time.Now().UTC(), err, domainagent.BaseAgent(nil))
		}

		stageAgent, ok := r.agents[stage]
		if !ok {
			return r.finishStageError(ctx, tracker, publisher, stage, stageInput, time.Now().UTC(), fmt.Errorf("missing agent for stage %s", stage), nil)
		}

		startedAt := time.Now().UTC()
		// Add agent name to metadata for stage_started event
		stageMetadata := cloneStringMap(stageInput.Metadata)
		if stageMetadata == nil {
			stageMetadata = map[string]string{}
		}
		stageMetadata["agent"] = stageAgentName(stage)
		publisher.emit(domainagent.EventStageStarted, stage, domainagent.StatusRunning, domainagent.Timing{StartedAt: startedAt}, nil, stageMetadata)

		// Get stage timeout
		stageTimeout := r.stageTimeouts[stage]
		if stageTimeout <= 0 {
			stageTimeout = 60 * time.Second // fallback default
		}

		// Create stage context with timeout (cancel is called explicitly in all code paths below)
		stageCtx, cancel := context.WithTimeout(ctx, stageTimeout)

		if err := stageAgent.Initialize(stageCtx); err != nil {
			cancel()
			// Check if graceful degradation is enabled and stage is non-critical
			if r.gracefulDegrade && !isCriticalStage(stage) {
				_ = r.handleStageDegradation(stage, stageInput, err, startedAt, tracker, publisher)
				degradedStages = append(degradedStages, string(stage))
				continue
			}
			return r.finishStageError(stageCtx, tracker, publisher, stage, stageInput, startedAt, err, stageAgent)
		}

		output, err := stageAgent.Execute(stageCtx, stageInput)
		if err != nil {
			cancel()
			if cleanupErr := stageAgent.Cleanup(ctx); cleanupErr != nil {
				err = errors.Join(err, cleanupErr)
			}
			// Check if it was a timeout
			if errors.Is(stageCtx.Err(), context.DeadlineExceeded) {
				err = fmt.Errorf("stage %s timed out after %s: %w", stage, stageTimeout, err)
			}
			// Check if graceful degradation is enabled and stage is non-critical
			if r.gracefulDegrade && !isCriticalStage(stage) {
				_ = r.handleStageDegradation(stage, stageInput, err, startedAt, tracker, publisher)
				degradedStages = append(degradedStages, string(stage))
				continue
			}
			return r.finishStageError(ctx, tracker, publisher, stage, stageInput, startedAt, err, stageAgent)
		}

		cancel()
		if err := stageAgent.Cleanup(ctx); err != nil {
			return r.finishStageError(ctx, tracker, publisher, stage, stageInput, startedAt, err, stageAgent)
		}

		completedAt := time.Now().UTC()
		timing := domainagent.Timing{
			StartedAt:   startedAt,
			CompletedAt: completedAt,
			Duration:    completedAt.Sub(startedAt),
		}

		stageState := stageAgent.GetState()
		stageState.Stage = stage
		stageState.Status = domainagent.StatusCompleted
		stageState.Timing = timing
		stageState.Input = cloneAgentInput(stageInput)
		stageState.Output = cloneAgentOutput(output)
		stageState.Error = nil
		stageState.Restore = stageInput.Restore

		tracker.completeStage(stageState, output)
		if err := r.persistSnapshot(tracker, stageState); err != nil {
			return r.finishPersistenceError(tracker, publisher, stage, stageInput, timing, err)
		}
		publisher.emit(domainagent.EventStageCompleted, stage, domainagent.StatusCompleted, timing, nil, stageEventMetadata(output))
		lastCompletedState = cloneAgentState(stageState)
		hasCompletedStage = true
	}

	completedAt := time.Now().UTC()
	tracker.completeRun(completedAt)
	if hasCompletedStage {
		if err := r.persistSnapshot(tracker, lastCompletedState); err != nil {
			return RunResult{
				Session:     tracker.snapshot(),
				FailedStage: lastCompletedState.Stage,
			}, err
		}
	}
	// Add degraded stages info to final metadata if any
	if len(degradedStages) > 0 {
		finalMetadata := cloneStringMap(tracker.state.Metadata)
		if finalMetadata == nil {
			finalMetadata = map[string]string{}
		}
		finalMetadata["degraded_stages"] = strings.Join(degradedStages, ",")
		tracker.state.Metadata = finalMetadata
	}
	publisher.emit(domainagent.EventRunCompleted, tracker.state.CurrentStage, domainagent.StatusCompleted, domainagent.Timing{CompletedAt: completedAt}, nil, tracker.state.Metadata)

	return RunResult{Session: tracker.snapshot()}, nil
}

// handleStageDegradation handles a failed non-critical stage with graceful degradation.
func (r *Runner) handleStageDegradation(stage domainagent.StageName, input domainagent.AgentInput, originalErr error, startedAt time.Time, tracker *sessionTracker, publisher eventPublisher) domainagent.AgentOutput {
	completedAt := time.Now().UTC()
	timing := domainagent.Timing{
		StartedAt:   startedAt,
		CompletedAt: completedAt,
		Duration:    completedAt.Sub(startedAt),
	}

	// Create fallback output
	output := fallbackOutputForStage(stage, input)

	// Create degraded stage state
	stageState := domainagent.AgentState{
		Stage:   stage,
		Status:  domainagent.StatusCompleted,
		Timing:  timing,
		Input:   cloneAgentInput(input),
		Output:  cloneAgentOutput(output),
		Restore: input.Restore,
		Error: &domainagent.ErrorDetail{
			Message:    originalErr.Error(),
			Code:       string(domainagent.ClassifyError(originalErr)),
			Stage:      stage,
			Retryable:  true,
			Suggestion: fmt.Sprintf("Stage %s failed but pipeline continued with degraded mode", stage),
		},
	}

	tracker.completeStage(stageState, output)

	// Emit degraded completion event
	degradedMetadata := cloneStringMap(output.Metadata)
	if degradedMetadata == nil {
		degradedMetadata = map[string]string{}
	}
	degradedMetadata["degraded"] = "true"
	degradedMetadata["original_error"] = originalErr.Error()
	publisher.emit(domainagent.EventStageCompleted, stage, domainagent.StatusCompleted, timing, stageState.Error, degradedMetadata)

	return output
}

func (r *Runner) finishStageError(ctx context.Context, tracker *sessionTracker, publisher eventPublisher, stage domainagent.StageName, input domainagent.AgentInput, startedAt time.Time, err error, stageAgent domainagent.BaseAgent) (RunResult, error) {
	completedAt := time.Now().UTC()
	timing := domainagent.Timing{
		StartedAt:   startedAt,
		CompletedAt: completedAt,
		Duration:    completedAt.Sub(startedAt),
	}

	// Classify and standardize the error
	errorCode := domainagent.ClassifyError(err)
	detail := domainagent.WrapAgentError(err, stage, errorCode)

	// Add timeout info if applicable
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		stageTimeout := r.stageTimeouts[stage]
		if stageTimeout <= 0 {
			stageTimeout = 60 * time.Second
		}
		detail.Code = string(domainagent.ErrCodeStageTimeout)
		detail.Message = fmt.Sprintf("stage %s timed out after %s: %s", stage, stageTimeout, err.Error())
		detail.Category = string(domainagent.ErrorCategoryTransient)
		detail.Suggestion = fmt.Sprintf("The %s stage took too long (limit: %s). Please try again.", stage, stageTimeout)
	}

	stageState := domainagent.AgentState{
		Stage:   stage,
		Status:  domainagent.StatusFailed,
		Timing:  timing,
		Input:   cloneAgentInput(input),
		Output:  domainagent.AgentOutput{Stage: stage},
		Error:   detail,
		Restore: input.Restore,
	}

	if stageAgent != nil {
		stageState = stageAgent.GetState()
		stageState.Stage = stage
		stageState.Timing = timing
		stageState.Input = cloneAgentInput(input)
		stageState.Error = detail
		stageState.Restore = input.Restore
	}

	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.Canceled) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
		stageState.Status = domainagent.StatusCanceled
		tracker.failStage(stageState, domainagent.StatusCanceled, detail)
		if persistErr := r.persistSnapshot(tracker, stageState); persistErr != nil {
			err = errors.Join(err, persistErr)
		}
		publisher.emit(domainagent.EventRunCanceled, stage, domainagent.StatusCanceled, timing, detail, input.Metadata)
		return RunResult{
			Session:     tracker.snapshot(),
			FailedStage: stage,
		}, err
	}

	stageState.Status = domainagent.StatusFailed
	tracker.failStage(stageState, domainagent.StatusFailed, detail)
	if persistErr := r.persistSnapshot(tracker, stageState); persistErr != nil {
		err = errors.Join(err, persistErr)
	}
	publisher.emit(domainagent.EventStageFailed, stage, domainagent.StatusFailed, timing, detail, input.Metadata)
	publisher.emit(domainagent.EventRunFailed, stage, domainagent.StatusFailed, timing, detail, input.Metadata)

	return RunResult{
		Session:     tracker.snapshot(),
		FailedStage: stage,
	}, err
}

func (r *Runner) finishPersistenceError(tracker *sessionTracker, publisher eventPublisher, stage domainagent.StageName, input domainagent.AgentInput, timing domainagent.Timing, err error) (RunResult, error) {
	detail := &domainagent.ErrorDetail{
		Message: err.Error(),
		Stage:   stage,
	}

	tracker.state.CurrentStage = stage
	tracker.state.FailedStage = stage
	tracker.state.Status = domainagent.StatusFailed
	tracker.state.Error = cloneErrorDetail(detail)
	tracker.state.UpdatedAt = timing.CompletedAt
	tracker.state.CompletedAt = timing.CompletedAt
	if count := len(tracker.state.StageStates); count > 0 && tracker.state.StageStates[count-1].Stage == stage {
		tracker.state.StageStates[count-1].Status = domainagent.StatusFailed
		tracker.state.StageStates[count-1].Error = cloneErrorDetail(detail)
		tracker.state.StageStates[count-1].Restore = input.Restore
	}

	publisher.emit(domainagent.EventStageFailed, stage, domainagent.StatusFailed, timing, detail, input.Metadata)
	publisher.emit(domainagent.EventRunFailed, stage, domainagent.StatusFailed, timing, detail, input.Metadata)

	return RunResult{
		Session:     tracker.snapshot(),
		FailedStage: stage,
	}, err
}

func (r *Runner) persistSnapshot(tracker *sessionTracker, state domainagent.AgentState) error {
	if r.snapshotStore == nil {
		return nil
	}
	return r.snapshotStore.Save(tracker.snapshot(), state)
}

func (r *Runner) resumeTracker(input domainagent.AgentInput) (*sessionTracker, []domainagent.StageName, error) {
	if strings.TrimSpace(input.SessionID) == "" {
		return nil, nil, ErrResumeRequiresSession
	}
	if r.snapshotStore == nil {
		return nil, nil, ErrResumeStoreMissing
	}

	searchStages := domainagent.CanonicalPipeline()
	for index := len(searchStages) - 1; index >= 0; index-- {
		stage := searchStages[index]
		snapshot, err := r.snapshotStore.Restore(input.SessionID, stage)
		switch {
		case err == nil:
		case errors.Is(err, agentstate.ErrSnapshotNotFound), errors.Is(err, agentstate.ErrInvalidSnapshot):
			continue
		default:
			return nil, nil, err
		}

		// Validate session state is resumable
		// Only sessions that failed, were canceled, or have remaining stages can be resumed
		sessionStatus := snapshot.Session.Status
		if sessionStatus == domainagent.StatusRunning {
			// A running session might be a stale state; skip to find a better restore point
			continue
		}

		// Validate the snapshot stage has completed status
		if snapshot.Stage.Status != domainagent.StatusCompleted {
			continue
		}

		// Validate pipeline integrity
		pipeline := r.pipelineForResume(snapshot, input.Metadata)
		if len(pipeline) == 0 {
			return nil, nil, fmt.Errorf("%w: session %s", ErrResumePipelineEmpty, input.SessionID)
		}

		// Validate essential session fields
		if strings.TrimSpace(snapshot.Session.SessionID) == "" {
			return nil, nil, fmt.Errorf("%w: session has empty session_id", ErrResumeSessionNotValid)
		}

		tracker := newRestoredSessionTracker(snapshot, input, pipeline)
		if err := r.restoreCompletedStates(snapshot); err != nil {
			return nil, nil, err
		}

		remaining := remainingPipeline(snapshot.Stage.Stage, pipeline)

		// If no remaining stages, check if this was a failed session that can be retried
		if len(remaining) == 0 && sessionStatus != domainagent.StatusFailed && sessionStatus != domainagent.StatusCanceled {
			// Session completed successfully with no remaining stages - nothing to resume
			continue
		}

		return tracker, remaining, nil
	}

	return nil, nil, fmt.Errorf("%w: %s", ErrResumeSnapshotMissing, input.SessionID)
}

func (r *Runner) restoreCompletedStates(snapshot agentstate.Snapshot) error {
	completedStates := make(map[domainagent.StageName]domainagent.AgentState, len(snapshot.Session.StageStates)+1)
	for _, state := range snapshot.Session.StageStates {
		if state.Status != domainagent.StatusCompleted {
			continue
		}
		completedStates[state.Stage] = cloneAgentState(state)
	}
	if snapshot.Stage.Status == domainagent.StatusCompleted {
		completedStates[snapshot.Stage.Stage] = cloneAgentState(agentStateFromSnapshot(snapshot.Stage))
	}

	restorePoint := snapshot.Stage.Stage
	if _, ok := completedStates[restorePoint]; !ok {
		return fmt.Errorf("orchestrator: restore point %s cannot be rehydrated safely: completed state missing", restorePoint)
	}

	restorePointAgent, ok := r.agents[restorePoint]
	if !ok || restorePointAgent == nil {
		return fmt.Errorf("orchestrator: restore point %s cannot be rehydrated safely: agent not registered", restorePoint)
	}

	for _, stage := range r.pipeline {
		state, ok := completedStates[stage]
		if !ok {
			continue
		}

		stageAgent, ok := r.agents[stage]
		if !ok || stageAgent == nil {
			if stage == restorePoint {
				return fmt.Errorf("orchestrator: restore point %s cannot be rehydrated safely: agent not registered", restorePoint)
			}
			continue
		}

		if err := stageAgent.RestoreState(cloneAgentState(state)); err != nil {
			if stage == restorePoint {
				return fmt.Errorf("orchestrator: restore point %s cannot be rehydrated safely: %w", restorePoint, err)
			}
			return fmt.Errorf("orchestrator: restore completed state for %s: %w", stage, err)
		}
	}

	return nil
}

func agentStateFromSnapshot(snapshot agentstate.StageSnapshot) domainagent.AgentState {
	return domainagent.AgentState{
		Stage:   snapshot.Stage,
		Status:  snapshot.Status,
		Timing:  snapshot.Timing,
		Input:   cloneAgentInput(snapshot.Input),
		Output:  cloneAgentOutput(snapshot.Output),
		Error:   cloneErrorDetail(snapshot.Error),
		Restore: snapshot.Restore,
	}
}

func remainingPipeline(current domainagent.StageName, pipeline []domainagent.StageName) []domainagent.StageName {
	for index, stage := range pipeline {
		if stage == current {
			return append([]domainagent.StageName(nil), pipeline[index+1:]...)
		}
	}
	return nil
}

func newRestoredSessionTracker(snapshot agentstate.Snapshot, input domainagent.AgentInput, pipeline []domainagent.StageName) *sessionTracker {
	restore := input.Restore
	if restore.SnapshotVersion == "" {
		restore.SnapshotVersion = snapshot.SchemaVersion
	}
	restore.RestoredFrom = snapshot.Stage.Stage
	if restore.RestoredAt.IsZero() {
		restore.RestoredAt = time.Now().UTC()
	}

	initialInput := cloneAgentInput(snapshot.Session.InitialInput)
	if strings.TrimSpace(initialInput.SessionID) == "" {
		initialInput = cloneAgentInput(snapshot.Stage.Input)
		initialInput.Stage = ""
	}
	initialInput.SessionID = snapshot.Session.SessionID
	if input.RequestID != "" {
		initialInput.RequestID = input.RequestID
	}

	currentInput := mergeAgentInput(snapshot.Stage.Input, snapshot.Stage.Output)
	currentInput.SessionID = snapshot.Session.SessionID
	currentInput.RequestID = nonEmpty(input.RequestID, snapshot.Session.RequestID)
	currentInput.Restore = restore

	state := domainagent.SessionState{
		SchemaVersion: snapshot.Session.SchemaVersion,
		SessionID:     snapshot.Session.SessionID,
		RequestID:     currentInput.RequestID,
		Status:        domainagent.StatusRunning,
		CurrentStage:  snapshot.Stage.Stage,
		Pipeline:      append([]domainagent.StageName(nil), snapshot.Session.Pipeline...),
		InitialInput:  initialInput,
		StageStates:   cloneAgentStates(snapshot.Session.StageStates),
		FinalOutput:   cloneAgentOutput(snapshot.Session.FinalOutput),
		Restore:       restore,
		Metadata:      mergeStringMaps(snapshot.Session.Metadata, input.Metadata),
		StartedAt:     snapshot.Session.StartedAt,
		UpdatedAt:     restore.RestoredAt,
	}
	if len(state.Pipeline) == 0 {
		state.Pipeline = append([]domainagent.StageName(nil), pipeline...)
	}

	return &sessionTracker{
		state:        state,
		currentInput: currentInput,
	}
}

func cloneAgentStates(states []domainagent.AgentState) []domainagent.AgentState {
	if len(states) == 0 {
		return nil
	}

	cloned := make([]domainagent.AgentState, len(states))
	for i, state := range states {
		cloned[i] = cloneAgentState(state)
	}
	return cloned
}

func nonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func prepareStageInput(stage domainagent.StageName, input, initial domainagent.AgentInput) domainagent.AgentInput {
	switch stage {
	case domainagent.StageStylist:
		// Stylist receives planner's output as Content
		// The Content is already set by mergeAgentInput in completeStage
		// Just ensure Stylist has all context needed
		prepared := cloneAgentInput(input)
		prepared.Stage = stage
		// Stylist needs the visual intent and content from planner
		if prepared.VisualIntent.Goal == "" && initial.VisualIntent.Goal != "" {
			prepared.VisualIntent = cloneVisualIntent(initial.VisualIntent)
		}
		return prepared

	case domainagent.StageVisualizer:
		// Visualizer receives stylist's output as Content
		// mergeAgentInput already sets Content from previous stage output
		prepared := cloneAgentInput(input)
		prepared.Stage = stage
		// Ensure visualizer has all references and artifacts
		if len(prepared.RetrievedReferences) == 0 && len(initial.RetrievedReferences) > 0 {
			prepared.RetrievedReferences = cloneReferences(initial.RetrievedReferences)
		}
		return prepared

	case domainagent.StageCritic:
		// Critic needs initial content for comparison
		prepared := cloneAgentInput(input)
		prepared.Stage = stage
		if strings.TrimSpace(initial.Content) != "" {
			metadata := cloneStringMap(prepared.Metadata)
			if metadata == nil {
				metadata = map[string]string{}
			}
			if _, exists := metadata["orchestrator.initial_content"]; !exists {
				metadata["orchestrator.initial_content"] = initial.Content
			}
			prepared.Metadata = metadata
		}
		return prepared

	default:
		// For other stages (Retriever, Planner), use input as-is
		prepared := cloneAgentInput(input)
		prepared.Stage = stage
		return prepared
	}
}

func cloneRegistry(agents map[domainagent.StageName]domainagent.BaseAgent) map[domainagent.StageName]domainagent.BaseAgent {
	if len(agents) == 0 {
		return map[domainagent.StageName]domainagent.BaseAgent{}
	}

	cloned := make(map[domainagent.StageName]domainagent.BaseAgent, len(agents))
	for stage, agent := range agents {
		cloned[stage] = agent
	}
	return cloned
}

func orderedPipeline(agents map[domainagent.StageName]domainagent.BaseAgent) []domainagent.StageName {
	if len(agents) == 0 {
		return nil
	}

	pipeline := make([]domainagent.StageName, 0, len(agents))
	for _, stage := range domainagent.CanonicalPipeline() {
		agent, ok := agents[stage]
		if !ok {
			continue
		}
		if agent == nil {
			continue
		}
		pipeline = append(pipeline, stage)
	}
	return pipeline
}

func (r *Runner) pipelineForInput(metadata map[string]string) []domainagent.StageName {
	return r.filterPipeline(metadata, r.pipeline)
}

func (r *Runner) pipelineForResume(snapshot agentstate.Snapshot, metadata map[string]string) []domainagent.StageName {
	base := snapshot.Session.Pipeline
	if len(base) == 0 {
		base = r.pipeline
	}
	return r.filterPipeline(metadata, base)
}

func (r *Runner) filterPipeline(metadata map[string]string, base []domainagent.StageName) []domainagent.StageName {
	if len(base) == 0 {
		return nil
	}

	mode := strings.TrimSpace(metadata["config.pipeline_mode"])
	if mode == "" {
		return append([]domainagent.StageName(nil), base...)
	}

	// Validate pipeline mode - fall back to full pipeline for invalid modes
	// while preserving metadata for observability
	if !domainagent.IsValidPipelineMode(mode) {
		return append([]domainagent.StageName(nil), base...)
	}

	allowed := map[domainagent.StageName]bool{}
	switch mode {
	case domainagent.PipelineModeFull:
		return append([]domainagent.StageName(nil), base...)
	case domainagent.PipelineModePlannerCritic:
		allowed[domainagent.StagePlanner] = true
		allowed[domainagent.StageCritic] = true
	case domainagent.PipelineModeVanilla:
		allowed[domainagent.StageVisualizer] = true
	default:
		return append([]domainagent.StageName(nil), base...)
	}

	pipeline := make([]domainagent.StageName, 0, len(base))
	for _, stage := range base {
		if allowed[stage] {
			pipeline = append(pipeline, stage)
		}
	}
	if len(pipeline) == 0 {
		return append([]domainagent.StageName(nil), base...)
	}
	return pipeline
}

func stageEventMetadata(output domainagent.AgentOutput) map[string]string {
	metadata := cloneStringMap(output.Metadata)
	if metadata == nil {
		metadata = map[string]string{}
	}

	if metadata["summary"] == "" {
		if summary := eventSummary(output); summary != "" {
			metadata["summary"] = summary
		}
	}
	if len(output.GeneratedArtifacts) > 0 {
		metadata["artifact_count"] = strconv.Itoa(len(output.GeneratedArtifacts))
		metadata["artifact_kinds"] = artifactKinds(output.GeneratedArtifacts)
	}

	if len(metadata) == 0 {
		return nil
	}
	return metadata
}

func eventSummary(output domainagent.AgentOutput) string {
	switch {
	case output.Content != "":
		return truncateSummary(output.Content, 160)
	case len(output.GeneratedArtifacts) > 0:
		return output.GeneratedArtifacts[0].ID
	default:
		return ""
	}
}

func truncateSummary(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return value[:limit]
}

func artifactKinds(artifacts []domainagent.Artifact) string {
	if len(artifacts) == 0 {
		return ""
	}

	kinds := make([]string, 0, len(artifacts))
	for _, artifact := range artifacts {
		kinds = append(kinds, string(artifact.Kind))
	}
	return strings.Join(kinds, ",")
}

// stageAgentName returns the human-readable agent name for a stage.
// e.g., "retriever" -> "Retriever"
func stageAgentName(stage domainagent.StageName) string {
	if len(stage) == 0 {
		return ""
	}
	return strings.ToUpper(string(stage[:1])) + string(stage[1:])
}

// isCriticalStage returns true if the stage is critical and cannot be skipped.
// Critical stages: planner, visualizer
// Non-critical stages (can be degraded): retriever, stylist, critic
func isCriticalStage(stage domainagent.StageName) bool {
	switch stage {
	case domainagent.StagePlanner, domainagent.StageVisualizer:
		return true
	case domainagent.StageRetriever, domainagent.StageStylist, domainagent.StageCritic:
		return false
	default:
		return true
	}
}

// fallbackOutputForStage returns a fallback output when a non-critical stage fails.
func fallbackOutputForStage(stage domainagent.StageName, input domainagent.AgentInput) domainagent.AgentOutput {
	output := domainagent.AgentOutput{
		Stage:        stage,
		VisualIntent: input.VisualIntent,
		Prompt:       input.Prompt,
		Metadata: map[string]string{
			"degraded": "true",
			"stage":    string(stage),
			"fallback": "true",
		},
	}

	switch stage {
	case domainagent.StageRetriever:
		// No references - planner will work without examples
		output.RetrievedReferences = nil
		output.Metadata["reason"] = "retriever_failed_no_references"
	case domainagent.StageStylist:
		// Pass through the planner output unchanged
		output.Content = input.Content
		output.GeneratedArtifacts = input.GeneratedArtifacts
		output.Metadata["reason"] = "stylist_failed_passthrough"
	case domainagent.StageCritic:
		// Accept the visualizer output as-is
		output.Content = "No critique performed due to critic stage failure."
		output.Metadata["reason"] = "critic_failed_no_critique"
	}

	return output
}

type eventPublisher struct {
	sessionID string
	requestID string
	sequence  int64
	out       chan<- domainagent.Event
}

func (p *eventPublisher) emit(eventType domainagent.EventType, stage domainagent.StageName, status domainagent.RunStatus, timing domainagent.Timing, errDetail *domainagent.ErrorDetail, metadata map[string]string) {
	p.sequence++
	p.out <- domainagent.Event{
		Sequence:   p.sequence,
		SessionID:  p.sessionID,
		RequestID:  p.requestID,
		Type:       eventType,
		Stage:      stage,
		Status:     status,
		OccurredAt: time.Now().UTC(),
		Timing:     timing,
		Error:      cloneErrorDetail(errDetail),
		Metadata:   cloneStringMap(metadata),
	}
}
