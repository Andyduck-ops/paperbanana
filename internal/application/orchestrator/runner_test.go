package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	criticagent "github.com/paperbanana/paperbanana/internal/application/agents/critic"
	planneragent "github.com/paperbanana/paperbanana/internal/application/agents/planner"
	retrieveragent "github.com/paperbanana/paperbanana/internal/application/agents/retriever"
	visualizeragent "github.com/paperbanana/paperbanana/internal/application/agents/visualizer"
	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
	agentstate "github.com/paperbanana/paperbanana/internal/infrastructure/agentstate"
	llminfra "github.com/paperbanana/paperbanana/internal/infrastructure/llm"
	sqlitePersistence "github.com/paperbanana/paperbanana/internal/infrastructure/persistence/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPipelineRunnerSerialStageOrder(t *testing.T) {
	var (
		mu        sync.Mutex
		callOrder []domainagent.StageName
	)

	runner := NewRunner(newStubRegistry(func(_ context.Context, input domainagent.AgentInput) (domainagent.AgentOutput, error) {
		mu.Lock()
		callOrder = append(callOrder, input.Stage)
		mu.Unlock()

		return domainagent.AgentOutput{
			Stage:        input.Stage,
			Content:      string(input.Stage) + "-output",
			VisualIntent: input.VisualIntent,
			Prompt:       input.Prompt,
			Metadata:     map[string]string{"stage": string(input.Stage)},
		}, nil
	}), WithEventBuffer(16))

	handle, err := runner.Start(context.Background(), testAgentInput())
	require.NoError(t, err)

	eventsCh := collectEvents(handle.Events())
	result, err := handle.Wait()
	events := <-eventsCh

	require.NoError(t, err)
	assert.Equal(t, activePipeline(), callOrder)
	require.Len(t, callOrder, 4)
	assert.Equal(t, domainagent.StageRetriever, callOrder[0])
	assert.Equal(t, domainagent.StagePlanner, callOrder[1])
	assert.Equal(t, domainagent.StageVisualizer, callOrder[2])
	assert.Equal(t, domainagent.StageCritic, callOrder[3])
	assert.Equal(t, domainagent.StatusCompleted, result.Session.Status)
	assert.Equal(t, activePipeline(), result.Session.Pipeline)
	assert.Equal(t, domainagent.StageCritic, result.Session.CurrentStage)
	assertStageEventOrder(t, events, domainagent.EventStageStarted, activePipeline())
	assertStageEventOrder(t, events, domainagent.EventStageCompleted, activePipeline())
}

func TestPipelineRunnerConcurrency(t *testing.T) {
	var (
		inFlight      int32
		maxConcurrent int32
	)

	runner := NewRunner(newStubRegistry(func(_ context.Context, input domainagent.AgentInput) (domainagent.AgentOutput, error) {
		current := atomic.AddInt32(&inFlight, 1)
		defer atomic.AddInt32(&inFlight, -1)

		for {
			observed := atomic.LoadInt32(&maxConcurrent)
			if current <= observed || atomic.CompareAndSwapInt32(&maxConcurrent, observed, current) {
				break
			}
		}

		time.Sleep(10 * time.Millisecond)

		return domainagent.AgentOutput{
			Stage:        input.Stage,
			Content:      "ok",
			VisualIntent: input.VisualIntent,
			Prompt:       input.Prompt,
		}, nil
	}), WithEventBuffer(16))

	handle, err := runner.Start(context.Background(), testAgentInput())
	require.NoError(t, err)

	eventsCh := collectEvents(handle.Events())
	_, err = handle.Wait()
	events := <-eventsCh

	require.NoError(t, err)
	assert.Equal(t, int32(1), maxConcurrent)
	assertStageEventOrder(t, events, domainagent.EventStageStarted, activePipeline())
}

