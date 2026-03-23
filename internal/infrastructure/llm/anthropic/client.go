package anthropic

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
)

const defaultBaseURL = "https://api.anthropic.com/v1"

type Client struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

func NewClient(apiKey, model string) (*Client, error) {
	return NewClientWithConfig(apiKey, "", model, 0, nil)
}

func NewClientWithConfig(apiKey, baseURL, model string, timeout time.Duration, httpClient *http.Client) (*Client, error) {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: timeoutOrDefault(timeout)}
	}

	return &Client{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
		client:  httpClient,
	}, nil
}

type anthropicRequest struct {
	Model       string             `json:"model"`
	System      string             `json:"system,omitempty"`
	Messages    []anthropicMessage `json:"messages"`
	MaxTokens   int                `json:"max_tokens"`
	Temperature float64            `json:"temperature,omitempty"`
	Metadata    map[string]string  `json:"metadata,omitempty"`
}

type anthropicMessage struct {
	Role    string             `json:"role"`
	Content []anthropicContent `json:"content"`
}

type anthropicResponse struct {
	Content      []anthropicContent `json:"content"`
	StopReason   string             `json:"stop_reason"`
	Usage        anthropicUsage     `json:"usage"`
	ErrorMessage string             `json:"error"`
}

type anthropicContent struct {
	Type   string                `json:"type"`
	Text   string                `json:"text,omitempty"`
	Source *anthropicImageSource `json:"source,omitempty"`
}

type anthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type anthropicImageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type,omitempty"`
	Data      string `json:"data,omitempty"`
}

func (c *Client) Generate(ctx context.Context, req domainllm.GenerateRequest) (*domainllm.GenerateResponse, error) {
	payload, err := buildAnthropicRequest(req, c.model)
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal anthropic request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build anthropic request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("anthropic request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("anthropic API error %d: %s", resp.StatusCode, string(raw))
	}

	var decoded anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return nil, fmt.Errorf("decode anthropic response: %w", err)
	}

	parts := buildAnthropicResponseParts(decoded.Content)

	return &domainllm.GenerateResponse{
		Content:      domainllm.CollectText(parts),
		Parts:        parts,
		TokensUsed:   decoded.Usage.InputTokens + decoded.Usage.OutputTokens,
		FinishReason: decoded.StopReason,
	}, nil
}

func (c *Client) GenerateStream(ctx context.Context, req domainllm.GenerateRequest) (<-chan domainllm.StreamChunk, <-chan error) {
	chunks := make(chan domainllm.StreamChunk)
	errs := make(chan error, 1)

	go func() {
		defer close(chunks)
		defer close(errs)

		resp, err := c.Generate(ctx, req)
		if err != nil {
			errs <- err
			return
		}

		chunks <- domainllm.StreamChunk{Content: resp.Content}
		chunks <- domainllm.StreamChunk{Done: true}
	}()

	return chunks, errs
}

func (c *Client) Provider() string {
	return "anthropic"
}

func timeoutOrDefault(timeout time.Duration) time.Duration {
	if timeout > 0 {
		return timeout
	}
	return 60 * time.Second
}

func maxTokensOrDefault(maxTokens int) int {
	if maxTokens > 0 {
		return maxTokens
	}
	return 1024
}

func buildAnthropicRequest(req domainllm.GenerateRequest, defaultModel string) (anthropicRequest, error) {
	if len(req.Messages) == 0 {
		return anthropicRequest{}, errors.New("anthropic request requires at least one message")
	}

	payload := anthropicRequest{
		Model:       domainllm.ResolveModel(req.Model, defaultModel),
		System:      req.SystemInstruction,
		MaxTokens:   maxTokensOrDefault(req.MaxTokens),
		Temperature: req.Temperature,
	}
	if req.PromptVersion != "" {
		payload.Metadata = map[string]string{"prompt_version": req.PromptVersion}
	}

	for _, message := range req.Messages {
		role, err := toAnthropicRole(message.Role)
		if err != nil {
			return anthropicRequest{}, err
		}

		content := make([]anthropicContent, 0, len(message.Parts))
		for _, part := range message.Parts {
			converted, err := toAnthropicContent(part)
			if err != nil {
				return anthropicRequest{}, err
			}
			content = append(content, converted)
		}

		payload.Messages = append(payload.Messages, anthropicMessage{
			Role:    role,
			Content: content,
		})
	}

	return payload, nil
}

func toAnthropicRole(role domainllm.Role) (string, error) {
	switch role {
	case domainllm.RoleUser:
		return "user", nil
	case domainllm.RoleAssistant:
		return "assistant", nil
	default:
		return "", fmt.Errorf("anthropic role %q is not supported", role)
	}
}

func toAnthropicContent(part domainllm.Part) (anthropicContent, error) {
	switch part.Type {
	case domainllm.PartTypeText:
		return anthropicContent{Type: "text", Text: part.Text}, nil
	case domainllm.PartTypeImage:
		if part.URL != "" {
			return anthropicContent{}, errors.New("anthropic image URL parts are not supported")
		}
		if part.MIMEType == "" || len(part.Data) == 0 {
			return anthropicContent{}, errors.New("anthropic image parts require mime type and data")
		}
		return anthropicContent{
			Type: "image",
			Source: &anthropicImageSource{
				Type:      "base64",
				MediaType: part.MIMEType,
				Data:      base64.StdEncoding.EncodeToString(part.Data),
			},
		}, nil
	default:
		return anthropicContent{}, fmt.Errorf("anthropic part type %q is not supported", part.Type)
	}
}

func buildAnthropicResponseParts(parts []anthropicContent) []domainllm.Part {
	var responseParts []domainllm.Part
	for _, part := range parts {
		if part.Type == "text" {
			responseParts = append(responseParts, domainllm.TextPart(part.Text))
		}
	}
	return responseParts
}
