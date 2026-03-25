package retriever

import (
	"encoding/json"
	"strings"
	"testing"

	domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShortlistCandidatesPrioritizesRelevantVisualIntent(t *testing.T) {
	input := domainagent.AgentInput{
		Content: "The method combines retrieval, planning, rendering, and revision.",
		VisualIntent: domainagent.VisualIntent{
			Mode: domainagent.VisualModeDiagram,
			Goal: "A workflow pipeline overview for an agentic academic figure system",
		},
	}

	candidates := []ReferenceExample{
		{
			ID:           "irrelevant",
			VisualIntent: "Microscopy analysis figure",
			Content:      json.RawMessage(`"cell segmentation and pathology review"`),
		},
		{
			ID:           "relevant",
			VisualIntent: "Agent workflow pipeline overview",
			Content:      json.RawMessage(`"retrieval planning rendering and critic revision loop"`),
		},
		{
			ID:           "secondary",
			VisualIntent: "System architecture diagram",
			Content:      json.RawMessage(`"module overview with planner and renderer"`),
		},
	}

	selected := shortlistCandidates(input, candidates, 2, 0)
	require.Len(t, selected, 2)
	assert.Equal(t, "relevant", selected[0].ID)
	assert.Equal(t, "secondary", selected[1].ID)
}

func TestBuildUserPromptCapsCandidatePoolForAutoRetrieval(t *testing.T) {
	input := domainagent.AgentInput{
		Content: "A paper about retrieval and planning for figure generation.",
		VisualIntent: domainagent.VisualIntent{
			Mode: domainagent.VisualModeDiagram,
			Goal: "Pipeline overview figure",
		},
	}

	candidates := make([]ReferenceExample, 0, 14)
	for i := 0; i < 14; i++ {
		candidates = append(candidates, ReferenceExample{
			ID:           "ref_" + strings.Repeat("x", i+1),
			VisualIntent: "Pipeline overview candidate",
			Content:      json.RawMessage(`"retrieval planning rendering revision"`),
		})
	}

	prompt, err := buildUserPrompt(input, candidates)
	require.NoError(t, err)
	assert.Equal(t, 12, strings.Count(prompt, "Candidate Diagram "))
}
