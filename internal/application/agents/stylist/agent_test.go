package stylist

import (
	"context"
	"testing"

	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStylistAgentInitializesCorrectly(t *testing.T) {
	client := &fakeLLMClient{}
	agent := NewAgent(client, Config{Model: "stylist-model"})

	err := agent.Initialize(context.Background())
	require.NoError(t, err)

	state := agent.GetState()
	assert.Equal(t, domainagent.StageStylist, state.Stage)
	assert.Equal(t, domainagent.StatusRunning, state.Status)
}

func TestStylistAgentExecuteReturnsEnhancedPlan(t *testing.T) {
	client := &fakeLLMClient{
		response: &domainllm.GenerateResponse{
			Content: "Enhanced visualization plan with NeurIPS 2025 style guidelines applied.",
		},
	}
	agent := NewAgent(client, Config{
		Model:       "stylist-model",
		Temperature: 0.3,
	})

	input := domainagent.AgentInput{
		SessionID: "session-01",
		RequestID: "request-01",
		Content:   "Original plan from planner",
		VisualIntent: domainagent.VisualIntent{
			Mode:  domainagent.VisualModeDiagram,
			Goal:  "Create a pipeline diagram",
			Style: "academic",
		},
		GeneratedArtifacts: []domainagent.Artifact{
			{
				ID:      "planner-plan",
				Kind:    domainagent.ArtifactKindPlan,
				Content: "Original plan from planner",
			},
		},
	}

	output, err := agent.Execute(context.Background(), input)
	require.NoError(t, err)

	assert.Equal(t, domainagent.StageStylist, output.Stage)
	assert.Equal(t, "Enhanced visualization plan with NeurIPS 2025 style guidelines applied.", output.Content)
	assert.Equal(t, input.VisualIntent, output.VisualIntent)
	assert.Equal(t, input.RetrievedReferences, output.RetrievedReferences)
}

func TestStylistAgentPreservesSemanticContent(t *testing.T) {
	client := &fakeLLMClient{
		response: &domainllm.GenerateResponse{
			Content: "Enhanced plan preserving original meaning",
		},
	}
	agent := NewAgent(client, Config{Model: "stylist-model"})

	input := domainagent.AgentInput{
		SessionID: "session-01",
		Content:   "Original semantic content",
		VisualIntent: domainagent.VisualIntent{
			Mode: domainagent.VisualModePlot,
			Goal: "Create a benchmark comparison chart",
		},
		RetrievedReferences: []domainagent.RetrievedReference{
			{ID: "ref-1", Title: "Reference 1"},
		},
		GeneratedArtifacts: []domainagent.Artifact{
			{ID: "artifact-1", Kind: domainagent.ArtifactKindPlan},
		},
		Metadata: map[string]string{"key": "value"},
	}

	output, err := agent.Execute(context.Background(), input)
	require.NoError(t, err)

	// Verify semantic content is preserved (passed through)
	assert.Equal(t, input.RetrievedReferences, output.RetrievedReferences)
	assert.Equal(t, input.VisualIntent, output.VisualIntent)
	assert.Equal(t, input.Metadata, output.Metadata)
}

func TestStylistAgentPromptContainsStyleGuide(t *testing.T) {
	client := &fakeLLMClient{
		response: &domainllm.GenerateResponse{
			Content: "Enhanced plan",
		},
	}
	agent := NewAgent(client, Config{Model: "stylist-model"})

	input := domainagent.AgentInput{
		Content: "Test plan",
		VisualIntent: domainagent.VisualIntent{
			Mode: domainagent.VisualModeDiagram,
		},
	}

	_, err := agent.Execute(context.Background(), input)
	require.NoError(t, err)

	require.Len(t, client.requests, 1)
	req := client.requests[0]

	// Check that the system instruction contains NeurIPS style guide reference
	assert.Contains(t, req.SystemInstruction, "NeurIPS 2025")
	assert.Contains(t, req.SystemInstruction, "Style Guidelines")
}

func TestStylistAgentPromptContainsVisualMode(t *testing.T) {
	client := &fakeLLMClient{
		response: &domainllm.GenerateResponse{
			Content: "Enhanced plan",
		},
	}
	agent := NewAgent(client, Config{Model: "stylist-model"})

	input := domainagent.AgentInput{
		Content: "Test plan",
		VisualIntent: domainagent.VisualIntent{
			Mode: domainagent.VisualModePlot,
		},
	}

	_, err := agent.Execute(context.Background(), input)
	require.NoError(t, err)

	require.Len(t, client.requests, 1)
	req := client.requests[0]

	// Check that the message contains the visual mode
	assert.Contains(t, domainllm.CollectText(req.Messages[0].Parts), "plot")
}

func TestStylistAgentHandlesDiagramAndPlotModes(t *testing.T) {
	tests := []struct {
		name string
		mode domainagent.VisualMode
	}{
		{"diagram mode", domainagent.VisualModeDiagram},
		{"plot mode", domainagent.VisualModePlot},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &fakeLLMClient{
				response: &domainllm.GenerateResponse{
					Content: "Enhanced plan for " + string(tt.mode),
				},
			}
			agent := NewAgent(client, Config{Model: "stylist-model"})

			input := domainagent.AgentInput{
				Content: "Test plan",
				VisualIntent: domainagent.VisualIntent{
					Mode: tt.mode,
				},
			}

			output, err := agent.Execute(context.Background(), input)
			require.NoError(t, err)
			assert.Contains(t, output.Content, string(tt.mode))
		})
	}
}