func TestRunnerPublishesCriticEvents(t *testing.T) {
	var plannerInput domainagent.AgentInput
	var visualizerInput domainagent.AgentInput
	var criticInput domainagent.AgentInput

	runner := NewRunner(newStubRegistry(func(_ context.Context, input domainagent.AgentInput) (domainagent.AgentOutput, error) {
		switch input.Stage {
		case domainagent.StageRetriever:
			return domainagent.AgentOutput{
				Stage:        input.Stage,
				VisualIntent: input.VisualIntent,
				Prompt:       input.Prompt,
				RetrievedReferences: []domainagent.RetrievedReference{
					{ID: "ref_1", Summary: "Retriever output"},
				},
				GeneratedArtifacts: []domainagent.Artifact{
					{
						ID:       "retriever-bundle",
						Kind:     domainagent.ArtifactKindReferenceBundle,
						MIMEType: "application/json",
						URI:      "memory://retriever/examples",
						Content:  `[{"id":"ref_1","visual_intent":"Agent framework overview","content":"Method section","path_to_gt_image":"ref_1.png"}]`,
					},
				},
				Metadata: map[string]string{"stage": "retriever"},
			}, nil
		case domainagent.StagePlanner:
			plannerInput = input
			return domainagent.AgentOutput{
				Stage:        input.Stage,
				Content:      "Detailed planner summary for the retrieved diagram.",
				VisualIntent: input.VisualIntent,
				Prompt:       input.Prompt,
				GeneratedArtifacts: append([]domainagent.Artifact(nil), append(input.GeneratedArtifacts,
					domainagent.Artifact{
						ID:       "planner-plan",
						Kind:     domainagent.ArtifactKindPlan,
						MIMEType: "text/plain",
						URI:      "memory://planner/current",
						Content:  "Detailed planner summary for the retrieved diagram.",
					},
				)...),
				Metadata: map[string]string{"summary": "Detailed planner summary for the retrieved diagram."},
			}, nil
		case domainagent.StageVisualizer:
			visualizerInput = input
			return domainagent.AgentOutput{
				Stage:        input.Stage,
				Content:      input.Content,
				VisualIntent: input.VisualIntent,
				Prompt:       input.Prompt,
				GeneratedArtifacts: append([]domainagent.Artifact(nil), append(input.GeneratedArtifacts,
					domainagent.Artifact{
						ID:       "visualizer-rendered",
						Kind:     domainagent.ArtifactKindRenderedFigure,
						MIMEType: "image/png",
						URI:      "memory://visualizer/current",
						Bytes:    []byte("rendered-image"),
					},
				)...),
				Metadata: map[string]string{"summary": "Visualizer produced a rendered figure artifact."},
			}, nil
		case domainagent.StageCritic:
			criticInput = input
			return domainagent.AgentOutput{
				Stage:        input.Stage,
				Content:      input.Content,
				VisualIntent: input.VisualIntent,
				Prompt:       input.Prompt,
				GeneratedArtifacts: append([]domainagent.Artifact(nil), append(input.GeneratedArtifacts,
					domainagent.Artifact{
						ID:       "critic-round-0",
						Kind:     domainagent.ArtifactKindCritique,
						MIMEType: "application/json",
						URI:      "memory://critic/current",
						Content:  `{"critic_suggestions":"No changes needed.","revised_description":"No changes needed."}`,
					},
				)...),
				CritiqueRounds: []domainagent.CritiqueRound{
					{
						Round:            0,
						Summary:          "No changes needed.",
						Accepted:         true,
						RequestedChanges: []string{"No changes needed."},
						EvaluatedAt:      time.Date(2026, time.March, 17, 6, 0, 0, 0, time.UTC),
					},
				},
				Metadata: map[string]string{
					"summary":         "Critic accepted the rendered figure without rerendering.",
					"reused_artifact": "true",
				},
			}, nil
		}

		return domainagent.AgentOutput{
			Stage:        input.Stage,
			Content:      string(input.Stage) + "-output",
			VisualIntent: input.VisualIntent,
			Prompt:       input.Prompt,
		}, nil
	}), WithEventBuffer(16))

	handle, err := runner.Start(context.Background(), testAgentInput())
	require.NoError(t, err)

	eventsCh := collectEvents(handle.Events())
	result, err := handle.Wait()
	events := <-eventsCh

	require.NoError(t, err)
	assert.Equal(t, domainagent.StatusCompleted, result.Session.Status)
	require.Len(t, plannerInput.RetrievedReferences, 1)
	assert.Equal(t, "ref_1", plannerInput.RetrievedReferences[0].ID)
	require.Len(t, plannerInput.GeneratedArtifacts, 1)
	assert.Equal(t, domainagent.ArtifactKindReferenceBundle, plannerInput.GeneratedArtifacts[0].Kind)
	assert.Equal(t, "Detailed planner summary for the retrieved diagram.", visualizerInput.Content)
	require.Len(t, visualizerInput.GeneratedArtifacts, 2)
	assert.Equal(t, domainagent.ArtifactKindReferenceBundle, visualizerInput.GeneratedArtifacts[0].Kind)
	assert.Equal(t, domainagent.ArtifactKindPlan, visualizerInput.GeneratedArtifacts[1].Kind)
	assert.Equal(t, "Detailed planner summary for the retrieved diagram.", criticInput.Content)
	assert.Equal(t, "Generate a figure", criticInput.Metadata["orchestrator.initial_content"])
	require.Len(t, criticInput.GeneratedArtifacts, 3)
	assert.Equal(t, domainagent.ArtifactKindReferenceBundle, criticInput.GeneratedArtifacts[0].Kind)
	assert.Equal(t, domainagent.ArtifactKindPlan, criticInput.GeneratedArtifacts[1].Kind)
	assert.Equal(t, domainagent.ArtifactKindRenderedFigure, criticInput.GeneratedArtifacts[2].Kind)
	require.Len(t, result.Session.FinalOutput.GeneratedArtifacts, 4)
	assert.Equal(t, domainagent.ArtifactKindCritique, result.Session.FinalOutput.GeneratedArtifacts[3].Kind)
	assertEventBefore(t, events, domainagent.EventStageStarted, domainagent.StageRetriever, domainagent.EventStageStarted, domainagent.StagePlanner)
	assertEventBefore(t, events, domainagent.EventStageCompleted, domainagent.StageRetriever, domainagent.EventStageStarted, domainagent.StagePlanner)
	assertEventBefore(t, events, domainagent.EventStageCompleted, domainagent.StagePlanner, domainagent.EventStageStarted, domainagent.StageVisualizer)
	assertEventBefore(t, events, domainagent.EventStageCompleted, domainagent.StageVisualizer, domainagent.EventStageStarted, domainagent.StageCritic)
	assertEventMetadataValue(t, events, domainagent.EventStageCompleted, domainagent.StageCritic, "summary", "Critic accepted the rendered figure without rerendering.")
	assertEventMetadataValue(t, events, domainagent.EventStageCompleted, domainagent.StageCritic, "artifact_count", "4")
	assertEventMetadataValue(t, events, domainagent.EventStageCompleted, domainagent.StageCritic, "artifact_kinds", "reference_bundle,plan,rendered_figure,critique")
}

