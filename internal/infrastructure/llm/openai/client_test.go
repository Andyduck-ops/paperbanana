package openai

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func TestOpenAIGenerateImageUsesImagesEndpointAndDecodesBase64(t *testing.T) {
	t.Parallel()

	var captured openaisdk.ImageRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/images/generations", r.URL.Path)

		require.NoError(t, json.NewDecoder(r.Body).Decode(&captured))
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(openaisdk.ImageResponse{
			Data: []openaisdk.ImageResponseDataInner{{
				B64JSON:       base64.StdEncoding.EncodeToString([]byte("png-image-bytes")),
				RevisedPrompt: "clean academic diagram",
			}},
			Usage: openaisdk.ImageResponseUsage{TotalTokens: 42},
		}))
	}))
	defer server.Close()

	client, err := NewClientWithConfig("test-key", server.URL, "gpt-image-1", 0, server.Client())
	require.NoError(t, err)

	resp, err := client.GenerateImage(context.Background(), domainllm.GenerateRequest{
		SystemInstruction: "Render a paper figure.",
		Model:             "gpt-image-1",
		Messages: []domainllm.Message{
			{Role: domainllm.RoleUser, Parts: []domainllm.Part{domainllm.TextPart("Create a clean system diagram.")}},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, "gpt-image-1", captured.Model)
	assert.Equal(t, openaisdk.CreateImageResponseFormatB64JSON, captured.ResponseFormat)
	assert.Contains(t, captured.Prompt, "Render a paper figure.")
	assert.Contains(t, captured.Prompt, "Create a clean system diagram.")
	assert.Equal(t, "clean academic diagram", resp.Content)
	assert.Equal(t, 42, resp.TokensUsed)
	require.Len(t, resp.Parts, 2)
	assert.Equal(t, domainllm.PartTypeText, resp.Parts[0].Type)
	assert.Equal(t, domainllm.PartTypeImage, resp.Parts[1].Type)
	assert.Equal(t, []byte("png-image-bytes"), resp.Parts[1].Data)
}

func TestOpenAIGenerateImageFallsBackToChatCompletionsForGeminiImageModels(t *testing.T) {
	t.Parallel()

	var chatCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/chat/completions":
			chatCalls++
			assert.Equal(t, http.MethodPost, r.Method)
			w.Header().Set("Content-Type", "application/json")
			require.NoError(t, json.NewEncoder(w).Encode(openaisdk.ChatCompletionResponse{
				Choices: []openaisdk.ChatCompletionChoice{{
					FinishReason: openaisdk.FinishReasonStop,
					Message: openaisdk.ChatCompletionMessage{
						Role:    openaisdk.ChatMessageRoleAssistant,
						Content: "![image](data:image/png;base64," + base64.StdEncoding.EncodeToString([]byte("chat-image-bytes")) + ")",
					},
				}},
				Usage: openaisdk.Usage{TotalTokens: 24},
			}))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client, err := NewClientWithConfig("test-key", server.URL, "gemini-3.1-flash-image-preview", 0, server.Client())
	require.NoError(t, err)

	resp, err := client.GenerateImage(context.Background(), domainllm.GenerateRequest{
		SystemInstruction: "Render a paper figure.",
		Model:             "gemini-3.1-flash-image-preview",
		Messages: []domainllm.Message{
			{Role: domainllm.RoleUser, Parts: []domainllm.Part{domainllm.TextPart("Create a clean system diagram.")}},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, 1, chatCalls)
	assert.Equal(t, "[generated image]", resp.Content)
	assert.Equal(t, 24, resp.TokensUsed)
	require.Len(t, resp.Parts, 2)
	assert.Equal(t, domainllm.PartTypeText, resp.Parts[0].Type)
	assert.Equal(t, "[generated image]", resp.Parts[0].Text)
	assert.Equal(t, domainllm.PartTypeImage, resp.Parts[1].Type)
	assert.Equal(t, []byte("chat-image-bytes"), resp.Parts[1].Data)
}

func TestBuildResponsePartsUsesMultiContentAndReasoningFallbacks(t *testing.T) {
	t.Run("multi content text", func(t *testing.T) {
		parts := buildResponseParts(openaisdk.ChatCompletionMessage{
			MultiContent: []openaisdk.ChatMessagePart{
				{Type: openaisdk.ChatMessagePartTypeText, Text: "planner output"},
			},
		})

		require.Len(t, parts, 1)
		assert.Equal(t, "planner output", parts[0].Text)
	})

	t.Run("reasoning fallback", func(t *testing.T) {
		parts := buildResponseParts(openaisdk.ChatCompletionMessage{
			ReasoningContent: "fallback reasoning content",
		})

		require.Len(t, parts, 1)
		assert.Equal(t, "fallback reasoning content", parts[0].Text)
	})
}
