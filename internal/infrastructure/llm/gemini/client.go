package gemini

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/generative-ai-go/genai"
	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type Client struct {
	client *genai.Client
	model  string
}

func NewClient(apiKey, model string) (*Client, error) {
	return NewClientWithHTTPClient(apiKey, model, nil)
}

func NewClientWithHTTPClient(apiKey, model string, httpClient *http.Client) (*Client, error) {
	clientOptions := []option.ClientOption{option.WithAPIKey(apiKey)}
	if httpClient != nil {
		clientOptions = append(clientOptions, option.WithHTTPClient(httpClient))
	}

	client, err := genai.NewClient(context.Background(), clientOptions...)
	if err != nil {
		return nil, fmt.Errorf("create gemini client: %w", err)
	}

	return &Client{client: client, model: model}, nil
}

func (c *Client) Generate(ctx context.Context, req domainllm.GenerateRequest) (*domainllm.GenerateResponse, error) {
	payload, err := buildGeminiGenerateConfig(req, c.model)
	if err != nil {
		return nil, err
	}

	model := c.client.GenerativeModel(domainllm.ResolveModel(req.Model, c.model))
	model.SetTemperature(float32(req.Temperature))
	if req.MaxTokens > 0 {
		model.SetMaxOutputTokens(int32(req.MaxTokens))
	}
	if payload.SystemInstruction != nil {
		model.SystemInstruction = payload.SystemInstruction
	}

	chat := model.StartChat()
	chat.History = payload.History

	resp, err := chat.SendMessage(ctx, payload.Parts...)
	if err != nil {
		return nil, fmt.Errorf("gemini generation failed: %w", err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return nil, errors.New("gemini response contained no candidates")
	}

	parts := collectGeminiParts(resp.Candidates[0].Content.Parts)

	return &domainllm.GenerateResponse{
		Content:      domainllm.CollectText(parts),
		Parts:        parts,
		TokensUsed:   geminiUsageTokens(resp.UsageMetadata),
		FinishReason: fmt.Sprint(resp.Candidates[0].FinishReason),
	}, nil
}

func (c *Client) GenerateStream(ctx context.Context, req domainllm.GenerateRequest) (<-chan domainllm.StreamChunk, <-chan error) {
	chunks := make(chan domainllm.StreamChunk)
	errs := make(chan error, 1)

	go func() {
		defer close(chunks)
		defer close(errs)

		payload, err := buildGeminiGenerateConfig(req, c.model)
		if err != nil {
			errs <- err
			return
		}

		model := c.client.GenerativeModel(domainllm.ResolveModel(req.Model, c.model))
		model.SetTemperature(float32(req.Temperature))
		if req.MaxTokens > 0 {
			model.SetMaxOutputTokens(int32(req.MaxTokens))
		}
		if payload.SystemInstruction != nil {
			model.SystemInstruction = payload.SystemInstruction
		}

		chat := model.StartChat()
		chat.History = payload.History

		iter := chat.SendMessageStream(ctx, payload.Parts...)
		for {
			resp, err := iter.Next()
			if errors.Is(err, iterator.Done) {
				chunks <- domainllm.StreamChunk{Done: true}
				return
			}
			if err != nil {
				errs <- fmt.Errorf("gemini stream failed: %w", err)
				return
			}
			if resp == nil || len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
				continue
			}

			content := domainllm.CollectText(collectGeminiParts(resp.Candidates[0].Content.Parts))
			if content != "" {
				chunks <- domainllm.StreamChunk{Content: content}
			}
		}
	}()

	return chunks, errs
}

func (c *Client) Provider() string {
	return "gemini"
}

type geminiGenerateConfig struct {
	Model             string
	PromptVersion     string
	SystemInstruction *genai.Content
	History           []*genai.Content
	Parts             []genai.Part
}

func buildGeminiGenerateConfig(req domainllm.GenerateRequest, defaultModel string) (*geminiGenerateConfig, error) {
	if len(req.Messages) == 0 {
		return nil, errors.New("gemini request requires at least one message")
	}

	contents := make([]*genai.Content, 0, len(req.Messages))
	for _, message := range req.Messages {
		content, err := toGeminiContent(message)
		if err != nil {
			return nil, err
		}
		contents = append(contents, content)
	}

	last := contents[len(contents)-1]
	if last.Role != "user" {
		return nil, errors.New("gemini request requires the final message to have user role")
	}

	payload := &geminiGenerateConfig{
		Model:         domainllm.ResolveModel(req.Model, defaultModel),
		PromptVersion: req.PromptVersion,
		History:       contents[:len(contents)-1],
		Parts:         last.Parts,
	}
	if req.SystemInstruction != "" {
		payload.SystemInstruction = genai.NewUserContent(genai.Text(req.SystemInstruction))
	}

	return payload, nil
}

func toGeminiContent(message domainllm.Message) (*genai.Content, error) {
	role, err := toGeminiRole(message.Role)
	if err != nil {
		return nil, err
	}

	parts := make([]genai.Part, 0, len(message.Parts))
	for _, part := range message.Parts {
		converted, err := toGeminiPart(part)
		if err != nil {
			return nil, err
		}
		parts = append(parts, converted)
	}

	return &genai.Content{Role: role, Parts: parts}, nil
}

func toGeminiRole(role domainllm.Role) (string, error) {
	switch role {
	case domainllm.RoleUser:
		return "user", nil
	case domainllm.RoleAssistant:
		return "model", nil
	default:
		return "", fmt.Errorf("gemini role %q is not supported", role)
	}
}

func toGeminiPart(part domainllm.Part) (genai.Part, error) {
	switch part.Type {
	case domainllm.PartTypeText:
		return genai.Text(part.Text), nil
	case domainllm.PartTypeImage:
		if part.URL != "" {
			return nil, errors.New("gemini image URL parts are not supported")
		}
		if part.MIMEType == "" || len(part.Data) == 0 {
			return nil, errors.New("gemini image parts require mime type and data")
		}
		return genai.Blob{MIMEType: part.MIMEType, Data: part.Data}, nil
	default:
		return nil, fmt.Errorf("gemini part type %q is not supported", part.Type)
	}
}

func collectGeminiParts(parts []genai.Part) []domainllm.Part {
	var content []domainllm.Part
	for _, part := range parts {
		switch value := part.(type) {
		case genai.Text:
			content = append(content, domainllm.TextPart(string(value)))
		case genai.Blob:
			content = append(content, domainllm.InlineImagePart(value.MIMEType, value.Data))
		}
	}
	return content
}

func geminiUsageTokens(metadata *genai.UsageMetadata) int {
	if metadata == nil {
		return 0
	}
	return int(metadata.TotalTokenCount)
}
