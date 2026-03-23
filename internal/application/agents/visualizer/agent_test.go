package visualizer

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	pbconfig "github.com/paperbanana/paperbanana/internal/config"
	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
	"github.com/paperbanana/paperbanana/internal/infrastructure/nodes/httpnode"
	"github.com/paperbanana/paperbanana/internal/infrastructure/resilience"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVisualizerExecute(t *testing.T) {
	t.Run("diagram", func(t *testing.T) {
		client := &fakeLLMClient{
			response: &domainllm.GenerateResponse{
				Parts: []domainllm.Part{
					domainllm.InlineImagePart("image/png", []byte("diagram-image")),
				},
			},
		}
		agent := NewAgent(client, Config{
			Model:       "visualizer-model",
			Temperature: 0.2,
		})

		output, err := agent.Execute(context.Background(), testInput(domainagent.VisualModeDiagram))
		require.NoError(t, err)
		require.Len(t, client.requests, 1)

		req := client.requests[0]
		assert.Equal(t, "visualizer-model", req.Model)
		assert.Equal(t, 0.2, req.Temperature)
		assert.Equal(t, PromptVersion, req.PromptVersion)
		require.Len(t, req.Messages, 1)
		require.Len(t, req.Messages[0].Parts, 1)
		assert.Contains(t, req.Messages[0].Parts[0].Text, "Render an image based on the following detailed description:")
		assert.Contains(t, req.Messages[0].Parts[0].Text, "do not include figure titles in the image")

		assert.Equal(t, domainagent.StageVisualizer, output.Stage)
		assert.Equal(t, "visualizer/diagram-system", output.Prompt.Template)
		assert.Equal(t, testInput(domainagent.VisualModeDiagram).Content, output.Content)
		assert.Equal(t, "llm-image", output.Metadata["execution_path"])

		require.Len(t, output.GeneratedArtifacts, 2)
		assert.Equal(t, domainagent.ArtifactKindPlan, output.GeneratedArtifacts[0].Kind)
		assert.Equal(t, domainagent.ArtifactKindRenderedFigure, output.GeneratedArtifacts[1].Kind)
		assert.Equal(t, []byte("diagram-image"), output.GeneratedArtifacts[1].Bytes)
		assert.Equal(t, "image/png", output.GeneratedArtifacts[1].MIMEType)
	})

	t.Run("plot", func(t *testing.T) {
		client := &fakeLLMClient{
			response: &domainllm.GenerateResponse{
				Content: "```python\nprint('plot')\n```",
			},
		}
		agent := NewAgent(client, Config{
			Model:        "plot-model",
			Temperature:  0.6,
			PlotExecutor: fakePlotExecutor{result: PlotExecutionResult{Bytes: []byte("plot-image"), MIMEType: "image/jpeg"}},
		})

		output, err := agent.Execute(context.Background(), testInput(domainagent.VisualModePlot))
		require.NoError(t, err)
		require.Len(t, client.requests, 1)

		req := client.requests[0]
		assert.Equal(t, "plot-model", req.Model)
		assert.Equal(t, 0.6, req.Temperature)
		assert.Equal(t, PromptVersion, req.PromptVersion)
		require.Len(t, req.Messages, 1)
		require.Len(t, req.Messages[0].Parts, 1)
		assert.Contains(t, req.Messages[0].Parts[0].Text, "Use python matplotlib to generate a statistical plot based on the following detailed description:")
		assert.Contains(t, req.Messages[0].Parts[0].Text, "Only provide the code without any explanations. Code:")

		assert.Equal(t, domainagent.StageVisualizer, output.Stage)
		assert.Equal(t, "visualizer/plot-system", output.Prompt.Template)
		assert.Equal(t, testInput(domainagent.VisualModePlot).Content, output.Content)
		assert.Equal(t, "llm-plot", output.Metadata["execution_path"])

		require.Len(t, output.GeneratedArtifacts, 3)
		assert.Equal(t, domainagent.ArtifactKindPlan, output.GeneratedArtifacts[0].Kind)
		assert.Equal(t, domainagent.ArtifactKindPromptTrace, output.GeneratedArtifacts[1].Kind)
		assert.Equal(t, "```python\nprint('plot')\n```", output.GeneratedArtifacts[1].Content)
		assert.Equal(t, domainagent.ArtifactKindRenderedFigure, output.GeneratedArtifacts[2].Kind)
		assert.Equal(t, []byte("plot-image"), output.GeneratedArtifacts[2].Bytes)
		assert.Equal(t, "image/jpeg", output.GeneratedArtifacts[2].MIMEType)
	})
}

func TestVisualizerPromptParity(t *testing.T) {
	diagramPrompt, err := SystemPrompt(domainagent.VisualModeDiagram)
	require.NoError(t, err)
	assert.Equal(t, loadFixture(t, "diagram_system.txt"), diagramPrompt)

	plotPrompt, err := SystemPrompt(domainagent.VisualModePlot)
	require.NoError(t, err)
	assert.Equal(t, loadFixture(t, "plot_system.txt"), plotPrompt)
}

