package llm

import (
	"context"

	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
)

type providerAliasClient struct {
	provider string
	wrapped  domainllm.LLMClient
}

func newProviderAliasClient(provider string, wrapped domainllm.LLMClient) domainllm.LLMClient {
	if wrapped == nil || provider == "" || provider == wrapped.Provider() {
		return wrapped
	}

	return &providerAliasClient{
		provider: provider,
		wrapped:  wrapped,
	}
}

func (c *providerAliasClient) Generate(ctx context.Context, req domainllm.GenerateRequest) (*domainllm.GenerateResponse, error) {
	return c.wrapped.Generate(ctx, req)
}

func (c *providerAliasClient) GenerateImage(ctx context.Context, req domainllm.GenerateRequest) (*domainllm.GenerateResponse, error) {
	imageClient, ok := c.wrapped.(domainllm.ImageGenerator)
	if !ok {
		return c.wrapped.Generate(ctx, req)
	}
	return imageClient.GenerateImage(ctx, req)
}

func (c *providerAliasClient) GenerateStream(ctx context.Context, req domainllm.GenerateRequest) (<-chan domainllm.StreamChunk, <-chan error) {
	return c.wrapped.GenerateStream(ctx, req)
}

func (c *providerAliasClient) Provider() string {
	return c.provider
}
