package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aiservice/internal/models"
	"github.com/aiservice/internal/providers"
	"github.com/firebase/genkit/go/ai"
)

// CachedLLMClient wraps an LLMClient with caching capabilities
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

func (c *CachedLLMClient) Summarize(ctx context.Context, parts []*ai.Part) (models.SummarizeResponse, error) {
	// Generate cache key from input parts
	cacheKey, err := c.generateCacheKey("summarize", parts)
	if err != nil {
		// If we can't generate a key, proceed without caching
		return c.client.Summarize(ctx, parts)
	}
	
	// Try to get from cache first
	if cachedValue, found := c.cache.Get(cacheKey); found {
		if response, ok := cachedValue.(models.SummarizeResponse); ok {
			return response, nil
		}
	}
	
	// Call the underlying client
	response, err := c.client.Summarize(ctx, parts)
	if err != nil {
		return models.SummarizeResponse{}, err
	}
	
	// Cache the result
	c.cache.Set(cacheKey, response, 1*time.Hour) // Cache for 1 hour
	
	return response, nil
}

func (c *CachedLLMClient) Structurize(ctx context.Context, parts []*ai.Part) (models.StructurizeResponse, error) {
	// Generate cache key from input parts
	cacheKey, err := c.generateCacheKey("structurize", parts)
	if err != nil {
		// If we can't generate a key, proceed without caching
		return c.client.Structurize(ctx, parts)
	}
	
	// Try to get from cache first
	if cachedValue, found := c.cache.Get(cacheKey); found {
		if response, ok := cachedValue.(models.StructurizeResponse); ok {
			return response, nil
		}
	}
	
	// Call the underlying client
	response, err := c.client.Structurize(ctx, parts)
	if err != nil {
		return models.StructurizeResponse{}, err
	}
	
	// Cache the result
	c.cache.Set(cacheKey, response, 1*time.Hour) // Cache for 1 hour
	
	return response, nil
}

// generateCacheKey creates a unique key based on the operation type and input parts
func (c *CachedLLMClient) generateCacheKey(operation string, parts []*ai.Part) (string, error) {
	// Convert parts to a comparable representation for hashing
	var partStrings []string
	for _, part := range parts {
		if part.IsText() {
			partStrings = append(partStrings, "text:"+part.Text)
		} else if part.IsMedia() || part.ContentType != "" {
			partStrings = append(partStrings, "media:"+part.ContentType+":"+part.Text)
		} else if part.Kind != 0 {
			// For other types, serialize the whole part
			serialized, err := json.Marshal(part)
			if err != nil {
				return "", fmt.Errorf("failed to serialize part: %w", err)
			}
			partStrings = append(partStrings, string(serialized))
		}
	}

	// Serialize the parts to create a consistent string representation
	partsBytes, err := json.Marshal(partStrings)
	if err != nil {
		return "", fmt.Errorf("failed to serialize parts: %w", err)
	}

	// Create a cache key combining operation and parts
	key := GenerateKey(operation, string(partsBytes))

	return fmt.Sprintf("llm:%s", key), nil
}

func (c *CachedLLMClient) GetName() string {
	return c.client.GetName()
}