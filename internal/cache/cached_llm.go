package cache

import (
	"context"

	"github.com/aiservice/internal/providers"
	"github.com/firebase/genkit/go/ai"
)

// CachedLLMClient wraps an LLMClient with caching capabilities
// Note: Caching is currently disabled, this is a placeholder for future implementation
type CachedLLMClient struct {
	client providers.LLMClient
	cache  Cache
}

// NewCachedLLMClient creates a new cached LLM client wrapper
func NewCachedLLMClient(client providers.LLMClient, cache Cache) *CachedLLMClient {
	return &CachedLLMClient{
		client: client,
		cache:  cache,
	}
}

func (c *CachedLLMClient) Summarize(ctx context.Context, parts []*ai.Part) (providers.SummarizeFlow, error) {
	// Caching disabled - pass through to underlying client
	return c.client.Summarize(ctx, parts)
}

func (c *CachedLLMClient) Structurize(ctx context.Context, parts []*ai.Part) (providers.StructurizeFlow, error) {
	// Caching disabled - pass through to underlying client
	return c.client.Structurize(ctx, parts)
}

func (c *CachedLLMClient) GetName() string {
	return c.client.GetName()
}
