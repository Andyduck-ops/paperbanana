package models

import (
	"context"
	"strings"

	"github.com/sashabaranov/go-openai"
	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
)

// OpenAIModelLister lists models from OpenAI.
type OpenAIModelLister struct {
	client *openai.Client
}

// NewOpenAIModelLister creates a new OpenAI model lister.
func NewOpenAIModelLister(apiKey, baseURL string) *OpenAIModelLister {
	config := openai.DefaultConfig(apiKey)
	if baseURL != "" {
		config.BaseURL = baseURL
	}
	return &OpenAIModelLister{client: openai.NewClientWithConfig(config)}
}

// ListModels lists available OpenAI models.
func (l *OpenAIModelLister) ListModels(ctx context.Context) ([]domainllm.ModelInfo, error) {
	models, err := l.client.ListModels(ctx)
	if err != nil {
		return nil, err
	}

	var result []domainllm.ModelInfo
	for _, m := range models.Models {
		// Filter to chat models
		if isChatModel(m.ID) {
			result = append(result, domainllm.ModelInfo{
				ID:       m.ID,
				Name:     m.ID,
				Provider: "openai",
			})
		}
	}
	return result, nil
}

func isChatModel(id string) bool {
	// Filter to GPT models
	return strings.HasPrefix(id, "gpt") || strings.HasPrefix(id, "chatgpt") || strings.HasPrefix(id, "o1") || strings.HasPrefix(id, "o3")
}
