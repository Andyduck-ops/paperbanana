package openrouter

import (
	"context"
	"os"
	"testing"

	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientBuildChatCompletionRequestSupportsPromptVersionAndImages(t *testing.T) {
	chatReq, err := buildChatCompletionRequest(domainllm.GenerateRequest{
		SystemInstruction: "You are the critic.",
		PromptVersion:     "critic-v1",
		Model:             "anthropic/claude-3.5-sonnet",
		Temperature:       0.1,
		MaxTokens:         256,
		Messages: []domainllm.Message{
			{
				Role: domainllm.RoleUser,
				Parts: []domainllm.Part{
					domainllm.TextPart("Critique this figure."),
					domainllm.InlineImagePart("image/jpeg", []byte("jpg-bytes")),
				},
			},
		},
	}, "fallback-model", false)
	require.NoError(t, err)

	require.Len(t, chatReq.Messages, 2)
	assert.Equal(t, "critic-v1", chatReq.Metadata["prompt_version"])
	assert.Equal(t, "You are the critic.", chatReq.Messages[0].Content)
	assert.Len(t, chatReq.Messages[1].MultiContent, 2)
	assert.Equal(t, "Critique this figure.", chatReq.Messages[1].MultiContent[0].Text)
	assert.Contains(t, chatReq.Messages[1].MultiContent[1].ImageURL.URL, "data:image/jpeg;base64,")
}

func TestOpenRouterGenerate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live OpenRouter call in short mode")
	}

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		t.Skip("OPENROUTER_API_KEY not set")
	}

	client, err := NewClient(apiKey, "anthropic/claude-3.5-sonnet")
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