func TestPipelineRunnerStopsOnStageError(t *testing.T) {
	plannerErr := errors.New("planner failed")

	var (
		mu        sync.Mutex
		callOrder []domainagent.StageName
	)

	runner := NewRunner(newStubRegistry(func(_ context.Context, input domainagent.AgentInput) (domainagent.AgentOutput, error) {
		mu.Lock()
		callOrder = append(callOrder, input.Stage)
		mu.Unlock()

		if input.Stage == domainagent.StagePlanner {
			return domainagent.AgentOutput{}, plannerErr
		}

		return domainagent.AgentOutput{
			Stage:        input.Stage,
			Content:      string(input.Stage) + "-output",
			VisualIntent: input.VisualIntent,
			Prompt:       input.Prompt,
		}, nil
	}), WithEventBuffer(16))

	handle, err := runner.Start(context.Background(), testAgentInput())
	require.NoError(t, err)

	eventsCh := collectEvents(handle.Events())
	result, err := handle.Wait()
	events := <-eventsCh

	require.ErrorIs(t, err, plannerErr)
	assert.Equal(t, []domainagent.StageName{domainagent.StageRetriever, domainagent.StagePlanner}, callOrder)
	assert.Equal(t, domainagent.StagePlanner, result.FailedStage)
	assert.Equal(t, domainagent.StatusFailed, result.Session.Status)
	assertHasEvent(t, events, domainagent.EventStageFailed, domainagent.StagePlanner)
	assertHasEvent(t, events, domainagent.EventRunFailed, domainagent.StagePlanner)
}

func TestRunnerResumeFromSnapshots(t *testing.T) {
	t.Parallel()

	var (
		mu          sync.Mutex
		callOrder   []domainagent.StageName
		stageInputs = map[domainagent.StageName]domainagent.AgentInput{}
	)

	rootDir := t.TempDir()
	store := agentstate.NewStore(rootDir)
	input := testAgentInput()
	input.SessionID = "resume-session"
	input.RequestID = "resume-request"

	retrieverState := domainagent.AgentState{
		Stage:  domainagent.StageRetriever,
		Status: domainagent.StatusCompleted,
		Timing: domainagent.Timing{
			StartedAt:   time.Date(2026, time.March, 17, 8, 0, 0, 0, time.UTC),
			CompletedAt: time.Date(2026, time.March, 17, 8, 0, 1, 0, time.UTC),
			Duration:    time.Second,
		},
		Input: input,
		Output: domainagent.AgentOutput{
			Stage:        domainagent.StageRetriever,
			VisualIntent: input.VisualIntent,
			Prompt:       input.Prompt,
		},
	}

	plannerInput := mergeAgentInput(input, retrieverState.Output)
	plannerInput.Stage = domainagent.StagePlanner
	plannerOutput := domainagent.AgentOutput{
		Stage:        domainagent.StagePlanner,
		Content:      "saved planner output",
		VisualIntent: input.VisualIntent,
		Prompt:       input.Prompt,
		GeneratedArtifacts: []domainagent.Artifact{
			{
				ID:       "planner-diagram-plan",
				Kind:     domainagent.ArtifactKindPlan,
				MIMEType: "text/plain",
				URI:      "memory://planner/diagram/plan",
				Content:  "saved planner output",
			},
		},
	}
	plannerState := domainagent.AgentState{
		Stage:  domainagent.StagePlanner,
		Status: domainagent.StatusCompleted,
		Timing: domainagent.Timing{
			StartedAt:   time.Date(2026, time.March, 17, 8, 0, 2, 0, time.UTC),
			CompletedAt: time.Date(2026, time.March, 17, 8, 0, 3, 0, time.UTC),
			Duration:    time.Second,
		},
		Input:  plannerInput,
		Output: plannerOutput,
	}

	session := domainagent.SessionState{
		SchemaVersion: sessionSchemaVersion,
		SessionID:     input.SessionID,
		RequestID:     input.RequestID,
		Status:        domainagent.StatusFailed,
		CurrentStage:  domainagent.StagePlanner,
		Pipeline:      activePipeline(),
		InitialInput:  input,
		StageStates:   []domainagent.AgentState{retrieverState, plannerState},
		FinalOutput:   plannerOutput,
		Error: &domainagent.ErrorDetail{
			Message: "visualizer interrupted before restart",
			Stage:   domainagent.StageVisualizer,
		},
		Metadata:    map[string]string{"session_origin": "snapshot"},
		StartedAt:   retrieverState.Timing.StartedAt,
		UpdatedAt:   plannerState.Timing.CompletedAt,
		CompletedAt: plannerState.Timing.CompletedAt,
	}
	require.NoError(t, store.Save(session, plannerState))

	runner := NewRunner(newStubRegistry(func(_ context.Context, stageInput domainagent.AgentInput) (domainagent.AgentOutput, error) {
		mu.Lock()
		callOrder = append(callOrder, stageInput.Stage)
		stageInputs[stageInput.Stage] = stageInput
		mu.Unlock()

		return domainagent.AgentOutput{
			Stage:              stageInput.Stage,
			Content:            fmt.Sprintf("%s-output", stageInput.Stage),
			VisualIntent:       stageInput.VisualIntent,
			Prompt:             stageInput.Prompt,
			GeneratedArtifacts: cloneArtifacts(stageInput.GeneratedArtifacts),
			Metadata:           map[string]string{"stage": string(stageInput.Stage)},
		}, nil
	}), WithEventBuffer(16), WithSnapshotStore(store))

	handle, err := runner.Resume(context.Background(), domainagent.AgentInput{
		SessionID: input.SessionID,
		RequestID: "resume-request-2",
		Content:   "ignored on resume",
	})
	require.NoError(t, err)

	eventsCh := collectEvents(handle.Events())
	result, err := handle.Wait()
	events := <-eventsCh

	require.NoError(t, err)
	assert.Equal(t, []domainagent.StageName{domainagent.StageVisualizer, domainagent.StageCritic}, callOrder)
	assert.Equal(t, "saved planner output", stageInputs[domainagent.StageVisualizer].Content)
	assert.Equal(t, domainagent.StagePlanner, stageInputs[domainagent.StageVisualizer].Restore.RestoredFrom)
	assert.Equal(t, input.Content, stageInputs[domainagent.StageCritic].Metadata["orchestrator.initial_content"])
	require.Len(t, result.Session.StageStates, 4)
	assert.Equal(t, domainagent.StageRetriever, result.Session.StageStates[0].Stage)
	assert.Equal(t, domainagent.StagePlanner, result.Session.StageStates[1].Stage)
	assert.Equal(t, domainagent.StageVisualizer, result.Session.StageStates[2].Stage)
	assert.Equal(t, domainagent.StageCritic, result.Session.StageStates[3].Stage)
	assert.Equal(t, domainagent.StatusCompleted, result.Session.Status)
	assertStageEventOrder(t, events, domainagent.EventStageStarted, []domainagent.StageName{domainagent.StageVisualizer, domainagent.StageCritic})
	assertStageEventOrder(t, events, domainagent.EventStageCompleted, []domainagent.StageName{domainagent.StageVisualizer, domainagent.StageCritic})
}

