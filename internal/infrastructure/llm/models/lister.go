package models

import (
	"fmt"

	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
)

// Provider that uses OpenAI-compatible API for model listing.
var openAICompatibleProviders = map[string]bool{
	"deepseek": true,
	"zhipu":    true,
	"moonshot": true,
	"qwen":     true,
	"doubao":   true,
	"baichuan": true,
	"minimax":  true,
	"yi":       true,
	"hunyuan":  true,
	"stepfun":  true,
	"silicon":  true,
	"ollama":   true,
}

// GetModelLister returns a model lister for the given provider.
func GetModelLister(provider, apiKey, baseURL string) (domainllm.ModelLister, error) {
	switch provider {
	case "gemini":
		return NewGeminiModelLister(apiKey), nil
	case "openai":
		return NewOpenAIModelLister(apiKey, baseURL), nil
	case "anthropic":
		return nil, fmt.Errorf("anthropic does not support model listing - please enter model ID manually")
	case "openrouter":
		return NewOpenRouterModelLister(apiKey), nil
	default:
		// Check if it's an OpenAI-compatible provider
		if openAICompatibleProviders[provider] {
			return NewOpenAIModelLister(apiKey, baseURL), nil
		}
		// For unknown/custom providers, try OpenAI-compatible
		if baseURL != "" {
			return NewOpenAIModelLister(apiKey, baseURL), nil
		}
		return nil, fmt.Errorf("unknown provider: %s - please enter model ID manually", provider)
	}
}
