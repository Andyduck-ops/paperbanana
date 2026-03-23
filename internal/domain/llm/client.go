package llm

import (
	"context"
	"strings"
)

// LLMClient defines the contract shared by all provider implementations.
type LLMClient interface {
	Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error)
	GenerateStream(ctx context.Context, req GenerateRequest) (<-chan StreamChunk, <-chan error)
	Provider() string
}

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

type PartType string

const (
	PartTypeText  PartType = "text"
	PartTypeImage PartType = "image"
)

type Message struct {
	Role  Role
	Parts []Part
}

type Part struct {
	Type     PartType
	Text     string
	MIMEType string
	Data     []byte
	URL      string
}

type GenerateRequest struct {
	SystemInstruction string
	Messages          []Message
	Model             string
	Temperature       float64
	MaxTokens         int
	PromptVersion     string
}

type GenerateResponse struct {
	Content      string
	Parts        []Part
	TokensUsed   int
	FinishReason string
}

type StreamChunk struct {
	Content string
	Done    bool
}

// ModelInfo describes a model exposed by a provider listing endpoint.
type ModelInfo struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Provider       string `json:"provider,omitempty"`
	Description    string `json:"description,omitempty"`
	MaxTokens      int    `json:"max_tokens,omitempty"`
	SupportsVision bool   `json:"supports_vision,omitempty"`
}

// ModelLister defines a provider-specific model listing capability.
type ModelLister interface {
	ListModels(ctx context.Context) ([]ModelInfo, error)
}

func ResolveModel(requestModel, defaultModel string) string {
	if requestModel != "" {
		return requestModel
	}
	return defaultModel
}

func TextPart(text string) Part {
	return Part{Type: PartTypeText, Text: text}
}

func InlineImagePart(mimeType string, data []byte) Part {
	return Part{Type: PartTypeImage, MIMEType: mimeType, Data: data}
}

func URLImagePart(mimeType, url string) Part {
	return Part{Type: PartTypeImage, MIMEType: mimeType, URL: url}
}

func CollectText(parts []Part) string {
	var builder strings.Builder
	for _, part := range parts {
		if part.Type == PartTypeText {
			builder.WriteString(part.Text)
		}
	}
	return builder.String()
}
