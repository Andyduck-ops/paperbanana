package openai

import (
	"context"
	"os"
	"testing"

	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
	openaisdk "github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientBuildChatCompletionRequestSupportsPromptVersionAndImages(t *testing.T) {
	chatReq, err := buildChatCompletionRequest(domainllm.GenerateRequest{
		SystemInstruction: "You are the planner.",
		PromptVersion:     "planner-v2",
		Model:             "gpt-4o-mini",
		Temperature:       0.3,
		MaxTokens:         128,
		Messages: []domainllm.Message{
			{
				Role: domainllm.RoleUser,
				Parts: []domainllm.Part{
					domainllm.TextPart("Example figure"),
					domainllm.InlineImagePart("image/png", []byte("png-bytes")),
				},
			},
		},
	}, "fallback-model", false)
	require.NoError(t, err)

	require.Len(t, chatReq.Messages, 2)
	assert.Equal(t, "gpt-4o-mini", chatReq.Model)
	assert.Equal(t, float32(0.3), chatReq.Temperature)
	assert.Equal(t, 128, chatReq.MaxTokens)
	assert.Equal(t, "planner-v2", chatReq.Metadata["prompt_version"])
	assert.Equal(t, "You are the planner.", chatReq.Messages[0].Content)
	assert.Len(t, chatReq.Messages[1].MultiContent, 2)
	assert.Equal(t, openaisdk.ChatMessagePartTypeText, chatReq.Messages[1].MultiContent[0].Type)
	assert.Equal(t, "Example figure", chatReq.Messages[1].MultiContent[0].Text)
	assert.Equal(t, openaisdk.ChatMessagePartTypeImageURL, chatReq.Messages[1].MultiContent[1].Type)
	assert.Contains(t, chatReq.Messages[1].MultiContent[1].ImageURL.URL, "data:image/png;base64,")
}

func TestOpenAIGenerate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live OpenAI call in short mode")
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	client, err := NewClient(apiKey, "gpt-4o-mini")
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
