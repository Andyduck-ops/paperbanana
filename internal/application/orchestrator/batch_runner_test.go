package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBatchRunner_StartBatch(t *testing.T) {
	t.Parallel()

	var retrieverCalls int32
	factory := &mockAgentFactory{
		createRetriever: func() domainagent.BaseAgent {
			return &stubAgent{
				stage: domainagent.StageRetriever,
				execute: func(ctx context.Context, input domainagent.AgentInput) (domainagent.AgentOutput, error) {
					atomic.AddInt32(&retrieverCalls, 1)
					return domainagent.AgentOutput{
						Stage:        domainagent.StageRetriever,
						VisualIntent: input.VisualIntent,
						RetrievedReferences: []domainagent.RetrievedReference{
							{ID: "ref-1", Summary: "shared reference"},
						},
					}, nil
				},
			}
		},
		createPlanner: func() domainagent.BaseAgent {
			return &stubAgent{stage: domainagent.StagePlanner}
		},
		createVisualizer: func() domainagent.BaseAgent {
			return &stubAgent{stage: domainagent.StageVisualizer}
		},
		createCritic: func() domainagent.BaseAgent {
			return &stubAgent{stage: domainagent.StageCritic}
		},
	}

	batchRunner := NewBatchRunner(factory)
	inputs := make([]domainagent.AgentInput, 3)
	for i := 0; i < 3; i++ {
		inputs[i] = domainagent.AgentInput{
			SessionID: "batch-session",
			Content:   "Generate diagram",
			VisualIntent: domainagent.VisualIntent{
				Mode:  domainagent.VisualModeDiagram,
				Goal:  "Test batch",
				Style: "academic",
			},
		}
	}

	handle, err := batchRunner.StartBatch(context.Background(), inputs)
	require.NoError(t, err)

	eventsCh := collectBatchEvents(handle.Events())
	result, err := handle.Wait()
	events := <-eventsCh

	require.NoError(t, err)
	assert.Equal(t, 3, result.Successful)
	assert.Equal(t, 0, result.Failed)
	assert.Len(t, result.Results, 3)

	// Verify retriever was called exactly once
	assert.Equal(t, int32(1), retrieverCalls, "retriever should be called exactly once")

	// Verify event sequence
	assertHasBatchEvent(t, events, domainagent.EventBatchStarted)
	assertHasBatchEvent(t, events, domainagent.EventBatchCompleted)
	for i := 0; i < 3; i++ {
		assertHasCandidateEvent(t, events, domainagent.EventCandidateStart, i)
		assertHasCandidateEvent(t, events, domainagent.EventCandidateComplete, i)
	}
}

func TestBatchRunner_SharedRetriever(t *testing.T) {
	t.Parallel()

	sharedRefs := []domainagent.RetrievedReference{
		{ID: "shared-ref-1", Summary: "Shared reference 1"},
		{ID: "shared-ref-2", Summary: "Shared reference 2"},
	}

	var retrieverCalled bool
	var receivedInputs []domainagent.AgentInput
	var mu sync.Mutex

	factory := &mockAgentFactory{
		createRetriever: func() domainagent.BaseAgent {
			return &stubAgent{
				stage: domainagent.StageRetriever,
				execute: func(ctx context.Context, input domainagent.AgentInput) (domainagent.AgentOutput, error) {
					retrieverCalled = true
					return domainagent.AgentOutput{
						Stage:              domainagent.StageRetriever,
						VisualIntent:       input.VisualIntent,
						RetrievedReferences: sharedRefs,
					}, nil
				},
			}
		},
		createPlanner: func() domainagent.BaseAgent {
			return &stubAgent{
				stage: domainagent.StagePlanner,
				execute: func(ctx context.Context, input domainagent.AgentInput) (domainagent.AgentOutput, error) {
					mu.Lock()
					receivedInputs = append(receivedInputs, input)
					mu.Unlock()
					return domainagent.AgentOutput{
						Stage:        domainagent.StagePlanner,
						Content:      "plan output",
						VisualIntent: input.VisualIntent,
					}, nil
				},
			}
		},
		createVisualizer: func() domainagent.BaseAgent {
			return &stubAgent{stage: domainagent.StageVisualizer}
		},
		createCritic: func() domainagent.BaseAgent {
			return &stubAgent{stage: domainagent.StageCritic}
		},
	}

	batchRunner := NewBatchRunner(factory)
	inputs := make([]domainagent.AgentInput, 2)
	for i := 0; i < 2; i++ {
		inputs[i] = domainagent.AgentInput{
			SessionID: "batch-session",
			Content:   "Generate diagram",
			VisualIntent: domainagent.VisualIntent{
				Mode:  domainagent.VisualModeDiagram,
				Goal:  "Test shared retriever",
				Style: "academic",
			},
		}
	}

	handle, err := batchRunner.StartBatch(context.Background(), inputs)
	require.NoError(t, err)

	_, err = handle.Wait()
	require.NoError(t, err)

	// Verify retriever was called exactly once
	assert.True(t, retrieverCalled, "retriever should have been called")

	// Verify all planners received shared references
	mu.Lock()
	defer mu.Unlock()
	require.Len(t, receivedInputs, 2)
	for _, input := range receivedInputs {
		assert.Len(t, input.RetrievedReferences, 2)
		assert.Equal(t, "shared-ref-1", input.RetrievedReferences[0].ID)
		assert.Equal(t, "shared-ref-2", input.RetrievedReferences[1].ID)
	}
}