func TestRunnerResumeFromSnapshotsRestoresAgentState(t *testing.T) {
	t.Parallel()

	store, input, _, plannerState := writeResumeSnapshot(t)
	registry := map[domainagent.StageName]domainagent.BaseAgent{
		domainagent.StageRetriever:  &stubAgent{stage: domainagent.StageRetriever},
		domainagent.StagePlanner:    &stubAgent{stage: domainagent.StagePlanner},
		domainagent.StageVisualizer: &stubAgent{stage: domainagent.StageVisualizer},
		domainagent.StageCritic:     &stubAgent{stage: domainagent.StageCritic},
	}
	runner := NewRunner(registry, WithSnapshotStore(store))

	handle, err := runner.Resume(context.Background(), domainagent.AgentInput{
		SessionID: input.SessionID,
		RequestID: "resume-request-2",
	})
	require.NoError(t, err)

	_, err = handle.Wait()
	require.NoError(t, err)

	retriever := registry[domainagent.StageRetriever].(*stubAgent)
	planner := registry[domainagent.StagePlanner].(*stubAgent)
	require.Len(t, retriever.restoreCalls, 1)
	require.Len(t, planner.restoreCalls, 1)
	assert.Equal(t, domainagent.StageRetriever, retriever.restoreCalls[0].Stage)
	assert.Equal(t, domainagent.StagePlanner, planner.restoreCalls[0].Stage)
	assert.Equal(t, plannerState.Output.Content, planner.restoreCalls[0].Output.Content)
}

func TestRunnerResumeFromSnapshotsPreservesRestoreMetadata(t *testing.T) {
	t.Parallel()

	store, input, _, _ := writeResumeSnapshot(t)

	var stageInputs = map[domainagent.StageName]domainagent.AgentInput{}
	runner := NewRunner(newStubRegistry(func(_ context.Context, stageInput domainagent.AgentInput) (domainagent.AgentOutput, error) {
		stageInputs[stageInput.Stage] = stageInput
		return domainagent.AgentOutput{
			Stage:              stageInput.Stage,
			Content:            fmt.Sprintf("%s-output", stageInput.Stage),
			VisualIntent:       stageInput.VisualIntent,
			Prompt:             stageInput.Prompt,
			GeneratedArtifacts: cloneArtifacts(stageInput.GeneratedArtifacts),
		}, nil
	}), WithSnapshotStore(store))

	handle, err := runner.Resume(context.Background(), domainagent.AgentInput{
		SessionID: input.SessionID,
		RequestID: "resume-request-2",
	})
	require.NoError(t, err)

	_, err = handle.Wait()
	require.NoError(t, err)

	restore := stageInputs[domainagent.StageVisualizer].Restore
	assert.Equal(t, agentstate.SnapshotSchemaVersion, restore.SnapshotVersion)
	assert.Equal(t, domainagent.StagePlanner, restore.RestoredFrom)
	assert.False(t, restore.RestoredAt.IsZero())
	assert.Equal(t, restore, stageInputs[domainagent.StageCritic].Restore)
}

