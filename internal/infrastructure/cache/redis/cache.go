package redis

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	domainllm "github.com/paperbanana/paperbanana/internal/domain/llm"
)

const DefaultTTL = 24 * time.Hour

var ErrCacheMiss = errors.New("llm cache miss")

type Store interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
}

type Cache struct {
	store Store
	ttl   time.Duration
}

type redisStore struct {
	client *goredis.Client
}

type cachedResponse struct {
	Content      string           `json:"content"`
	Parts        []domainllm.Part `json:"parts"`
	TokensUsed   int              `json:"tokens_used"`
	FinishReason string           `json:"finish_reason"`
}

type canonicalRequest struct {
	SystemInstruction string              `json:"system_instruction"`
	Messages          []domainllm.Message `json:"messages"`
	Temperature       float64             `json:"temperature"`
	MaxTokens         int                 `json:"max_tokens"`
	PromptVersion     string              `json:"prompt_version"`
}

func NewCache(store Store) *Cache {
	return &Cache{
		store: store,
		ttl:   DefaultTTL,
	}
}

func NewStore(client *goredis.Client) Store {
	return &redisStore{client: client}
}

func CacheKey(provider, model string, req domainllm.GenerateRequest) (string, error) {
	payload, err := json.Marshal(canonicalRequest{
		SystemInstruction: req.SystemInstruction,
		Messages:          req.Messages,
		Temperature:       req.Temperature,
		MaxTokens:         req.MaxTokens,
		PromptVersion:     req.PromptVersion,
	})
	if err != nil {
		return "", fmt.Errorf("marshal cache key payload: %w", err)
	}

	sum := sha256.Sum256(payload)
	return fmt.Sprintf("llm:cache:%s:%s:%s", provider, model, hex.EncodeToString(sum[:])), nil
}

func (c *Cache) Get(ctx context.Context, provider, model string, req domainllm.GenerateRequest) (*domainllm.GenerateResponse, bool, error) {
	key, err := CacheKey(provider, model, req)
	if err != nil {
		return nil, false, err
	}

	payload, err := c.store.Get(ctx, key)
	if err != nil {
		if errors.Is(err, ErrCacheMiss) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("read cached response: %w", err)
	}

	var cached cachedResponse
	if err := json.Unmarshal(payload, &cached); err != nil {
		return nil, false, fmt.Errorf("decode cached response: %w", err)
	}

	return &domainllm.GenerateResponse{
		Content:      cached.Content,
		Parts:        cached.Parts,
		TokensUsed:   cached.TokensUsed,
		FinishReason: cached.FinishReason,
	}, true, nil
}

func (c *Cache) Set(ctx context.Context, provider, model string, req domainllm.GenerateRequest, resp *domainllm.GenerateResponse) error {
	key, err := CacheKey(provider, model, req)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(cachedResponse{
		Content:      resp.Content,
		Parts:        resp.Parts,
		TokensUsed:   resp.TokensUsed,
		FinishReason: resp.FinishReason,
	})
	if err != nil {
		return fmt.Errorf("encode cached response: %w", err)
	}

	if err := c.store.Set(ctx, key, payload, c.ttl); err != nil {
		return fmt.Errorf("write cached response: %w", err)
	}

	return nil
}

func (s *redisStore) Get(ctx context.Context, key string) ([]byte, error) {
	value, err := s.client.Get(ctx, key).Bytes()
	if errors.Is(err, goredis.Nil) {
		return nil, ErrCacheMiss
	}
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (s *redisStore) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return s.client.Set(ctx, key, value, ttl).Err()
}
