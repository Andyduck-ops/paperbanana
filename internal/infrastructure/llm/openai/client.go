package openai

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
	openaisdk "github.com/sashabaranov/go-openai"
)

var inlineImageDataURIRegex = regexp.MustCompile(`data:(image/[a-zA-Z0-9.+-]+);base64,([A-Za-z0-9+/=]+)`)

type Client struct {
	client     *openaisdk.Client
	httpClient openaisdk.HTTPDoer
	model      string
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
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = http.DefaultClient
	}

	return &Client{
		client:     openaisdk.NewClientWithConfig(cfg),
		httpClient: cfg.HTTPClient,
		model:      model,
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

	parts := buildResponseParts(resp.Choices[0].Message)

	return &domainllm.GenerateResponse{
		Content:      domainllm.CollectText(parts),
		Parts:        parts,
		TokensUsed:   resp.Usage.TotalTokens,
		FinishReason: string(resp.Choices[0].FinishReason),
	}, nil
}

func (c *Client) GenerateImage(ctx context.Context, req domainllm.GenerateRequest) (*domainllm.GenerateResponse, error) {
	model := domainllm.ResolveModel(req.Model, c.model)
	if shouldUseChatImageGeneration(model) {
		req.Model = model
		return c.generateImageViaChatCompletion(ctx, req)
	}

	prompt, err := buildImagePrompt(req)
	if err != nil {
		return nil, err
	}

	imageReq := openaisdk.ImageRequest{
		Model:          model,
		Prompt:         prompt,
		ResponseFormat: openaisdk.CreateImageResponseFormatB64JSON,
	}

	resp, err := c.client.CreateImage(ctx, imageReq)
	if err != nil {
		return nil, fmt.Errorf("openai image generation failed: %w", err)
	}
	if len(resp.Data) == 0 {
		return nil, errors.New("openai image response contained no data")
	}

	parts, content, err := c.buildImageResponseParts(ctx, resp)
	if err != nil {
		return nil, err
	}

	return &domainllm.GenerateResponse{
		Content:      content,
		Parts:        parts,
		TokensUsed:   resp.Usage.TotalTokens,
		FinishReason: "image_generation",
	}, nil
}

func shouldUseChatImageGeneration(model string) bool {
	model = strings.ToLower(strings.TrimSpace(model))
	return strings.HasPrefix(model, "gemini-") && strings.Contains(model, "image")
}

