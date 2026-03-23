package planner

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlannerExecute(t *testing.T) {
	client := &fakeLLMClient{
		response: &domainllm.GenerateResponse{
			Content: "Detailed diagram plan with labeled modules and arrows.",
		},
	}
	agent := NewAgent(client, Config{
		Model:       "planner-model",
		Temperature: 0.4,
		LoadExampleImage: func(mode domainagent.VisualMode, path string) ([]byte, string, error) {
			require.Equal(t, domainagent.VisualModeDiagram, mode)
			return []byte("diagram-image:" + path), "image/png", nil
		},
	})

	output, err := agent.Execute(context.Background(), testInput(t, domainagent.VisualModeDiagram))
	require.NoError(t, err)
	require.Len(t, client.requests, 1)

	req := client.requests[0]
	require.Len(t, req.Messages, 1)
	assert.Equal(t, domainllm.RoleUser, req.Messages[0].Role)
	assert.Equal(t, "planner-model", req.Model)
	assert.Equal(t, 0.4, req.Temperature)
	assert.Equal(t, PromptVersion, req.PromptVersion)
	assert.Equal(t, "planner/diagram-system", output.Prompt.Template)

	assert.Equal(t, domainagent.StagePlanner, output.Stage)
	assert.Equal(t, "Detailed diagram plan with labeled modules and arrows.", output.Content)
	assert.Equal(t, testInput(t, domainagent.VisualModeDiagram).RetrievedReferences, output.RetrievedReferences)
	assert.Equal(t, "Detailed diagram plan with labeled modules and arrows.", output.Metadata["summary"])
	require.Len(t, output.GeneratedArtifacts, 2)
	assert.Equal(t, domainagent.ArtifactKindReferenceBundle, output.GeneratedArtifacts[0].Kind)
	assert.Equal(t, domainagent.ArtifactKindPlan, output.GeneratedArtifacts[1].Kind)
	assert.Equal(t, "Detailed diagram plan with labeled modules and arrows.", output.GeneratedArtifacts[1].Content)
	assert.Equal(t, "memory://planner/diagram/plan", output.GeneratedArtifacts[1].URI)
}

func TestPlannerPromptParity(t *testing.T) {
	diagramPrompt, err := SystemPrompt(domainagent.VisualModeDiagram)
	require.NoError(t, err)
	assert.Equal(t, loadFixture(t, "diagram_system.txt"), diagramPrompt)

	plotPrompt, err := SystemPrompt(domainagent.VisualModePlot)
	require.NoError(t, err)
	assert.Equal(t, loadFixture(t, "plot_system.txt"), plotPrompt)
}

func TestPlannerBuildsMultimodalRequest(t *testing.T) {
	t.Run("diagram", func(t *testing.T) {
		client := &fakeLLMClient{
			response: &domainllm.GenerateResponse{Content: "diagram plan"},
		}
		agent := NewAgent(client, Config{
			LoadExampleImage: func(mode domainagent.VisualMode, path string) ([]byte, string, error) {
				require.Equal(t, domainagent.VisualModeDiagram, mode)
				return []byte("image:" + path), "image/png", nil
			},
		})

		_, err := agent.Execute(context.Background(), testInput(t, domainagent.VisualModeDiagram))
		require.NoError(t, err)

		req := client.requests[0]
		require.Len(t, req.Messages, 1)
		require.Len(t, req.Messages[0].Parts, 5)
		assert.Equal(t, domainllm.PartTypeText, req.Messages[0].Parts[0].Type)
		assert.Contains(t, req.Messages[0].Parts[0].Text, "Example 1:")
		assert.Contains(t, req.Messages[0].Parts[0].Text, "Methodology Section: Method section for agent framework.")
		assert.Contains(t, req.Messages[0].Parts[0].Text, "Diagram Caption: Agent framework overview")
		assert.Equal(t, domainllm.PartTypeImage, req.Messages[0].Parts[1].Type)
		assert.Equal(t, []byte("image:ref_1.png"), req.Messages[0].Parts[1].Data)
		assert.Equal(t, domainllm.PartTypeText, req.Messages[0].Parts[2].Type)
		assert.Equal(t, domainllm.PartTypeImage, req.Messages[0].Parts[3].Type)
		assert.Contains(t, req.Messages[0].Parts[4].Text, "Now, based on the following methodology section and diagram caption")
		assert.Contains(t, req.Messages[0].Parts[4].Text, "Detailed description of the target figure to be generated (do not include figure titles):")
	})

	t.Run("plot", func(t *testing.T) {
		client := &fakeLLMClient{
			response: &domainllm.GenerateResponse{Content: "plot plan"},
		}
		agent := NewAgent(client, Config{
			LoadExampleImage: func(mode domainagent.VisualMode, path string) ([]byte, string, error) {
				require.Equal(t, domainagent.VisualModePlot, mode)
				return []byte("plot-image:" + path), "image/png", nil
			},
		})

		_, err := agent.Execute(context.Background(), testInput(t, domainagent.VisualModePlot))
		require.NoError(t, err)

		req := client.requests[0]
		require.Len(t, req.Messages, 1)
		require.Len(t, req.Messages[0].Parts, 5)
		assert.Contains(t, req.Messages[0].Parts[0].Text, "Plot Raw Data: {\"series\":[1,2,3]}")
		assert.Contains(t, req.Messages[0].Parts[0].Text, "Visual Intent of the Desired Plot: Grouped bar chart comparing three methods")
		assert.Equal(t, []byte("plot-image:ref_10.png"), req.Messages[0].Parts[1].Data)
		assert.Contains(t, req.Messages[0].Parts[4].Text, "Now, based on the following plot raw data and visual intent of the desired plot")
		assert.Contains(t, req.Messages[0].Parts[4].Text, "Detailed description of the target figure to be generated:")
		assert.NotContains(t, req.Messages[0].Parts[4].Text, "do not include figure titles")
	})
}

