package models

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
)

// OpenRouterModelLister lists models from OpenRouter.
type OpenRouterModelLister struct {
	apiKey string
}

// NewOpenRouterModelLister creates a new OpenRouter model lister.
func NewOpenRouterModelLister(apiKey string) *OpenRouterModelLister {
	return &OpenRouterModelLister{apiKey: apiKey}
}

// ListModels lists available OpenRouter models.
func (l *OpenRouterModelLister) ListModels(ctx context.Context) ([]domainllm.ModelInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://openrouter.ai/api/v1/models", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+l.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("openrouter API error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	models := make([]domainllm.ModelInfo, len(result.Data))
	for i, m := range result.Data {
		models[i] = domainllm.ModelInfo{
			ID:       m.ID,
			Name:     m.Name,
			Provider: "openrouter",
		}
	}
	return models, nil
}