func TestVisualizerReusesPriorArtifactWhenCriticRequestsNoChange(t *testing.T) {
	client := &fakeLLMClient{}
	agent := NewAgent(client, Config{
		PlotExecutor: fakePlotExecutor{result: PlotExecutionResult{Bytes: []byte("new-image"), MIMEType: "image/jpeg"}},
	})

	input := testInput(domainagent.VisualModeDiagram)
	input.GeneratedArtifacts = append(input.GeneratedArtifacts, domainagent.Artifact{
		ID:       "visualizer-diagram-rendered",
		Kind:     domainagent.ArtifactKindRenderedFigure,
		MIMEType: "image/png",
		URI:      "memory://visualizer/diagram/rendered",
		Bytes:    []byte("prior-image"),
	})
	input.CritiqueRounds = []domainagent.CritiqueRound{
		{
			Round:            1,
			Summary:          "Looks good.",
			RequestedChanges: []string{"No changes needed."},
			EvaluatedAt:      time.Date(2026, time.March, 16, 16, 0, 0, 0, time.UTC),
		},
	}

	output, err := agent.Execute(context.Background(), input)
	require.NoError(t, err)
	assert.Empty(t, client.requests)
	assert.Equal(t, "true", output.Metadata["reused_artifact"])
	require.Len(t, output.GeneratedArtifacts, 2)
	assert.Equal(t, []byte("prior-image"), output.GeneratedArtifacts[1].Bytes)
}

func TestVisualizerInvokesConfiguredNode(t *testing.T) {
	const imagePayload = "node-rendered-image"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "token-123", r.Header.Get("X-API-Key"))

		var payload map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		assert.Equal(t, "Create a clean comparison diagram for the agent pipeline.", payload["prompt"])
		assert.Equal(t, "diagram", payload["mode"])
		assert.Equal(t, "Agent pipeline comparison", payload["goal"])
		assert.Equal(t, "request-01", payload["request_id"])

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"image_base64":"` + base64.StdEncoding.EncodeToString([]byte(imagePayload)) + `","mime_type":"image/png","summary":"configured node render"}`))
	}))
	defer server.Close()

	configPath := filepath.Join(t.TempDir(), "custom_nodes.yaml")
	err := os.WriteFile(configPath, []byte(`
custom_nodes:
  - name: external_visualizer
    url: `+server.URL+`
    method: POST
    headers:
      X-API-Key: token-123
    request_template:
      prompt: "{{content}}"
      mode: "{{visual_intent.mode}}"
      goal: "{{visual_intent.goal}}"
      request_id: "{{request_id}}"
    response_parser: json_path
    response_selectors:
      image_base64: $.image_base64
      mime_type: $.mime_type
      summary: $.summary
`), 0o644)
	require.NoError(t, err)

	catalog, err := pbconfig.LoadNodeConfig(configPath)
	require.NoError(t, err)

	agent := NewAgent(nil, Config{
		NodeCatalog: catalog,
		NodeAdapter: httpnode.NewAdapter(resilience.NewResilientClient("visualizer-node", time.Second)),
	})

	input := testInput(domainagent.VisualModeDiagram)
	input.Metadata["visualizer.node_name"] = "external_visualizer"

	output, err := agent.Execute(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, "configured-node", output.Metadata["execution_path"])
	require.Len(t, output.GeneratedArtifacts, 2)
	assert.Equal(t, []byte(imagePayload), output.GeneratedArtifacts[1].Bytes)
	assert.Equal(t, "image/png", output.GeneratedArtifacts[1].MIMEType)
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

type fakePlotExecutor struct {
	result PlotExecutionResult
	err    error
}

func (f fakePlotExecutor) Execute(context.Context, string) (PlotExecutionResult, error) {
	return f.result, f.err
}

func testInput(mode domainagent.VisualMode) domainagent.AgentInput {
	goal := "Agent pipeline comparison"
	content := "Create a clean comparison diagram for the agent pipeline."
	if mode == domainagent.VisualModePlot {
		goal = "Grouped benchmark comparison"
		content = "Render a grouped bar chart that compares Retriever, Planner, and Visualizer latency across two datasets."
	}

	return domainagent.AgentInput{
		SessionID: "session-01",
		RequestID: "request-01",
		Content:   content,
		VisualIntent: domainagent.VisualIntent{
			Mode:  mode,
			Goal:  goal,
			Style: "academic",
		},
		GeneratedArtifacts: []domainagent.Artifact{
			{
				ID:       "planner-plan",
				Kind:     domainagent.ArtifactKindPlan,
				MIMEType: "text/plain",
				URI:      "memory://planner/current",
				Content:  content,
			},
		},
		Metadata: map[string]string{},
	}
}

func loadFixture(t *testing.T, name string) string {
	t.Helper()

	path := filepath.Join("..", "..", "..", "..", "testdata", "legacy_prompts", "visualizer", name)
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(data)
}
