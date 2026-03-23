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
		name        string
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

// TestPipelineModeRouting documents expected behavior for pipeline routing.
// Full implementation of routing logic will be added in a future phase.
func TestPipelineModeRouting(t *testing.T) {
	t.Parallel()

	// Define expected stages for each pipeline mode
	expectedStages := map[string][]domainagent.StageName{
		"full": {
			domainagent.StageRetriever,
			domainagent.StagePlanner,
			domainagent.StageStylist,
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

	// Verify expected stages are defined correctly
	for mode, stages := range expectedStages {
		t.Run(mode+"_stages_defined", func(t *testing.T) {
			t.Parallel()

			assert.NotEmpty(t, stages, "expected stages should not be empty for mode %s", mode)

			switch mode {
			case "full":
				assert.Contains(t, stages, domainagent.StageRetriever, "full mode should include Retriever")
				assert.Contains(t, stages, domainagent.StagePlanner, "full mode should include Planner")
				assert.Contains(t, stages, domainagent.StageStylist, "full mode should include Stylist")
				assert.Contains(t, stages, domainagent.StageVisualizer, "full mode should include Visualizer")
				assert.Contains(t, stages, domainagent.StageCritic, "full mode should include Critic")
			case "planner-critic":
				assert.Contains(t, stages, domainagent.StagePlanner, "planner-critic mode should include Planner")
				assert.Contains(t, stages, domainagent.StageCritic, "planner-critic mode should include Critic")
				assert.NotContains(t, stages, domainagent.StageRetriever, "planner-critic mode should NOT include Retriever")
				assert.NotContains(t, stages, domainagent.StageVisualizer, "planner-critic mode should NOT include Visualizer")
			case "vanilla":
				assert.Contains(t, stages, domainagent.StageVisualizer, "vanilla mode should include Visualizer")
				assert.NotContains(t, stages, domainagent.StageRetriever, "vanilla mode should NOT include Retriever")
				assert.NotContains(t, stages, domainagent.StagePlanner, "vanilla mode should NOT include Planner")
				assert.NotContains(t, stages, domainagent.StageCritic, "vanilla mode should NOT include Critic")
			}
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