func TestRunnerResumeSkipsRestoreForMissingAgentRegistration(t *testing.T) {
	t.Parallel()

	store, input, retrieverState, _ := writeResumeSnapshot(t)

	t.Run("skips completed stages without a registered agent", func(t *testing.T) {
		t.Parallel()

		planner := &stubAgent{stage: domainagent.StagePlanner}
		runner := NewRunner(map[domainagent.StageName]domainagent.BaseAgent{
			domainagent.StagePlanner:    planner,
			domainagent.StageVisualizer: &stubAgent{stage: domainagent.StageVisualizer},
			domainagent.StageCritic:     &stubAgent{stage: domainagent.StageCritic},
		}, WithSnapshotStore(store))

		handle, err := runner.Resume(context.Background(), domainagent.AgentInput{
			SessionID: input.SessionID,
			RequestID: "resume-request-2",
		})
		require.NoError(t, err)

		_, err = handle.Wait()
		require.NoError(t, err)
		require.Len(t, planner.restoreCalls, 1)
		assert.Equal(t, domainagent.StagePlanner, planner.restoreCalls[0].Stage)
	})

	t.Run("fails when the restore point agent is not registered", func(t *testing.T) {
		t.Parallel()

		runner := NewRunner(map[domainagent.StageName]domainagent.BaseAgent{
			domainagent.StageRetriever: &stubAgent{stage: domainagent.StageRetriever},
			domainagent.StageVisualizer: &stubAgent{
				stage: domainagent.StageVisualizer,
			},
			domainagent.StageCritic: &stubAgent{stage: domainagent.StageCritic},
		}, WithSnapshotStore(store))

		_, err := runner.Resume(context.Background(), domainagent.AgentInput{
			SessionID: input.SessionID,
			RequestID: "resume-request-2",
		})
		require.Error(t, err)
		assert.ErrorContains(t, err, "restore point")
		assert.ErrorContains(t, err, string(domainagent.StagePlanner))
	})

	assert.Equal(t, domainagent.StageRetriever, retrieverState.Stage)
}

func TestRunnerUsesCachedLLMBoundary(t *testing.T) {
	t.Parallel()

	upstream := &runnerLLMProvider{
		provider: "runner-provider",
		responses: map[string]*domainllm.GenerateResponse{
			retrieveragent.PromptVersion: {
				Content: `{"top10_diagrams":[]}`,
			},
			planneragent.PromptVersion: {
				Content: "cached planner output",
			},
			visualizeragent.PromptVersion: {
				Parts: []domainllm.Part{
					domainllm.InlineImagePart("image/png", []byte("cached-image")),
				},
			},
			criticagent.PromptVersion: {
				Content: `{"critic_suggestions":"No changes needed.","revised_description":"No changes needed."}`,
			},
		},
	}
	cache := &runnerLLMCache{}
	client := llminfra.NewCachedClient("runner-provider", "runner-model", upstream, cache)

	runner := NewCanonicalRunner(
		retrieveragent.NewAgent(client, retrieveragent.Config{Mode: retrieveragent.RetrievalModeNone, Model: "runner-model"}),
		planneragent.NewAgent(client, planneragent.Config{Model: "runner-model"}),
		nil, // stylist is optional
		visualizeragent.NewAgent(client, visualizeragent.Config{Model: "runner-model"}),
		criticagent.NewAgent(client, criticagent.Config{Model: "runner-model"}),
	)

	firstHandle, err := runner.Start(context.Background(), testAgentInput())
	require.NoError(t, err)
	_, err = firstHandle.Wait()
	require.NoError(t, err)

	secondInput := testAgentInput()
	secondInput.SessionID = "session-02"
	secondInput.RequestID = "request-02"

	secondHandle, err := runner.Start(context.Background(), secondInput)
	require.NoError(t, err)
	_, err = secondHandle.Wait()
	require.NoError(t, err)

	assert.Equal(t, 1, upstream.calls[planneragent.PromptVersion])
	assert.Equal(t, 1, upstream.calls[visualizeragent.PromptVersion])
	assert.Equal(t, 1, upstream.calls[criticagent.PromptVersion])
	assert.Equal(t, 3, cache.sets)
}

type stubAgent struct {
	stage        domainagent.StageName
	execute      func(context.Context, domainagent.AgentInput) (domainagent.AgentOutput, error)
	state        domainagent.AgentState
	restoreCalls []domainagent.AgentState
	restoreErr   error
}

func (a *stubAgent) Initialize(context.Context) error {
	return nil
}

