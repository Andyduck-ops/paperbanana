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

func TestSharedBytesBasic(t *testing.T) {
	t.Run("creates nil for empty data", func(t *testing.T) {
		sb := NewSharedBytes(nil)
		assert.Nil(t, sb)

		sb = NewSharedBytes([]byte{})
		assert.Nil(t, sb)
	})

	t.Run("creates and returns data", func(t *testing.T) {
		data := []byte("hello world")
		sb := NewSharedBytes(data)
		require.NotNil(t, sb)

		assert.Equal(t, data, sb.Data())
		assert.Equal(t, 11, sb.Len())
		assert.Equal(t, int32(1), sb.RefCount())
	})

	t.Run("retain increments count", func(t *testing.T) {
		sb := NewSharedBytes([]byte("test"))
		require.Equal(t, int32(1), sb.RefCount())

		sb2 := sb.Retain()
		assert.Same(t, sb, sb2)
		assert.Equal(t, int32(2), sb.RefCount())

		sb3 := sb.Clone() // Clone is alias for Retain
		assert.Same(t, sb, sb3)
		assert.Equal(t, int32(3), sb.RefCount())
	})

	t.Run("release decrements count", func(t *testing.T) {
		sb := NewSharedBytes([]byte("test"))
		sb.Retain()
		require.Equal(t, int32(2), sb.RefCount())

		released := sb.Release()
		assert.False(t, released) // Not released yet
		assert.Equal(t, int32(1), sb.RefCount())
		assert.Equal(t, []byte("test"), sb.Data()) // Data still available
	})

	t.Run("release clears data when count reaches zero", func(t *testing.T) {
		sb := NewSharedBytes([]byte("test"))
		released := sb.Release()
		assert.True(t, released)
		assert.Equal(t, int32(0), sb.RefCount())
		assert.Nil(t, sb.Data())
	})

	t.Run("nil shared bytes operations", func(t *testing.T) {
		var sb *SharedBytes
		assert.Nil(t, sb.Retain())
		assert.False(t, sb.Release())
		assert.Nil(t, sb.Data())
		assert.Equal(t, 0, sb.Len())
		assert.Equal(t, int32(0), sb.RefCount())
	})
}

func TestArtifactClone(t *testing.T) {
	t.Run("clones artifact with shared bytes", func(t *testing.T) {
		data := []byte("image data here")
		original := Artifact{
			ID:       "test-1",
			Kind:     ArtifactKindRenderedFigure,
			MIMEType: "image/png",
			URI:      "memory://test/1",
			Metadata: map[string]string{"dpi": "300"},
		}
		original.SetBytes(data)

		require.NotNil(t, original.Shared)
		require.Equal(t, int32(1), original.Shared.RefCount())

		cloned := original.Clone()

		// Clone should share the same SharedBytes
		assert.Same(t, original.Shared, cloned.Shared)
		assert.Equal(t, int32(2), original.Shared.RefCount())
		assert.Equal(t, data, cloned.GetBytes())

		// Metadata should be copied
		assert.Equal(t, original.Metadata, cloned.Metadata)
		// Check metadata is a copy (different map but same content)
		cloned.Metadata["test"] = "value"
		_, exists := original.Metadata["test"]
		assert.False(t, exists) // Original not modified
	})

	t.Run("clones artifact with legacy bytes", func(t *testing.T) {
		data := []byte("legacy data")
		original := Artifact{
			ID:       "test-2",
			Kind:     ArtifactKindRenderedFigure,
			MIMEType: "image/png",
			URI:      "memory://test/2",
			Bytes:    data,
			Metadata: map[string]string{"version": "1"},
		}

		cloned := original.Clone()

		// When Shared is nil, Clone() does shallow copy of Bytes slice header
		assert.Nil(t, original.Shared)
		assert.Equal(t, data, cloned.Bytes)
		// Note: Bytes is NOT deep copied by Clone() - use SetBytes() for SharedBytes
		// This is intentional: legacy code paths handle their own deep copy logic
	})

	t.Run("nil metadata handling", func(t *testing.T) {
		original := Artifact{
			ID:   "test-3",
			Kind: ArtifactKindPlan,
		}
		original.SetBytes([]byte("test"))

		cloned := original.Clone()
		assert.Nil(t, cloned.Metadata)
	})

	t.Run("empty metadata handling", func(t *testing.T) {
		original := Artifact{
			ID:       "test-4",
			Kind:     ArtifactKindPlan,
			Metadata: map[string]string{},
		}
		original.SetBytes([]byte("test"))

		cloned := original.Clone()
		assert.Nil(t, cloned.Metadata) // cloneStringMap returns nil for empty maps
	})
}

func TestArtifactGetBytesSetBytes(t *testing.T) {
	t.Run("GetBytes prefers SharedBytes", func(t *testing.T) {
		a := Artifact{
			ID:    "test",
			Bytes: []byte("legacy"),
		}
		a.SetBytes([]byte("new shared"))

		// GetBytes should return SharedBytes data
		assert.Equal(t, []byte("new shared"), a.GetBytes())
		// Legacy Bytes field is also updated by SetBytes
		assert.Equal(t, []byte("new shared"), a.Bytes)
	})

	t.Run("GetBytes falls back to legacy Bytes", func(t *testing.T) {
		a := Artifact{
			ID:    "test",
			Bytes: []byte("legacy"),
		}

		assert.Equal(t, []byte("legacy"), a.GetBytes())
	})

	t.Run("SetBytes syncs both fields", func(t *testing.T) {
		a := Artifact{}
		a.SetBytes([]byte("test"))

		assert.NotNil(t, a.Shared)
		assert.Equal(t, []byte("test"), a.Bytes)
		assert.Equal(t, []byte("test"), a.Shared.Data())
	})
}

func TestSharedBytesJSONCompatibility(t *testing.T) {
	t.Run("artifact with SharedBytes serializes via Bytes field", func(t *testing.T) {
		data := []byte("binary data")
		original := Artifact{
			ID:       "test-json",
			Kind:     ArtifactKindRenderedFigure,
			MIMEType: "image/png",
			URI:      "memory://test/json",
		}
		original.SetBytes(data)

		encoded, err := json.Marshal(original)
		require.NoError(t, err)

		var restored Artifact
		require.NoError(t, json.Unmarshal(encoded, &restored))

		// JSON restores to Bytes field (SharedBytes is not serialized)
		assert.Equal(t, data, restored.Bytes)
		assert.Equal(t, original.ID, restored.ID)
		assert.Equal(t, original.Kind, restored.Kind)
		assert.Equal(t, original.MIMEType, restored.MIMEType)
	})
}

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
