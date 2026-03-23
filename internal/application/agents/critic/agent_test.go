package critic

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCriticRounds(t *testing.T) {
	client := &fakeLLMClient{
		responses: []*domainllm.GenerateResponse{
			{
				Content: `{"critic_suggestions":"Tighten the module labels and clarify the merge arrow.","revised_description":"Revised diagram description with corrected labels and clearer merge arrow."}`,
			},
			{
				Content: `{"critic_suggestions":"No changes needed.","revised_description":"No changes needed."}`,
			},
		},
	}
	revisionAgent := &stubRevisionAgent{
		outputs: []domainagent.AgentOutput{
			{
				Stage:        domainagent.StageVisualizer,
				Content:      "Revised diagram description with corrected labels and clearer merge arrow.",
				VisualIntent: testInput(domainagent.VisualModeDiagram).VisualIntent,
				GeneratedArtifacts: []domainagent.Artifact{
					planArtifact("Revised diagram description with corrected labels and clearer merge arrow."),
					renderedArtifact([]byte("revised-image")),
				},
				Metadata: map[string]string{"summary": "rerendered figure"},
			},
		},
	}
	now := time.Date(2026, time.March, 17, 8, 0, 0, 0, time.UTC)

	agent := NewAgent(client, Config{
		Model:         "critic-model",
		Temperature:   0.3,
		MaxRounds:     3,
		RevisionAgent: revisionAgent,
		Now: func() time.Time {
			return now
		},
	})

	output, err := agent.Execute(context.Background(), testInput(domainagent.VisualModeDiagram))
	require.NoError(t, err)

	require.Len(t, client.requests, 2)
	assert.Equal(t, "critic-model", client.requests[0].Model)
	assert.Equal(t, 0.3, client.requests[0].Temperature)
	assert.Equal(t, PromptVersion, client.requests[0].PromptVersion)
	require.Len(t, revisionAgent.inputs, 1)
	assert.Equal(t, "Revised diagram description with corrected labels and clearer merge arrow.", revisionAgent.inputs[0].Content)

	assert.Equal(t, domainagent.StageCritic, output.Stage)
	assert.Equal(t, "Revised diagram description with corrected labels and clearer merge arrow.", output.Content)
	assert.Equal(t, "false", output.Metadata["reused_artifact"])
	require.Len(t, output.CritiqueRounds, 2)
	assert.False(t, output.CritiqueRounds[0].Accepted)
	assert.Equal(t, []string{"Tighten the module labels and clarify the merge arrow."}, output.CritiqueRounds[0].RequestedChanges)
	assert.True(t, output.CritiqueRounds[1].Accepted)
	assert.Equal(t, now, output.CritiqueRounds[0].EvaluatedAt)
	assert.Equal(t, now, output.CritiqueRounds[1].EvaluatedAt)
	require.Len(t, output.GeneratedArtifacts, 4)
	assert.Equal(t, domainagent.ArtifactKindCritique, output.GeneratedArtifacts[2].Kind)
	assert.Equal(t, domainagent.ArtifactKindRenderedFigure, output.GeneratedArtifacts[3].Kind)
	assert.Equal(t, []byte("revised-image"), output.GeneratedArtifacts[3].Bytes)
}

func TestCriticPromptParity(t *testing.T) {
	diagramPrompt, err := SystemPrompt(domainagent.VisualModeDiagram)
	require.NoError(t, err)
	assert.Equal(t, loadFixture(t, "diagram_system.txt"), diagramPrompt)

	plotPrompt, err := SystemPrompt(domainagent.VisualModePlot)
	require.NoError(t, err)
	assert.Equal(t, loadFixture(t, "plot_system.txt"), plotPrompt)
}

func TestCriticReusesPriorArtifactsOnNoChange(t *testing.T) {
	client := &fakeLLMClient{
		responses: []*domainllm.GenerateResponse{
			{
				Content: `{"critic_suggestions":"No changes needed.","revised_description":"No changes needed."}`,
			},
		},
	}
	revisionAgent := &stubRevisionAgent{}
	agent := NewAgent(client, Config{
		MaxRounds:     2,
		RevisionAgent: revisionAgent,
		Now: func() time.Time {
			return time.Date(2026, time.March, 17, 8, 30, 0, 0, time.UTC)
		},
	})

	output, err := agent.Execute(context.Background(), testInput(domainagent.VisualModeDiagram))
	require.NoError(t, err)

	assert.Len(t, client.requests, 1)
	assert.Empty(t, revisionAgent.inputs)
	assert.Equal(t, "true", output.Metadata["reused_artifact"])
	assert.Equal(t, "Current diagram description.", output.Content)
	require.Len(t, output.CritiqueRounds, 1)
	assert.True(t, output.CritiqueRounds[0].Accepted)
	require.Len(t, output.GeneratedArtifacts, 4)
	assert.Equal(t, domainagent.ArtifactKindRenderedFigure, output.GeneratedArtifacts[3].Kind)
	assert.Equal(t, []byte("prior-image"), output.GeneratedArtifacts[3].Bytes)
}

