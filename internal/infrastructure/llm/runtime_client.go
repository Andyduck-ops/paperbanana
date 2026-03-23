package llm

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	pbconfig "github.com/paperbanana/paperbanana/internal/config"
	domainconfig "github.com/paperbanana/paperbanana/internal/domain/config"
	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
)

type RuntimePurpose string

const (
	RuntimePurposeQuery RuntimePurpose = "query"
	RuntimePurposeGen   RuntimePurpose = "gen"
)

type RuntimeClient struct {
	purpose             RuntimePurpose
	startupProviderName string
	startupConfig       pbconfig.ProviderConfig
	options             ClientOptions
	providerRepo        domainconfig.ProviderRepository
	apiKeyRepo          domainconfig.APIKeyRepository
}

func NewRuntimeClient(
	purpose RuntimePurpose,
	startupProviderName string,
	startupConfig pbconfig.ProviderConfig,
	options ClientOptions,
	providerRepo domainconfig.ProviderRepository,
	apiKeyRepo domainconfig.APIKeyRepository,
) *RuntimeClient {
	return &RuntimeClient{
		purpose:             purpose,
		startupProviderName: startupProviderName,
		startupConfig:       startupConfig,
		options:             options,
		providerRepo:        providerRepo,
		apiKeyRepo:          apiKeyRepo,
	}
}

func (c *RuntimeClient) Generate(ctx context.Context, req domainllm.GenerateRequest) (*domainllm.GenerateResponse, error) {
	client, resolvedReq, err := c.resolveClient(ctx, req)
	if err != nil {
		return nil, err
	}
	return client.Generate(ctx, resolvedReq)
}

func (c *RuntimeClient) GenerateStream(ctx context.Context, req domainllm.GenerateRequest) (<-chan domainllm.StreamChunk, <-chan error) {
	chunks := make(chan domainllm.StreamChunk)
	errs := make(chan error, 1)

	go func() {
		defer close(chunks)
		defer close(errs)

		client, resolvedReq, err := c.resolveClient(ctx, req)
		if err != nil {
			errs <- err
			return
		}

		upstreamChunks, upstreamErrs := client.GenerateStream(ctx, resolvedReq)
		for upstreamChunks != nil || upstreamErrs != nil {
			select {
			case chunk, ok := <-upstreamChunks:
				if !ok {
					upstreamChunks = nil
					continue
				}
				chunks <- chunk
			case err, ok := <-upstreamErrs:
				if !ok {
					upstreamErrs = nil
					continue
				}
				if err != nil {
					errs <- err
				}
			case <-ctx.Done():
				errs <- ctx.Err()
				return
			}
		}
	}()

	return chunks, errs
}

func (c *RuntimeClient) Provider() string {
	return string(c.purpose)
}

func (c *RuntimeClient) resolveClient(ctx context.Context, req domainllm.GenerateRequest) (domainllm.LLMClient, domainllm.GenerateRequest, error) {
	selectedProvider, selectedModel := splitProviderModel(req.Model)

	if strings.TrimSpace(selectedProvider) != "" {
		provider, err := c.lookupProvider(selectedProvider)
		if err != nil {
			return nil, req, err
		}
		model := strings.TrimSpace(selectedModel)
		if model == "" {
			model = c.providerModel(provider)
		}
		client, err := c.buildProviderClient(ctx, provider, model)
		if err != nil {
			return nil, req, err
		}
		req.Model = model
		return client, req, nil
	}

	provider, err := c.resolveDefaultProvider()
	switch {
	case err == nil:
		model := c.providerModel(provider)
		client, buildErr := c.buildProviderClient(ctx, provider, model)
		if buildErr != nil {
			return nil, req, buildErr
		}
		req.Model = model
		return client, req, nil
	case errors.Is(err, errDefaultProviderNotFound):
		client, buildErr := c.buildStartupClient(req.Model)
		if buildErr != nil {
			return nil, req, buildErr
		}
		if strings.TrimSpace(req.Model) == "" {
			req.Model = c.startupConfig.Model
		}
		return client, req, nil
	default:
		return nil, req, err
	}
}

var errDefaultProviderNotFound = errors.New("runtime llm: default provider not found")

func (c *RuntimeClient) resolveDefaultProvider() (*domainconfig.Provider, error) {
	if c.providerRepo == nil {
		return nil, errDefaultProviderNotFound
	}

	provider, err := c.providerRepo.GetDefault()
	if err != nil || provider == nil {
		return nil, errDefaultProviderNotFound
	}
	return provider, nil
}

func (c *RuntimeClient) lookupProvider(value string) (*domainconfig.Provider, error) {
	if c.providerRepo == nil {
		return nil, fmt.Errorf("provider %s not found", value)
	}

	if provider, err := c.providerRepo.GetByID(value); err == nil && provider != nil {
		return provider, nil
	}
	if provider, err := c.providerRepo.GetByName(value); err == nil && provider != nil {
		return provider, nil
	}

	return nil, fmt.Errorf("provider %s not found", value)
}

func (c *RuntimeClient) providerModel(provider *domainconfig.Provider) string {
	if provider == nil {
		return ""
	}
	if c.purpose == RuntimePurposeGen {
		if value := strings.TrimSpace(provider.GenModel); value != "" {
			return value
		}
	}
	if value := strings.TrimSpace(provider.QueryModel); value != "" {
		return value
	}
	return provider.GetDefaultModel()
}

func (c *RuntimeClient) buildProviderClient(ctx context.Context, provider *domainconfig.Provider, model string) (domainllm.LLMClient, error) {
	if provider == nil {
		return nil, errors.New("provider is required")
	}
	if c.apiKeyRepo == nil {
		return nil, fmt.Errorf("provider %s has no key repository configured", provider.Name)
	}

	_, plaintext, err := c.apiKeyRepo.GetNextKey(ctx, provider.ID)
	if err != nil || strings.TrimSpace(plaintext) == "" {
		return nil, fmt.Errorf("provider %s has no active API key configured", provider.Name)
	}

	timeout := time.Duration(provider.TimeoutMs) * time.Millisecond
	if timeout <= 0 {
		timeout = c.startupConfig.Timeout
	}

	providerName := provider.Name
	if provider.Type != "" {
		providerName = string(provider.Type)
	}

	return NewLLMClientWithOptions(providerName, pbconfig.ProviderConfig{
		APIKey:  plaintext,
		BaseURL: provider.APIHost,
		Model:   model,
		Timeout: timeout,
	}, c.options)
}

func (c *RuntimeClient) buildStartupClient(model string) (domainllm.LLMClient, error) {
	if strings.TrimSpace(c.startupConfig.APIKey) == "" {
		return nil, fmt.Errorf(
			"default provider %s has no API key configured; open Settings to add a key before generating",
			c.startupProviderName,
		)
	}

	cfg := c.startupConfig
	if strings.TrimSpace(model) != "" {
		cfg.Model = model
	}
	return NewLLMClientWithOptions(c.startupProviderName, cfg, c.options)
}

func splitProviderModel(value string) (string, string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", ""
	}

	index := strings.Index(value, ":")
	if index <= 0 || index >= len(value)-1 {
		return "", value
	}
	return strings.TrimSpace(value[:index]), strings.TrimSpace(value[index+1:])
}
