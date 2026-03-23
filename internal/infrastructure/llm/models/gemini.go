package models

import (
	"context"

	"github.com/google/generative-ai-go/genai"
	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
	"google.golang.org/api/option"
)

// GeminiModelLister lists models from Google Gemini.
type GeminiModelLister struct {
	apiKey string
}

// NewGeminiModelLister creates a new Gemini model lister.
func NewGeminiModelLister(apiKey string) *GeminiModelLister {
	return &GeminiModelLister{apiKey: apiKey}
}

// ListModels lists available Gemini models.
func (l *GeminiModelLister) ListModels(ctx context.Context) ([]domainllm.ModelInfo, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(l.apiKey))
	if err != nil {
		return nil, err
	}
	defer client.Close()

	iter := client.ListModels(ctx)
	var result []domainllm.ModelInfo

	for {
		m, err := iter.Next()
		if err != nil {
			break
		}
		// Filter to generative models only
		if len(m.SupportedGenerationMethods) > 0 {
			result = append(result, domainllm.ModelInfo{
				ID:          m.Name,
				Name:        m.DisplayName,
				Provider:    "gemini",
				Description: m.Description,
			})
		}
	}
	return result, nil
}
