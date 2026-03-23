package llm

import (
	"fmt"
	"net/http"

	pbconfig "github.com/paperbanana/paperbanana/internal/config"
	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
	"github.com/paperbanana/paperbanana/internal/infrastructure/llm/anthropic"
	"github.com/paperbanana/paperbanana/internal/infrastructure/llm/gemini"
	openaiclient "github.com/paperbanana/paperbanana/internal/infrastructure/llm/openai"
	"github.com/paperbanana/paperbanana/internal/infrastructure/llm/openrouter"
)

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
	client, err := newRawLLMClient(provider, cfg, options.HTTPClient)
	if err != nil {
		return nil, err
	}

	if options.Cache != nil {
		return NewCachedClient(provider, cfg.Model, client, options.Cache), nil
	}

	return client, nil
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
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}
}