func (a *stubAgent) Execute(ctx context.Context, input domainagent.AgentInput) (domainagent.AgentOutput, error) {
	a.state.Input = input
	a.state.Stage = a.stage
	a.state.Status = domainagent.StatusRunning

	if a.execute == nil {
		return domainagent.AgentOutput{
			Stage:        a.stage,
			Content:      string(a.stage),
			VisualIntent: input.VisualIntent,
			Prompt:       input.Prompt,
		}, nil
	}

	output, err := a.execute(ctx, input)
	if err == nil {
		a.state.Output = output
	}
	return output, err
}

func (a *stubAgent) Cleanup(context.Context) error {
	return nil
}

func (a *stubAgent) GetState() domainagent.AgentState {
	return a.state
}

func (a *stubAgent) RestoreState(state domainagent.AgentState) error {
	a.restoreCalls = append(a.restoreCalls, state)
	if a.restoreErr != nil {
		return a.restoreErr
	}
	a.state = state
	return nil
}

func writeResumeSnapshot(t *testing.T) (SnapshotStore, domainagent.AgentInput, domainagent.AgentState, domainagent.AgentState) {
	t.Helper()

	rootDir := t.TempDir()
	store := agentstate.NewStore(rootDir)
	input := testAgentInput()
	input.SessionID = "resume-session"
	input.RequestID = "resume-request"

	retrieverState := domainagent.AgentState{
		Stage:  domainagent.StageRetriever,
		Status: domainagent.StatusCompleted,
		Timing: domainagent.Timing{
			StartedAt:   time.Date(2026, time.March, 17, 8, 0, 0, 0, time.UTC),
			CompletedAt: time.Date(2026, time.March, 17, 8, 0, 1, 0, time.UTC),
			Duration:    time.Second,
		},
		Input: input,
		Output: domainagent.AgentOutput{
			Stage:        domainagent.StageRetriever,
			VisualIntent: input.VisualIntent,
			Prompt:       input.Prompt,
		},
	}

	plannerInput := mergeAgentInput(input, retrieverState.Output)
	plannerInput.Stage = domainagent.StagePlanner
	plannerOutput := domainagent.AgentOutput{
		Stage:        domainagent.StagePlanner,
		Content:      "saved planner output",
		VisualIntent: input.VisualIntent,
		Prompt:       input.Prompt,
		GeneratedArtifacts: []domainagent.Artifact{
			{
				ID:       "planner-diagram-plan",
				Kind:     domainagent.ArtifactKindPlan,
				MIMEType: "text/plain",
				URI:      "memory://planner/diagram/plan",
				Content:  "saved planner output",
			},
		},
	}
	plannerState := domainagent.AgentState{
		Stage:  domainagent.StagePlanner,
		Status: domainagent.StatusCompleted,
		Timing: domainagent.Timing{
			StartedAt:   time.Date(2026, time.March, 17, 8, 0, 2, 0, time.UTC),
			CompletedAt: time.Date(2026, time.March, 17, 8, 0, 3, 0, time.UTC),
			Duration:    time.Second,
		},
		Input:  plannerInput,
		Output: plannerOutput,
	}

	session := domainagent.SessionState{
		SchemaVersion: sessionSchemaVersion,
		SessionID:     input.SessionID,
		RequestID:     input.RequestID,
		Status:        domainagent.StatusFailed,
		CurrentStage:  domainagent.StagePlanner,
		Pipeline:      activePipeline(),
		InitialInput:  input,
		StageStates:   []domainagent.AgentState{retrieverState, plannerState},
		FinalOutput:   plannerOutput,
		Error: &domainagent.ErrorDetail{
			Message: "visualizer interrupted before restart",
			Stage:   domainagent.StageVisualizer,
		},
		Metadata:    map[string]string{"session_origin": "snapshot"},
		StartedAt:   retrieverState.Timing.StartedAt,
		UpdatedAt:   plannerState.Timing.CompletedAt,
		CompletedAt: plannerState.Timing.CompletedAt,
	}
	require.NoError(t, store.Save(session, plannerState))

	return store, input, retrieverState, plannerState
}

func newStubRegistry(execute func(context.Context, domainagent.AgentInput) (domainagent.AgentOutput, error)) map[domainagent.StageName]domainagent.BaseAgent {
	return map[domainagent.StageName]domainagent.BaseAgent{
		domainagent.StageRetriever:  &stubAgent{stage: domainagent.StageRetriever, execute: execute},
		domainagent.StagePlanner:    &stubAgent{stage: domainagent.StagePlanner, execute: execute},
		domainagent.StageVisualizer: &stubAgent{stage: domainagent.StageVisualizer, execute: execute},
		domainagent.StageCritic:     &stubAgent{stage: domainagent.StageCritic, execute: execute},
	}
}

func activePipeline() []domainagent.StageName {
	return []domainagent.StageName{
		domainagent.StageRetriever,
		domainagent.StagePlanner,
		domainagent.StageVisualizer,
		domainagent.StageCritic,
	}
}

func testAgentInput() domainagent.AgentInput {
	return domainagent.AgentInput{
		SessionID: "session-01",
		RequestID: "request-01",
		Content:   "Generate a figure",
		VisualIntent: domainagent.VisualIntent{
			Mode:  domainagent.VisualModeDiagram,
			Goal:  "Summarize the pipeline",
			Style: "academic",
		},
		Prompt: domainagent.PromptMetadata{
			SystemInstruction: "You are a test agent.",
			Version:           "test-v1",
			Template:          "test/template",
		},
	}
}

