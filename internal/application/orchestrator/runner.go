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
)

type SnapshotStore interface {
	Save(session domainagent.SessionState, state domainagent.AgentState) error
	Restore(sessionID string, stage domainagent.StageName) (agentstate.Snapshot, error)
}

type Runner struct {
	agents        map[domainagent.StageName]domainagent.BaseAgent
	pipeline      []domainagent.StageName
	eventBuffer   int
	snapshotStore SnapshotStore
}

func NewCanonicalRunner(retriever, planner, stylist, visualizer, critic domainagent.BaseAgent, opts ...RunnerOption) *Runner {
	agents := map[domainagent.StageName]domainagent.BaseAgent{
		domainagent.StageRetriever:  retriever,
		domainagent.StagePlanner:    planner,
		domainagent.StageVisualizer: visualizer,
		domainagent.StageCritic:     critic,
	}
	if stylist != nil {
		agents[domainagent.StageStylist] = stylist
	}
	return NewRunner(agents, opts...)
}

func NewRunner(agents map[domainagent.StageName]domainagent.BaseAgent, opts ...RunnerOption) *Runner {
	runner := &Runner{
		agents:      cloneRegistry(agents),
		pipeline:    orderedPipeline(agents),
		eventBuffer: 32,
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

func (r *Runner) Start(ctx context.Context, input domainagent.AgentInput) (*RunHandle, error) {
	tracker := newSessionTracker(input, r.pipeline)
	return r.startWithTracker(ctx, tracker, r.pipeline)
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
		publisher.emit(domainagent.EventStageStarted, stage, domainagent.StatusRunning, domainagent.Timing{StartedAt: startedAt}, nil, stageInput.Metadata)

		if err := stageAgent.Initialize(ctx); err != nil {
			return r.finishStageError(ctx, tracker, publisher, stage, stageInput, startedAt, err, stageAgent)
		}

		output, err := stageAgent.Execute(ctx, stageInput)
		if err != nil {
			if cleanupErr := stageAgent.Cleanup(ctx); cleanupErr != nil {
				err = errors.Join(err, cleanupErr)
			}
			return r.finishStageError(ctx, tracker, publisher, stage, stageInput, startedAt, err, stageAgent)
		}

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
	}

	completedAt := time.Now().UTC()
	tracker.completeRun(completedAt)
	publisher.emit(domainagent.EventRunCompleted, tracker.state.CurrentStage, domainagent.StatusCompleted, domainagent.Timing{CompletedAt: completedAt}, nil, tracker.state.Metadata)

	return RunResult{Session: tracker.snapshot()}, nil
}

func (r *Runner) finishStageError(ctx context.Context, tracker *sessionTracker, publisher eventPublisher, stage domainagent.StageName, input domainagent.AgentInput, startedAt time.Time, err error, stageAgent domainagent.BaseAgent) (RunResult, error) {
	completedAt := time.Now().UTC()
	timing := domainagent.Timing{
		StartedAt:   startedAt,
		CompletedAt: completedAt,
		Duration:    completedAt.Sub(startedAt),
	}
	detail := &domainagent.ErrorDetail{
		Message: err.Error(),
		Stage:   stage,
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

		if snapshot.Stage.Status != domainagent.StatusCompleted {
			continue
		}

		tracker := newRestoredSessionTracker(snapshot, input, r.pipeline)
		if err := r.restoreCompletedStates(snapshot); err != nil {
			return nil, nil, err
		}
		return tracker, remainingPipeline(snapshot.Stage.Stage, r.pipeline), nil
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
	if stage != domainagent.StageCritic || strings.TrimSpace(initial.Content) == "" {
		return input
	}

	metadata := cloneStringMap(input.Metadata)
	if metadata == nil {
		metadata = map[string]string{}
	}
	if _, exists := metadata["orchestrator.initial_content"]; !exists {
		metadata["orchestrator.initial_content"] = initial.Content
	}
	input.Metadata = metadata
	return input
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
		if !ok || agent == nil {
			continue
		}
		pipeline = append(pipeline, stage)
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
