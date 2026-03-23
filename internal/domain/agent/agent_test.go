package agent

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBaseAgentLifecycle(t *testing.T) {
	contract := reflect.TypeOf((*BaseAgent)(nil)).Elem()
	ctxType := reflect.TypeOf((*context.Context)(nil)).Elem()
	errType := reflect.TypeOf((*error)(nil)).Elem()

	require.Equal(t, 5, contract.NumMethod())

	cases := []struct {
		name    string
		inputs  []reflect.Type
		outputs []reflect.Type
	}{
		{
			name:    "Initialize",
			inputs:  []reflect.Type{ctxType},
			outputs: []reflect.Type{errType},
		},
		{
			name:    "Execute",
			inputs:  []reflect.Type{ctxType, reflect.TypeOf(AgentInput{})},
			outputs: []reflect.Type{reflect.TypeOf(AgentOutput{}), errType},
		},
		{
			name:    "Cleanup",
			inputs:  []reflect.Type{ctxType},
			outputs: []reflect.Type{errType},
		},
		{
			name:    "GetState",
			inputs:  nil,
			outputs: []reflect.Type{reflect.TypeOf(AgentState{})},
		},
		{
			name:    "RestoreState",
			inputs:  []reflect.Type{reflect.TypeOf(AgentState{})},
			outputs: []reflect.Type{errType},
		},
	}

	for _, tc := range cases {
		method, ok := contract.MethodByName(tc.name)
		require.Truef(t, ok, "expected method %s to exist", tc.name)
		require.Equalf(t, len(tc.inputs), method.Type.NumIn(), "unexpected input count for %s", tc.name)
		require.Equalf(t, len(tc.outputs), method.Type.NumOut(), "unexpected output count for %s", tc.name)

		for i, want := range tc.inputs {
			assert.Equalf(t, want, method.Type.In(i), "unexpected input %d for %s", i, tc.name)
		}

		for i, want := range tc.outputs {
			assert.Equalf(t, want, method.Type.Out(i), "unexpected output %d for %s", i, tc.name)
		}
	}
}

func TestAgentStateRoundTrips(t *testing.T) {
	now := time.Date(2026, time.March, 16, 16, 0, 0, 0, time.UTC)

	state := SessionState{
		SchemaVersion: "agent-session/v1",
		SessionID:     "session-01",
		RequestID:     "request-01",
		Status:        StatusCompleted,
		CurrentStage:  StageCritic,
		Pipeline:      CanonicalPipeline(),
		StartedAt:     now,
		UpdatedAt:     now.Add(5 * time.Minute),
		CompletedAt:   now.Add(9 * time.Minute),
		InitialInput: AgentInput{
			SessionID: "session-01",
			RequestID: "request-01",
			Stage:     StageRetriever,
			Content:   "Create an academic figure describing model evaluation.",
			Messages: []domainllm.Message{
				{
					Role:  domainllm.RoleUser,
					Parts: []domainllm.Part{domainllm.TextPart("Create an academic figure describing model evaluation.")},
				},
			},
			VisualIntent: VisualIntent{
				Mode:        VisualModeDiagram,
				Goal:        "Summarize the evaluation workflow",
				Audience:    "ML researchers",
				Style:       "Nature Methods",
				Constraints: []string{"Use concise labels", "Prefer vector-safe shapes"},
			},
			RetrievedReferences: []RetrievedReference{
				{
					ID:          "ref-1",
					Title:       "PaperBanana benchmark example",
					Source:      "paperbanana-bench",
					URI:         "https://example.com/ref-1",
					Summary:     "Shows a four-stage evaluation flow.",
					Score:       0.92,
					Snippets:    []string{"retriever summary", "diagram layout"},
					RetrievedAt: now,
				},
			},
			Prompt: PromptMetadata{
				SystemInstruction: "You are the retriever.",
				Version:           "retriever-v1",
				Template:          "retriever/default",
				Variables:         map[string]string{"task": "diagram"},
			},
			GeneratedArtifacts: []Artifact{
				{
					ID:       "artifact-1",
					Kind:     ArtifactKindReferenceBundle,
					MIMEType: "application/json",
					URI:      "memory://references/1",
					Content:  "{\"references\":1}",
					Metadata: map[string]string{"provider": "memory"},
				},
			},
			CritiqueRounds: []CritiqueRound{
				{
					Round:            1,
					Summary:          "Label alignment looks good.",
					Accepted:         true,
					RequestedChanges: []string{"None"},
					EvaluatedAt:      now.Add(8 * time.Minute),
				},
			},
			Restore: RestoreMetadata{
				SnapshotVersion: "agent-session/v1",
				RestoredFrom:    StagePlanner,
				RestoredAt:      now.Add(-2 * time.Minute),
				ResumeToken:     "resume-01",
			},
			Metadata: map[string]string{"locale": "zh-CN"},
		},
		StageStates: []AgentState{
			{
				Stage:  StageRetriever,
				Status: StatusCompleted,
				Timing: Timing{
					StartedAt:   now,
					CompletedAt: now.Add(time.Minute),
					Duration:    time.Minute,
				},
				Input: AgentInput{
					SessionID: "session-01",
					RequestID: "request-01",
					Stage:     StageRetriever,
					Content:   "Create an academic figure describing model evaluation.",
				},
				Output: AgentOutput{
					Stage: StageRetriever,
					RetrievedReferences: []RetrievedReference{
						{
							ID:          "ref-1",
							Title:       "PaperBanana benchmark example",
							Source:      "paperbanana-bench",
							URI:         "https://example.com/ref-1",
							Summary:     "Shows a four-stage evaluation flow.",
							Score:       0.92,
							Snippets:    []string{"retriever summary"},
							RetrievedAt: now,
						},
					},
					Prompt: PromptMetadata{
						SystemInstruction: "You are the retriever.",
						Version:           "retriever-v1",
						Template:          "retriever/default",
					},
				},
			},
		},
		FinalOutput: AgentOutput{
			Stage: StageCritic,
			GeneratedArtifacts: []Artifact{
				{
					ID:       "artifact-final",
					Kind:     ArtifactKindRenderedFigure,
					MIMEType: "image/png",
					URI:      "memory://figures/final",
					Metadata: map[string]string{"dpi": "300"},
				},
			},
			CritiqueRounds: []CritiqueRound{
				{
					Round:            2,
					Summary:          "Final visualization approved.",
					Accepted:         true,
					RequestedChanges: []string{"Increase axis label contrast"},
					EvaluatedAt:      now.Add(9 * time.Minute),
				},
			},
		},
		Error: &ErrorDetail{
			Message:   "",
			Code:      "",
			Retryable: false,
			Stage:     "",
		},
		Restore: RestoreMetadata{
			SnapshotVersion: "agent-session/v1",
			RestoredFrom:    StagePlanner,
			RestoredAt:      now.Add(-2 * time.Minute),
			ResumeToken:     "resume-01",
		},
		Metadata: map[string]string{"pipeline": "serial"},
	}

	encoded, err := json.Marshal(state)
	require.NoError(t, err)

	var restored SessionState
	require.NoError(t, json.Unmarshal(encoded, &restored))
	assert.Equal(t, state, restored)
}
