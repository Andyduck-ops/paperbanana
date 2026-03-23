package llm

import (
	"context"
	"fmt"
	"strings"

	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
)

// UnavailableClient lets the server start before a provider is configured.
// Generate requests fail with a clear setup error instead of blocking startup.
type UnavailableClient struct {
	provider string
	reason   string
}

func NewUnavailableClient(provider, reason string) *UnavailableClient {
	return &UnavailableClient{
		provider: provider,
		reason:   strings.TrimSpace(reason),
	}
}

func (c *UnavailableClient) Generate(_ context.Context, _ domainllm.GenerateRequest) (*domainllm.GenerateResponse, error) {
	return nil, fmt.Errorf("%s", c.message())
}

func (c *UnavailableClient) GenerateStream(_ context.Context, _ domainllm.GenerateRequest) (<-chan domainllm.StreamChunk, <-chan error) {
	chunks := make(chan domainllm.StreamChunk)
	errs := make(chan error, 1)

	go func() {
		defer close(chunks)
		defer close(errs)
		errs <- fmt.Errorf("%s", c.message())
	}()

	return chunks, errs
}

func (c *UnavailableClient) Provider() string {
	return c.provider
}

func (c *UnavailableClient) message() string {
	if c.reason != "" {
		return c.reason
	}
	if c.provider == "" {
		return "no LLM provider is configured"
	}
	return fmt.Sprintf("provider %s is not configured", c.provider)
}
