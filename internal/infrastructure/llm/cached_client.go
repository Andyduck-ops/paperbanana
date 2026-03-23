package llm

import (
	"context"
	"fmt"

	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
)

type ResponseCache interface {
	Get(ctx context.Context, provider, model string, req domainllm.GenerateRequest) (*domainllm.GenerateResponse, bool, error)
	Set(ctx context.Context, provider, model string, req domainllm.GenerateRequest, resp *domainllm.GenerateResponse) error
}

type CachedClient struct {
	provider     string
	defaultModel string
	wrapped      domainllm.LLMClient
	cache        ResponseCache
}

func NewCachedClient(provider, defaultModel string, wrapped domainllm.LLMClient, cache ResponseCache) *CachedClient {
	return &CachedClient{
		provider:     provider,
		defaultModel: defaultModel,
		wrapped:      wrapped,
		cache:        cache,
	}
}

func (c *CachedClient) Generate(ctx context.Context, req domainllm.GenerateRequest) (*domainllm.GenerateResponse, error) {
	model := domainllm.ResolveModel(req.Model, c.defaultModel)

	cached, ok, err := c.cache.Get(ctx, c.provider, model, req)
	if err != nil {
		return nil, fmt.Errorf("read llm cache: %w", err)
	}
	if ok {
		return cached, nil
	}

	resp, err := c.wrapped.Generate(ctx, req)
	if err != nil {
		return nil, err
	}

	if err := c.cache.Set(ctx, c.provider, model, req, resp); err != nil {
		return nil, fmt.Errorf("write llm cache: %w", err)
	}

	return resp, nil
}

func (c *CachedClient) GenerateStream(ctx context.Context, req domainllm.GenerateRequest) (<-chan domainllm.StreamChunk, <-chan error) {
	return c.wrapped.GenerateStream(ctx, req)
}

func (c *CachedClient) Provider() string {
	if c.provider != "" {
		return c.provider
	}
	return c.wrapped.Provider()
}
