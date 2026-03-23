package gemini

import (
	"context"
	"os"
	"testing"

	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientBuildGeminiGenerateConfigSupportsPromptVersionAndImages(t *testing.T) {
	payload, err := buildGeminiGenerateConfig(domainllm.GenerateRequest{
		SystemInstruction: "You are the planner.",
		PromptVersion:     "planner-v5",
		Model:             "gemini-2.0-flash-exp",
		Temperature:       0.4,
		MaxTokens:         300,
		Messages: []domainllm.Message{
			{
				Role: domainllm.RoleUser,
				Parts: []domainllm.Part{
					domainllm.TextPart("Plan the figure."),
					domainllm.InlineImagePart("image/png", []byte("png-bytes")),
				},
			},
		},
	}, "fallback-model")
	require.NoError(t, err)

	require.NotNil(t, payload.SystemInstruction)
	assert.Equal(t, "planner-v5", payload.PromptVersion)
	assert.Len(t, payload.History, 0)
	assert.Len(t, payload.Parts, 2)
	assert.Equal(t, "gemini-2.0-flash-exp", payload.Model)
}

func TestGeminiGenerate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live Gemini call in short mode")
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set")
	}

	client, err := NewClient(apiKey, "gemini-2.0-flash-exp")
	require.NoError(t, err)

	resp, err := client.Generate(context.Background(), domainllm.GenerateRequest{
		Messages: []domainllm.Message{
			{Role: domainllm.RoleUser, Parts: []domainllm.Part{domainllm.TextPart("Say hello in 5 words")}},
		},
		Temperature: 0.2,
		MaxTokens:   32,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Content)
}
