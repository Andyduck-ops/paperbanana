package anthropic

import (
	"context"
	"os"
	"testing"

	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientBuildAnthropicRequestSupportsPromptVersionAndImages(t *testing.T) {
	payload, err := buildAnthropicRequest(domainllm.GenerateRequest{
		SystemInstruction: "You are the critic.",
		PromptVersion:     "critic-v3",
		Model:             "claude-3-5-sonnet-20241022",
		Temperature:       0.2,
		MaxTokens:         512,
		Messages: []domainllm.Message{
			{
				Role: domainllm.RoleUser,
				Parts: []domainllm.Part{
					domainllm.TextPart("Review the chart."),
					domainllm.InlineImagePart("image/png", []byte("png-bytes")),
				},
			},
		},
	}, "fallback-model")
	require.NoError(t, err)

	assert.Equal(t, "You are the critic.", payload.System)
	assert.Equal(t, "critic-v3", payload.Metadata["prompt_version"])
	require.Len(t, payload.Messages, 1)
	require.Len(t, payload.Messages[0].Content, 2)
	assert.Equal(t, "text", payload.Messages[0].Content[0].Type)
	assert.Equal(t, "Review the chart.", payload.Messages[0].Content[0].Text)
	assert.Equal(t, "image", payload.Messages[0].Content[1].Type)
	require.NotNil(t, payload.Messages[0].Content[1].Source)
	assert.Equal(t, "base64", payload.Messages[0].Content[1].Source.Type)
	assert.Equal(t, "image/png", payload.Messages[0].Content[1].Source.MediaType)
}

func TestAnthropicGenerate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live Anthropic call in short mode")
	}

	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}

	client, err := NewClient(apiKey, "claude-3-5-sonnet-20241022")
	require.NoError(t, err)

	resp, err := client.Generate(context.Background(), domainllm.GenerateRequest{
		Messages: []domainllm.Message{
			{Role: domainllm.RoleUser, Parts: []domainllm.Part{domainllm.TextPart("Say hello in 5 words")}},
		},
		Temperature: 0.2,
		MaxTokens:   64,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Content)
}
