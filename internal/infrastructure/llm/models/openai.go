package models

import (
	"context"
	"sort"

	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
	"github.com/sashabaranov/go-openai"
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

	result := make([]domainllm.ModelInfo, 0, len(models.Models))
	for _, m := range models.Models {
		if m.ID == "" {
			continue
		}
		result = append(result, domainllm.ModelInfo{
			ID:       m.ID,
			Name:     m.ID,
			Provider: "openai",
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})

	return result, nil
}