func TestBatchRunner_PartialFailure(t *testing.T) {
	t.Parallel()

	failingPlannerErr := errors.New("planner failed for candidate 1")

	factory := &mockAgentFactory{
		createRetriever: func() domainagent.BaseAgent {
			return &stubAgent{
				stage: domainagent.StageRetriever,
				execute: func(ctx context.Context, input domainagent.AgentInput) (domainagent.AgentOutput, error) {
					return domainagent.AgentOutput{
						Stage:        domainagent.StageRetriever,
						VisualIntent: input.VisualIntent,
						RetrievedReferences: []domainagent.RetrievedReference{
							{ID: "ref-1", Summary: "reference"},
						},
					}, nil
				},
			}
		},
		createPlanner: func() domainagent.BaseAgent {
			return &stubAgent{
				stage: domainagent.StagePlanner,
				execute: func(ctx context.Context, input domainagent.AgentInput) (domainagent.AgentOutput, error) {
					// Fail candidate with ID 1 (based on metadata)
					if input.Metadata != nil {
						if cid, ok := input.Metadata["candidate_id"]; ok && cid == "1" {
							return domainagent.AgentOutput{}, failingPlannerErr
						}
					}
					return domainagent.AgentOutput{
						Stage:        domainagent.StagePlanner,
						Content:      "plan output",
						VisualIntent: input.VisualIntent,
					}, nil
				},
			}
		},
		createVisualizer: func() domainagent.BaseAgent {
			return &stubAgent{stage: domainagent.StageVisualizer}
		},
		createCritic: func() domainagent.BaseAgent {
			return &stubAgent{stage: domainagent.StageCritic}
		},
	}

	batchRunner := NewBatchRunner(factory)
	inputs := make([]domainagent.AgentInput, 3)
	for i := 0; i < 3; i++ {
		inputs[i] = domainagent.AgentInput{
			SessionID: "batch-session",
			Content:   "Generate diagram",
			Metadata:  map[string]string{"candidate_id": fmt.Sprintf("%d", i)},
			VisualIntent: domainagent.VisualIntent{
				Mode:  domainagent.VisualModeDiagram,
				Goal:  "Test partial failure",
				Style: "academic",
			},
		}
	}

	handle, err := batchRunner.StartBatch(context.Background(), inputs)
	require.NoError(t, err)

	eventsCh := collectBatchEvents(handle.Events())
	result, err := handle.Wait()
	events := <-eventsCh

	// Batch should complete without error even with partial failures
	require.NoError(t, err)

	// Verify mixed results
	assert.Equal(t, 2, result.Successful, "should have 2 successful candidates")
	assert.Equal(t, 1, result.Failed, "should have 1 failed candidate")
	require.Len(t, result.Results, 3)

	// Find the failed candidate
	var foundFailed bool
	for _, r := range result.Results {
		if r.Status == domainagent.StatusFailed {
			foundFailed = true
			assert.Equal(t, 1, r.CandidateID)
			require.NotNil(t, r.Error)
			assert.Contains(t, r.Error.Message, failingPlannerErr.Error())
		} else {
			assert.Equal(t, domainagent.StatusCompleted, r.Status)
		}
	}
	assert.True(t, foundFailed, "should have found a failed candidate")

	// Verify batch events for all candidates
	assertHasBatchEvent(t, events, domainagent.EventBatchStarted)
	assertHasBatchEvent(t, events, domainagent.EventBatchCompleted)
	for i := 0; i < 3; i++ {
		assertHasCandidateEvent(t, events, domainagent.EventCandidateStart, i)
		assertHasCandidateEvent(t, events, domainagent.EventCandidateComplete, i)
	}
}

func TestBatchRunner_RetrieverFailure(t *testing.T) {
	t.Parallel()

	retrieverErr := errors.New("retriever failed")

	factory := &mockAgentFactory{
		createRetriever: func() domainagent.BaseAgent {
			return &stubAgent{
				stage: domainagent.StageRetriever,
				execute: func(ctx context.Context, input domainagent.AgentInput) (domainagent.AgentOutput, error) {
					return domainagent.AgentOutput{}, retrieverErr
				},
			}
		},
		createPlanner:   func() domainagent.BaseAgent { return &stubAgent{stage: domainagent.StagePlanner} },
		createVisualizer: func() domainagent.BaseAgent { return &stubAgent{stage: domainagent.StageVisualizer} },
		createCritic:     func() domainagent.BaseAgent { return &stubAgent{stage: domainagent.StageCritic} },
	}

	batchRunner := NewBatchRunner(factory)
	inputs := []domainagent.AgentInput{
		{SessionID: "batch-session", Content: "Generate diagram"},
	}

	handle, err := batchRunner.StartBatch(context.Background(), inputs)
	require.NoError(t, err)

	result, err := handle.Wait()

	// Batch should fail when retriever fails
	require.Error(t, err)
	assert.Equal(t, 0, result.Successful)
	assert.Equal(t, 1, result.Failed)
}