type fakeLLMClient struct {
	requests  []domainllm.GenerateRequest
	responses []*domainllm.GenerateResponse
	err       error
}

func (f *fakeLLMClient) Generate(_ context.Context, req domainllm.GenerateRequest) (*domainllm.GenerateResponse, error) {
	f.requests = append(f.requests, req)
	if f.err != nil {
		return nil, f.err
	}
	if len(f.responses) == 0 {
		return &domainllm.GenerateResponse{}, nil
	}
	response := f.responses[0]
	f.responses = f.responses[1:]
	return response, nil
}

func (f *fakeLLMClient) GenerateStream(context.Context, domainllm.GenerateRequest) (<-chan domainllm.StreamChunk, <-chan error) {
	return nil, nil
}

func (f *fakeLLMClient) Provider() string {
	return "fake"
}

type stubRevisionAgent struct {
	inputs  []domainagent.AgentInput
	outputs []domainagent.AgentOutput
	err     error
}

func (s *stubRevisionAgent) Initialize(context.Context) error {
	return nil
}

func (s *stubRevisionAgent) Execute(_ context.Context, input domainagent.AgentInput) (domainagent.AgentOutput, error) {
	s.inputs = append(s.inputs, input)
	if s.err != nil {
		return domainagent.AgentOutput{}, s.err
	}
	if len(s.outputs) == 0 {
		return domainagent.AgentOutput{
			Stage:        domainagent.StageVisualizer,
			Content:      input.Content,
			VisualIntent: input.VisualIntent,
			GeneratedArtifacts: []domainagent.Artifact{
				planArtifact(input.Content),
				renderedArtifact([]byte("rerendered-image")),
			},
		}, nil
	}
	output := s.outputs[0]
	s.outputs = s.outputs[1:]
	return output, nil
}

func (s *stubRevisionAgent) Cleanup(context.Context) error {
	return nil
}

func (s *stubRevisionAgent) GetState() domainagent.AgentState {
	return domainagent.AgentState{}
}

func (s *stubRevisionAgent) RestoreState(domainagent.AgentState) error {
	return nil
}

func testInput(mode domainagent.VisualMode) domainagent.AgentInput {
	return domainagent.AgentInput{
		SessionID: "session-critic",
		RequestID: "request-critic",
		Content:   currentDescription(mode),
		VisualIntent: domainagent.VisualIntent{
			Mode:  mode,
			Goal:  currentGoal(mode),
			Style: "academic",
		},
		GeneratedArtifacts: []domainagent.Artifact{
			planArtifact(currentDescription(mode)),
			renderedArtifact([]byte("prior-image")),
		},
		Metadata: map[string]string{
			"critic.source_content": currentSourceContent(mode),
		},
	}
}

func planArtifact(content string) domainagent.Artifact {
	return domainagent.Artifact{
		ID:       "planner-current-plan",
		Kind:     domainagent.ArtifactKindPlan,
		MIMEType: "text/plain",
		URI:      "memory://planner/current",
		Content:  content,
	}
}

func renderedArtifact(bytes []byte) domainagent.Artifact {
	return domainagent.Artifact{
		ID:       "visualizer-current-rendered",
		Kind:     domainagent.ArtifactKindRenderedFigure,
		MIMEType: "image/png",
		URI:      "memory://visualizer/current",
		Bytes:    bytes,
	}
}

func currentDescription(mode domainagent.VisualMode) string {
	if mode == domainagent.VisualModePlot {
		return "Current plot description."
	}
	return "Current diagram description."
}

func currentGoal(mode domainagent.VisualMode) string {
	if mode == domainagent.VisualModePlot {
		return "Grouped bar chart comparing benchmark models"
	}
	return "Figure 2: Agent pipeline overview"
}

func currentSourceContent(mode domainagent.VisualMode) string {
	if mode == domainagent.VisualModePlot {
		return `{"series":{"Method A":[1,2,3],"Method B":[2,3,4]}}`
	}
	return "The methodology section describes a retriever, planner, visualizer, and critic loop."
}

func loadFixture(t *testing.T, name string) string {
	t.Helper()

	path := filepath.Join("..", "..", "..", "..", "testdata", "legacy_prompts", "critic", name)
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	return strings.TrimSpace(string(data))
}
