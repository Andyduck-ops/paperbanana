package openai

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
	openaisdk "github.com/sashabaranov/go-openai"
)

type Client struct {
	client *openaisdk.Client
	model  string
}

func NewClient(apiKey, model string) (*Client, error) {
	return NewClientWithConfig(apiKey, "", model, 0, nil)
}

func NewClientWithConfig(apiKey, baseURL, model string, timeout time.Duration, httpClient *http.Client) (*Client, error) {
	cfg := openaisdk.DefaultConfig(apiKey)
	if baseURL != "" {
		cfg.BaseURL = baseURL
	}
	if httpClient != nil {
		cfg.HTTPClient = httpClient
	} else if timeout > 0 {
		cfg.HTTPClient = &http.Client{Timeout: timeout}
	}

	return &Client{
		client: openaisdk.NewClientWithConfig(cfg),
		model:  model,
	}, nil
}

func (c *Client) Generate(ctx context.Context, req domainllm.GenerateRequest) (*domainllm.GenerateResponse, error) {
	chatReq, err := buildChatCompletionRequest(req, c.model, false)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.CreateChatCompletion(ctx, chatReq)
	if err != nil {
		return nil, fmt.Errorf("openai generation failed: %w", err)
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("openai response contained no choices")
	}

	parts := buildResponseParts(resp.Choices[0].Message.Content)

	return &domainllm.GenerateResponse{
		Content:      domainllm.CollectText(parts),
		Parts:        parts,
		TokensUsed:   resp.Usage.TotalTokens,
		FinishReason: string(resp.Choices[0].FinishReason),
	}, nil
}

func (c *Client) GenerateStream(ctx context.Context, req domainllm.GenerateRequest) (<-chan domainllm.StreamChunk, <-chan error) {
	chunks := make(chan domainllm.StreamChunk)
	errs := make(chan error, 1)

	go func() {
		defer close(chunks)
		defer close(errs)

		chatReq, err := buildChatCompletionRequest(req, c.model, true)
		if err != nil {
			errs <- err
			return
		}

		stream, err := c.client.CreateChatCompletionStream(ctx, chatReq)
		if err != nil {
			errs <- fmt.Errorf("openai stream failed: %w", err)
			return
		}
		defer stream.Close()

		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				chunks <- domainllm.StreamChunk{Done: true}
				return
			}
			if err != nil {
				errs <- err
				return
			}
			if len(resp.Choices) == 0 {
				continue
			}

			content := resp.Choices[0].Delta.Content
			if content != "" {
				chunks <- domainllm.StreamChunk{Content: content}
			}
		}
	}()

	return chunks, errs
}

func (c *Client) Provider() string {
	return "openai"
}

func buildChatCompletionRequest(req domainllm.GenerateRequest, defaultModel string, stream bool) (openaisdk.ChatCompletionRequest, error) {
	if len(req.Messages) == 0 {
		return openaisdk.ChatCompletionRequest{}, errors.New("openai request requires at least one message")
	}

	chatReq := openaisdk.ChatCompletionRequest{
		Model:       openaisdk.GPT4oMini,
		Temperature: float32(req.Temperature),
		Stream:      stream,
	}
	if model := domainllm.ResolveModel(req.Model, defaultModel); model != "" {
		chatReq.Model = model
	}
	if req.MaxTokens > 0 {
		chatReq.MaxTokens = req.MaxTokens
	}
	if req.PromptVersion != "" {
		chatReq.Metadata = map[string]string{"prompt_version": req.PromptVersion}
	}
	if req.SystemInstruction != "" {
		chatReq.Messages = append(chatReq.Messages, openaisdk.ChatCompletionMessage{
			Role:    openaisdk.ChatMessageRoleSystem,
			Content: req.SystemInstruction,
		})
	}

	for _, message := range req.Messages {
		converted, err := toOpenAIMessage(message)
		if err != nil {
			return openaisdk.ChatCompletionRequest{}, err
		}
		chatReq.Messages = append(chatReq.Messages, converted)
	}

	return chatReq, nil
}

func toOpenAIMessage(message domainllm.Message) (openaisdk.ChatCompletionMessage, error) {
	role, err := toOpenAIRole(message.Role)
	if err != nil {
		return openaisdk.ChatCompletionMessage{}, err
	}

	converted := openaisdk.ChatCompletionMessage{Role: role}
	if len(message.Parts) == 0 {
		return converted, nil
	}

	for _, part := range message.Parts {
		switch part.Type {
		case domainllm.PartTypeText:
			converted.MultiContent = append(converted.MultiContent, openaisdk.ChatMessagePart{
				Type: openaisdk.ChatMessagePartTypeText,
				Text: part.Text,
			})
		case domainllm.PartTypeImage:
			imageURL, err := toOpenAIImageURL(part)
			if err != nil {
				return openaisdk.ChatCompletionMessage{}, err
			}
			converted.MultiContent = append(converted.MultiContent, openaisdk.ChatMessagePart{
				Type:     openaisdk.ChatMessagePartTypeImageURL,
				ImageURL: imageURL,
			})
		default:
			return openaisdk.ChatCompletionMessage{}, fmt.Errorf("openai message part type %q is not supported", part.Type)
		}
	}

	if len(converted.MultiContent) == 1 && converted.MultiContent[0].Type == openaisdk.ChatMessagePartTypeText {
		converted.Content = converted.MultiContent[0].Text
		converted.MultiContent = nil
	}

	return converted, nil
}

func toOpenAIRole(role domainllm.Role) (string, error) {
	switch role {
	case domainllm.RoleUser:
		return openaisdk.ChatMessageRoleUser, nil
	case domainllm.RoleAssistant:
		return openaisdk.ChatMessageRoleAssistant, nil
	default:
		return "", fmt.Errorf("openai role %q is not supported", role)
	}
}

func toOpenAIImageURL(part domainllm.Part) (*openaisdk.ChatMessageImageURL, error) {
	if part.URL != "" {
		return &openaisdk.ChatMessageImageURL{URL: part.URL}, nil
	}
	if len(part.Data) == 0 || part.MIMEType == "" {
		return nil, errors.New("openai image parts require mime type and data")
	}

	return &openaisdk.ChatMessageImageURL{
		URL: fmt.Sprintf("data:%s;base64,%s", part.MIMEType, base64.StdEncoding.EncodeToString(part.Data)),
	}, nil
}

func buildResponseParts(content string) []domainllm.Part {
	if content == "" {
		return nil
	}
	return []domainllm.Part{domainllm.TextPart(content)}
}
