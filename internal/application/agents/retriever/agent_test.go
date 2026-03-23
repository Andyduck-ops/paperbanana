package retriever

import (
	"context"
	"encoding/json"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRetrieverExecute(t *testing.T) {
	now := time.Date(2026, time.March, 16, 16, 30, 0, 0, time.UTC)
	store := memoryStore{
		candidates: map[domainagent.VisualMode][]ReferenceExample{
			domainagent.VisualModeDiagram: {
				{
					ID:            "ref_1",
					VisualIntent:  "Agent framework overview",
					Content:       json.RawMessage(`"A methodology section about agents."`),
					PathToGTImage: "diagram/ref_1.png",
				},
				{
					ID:            "ref_2",
					VisualIntent:  "Training pipeline",
					Content:       json.RawMessage(`"A methodology section about training."`),
					PathToGTImage: "diagram/ref_2.png",
				},
				{
					ID:            "ref_3",
					VisualIntent:  "Evaluation chart",
					Content:       json.RawMessage(`"A methodology section about evaluation."`),
					PathToGTImage: "diagram/ref_3.png",
				},
			},
			domainagent.VisualModePlot: {
				{
					ID:            "ref_10",
					VisualIntent:  "Bar chart comparing methods",
					Content:       json.RawMessage(`{"series":[1,2,3]}`),
					PathToGTImage: "plot/ref_10.png",
				},
				{
					ID:            "ref_11",
					VisualIntent:  "Scatter plot with clusters",
					Content:       json.RawMessage(`{"points":[[1,2],[2,3]]}`),
					PathToGTImage: "plot/ref_11.png",
				},
				{
					ID:            "ref_12",
					VisualIntent:  "Line chart with uncertainty",
					Content:       json.RawMessage(`{"series":[[1,2],[2,4]]}`),
					PathToGTImage: "plot/ref_12.png",
				},
			},
		},
		manual: map[domainagent.VisualMode][]ReferenceExample{
			domainagent.VisualModeDiagram: {
				{
					ID:            "ref_20",
					VisualIntent:  "Manual diagram example",
					Content:       json.RawMessage(`"A curated few-shot example."`),
					PathToGTImage: "diagram/ref_20.png",
				},
				{
					ID:            "ref_21",
					VisualIntent:  "Second manual diagram example",
					Content:       json.RawMessage(`"A second curated few-shot example."`),
					PathToGTImage: "diagram/ref_21.png",
				},
			},
		},
	}

	t.Run("none", func(t *testing.T) {
		client := &fakeLLMClient{}
		agent := NewAgent(client, Config{
			Mode:  RetrievalModeNone,
			Store: store,
			Now:   func() time.Time { return now },
		})

		output, err := agent.Execute(context.Background(), testInput(domainagent.VisualModeDiagram))
		require.NoError(t, err)
		assert.Empty(t, output.RetrievedReferences)
		assert.Empty(t, output.GeneratedArtifacts)
		assert.Empty(t, client.requests)
	})

	t.Run("manual", func(t *testing.T) {
		client := &fakeLLMClient{}
		agent := NewAgent(client, Config{
			Mode:  RetrievalModeManual,
			Store: store,
			Now:   func() time.Time { return now },
		})

		output, err := agent.Execute(context.Background(), testInput(domainagent.VisualModeDiagram))
		require.NoError(t, err)
		require.Len(t, output.RetrievedReferences, 2)
		assert.Equal(t, []string{"ref_20", "ref_21"}, referenceIDs(output.RetrievedReferences))
		assert.Equal(t, domainagent.StageRetriever, output.Stage)
		assert.Len(t, output.GeneratedArtifacts, 1)
		assert.Equal(t, domainagent.ArtifactKindReferenceBundle, output.GeneratedArtifacts[0].Kind)

		var examples []ReferenceExample
		require.NoError(t, json.Unmarshal([]byte(output.GeneratedArtifacts[0].Content), &examples))
		assert.Equal(t, []string{"ref_20", "ref_21"}, exampleIDs(examples))
		assert.Empty(t, client.requests)
	})

	t.Run("random", func(t *testing.T) {
		client := &fakeLLMClient{}
		agent := NewAgent(client, Config{
			Mode:   RetrievalModeRandom,
			Store:  store,
			Random: rand.New(rand.NewSource(7)),
			Now:    func() time.Time { return now },
		})

		output, err := agent.Execute(context.Background(), testInput(domainagent.VisualModePlot))
		require.NoError(t, err)
		require.Len(t, output.RetrievedReferences, 3)
		assert.ElementsMatch(t, []string{"ref_10", "ref_11", "ref_12"}, referenceIDs(output.RetrievedReferences))
		assert.Len(t, output.GeneratedArtifacts, 1)
		assert.Empty(t, client.requests)
	})

	t.Run("auto", func(t *testing.T) {
		client := &fakeLLMClient{
			response: &domainllm.GenerateResponse{
				Content: `{"top10_diagrams":["ref_2","ref_1"]}`,
			},
		}
		agent := NewAgent(client, Config{
			Mode:            RetrievalModeAuto,
			Store:           store,
			Model:           "retriever-model",
			Temperature:     0.3,
			MaxOutputTokens: 512,
			Now:             func() time.Time { return now },
		})

		output, err := agent.Execute(context.Background(), testInput(domainagent.VisualModeDiagram))
		require.NoError(t, err)
		require.Len(t, client.requests, 1)
		assert.Equal(t, []string{"ref_2", "ref_1"}, referenceIDs(output.RetrievedReferences))
		assert.Len(t, output.GeneratedArtifacts, 1)
		assert.Equal(t, PromptVersion, output.Prompt.Version)
		assert.Equal(t, "retriever/diagram-system", output.Prompt.Template)

		req := client.requests[0]
		expectedSystem, err := SystemPrompt(domainagent.VisualModeDiagram)
		require.NoError(t, err)
		assert.Equal(t, expectedSystem, req.SystemInstruction)
		assert.Equal(t, PromptVersion, req.PromptVersion)
		assert.Equal(t, "retriever-model", req.Model)
		assert.Equal(t, 0.3, req.Temperature)
		assert.Equal(t, 512, req.MaxTokens)
		require.Len(t, req.Messages, 1)
		assert.Equal(t, domainllm.RoleUser, req.Messages[0].Role)
		require.Len(t, req.Messages[0].Parts, 1)
		assert.Equal(t, domainllm.PartTypeText, req.Messages[0].Parts[0].Type)
		assert.Contains(t, req.Messages[0].Parts[0].Text, "**Target Input**")
		assert.Contains(t, req.Messages[0].Parts[0].Text, "Candidate Diagram 1:")
		assert.Contains(t, req.Messages[0].Parts[0].Text, "select the Top 10 most relevant diagrams")
	})
}

func TestRetrieverPromptParity(t *testing.T) {
	diagramPrompt, err := SystemPrompt(domainagent.VisualModeDiagram)
	require.NoError(t, err)
	assert.Equal(t, loadFixture(t, "diagram_system.txt"), diagramPrompt)

	plotPrompt, err := SystemPrompt(domainagent.VisualModePlot)
	require.NoError(t, err)
	assert.Equal(t, loadFixture(t, "plot_system.txt"), plotPrompt)
}

func TestRetrieverParsesTop10References(t *testing.T) {
	t.Run("diagram fenced json", func(t *testing.T) {
		ids := ParseTopReferences("```json\n{\"top10_diagrams\": [\"ref_1\", \"ref_2\"]}\n```", domainagent.VisualModeDiagram)
		assert.Equal(t, []string{"ref_1", "ref_2"}, ids)
	})

	t.Run("plot partial array", func(t *testing.T) {
		ids := ParseTopReferences(`{"top10_plots": ["ref_10", "ref_11", "ref_12"`, domainagent.VisualModePlot)
		assert.Equal(t, []string{"ref_10", "ref_11", "ref_12"}, ids)
	})

	t.Run("garbage is safe", func(t *testing.T) {
		ids := ParseTopReferences("definitely not json", domainagent.VisualModeDiagram)
		assert.Empty(t, ids)
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

type memoryStore struct {
	candidates map[domainagent.VisualMode][]ReferenceExample
	manual     map[domainagent.VisualMode][]ReferenceExample
}

func (s memoryStore) Candidates(_ context.Context, mode domainagent.VisualMode) ([]ReferenceExample, error) {
	return append([]ReferenceExample(nil), s.candidates[mode]...), nil
}

func (s memoryStore) ManualExamples(_ context.Context, mode domainagent.VisualMode) ([]ReferenceExample, error) {
	return append([]ReferenceExample(nil), s.manual[mode]...), nil
}

func testInput(mode domainagent.VisualMode) domainagent.AgentInput {
	return domainagent.AgentInput{
		SessionID: "session-01",
		RequestID: "request-01",
		Content:   "Detailed methodology text",
		VisualIntent: domainagent.VisualIntent{
			Mode:  mode,
			Goal:  "Compose an academic figure",
			Style: "academic",
		},
	}
}

func loadFixture(t *testing.T, name string) string {
	t.Helper()

	path := filepath.Join("..", "..", "..", "..", "testdata", "legacy_prompts", "retriever", name)
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(data)
}

func referenceIDs(references []domainagent.RetrievedReference) []string {
	ids := make([]string, 0, len(references))
	for _, reference := range references {
		ids = append(ids, reference.ID)
	}
	return ids
}

func exampleIDs(examples []ReferenceExample) []string {
	ids := make([]string, 0, len(examples))
	for _, example := range examples {
		ids = append(ids, example.ID)
	}
	return ids
}