func TestStylistAgentCleanup(t *testing.T) {
	agent := NewAgent(&fakeLLMClient{}, Config{})

	err := agent.Cleanup(context.Background())
	assert.NoError(t, err)
}

func TestStylistAgentGetState(t *testing.T) {
	client := &fakeLLMClient{}
	agent := NewAgent(client, Config{})

	_ = agent.Initialize(context.Background())
	state := agent.GetState()

	assert.Equal(t, domainagent.StageStylist, state.Stage)
}

func TestStylistAgentRestoreState(t *testing.T) {
	agent := NewAgent(&fakeLLMClient{}, Config{})

	savedState := domainagent.AgentState{
		Stage:  domainagent.StageStylist,
		Status: domainagent.StatusCompleted,
		Output: domainagent.AgentOutput{
			Content: "saved content",
		},
	}

	err := agent.RestoreState(savedState)
	require.NoError(t, err)

	state := agent.GetState()
	assert.Equal(t, domainagent.StageStylist, state.Stage)
	assert.Equal(t, domainagent.StatusCompleted, state.Status)
	assert.Equal(t, "saved content", state.Output.Content)
}

func TestStylistAgentGeneratesPlanArtifact(t *testing.T) {
	client := &fakeLLMClient{
		response: &domainllm.GenerateResponse{
			Content: "Enhanced plan with style",
		},
	}
	agent := NewAgent(client, Config{Model: "stylist-model"})

	input := domainagent.AgentInput{
		SessionID: "session-01",
		Content:   "Original plan",
		VisualIntent: domainagent.VisualIntent{
			Mode: domainagent.VisualModeDiagram,
		},
		GeneratedArtifacts: []domainagent.Artifact{
			{
				ID:      "planner-plan",
				Kind:    domainagent.ArtifactKindPlan,
				Content: "Original plan",
			},
		},
	}

	output, err := agent.Execute(context.Background(), input)
	require.NoError(t, err)

	// Find the stylist's plan artifact
	var stylistPlan *domainagent.Artifact
	for i, artifact := range output.GeneratedArtifacts {
		if artifact.ID == "stylist-diagram-plan" {
			stylistPlan = &output.GeneratedArtifacts[i]
			break
		}
	}

	require.NotNil(t, stylistPlan, "Expected stylist to generate a plan artifact")
	assert.Equal(t, domainagent.ArtifactKindPlan, stylistPlan.Kind)
	assert.Equal(t, "Enhanced plan with style", stylistPlan.Content)
	assert.Contains(t, stylistPlan.URI, "stylist")
}

// fakeLLMClient implements domainllm.LLMClient for testing
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