func (c *Client) generateImageViaChatCompletion(ctx context.Context, req domainllm.GenerateRequest) (*domainllm.GenerateResponse, error) {
	chatReq, err := buildChatCompletionRequest(req, c.model, false)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.CreateChatCompletion(ctx, chatReq)
	if err != nil {
		return nil, fmt.Errorf("openai chat image generation failed: %w", err)
	}
	if len(resp.Choices) == 0 {
		return nil, errors.New("openai chat image generation returned no choices")
	}

	parts, content, err := buildChatImageResponseParts(resp.Choices[0].Message)
	if err != nil {
		return nil, err
	}
	parts, err = c.resolveImageURLs(ctx, parts)
	if err != nil {
		return nil, err
	}

	return &domainllm.GenerateResponse{
		Content:      content,
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

func buildImagePrompt(req domainllm.GenerateRequest) (string, error) {
	var sections []string
	if system := strings.TrimSpace(req.SystemInstruction); system != "" {
		sections = append(sections, system)
	}

	for _, message := range req.Messages {
		var textParts []string
		for _, part := range message.Parts {
			switch part.Type {
			case domainllm.PartTypeText:
				if text := strings.TrimSpace(part.Text); text != "" {
					textParts = append(textParts, text)
				}
			case domainllm.PartTypeImage:
				return "", errors.New("openai image generation does not support image input in generation mode")
			default:
				return "", fmt.Errorf("openai image generation does not support part type %q", part.Type)
			}
		}
		if len(textParts) == 0 {
			continue
		}

		role := strings.TrimSpace(string(message.Role))
		if role != "" {
			sections = append(sections, strings.ToUpper(role)+":\n"+strings.Join(textParts, "\n\n"))
			continue
		}
		sections = append(sections, strings.Join(textParts, "\n\n"))
	}

	prompt := strings.TrimSpace(strings.Join(sections, "\n\n"))
	if prompt == "" {
		return "", errors.New("openai image generation requires a text prompt")
	}
	return prompt, nil
}

func (c *Client) buildImageResponseParts(ctx context.Context, resp openaisdk.ImageResponse) ([]domainllm.Part, string, error) {
	parts := make([]domainllm.Part, 0, len(resp.Data)*2)
	var content string

	for _, item := range resp.Data {
		if revised := strings.TrimSpace(item.RevisedPrompt); revised != "" {
			if content == "" {
				content = revised
			}
			parts = append(parts, domainllm.TextPart(revised))
		}

		switch {
		case strings.TrimSpace(item.B64JSON) != "":
			decoded, err := base64.StdEncoding.DecodeString(item.B64JSON)
			if err != nil {
				return nil, "", fmt.Errorf("decode openai image payload: %w", err)
			}
			parts = append(parts, domainllm.InlineImagePart(detectImageMIMEType(decoded), decoded))
		case strings.TrimSpace(item.URL) != "":
			mimeType, body, err := c.downloadImage(ctx, item.URL)
			if err != nil {
				return nil, "", err
			}
			parts = append(parts, domainllm.InlineImagePart(mimeType, body))
		}
	}

	if len(parts) == 0 {
		return nil, "", errors.New("openai image response contained no image payload")
	}
	return parts, content, nil
}

func (c *Client) downloadImage(ctx context.Context, url string) (string, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", nil, fmt.Errorf("build image download request: %w", err)
	}

	httpClient := c.httpClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("download openai image payload: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", nil, fmt.Errorf("download openai image payload: unexpected status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("read openai image payload: %w", err)
	}
	if len(body) == 0 {
		return "", nil, errors.New("download openai image payload: empty body")
	}

	mimeType := strings.TrimSpace(resp.Header.Get("Content-Type"))
	if mimeType == "" {
		mimeType = detectImageMIMEType(body)
	}
	return mimeType, body, nil
}

func (c *Client) resolveImageURLs(ctx context.Context, parts []domainllm.Part) ([]domainllm.Part, error) {
	if len(parts) == 0 {
		return nil, nil
	}

	resolved := make([]domainllm.Part, len(parts))
	for i, part := range parts {
		resolved[i] = part
		if part.Type != domainllm.PartTypeImage || len(part.Data) > 0 || strings.TrimSpace(part.URL) == "" {
			continue
		}

		mimeType, body, err := c.downloadImage(ctx, part.URL)
		if err != nil {
			return nil, err
		}
		resolved[i] = domainllm.InlineImagePart(mimeType, body)
	}

	return resolved, nil
}

func detectImageMIMEType(data []byte) string {
	if len(data) == 0 {
		return "image/png"
	}

	mimeType := http.DetectContentType(data)
	if strings.HasPrefix(mimeType, "image/") {
		return mimeType
	}
	return "image/png"
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

func buildResponseParts(message openaisdk.ChatCompletionMessage) []domainllm.Part {
	var parts []domainllm.Part

	if content := strings.TrimSpace(message.Content); content != "" {
		parts = append(parts, domainllm.TextPart(content))
	}

	for _, item := range message.MultiContent {
		if item.Type != openaisdk.ChatMessagePartTypeText {
			continue
		}
		if text := strings.TrimSpace(item.Text); text != "" {
			parts = append(parts, domainllm.TextPart(text))
		}
	}

	if len(parts) == 0 {
		if reasoning := strings.TrimSpace(message.ReasoningContent); reasoning != "" {
			parts = append(parts, domainllm.TextPart(reasoning))
		}
	}

	if len(parts) == 0 {
		if refusal := strings.TrimSpace(message.Refusal); refusal != "" {
			parts = append(parts, domainllm.TextPart(refusal))
		}
	}

	return parts
}

func buildChatImageResponseParts(message openaisdk.ChatCompletionMessage) ([]domainllm.Part, string, error) {
	textSegments := make([]string, 0, 1+len(message.MultiContent))
	parts := make([]domainllm.Part, 0, len(message.MultiContent)+1)
	if content := strings.TrimSpace(message.Content); content != "" {
		textSegments = append(textSegments, content)
	}
	for _, item := range message.MultiContent {
		switch item.Type {
		case openaisdk.ChatMessagePartTypeText:
			if text := strings.TrimSpace(item.Text); text != "" {
				textSegments = append(textSegments, text)
			}
		case openaisdk.ChatMessagePartTypeImageURL:
			imageParts, err := imagePartsFromURL(item.ImageURL)
			if err != nil {
				return nil, "", err
			}
			parts = append(parts, imageParts...)
		}
	}

	contentParts := make([]string, 0, len(textSegments))
	for _, text := range textSegments {
		imageParts, sanitized, err := extractInlineImagesFromText(text)
		if err != nil {
			return nil, "", err
		}
		if sanitized != "" {
			contentParts = append(contentParts, sanitized)
			parts = append(parts, domainllm.TextPart(sanitized))
		}
		parts = append(parts, imageParts...)
	}

	if len(parts) == 0 {
		return nil, "", errors.New("openai chat image generation returned no image payload")
	}

	content := strings.TrimSpace(strings.Join(contentParts, "\n"))
	if content == "" {
		content = "generated image via chat completion"
	}
	return parts, content, nil
}

func imagePartsFromURL(imageURL *openaisdk.ChatMessageImageURL) ([]domainllm.Part, error) {
	if imageURL == nil || strings.TrimSpace(imageURL.URL) == "" {
		return nil, nil
	}

	if matches := inlineImageDataURIRegex.FindStringSubmatch(imageURL.URL); len(matches) == 3 {
		decoded, err := base64.StdEncoding.DecodeString(matches[2])
		if err != nil {
			return nil, fmt.Errorf("decode chat image payload: %w", err)
		}
		return []domainllm.Part{domainllm.InlineImagePart(matches[1], decoded)}, nil
	}

	return []domainllm.Part{domainllm.URLImagePart("", imageURL.URL)}, nil
}

func extractInlineImagesFromText(text string) ([]domainllm.Part, string, error) {
	matches := inlineImageDataURIRegex.FindAllStringSubmatch(text, -1)
	if len(matches) == 0 {
		return nil, strings.TrimSpace(text), nil
	}

	parts := make([]domainllm.Part, 0, len(matches))
	for _, match := range matches {
		decoded, err := base64.StdEncoding.DecodeString(match[2])
		if err != nil {
			return nil, "", fmt.Errorf("decode chat image payload: %w", err)
		}
		parts = append(parts, domainllm.InlineImagePart(match[1], decoded))
	}

	sanitized := inlineImageDataURIRegex.ReplaceAllString(text, "[inline-image]")
	sanitized = strings.ReplaceAll(sanitized, "![image]([inline-image])", "[generated image]")
	sanitized = strings.TrimSpace(sanitized)
	return parts, sanitized, nil
}