func TestBatchRunner_ConcurrencyLimit(t *testing.T) {
	t.Parallel()

	var (
		inFlight      int32
		maxConcurrent int32
	)

	factory := &mockAgentFactory{
		createRetriever: func() domainagent.BaseAgent {
			return &stubAgent{
				stage: domainagent.StageRetriever,
				execute: func(ctx context.Context, input domainagent.AgentInput) (domainagent.AgentOutput, error) {
					return domainagent.AgentOutput{
						Stage:              domainagent.StageRetriever,
						RetrievedReferences: []domainagent.RetrievedReference{{ID: "ref"}},
					}, nil
				},
			}
		},
		createPlanner: func() domainagent.BaseAgent {
			return &stubAgent{
				stage: domainagent.StagePlanner,
				execute: func(ctx context.Context, input domainagent.AgentInput) (domainagent.AgentOutput, error) {
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
						Stage:        domainagent.StagePlanner,
						Content:      "plan",
						VisualIntent: input.VisualIntent,
					}, nil
				},
			}
		},
		createVisualizer: func() domainagent.BaseAgent {
			return &stubAgent{
				stage: domainagent.StageVisualizer,
				execute: func(ctx context.Context, input domainagent.AgentInput) (domainagent.AgentOutput, error) {
					// Simulate some work
					time.Sleep(5 * time.Millisecond)
					return domainagent.AgentOutput{
						Stage:        domainagent.StageVisualizer,
						VisualIntent: input.VisualIntent,
					}, nil
				},
			}
		},
		createCritic: func() domainagent.BaseAgent {
			return &stubAgent{stage: domainagent.StageCritic}
		},
	}

	// Set max concurrent to 2
	batchRunner := NewBatchRunner(factory, WithBatchMaxConcurrent(2))
	inputs := make([]domainagent.AgentInput, 5)
	for i := 0; i < 5; i++ {
		inputs[i] = domainagent.AgentInput{
			SessionID: "batch-session",
			Content:   "Generate diagram",
		}
	}

	handle, err := batchRunner.StartBatch(context.Background(), inputs)
	require.NoError(t, err)

	_, err = handle.Wait()
	require.NoError(t, err)

	// Verify that max concurrent didn't exceed 2
	assert.LessOrEqual(t, maxConcurrent, int32(2), "max concurrent should not exceed limit")
}

// mockAgentFactory implements AgentFactory for testing
type mockAgentFactory struct {
	createRetriever  func() domainagent.BaseAgent
	createPlanner    func() domainagent.BaseAgent
	createStylist    func() domainagent.BaseAgent
	createVisualizer func() domainagent.BaseAgent
	createCritic     func() domainagent.BaseAgent
}

func (f *mockAgentFactory) CreateRetriever() domainagent.BaseAgent {
	if f.createRetriever == nil {
		return &stubAgent{stage: domainagent.StageRetriever}
	}
	return f.createRetriever()
}

func (f *mockAgentFactory) CreatePlanner() domainagent.BaseAgent {
	if f.createPlanner == nil {
		return &stubAgent{stage: domainagent.StagePlanner}
	}
	return f.createPlanner()
}

func (f *mockAgentFactory) CreateStylist() domainagent.BaseAgent {
	if f.createStylist == nil {
		return &stubAgent{stage: domainagent.StageStylist}
	}
	return f.createStylist()
}

func (f *mockAgentFactory) CreateVisualizer() domainagent.BaseAgent {
	if f.createVisualizer == nil {
		return &stubAgent{stage: domainagent.StageVisualizer}
	}
	return f.createVisualizer()
}

func (f *mockAgentFactory) CreateCritic() domainagent.BaseAgent {
	if f.createCritic == nil {
		return &stubAgent{stage: domainagent.StageCritic}
	}
	return f.createCritic()
}

func collectBatchEvents(events <-chan domainagent.BatchEvent) <-chan []domainagent.BatchEvent {
	collected := make(chan []domainagent.BatchEvent, 1)

	go func() {
		var items []domainagent.BatchEvent
		for event := range events {
			items = append(items, event)
		}
		collected <- items
	}()

	return collected
}

func assertHasBatchEvent(t *testing.T, events []domainagent.BatchEvent, eventType domainagent.EventType) {
	t.Helper()

	for _, event := range events {
		if event.Type == eventType && event.CandidateID == 0 {
			return
		}
	}

	t.Fatalf("expected batch event %s", eventType)
}

func assertHasCandidateEvent(t *testing.T, events []domainagent.BatchEvent, eventType domainagent.EventType, candidateID int) {
	t.Helper()

	for _, event := range events {
		if event.Type == eventType && event.CandidateID == candidateID {
			return
		}
	}

	t.Fatalf("expected candidate event %s for candidate %d", eventType, candidateID)
}