func collectEvents(events <-chan domainagent.Event) <-chan []domainagent.Event {
	collected := make(chan []domainagent.Event, 1)

	go func() {
		var items []domainagent.Event
		for event := range events {
			items = append(items, event)
		}
		collected <- items
	}()

	return collected
}

func assertStageEventOrder(t *testing.T, events []domainagent.Event, eventType domainagent.EventType, want []domainagent.StageName) {
	t.Helper()

	var got []domainagent.StageName
	for _, event := range events {
		if event.Type == eventType {
			got = append(got, event.Stage)
		}
	}

	assert.Equal(t, want, got)
}

func assertHasEvent(t *testing.T, events []domainagent.Event, eventType domainagent.EventType, stage domainagent.StageName) {
	t.Helper()

	for _, event := range events {
		if event.Type == eventType && event.Stage == stage {
			return
		}
	}

	t.Fatalf("expected %s event for stage %s", eventType, stage)
}

func assertEventBefore(t *testing.T, events []domainagent.Event, firstType domainagent.EventType, firstStage domainagent.StageName, secondType domainagent.EventType, secondStage domainagent.StageName) {
	t.Helper()

	firstIndex := -1
	secondIndex := -1
	for index, event := range events {
		if firstIndex == -1 && event.Type == firstType && event.Stage == firstStage {
			firstIndex = index
		}
		if secondIndex == -1 && event.Type == secondType && event.Stage == secondStage {
			secondIndex = index
		}
	}

	require.NotEqual(t, -1, firstIndex, "missing %s for %s", firstType, firstStage)
	require.NotEqual(t, -1, secondIndex, "missing %s for %s", secondType, secondStage)
	assert.Less(t, firstIndex, secondIndex)
}

func assertEventMetadataValue(t *testing.T, events []domainagent.Event, eventType domainagent.EventType, stage domainagent.StageName, key, want string) {
	t.Helper()

	for _, event := range events {
		if event.Type == eventType && event.Stage == stage {
			require.NotNil(t, event.Metadata)
			assert.Equal(t, want, event.Metadata[key])
			return
		}
	}

	t.Fatalf("expected %s event for stage %s", eventType, stage)
}

type runnerLLMProvider struct {
	provider  string
	mu        sync.Mutex
	calls     map[string]int
	responses map[string]*domainllm.GenerateResponse
}

func (m *runnerLLMProvider) Generate(_ context.Context, req domainllm.GenerateRequest) (*domainllm.GenerateResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.calls == nil {
		m.calls = map[string]int{}
	}
	m.calls[req.PromptVersion]++

	resp, ok := m.responses[req.PromptVersion]
	if !ok {
		return nil, fmt.Errorf("unexpected prompt version %q", req.PromptVersion)
	}
	return cloneGenerateResponse(resp), nil
}

func (m *runnerLLMProvider) GenerateStream(_ context.Context, _ domainllm.GenerateRequest) (<-chan domainllm.StreamChunk, <-chan error) {
	chunks := make(chan domainllm.StreamChunk)
	errs := make(chan error)
	close(chunks)
	close(errs)
	return chunks, errs
}

func (m *runnerLLMProvider) Provider() string {
	return m.provider
}

type runnerLLMCache struct {
	mu     sync.Mutex
	values map[string]*domainllm.GenerateResponse
	sets   int
}

func (c *runnerLLMCache) Get(_ context.Context, provider, model string, req domainllm.GenerateRequest) (*domainllm.GenerateResponse, bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	value, ok := c.values[c.cacheKey(provider, model, req)]
	if !ok {
		return nil, false, nil
	}
	return cloneGenerateResponse(value), true, nil
}

func (c *runnerLLMCache) Set(_ context.Context, provider, model string, req domainllm.GenerateRequest, resp *domainllm.GenerateResponse) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.values == nil {
		c.values = map[string]*domainllm.GenerateResponse{}
	}
	c.values[c.cacheKey(provider, model, req)] = cloneGenerateResponse(resp)
	c.sets++
	return nil
}

func (c *runnerLLMCache) cacheKey(provider, model string, req domainllm.GenerateRequest) string {
	return provider + "|" + model + "|" + req.PromptVersion + "|" + req.SystemInstruction + "|" + domainllm.CollectText(req.Messages[0].Parts)
}

func cloneGenerateResponse(resp *domainllm.GenerateResponse) *domainllm.GenerateResponse {
	if resp == nil {
		return nil
	}

	cloned := *resp
	if len(resp.Parts) > 0 {
		cloned.Parts = append([]domainllm.Part(nil), resp.Parts...)
		for i := range cloned.Parts {
			cloned.Parts[i].Data = append([]byte(nil), resp.Parts[i].Data...)
		}
	}
	return &cloned
}

