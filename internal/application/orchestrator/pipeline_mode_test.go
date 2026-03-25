package orchestrator

import (
	"context"
	"sync"
	"testing"

	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPipelineModeMetadata verifies that pipeline_mode is correctly passed through metadata.
// This validates the infrastructure is ready for pipeline routing.
func TestPipelineModeMetadata(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		pipelineMode string
	}{
		{name: "full pipeline", pipelineMode: "full"},
		{name: "planner-critic mode", pipelineMode: "planner-critic"},
		{name: "vanilla mode", pipelineMode: "vanilla"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var (
				mu           sync.Mutex
				receivedMeta map[string]string
			)

			runner := NewRunner(newStubRegistry(func(_ context.Context, input domainagent.AgentInput) (domainagent.AgentOutput, error) {
				mu.Lock()
				defer mu.Unlock()
				if receivedMeta == nil {
					receivedMeta = make(map[string]string)
				}
				for k, v := range input.Metadata {
					receivedMeta[k] = v
				}

				return domainagent.AgentOutput{
					Stage:        input.Stage,
					Content:      string(input.Stage) + "-output",
					VisualIntent: input.VisualIntent,
					Prompt:       input.Prompt,
				}, nil
			}), WithEventBuffer(16))

			input := testAgentInput()
			input.Metadata = map[string]string{
				"config.pipeline_mode": tc.pipelineMode,
			}

			handle, err := runner.Start(context.Background(), input)
			require.NoError(t, err)

			_, err = handle.Wait()
			require.NoError(t, err)

			// Verify pipeline_mode is propagated through metadata
			assert.Equal(t, tc.pipelineMode, receivedMeta["config.pipeline_mode"],
				"pipeline_mode should be propagated through agent input metadata")
		})
	}
}

// TestPipelineModeRouting verifies that the runner executes only the stages
// enabled by the requested pipeline mode.
func TestPipelineModeRouting(t *testing.T) {
	t.Parallel()

	// Define expected stages for each pipeline mode
	expectedStages := map[string][]domainagent.StageName{
		"full": {
			domainagent.StageRetriever,
			domainagent.StagePlanner,
			domainagent.StageVisualizer,
			domainagent.StageCritic,
		},
		"planner-critic": {
			domainagent.StagePlanner,
			domainagent.StageCritic,
		},
		"vanilla": {
			domainagent.StageVisualizer,
		},
	}

	for mode, stages := range expectedStages {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()

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
				}, nil
			}), WithEventBuffer(16))

			input := testAgentInput()
			input.Metadata = map[string]string{
				"config.pipeline_mode": mode,
			}

			handle, err := runner.Start(context.Background(), input)
			require.NoError(t, err)

			result, err := handle.Wait()
			require.NoError(t, err)

			assert.Equal(t, stages, callOrder)
			assert.Equal(t, stages, result.Session.Pipeline)
		})
	}
}

// TestCanonicalPipelineIncludesStylist verifies that the canonical pipeline includes the Stylist stage.
func TestCanonicalPipelineIncludesStylist(t *testing.T) {
	pipeline := domainagent.CanonicalPipeline()

	assert.Contains(t, pipeline, domainagent.StageStylist,
		"canonical pipeline should include Stylist stage")

	// Verify Stylist is between Planner and Visualizer
	var plannerIdx, stylistIdx, visualizerIdx int = -1, -1, -1
	for i, stage := range pipeline {
		switch stage {
		case domainagent.StagePlanner:
			plannerIdx = i
		case domainagent.StageStylist:
			stylistIdx = i
		case domainagent.StageVisualizer:
			visualizerIdx = i
		}
	}

	require.True(t, plannerIdx >= 0, "Planner should be in pipeline")
	require.True(t, stylistIdx >= 0, "Stylist should be in pipeline")
	require.True(t, visualizerIdx >= 0, "Visualizer should be in pipeline")

	assert.True(t, plannerIdx < stylistIdx, "Planner should run before Stylist")
	assert.True(t, stylistIdx < visualizerIdx, "Stylist should run before Visualizer")
}

// TestInvalidPipelineModeFallback verifies that invalid pipeline modes fall back to full pipeline.
func TestInvalidPipelineModeFallback(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		pipelineMode string
		expectFull   bool
	}{
		{name: "empty mode falls back to full", pipelineMode: "", expectFull: true},
		{name: "invalid mode falls back to full", pipelineMode: "invalid-mode", expectFull: true},
		{name: "random string falls back to full", pipelineMode: "random", expectFull: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

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
				}, nil
			}), WithEventBuffer(16))

			input := testAgentInput()
			input.Metadata = map[string]string{
				"config.pipeline_mode": tc.pipelineMode,
			}

			handle, err := runner.Start(context.Background(), input)
			require.NoError(t, err)

			result, err := handle.Wait()
			require.NoError(t, err)

			// Invalid modes should fall back to the full pipeline
			fullPipeline := []domainagent.StageName{
				domainagent.StageRetriever,
				domainagent.StagePlanner,
				domainagent.StageStylist,
				domainagent.StageVisualizer,
				domainagent.StageCritic,
			}
			assert.Equal(t, fullPipeline, callOrder, "invalid mode should execute full pipeline")
			assert.Equal(t, fullPipeline, result.Session.Pipeline, "session should reflect full pipeline")
		})
	}
}

// TestPipelineModeValidation verifies the IsValidPipelineMode function.
func TestPipelineModeValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		mode     string
		expected bool
	}{
		{"full", true},
		{"planner-critic", true},
		{"vanilla", true},
		{"", false},
		{"invalid", false},
		{"FULL", false}, // case sensitive
		{"Vanilla", false},
	}

	for _, tc := range tests {
		t.Run(tc.mode, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, domainagent.IsValidPipelineMode(tc.mode))
		})
	}
}