type fakeLLMClient struct {
	requests []domainllm.GenerateRequest
	response *domainllm.GenerateResponse
	err      error
}

func (f *fakeLLMClient) Generate(_ context.Context, req domainllm.GenerateRequest) (*domainllm.GenerateResponse, error) {
	f.requests = append(f.requests, req)
	if f.err != nil {
		return nil, f.err
	}
	if f.response != nil {
		return f.response, nil
	}
	return &domainllm.GenerateResponse{}, nil
}

func (f *fakeLLMClient) GenerateStream(context.Context, domainllm.GenerateRequest) (<-chan domainllm.StreamChunk, <-chan error) {
	return nil, nil
}

func (f *fakeLLMClient) Provider() string {
	return "fake"
}

func testInput(t *testing.T, mode domainagent.VisualMode) domainagent.AgentInput {
	t.Helper()

	examples := []referenceExample{
		{
			ID:            exampleIDForMode(mode, 1),
			VisualIntent:  exampleIntentForMode(mode, 1),
			Content:       exampleContentForMode(mode, 1),
			PathToGTImage: examplePathForMode(mode, 1),
		},
		{
			ID:            exampleIDForMode(mode, 2),
			VisualIntent:  exampleIntentForMode(mode, 2),
			Content:       exampleContentForMode(mode, 2),
			PathToGTImage: examplePathForMode(mode, 2),
		},
	}
	content, err := json.Marshal(examples)
	require.NoError(t, err)

	return domainagent.AgentInput{
		SessionID: "session-01",
		RequestID: "request-01",
		Content:   currentContentForMode(mode),
		VisualIntent: domainagent.VisualIntent{
			Mode:  mode,
			Goal:  currentIntentForMode(mode),
			Style: "academic",
		},
		RetrievedReferences: []domainagent.RetrievedReference{
			{ID: exampleIDForMode(mode, 1), Title: exampleIntentForMode(mode, 1), URI: examplePathForMode(mode, 1)},
			{ID: exampleIDForMode(mode, 2), Title: exampleIntentForMode(mode, 2), URI: examplePathForMode(mode, 2)},
		},
		GeneratedArtifacts: []domainagent.Artifact{
			{
				ID:       "retriever-bundle",
				Kind:     domainagent.ArtifactKindReferenceBundle,
				MIMEType: "application/json",
				URI:      "memory://retriever/examples",
				Content:  string(content),
			},
		},
	}
}

func loadFixture(t *testing.T, name string) string {
	t.Helper()

	path := filepath.Join("..", "..", "..", "..", "testdata", "legacy_prompts", "planner", name)
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(data)
}

func exampleIDForMode(mode domainagent.VisualMode, index int) string {
	if mode == domainagent.VisualModePlot {
		if index == 1 {
			return "ref_10"
		}
		return "ref_11"
	}
	if index == 1 {
		return "ref_1"
	}
	return "ref_2"
}

func exampleIntentForMode(mode domainagent.VisualMode, index int) string {
	if mode == domainagent.VisualModePlot {
		if index == 1 {
			return "Grouped bar chart comparing three methods"
		}
		return "Scatter plot with confidence bands"
	}
	if index == 1 {
		return "Agent framework overview"
	}
	return "Training pipeline"
}

func exampleContentForMode(mode domainagent.VisualMode, index int) json.RawMessage {
	if mode == domainagent.VisualModePlot {
		if index == 1 {
			return json.RawMessage(`{"series":[1,2,3]}`)
		}
		return json.RawMessage(`{"points":[[1,2],[2,3]]}`)
	}
	if index == 1 {
		return json.RawMessage(`"Method section for agent framework."`)
	}
	return json.RawMessage(`"Method section for a training pipeline."`)
}

func examplePathForMode(mode domainagent.VisualMode, index int) string {
	if mode == domainagent.VisualModePlot {
		if index == 1 {
			return "ref_10.png"
		}
		return "ref_11.png"
	}
	if index == 1 {
		return "ref_1.png"
	}
	return "ref_2.png"
}

func currentContentForMode(mode domainagent.VisualMode) string {
	if mode == domainagent.VisualModePlot {
		return "{\"categories\":[\"A\",\"B\",\"C\"],\"series\":{\"Method A\":[1,2,3],\"Method B\":[2,3,4]}}"
	}
	return "The methodology section describes a retriever, planner, and visualizer pipeline."
}

func currentIntentForMode(mode domainagent.VisualMode) string {
	if mode == domainagent.VisualModePlot {
		return "Grouped bar chart for benchmark comparison"
	}
	return "Diagram caption for the pipeline overview"
}