func TestRunnerResumeFromPersistentStore(t *testing.T) {
	t.Parallel()

	var (
		mu          sync.Mutex
		callOrder   []domainagent.StageName
		stageInputs = map[domainagent.StageName]domainagent.AgentInput{}
	)

	// Create in-memory SQLite database
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&sqlitePersistence.SessionModel{}))

	sessionRepo := sqlitePersistence.NewSessionRepository(db)
	persistentStore := sqlitePersistence.NewPersistentSnapshotStore(sessionRepo)

	input := testAgentInput()
	input.SessionID = "persistent-resume-session"
	input.RequestID = "persistent-resume-request"

	// Build the initial session with retriever and planner completed
	retrieverState := domainagent.AgentState{
		Stage:  domainagent.StageRetriever,
		Status: domainagent.StatusCompleted,
		Timing: domainagent.Timing{
			StartedAt:   time.Date(2026, time.March, 17, 10, 0, 0, 0, time.UTC),
			CompletedAt: time.Date(2026, time.March, 17, 10, 0, 1, 0, time.UTC),
			Duration:    time.Second,
		},
		Input: input,
		Output: domainagent.AgentOutput{
			Stage:        domainagent.StageRetriever,
			VisualIntent: input.VisualIntent,
			Prompt:       input.Prompt,
		},
	}

	plannerInput := mergeAgentInput(input, retrieverState.Output)
	plannerInput.Stage = domainagent.StagePlanner
	plannerOutput := domainagent.AgentOutput{
		Stage:        domainagent.StagePlanner,
		Content:      "persistent planner output",
		VisualIntent: input.VisualIntent,
		Prompt:       input.Prompt,
		GeneratedArtifacts: []domainagent.Artifact{
			{
				ID:       "planner-persistent-plan",
				Kind:     domainagent.ArtifactKindPlan,
				MIMEType: "text/plain",
				Content:  "persistent planner output",
			},
		},
	}
	plannerState := domainagent.AgentState{
		Stage:  domainagent.StagePlanner,
		Status: domainagent.StatusCompleted,
		Timing: domainagent.Timing{
			StartedAt:   time.Date(2026, time.March, 17, 10, 0, 2, 0, time.UTC),
			CompletedAt: time.Date(2026, time.March, 17, 10, 0, 3, 0, time.UTC),
			Duration:    time.Second,
		},
		Input:  plannerInput,
		Output: plannerOutput,
	}

	session := domainagent.SessionState{
		SchemaVersion: sessionSchemaVersion,
		SessionID:     input.SessionID,
		RequestID:     input.RequestID,
		Status:        domainagent.StatusFailed,
		CurrentStage:  domainagent.StagePlanner,
		Pipeline:      activePipeline(),
		InitialInput:  input,
		StageStates:   []domainagent.AgentState{retrieverState, plannerState},
		FinalOutput:   plannerOutput,
		Error: &domainagent.ErrorDetail{
			Message: "visualizer interrupted",
			Stage:   domainagent.StageVisualizer,
		},
		Metadata:    map[string]string{"session_origin": "persistent"},
		StartedAt:   retrieverState.Timing.StartedAt,
		UpdatedAt:   plannerState.Timing.CompletedAt,
		CompletedAt: plannerState.Timing.CompletedAt,
	}

	// Save to persistent store
	err = persistentStore.Save(session, plannerState)
	require.NoError(t, err)

	// Create runner with persistent snapshot store
	runner := NewRunner(newStubRegistry(func(_ context.Context, stageInput domainagent.AgentInput) (domainagent.AgentOutput, error) {
		mu.Lock()
		callOrder = append(callOrder, stageInput.Stage)
		stageInputs[stageInput.Stage] = stageInput
		mu.Unlock()

		return domainagent.AgentOutput{
			Stage:              stageInput.Stage,
			Content:            fmt.Sprintf("%s-output", stageInput.Stage),
			VisualIntent:       stageInput.VisualIntent,
			Prompt:             stageInput.Prompt,
			GeneratedArtifacts: cloneArtifacts(stageInput.GeneratedArtifacts),
			Metadata:           map[string]string{"stage": string(stageInput.Stage)},
		}, nil
	}), WithEventBuffer(16), WithSnapshotStore(persistentStore))

	handle, err := runner.Resume(context.Background(), domainagent.AgentInput{
		SessionID: input.SessionID,
		RequestID: "resume-request-3",
	})
	require.NoError(t, err)

	eventsCh := collectEvents(handle.Events())
	result, err := handle.Wait()
	events := <-eventsCh

	require.NoError(t, err)
	assert.Equal(t, []domainagent.StageName{domainagent.StageVisualizer, domainagent.StageCritic}, callOrder)
	assert.Equal(t, "persistent planner output", stageInputs[domainagent.StageVisualizer].Content)
	assert.Equal(t, domainagent.StagePlanner, stageInputs[domainagent.StageVisualizer].Restore.RestoredFrom)
	require.Len(t, result.Session.StageStates, 4)
	assert.Equal(t, domainagent.StatusCompleted, result.Session.Status)
	assertStageEventOrder(t, events, domainagent.EventStageStarted, []domainagent.StageName{domainagent.StageVisualizer, domainagent.StageCritic})
	assertStageEventOrder(t, events, domainagent.EventStageCompleted, []domainagent.StageName{domainagent.StageVisualizer, domainagent.StageCritic})
}
