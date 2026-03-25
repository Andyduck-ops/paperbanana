package llm

import (
	"fmt"
	"net/http"
	"strings"

	pbconfig "github.com/paperbanana/paperbanana/internal/config"
	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
	"github.com/paperbanana/paperbanana/internal/infrastructure/llm/anthropic"
	"github.com/paperbanana/paperbanana/internal/infrastructure/llm/gemini"
	openaiclient "github.com/paperbanana/paperbanana/internal/infrastructure/llm/openai"
	"github.com/paperbanana/paperbanana/internal/infrastructure/llm/openrouter"
	"github.com/paperbanana/paperbanana/internal/infrastructure/resilience"
)

var openAICompatibleProviders = map[string]bool{
	"deepseek":          true,
	"zhipu":             true,
	"moonshot":          true,
	"qwen":              true,
	"doubao":            true,
	"baichuan":          true,
	"minimax":           true,
	"yi":                true,
	"hunyuan":           true,
	"stepfun":           true,
	"silicon":           true,
	"ollama":            true,
	"grok":              true,
	"xai":               true,
	"openai-compatible": true,
}

type ClientOptions struct {
	HTTPClient *http.Client
	Cache      ResponseCache
}

func NewLLMClient(provider, apiKey, model string) (domainllm.LLMClient, error) {
	return NewLLMClientWithOptions(provider, pbconfig.ProviderConfig{
		APIKey: apiKey,
		Model:  model,
	}, ClientOptions{})
}

func NewLLMClientFromConfig(provider string, cfg pbconfig.ProviderConfig) (domainllm.LLMClient, error) {
	return NewLLMClientWithOptions(provider, cfg, ClientOptions{})
}

func NewLLMClientWithHTTPClient(provider string, cfg pbconfig.ProviderConfig, httpClient *http.Client) (domainllm.LLMClient, error) {
	return NewLLMClientWithOptions(provider, cfg, ClientOptions{HTTPClient: httpClient})
}

func NewLLMClientWithOptions(provider string, cfg pbconfig.ProviderConfig, options ClientOptions) (domainllm.LLMClient, error) {
	httpClient := options.HTTPClient
	if httpClient == nil {
		httpClient = resilience.NewResilientClient(resilientClientName(provider, cfg), cfg.Timeout).HTTPClient()
	}

	client, err := newRawLLMClient(provider, cfg, httpClient)
	if err != nil {
		return nil, err
	}

	if options.Cache != nil {
		return NewCachedClient(provider, cfg.Model, client, options.Cache), nil
	}

	return client, nil
}

func resilientClientName(provider string, cfg pbconfig.ProviderConfig) string {
	parts := []string{"llm"}
	if value := strings.TrimSpace(provider); value != "" {
		parts = append(parts, value)
	}
	if value := strings.TrimSpace(cfg.Model); value != "" {
		parts = append(parts, value)
	}
	if value := strings.TrimSpace(cfg.BaseURL); value != "" {
		parts = append(parts, value)
	}
	return strings.Join(parts, ":")
}

func newRawLLMClient(provider string, cfg pbconfig.ProviderConfig, httpClient *http.Client) (domainllm.LLMClient, error) {
	switch provider {
	case "gemini":
		return gemini.NewClientWithHTTPClient(cfg.APIKey, cfg.Model, httpClient)
	case "openai":
		return openaiclient.NewClientWithConfig(cfg.APIKey, cfg.BaseURL, cfg.Model, cfg.Timeout, httpClient)
	case "anthropic":
		return anthropic.NewClientWithConfig(cfg.APIKey, cfg.BaseURL, cfg.Model, cfg.Timeout, httpClient)
	case "openrouter":
		return openrouter.NewClientWithConfig(cfg.APIKey, cfg.BaseURL, cfg.Model, cfg.Timeout, httpClient)
	default:
		if openAICompatibleProviders[provider] || cfg.BaseURL != "" {
			client, err := openaiclient.NewClientWithConfig(cfg.APIKey, cfg.BaseURL, cfg.Model, cfg.Timeout, httpClient)
			if err != nil {
				return nil, err
			}
			return newProviderAliasClient(provider, client), nil
		}
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}
}
