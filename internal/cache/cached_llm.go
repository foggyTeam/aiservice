package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aiservice/internal/providers"
	"github.com/firebase/genkit/go/ai"
)

const (
	// DefaultTTL for LLM cache entries
	DefaultTTL = 30 * time.Minute

	// MaxCacheSize in bytes (1GB)
	MaxCacheSize = 1 << 30 // 1024 * 1024 * 1024 bytes
)

// CachedLLMClient wraps an LLMClient with caching capabilities
type CachedLLMClient struct {
	client      providers.LLMClient
	cache       Cache
	mu          sync.RWMutex
	currentSize int64 // Current cache size in bytes (atomic)
	hits        int64 // Cache hits counter (atomic)
	misses      int64 // Cache misses counter (atomic)
}

// cacheEntry represents a cached LLM response with size tracking
type cacheEntry struct {
	Size      int64
	Timestamp time.Time
}

// NewCachedLLMClient creates a new cached LLM client wrapper
func NewCachedLLMClient(client providers.LLMClient, cache Cache) *CachedLLMClient {
	return &CachedLLMClient{
		client: client,
		cache:  cache,
	}
}

// Summarize with caching
func (c *CachedLLMClient) Summarize(ctx context.Context, parts []*ai.Part) (providers.SummarizeFlow, error) {
	// Generate cache key
	cacheKey, err := c.generateCacheKey("summarize", parts)
	if err != nil {
		slog.Warn("LLM cache key generation failed, bypassing cache", "err", err)
		return c.client.Summarize(ctx, parts)
	}

	// Try to get from cache
	c.mu.RLock()
	if cachedValue, found := c.cache.Get(cacheKey); found {
		c.mu.RUnlock()
		if response, ok := cachedValue.(providers.SummarizeFlow); ok {
			atomic.AddInt64(&c.hits, 1)
			slog.Info("LLM cache HIT", "operation", "summarize", "key", c.shortKey(cacheKey))
			return response, nil
		}
		c.mu.RLock()
	}
	c.mu.RUnlock()

	// Cache miss - call the underlying client
	atomic.AddInt64(&c.misses, 1)
	slog.Info("LLM cache MISS", "operation", "summarize", "key", c.shortKey(cacheKey))

	response, err := c.client.Summarize(ctx, parts)
	if err != nil {
		return providers.SummarizeFlow{}, err
	}

	// Cache the result with size tracking
	entrySize := c.estimateSize(response)
	c.mu.Lock()
	if c.currentSize+entrySize <= MaxCacheSize {
		c.cache.Set(cacheKey, response, DefaultTTL)
		c.currentSize += entrySize
		slog.Debug("LLM response cached", "key", c.shortKey(cacheKey), "size", entrySize, "totalSize", c.currentSize)
	} else {
		slog.Debug("LLM cache full, skipping cache", "currentSize", c.currentSize, "entrySize", entrySize, "maxSize", MaxCacheSize)
	}
	c.mu.Unlock()

	return response, nil
}

// Structurize with caching
func (c *CachedLLMClient) Structurize(ctx context.Context, parts []*ai.Part) (providers.StructurizeFlow, error) {
	// Generate cache key
	cacheKey, err := c.generateCacheKey("structurize", parts)
	if err != nil {
		slog.Warn("LLM cache key generation failed, bypassing cache", "err", err)
		return c.client.Structurize(ctx, parts)
	}

	// Try to get from cache
	c.mu.RLock()
	if cachedValue, found := c.cache.Get(cacheKey); found {
		c.mu.RUnlock()
		if response, ok := cachedValue.(providers.StructurizeFlow); ok {
			atomic.AddInt64(&c.hits, 1)
			slog.Info("LLM cache HIT", "operation", "structurize", "key", c.shortKey(cacheKey))
			return response, nil
		}
		c.mu.RLock()
	}
	c.mu.RUnlock()

	// Cache miss - call the underlying client
	atomic.AddInt64(&c.misses, 1)
	slog.Info("LLM cache MISS", "operation", "structurize", "key", c.shortKey(cacheKey))

	response, err := c.client.Structurize(ctx, parts)
	if err != nil {
		return providers.StructurizeFlow{}, err
	}

	// Cache the result with size tracking
	entrySize := c.estimateSize(response)
	c.mu.Lock()
	if c.currentSize+entrySize <= MaxCacheSize {
		c.cache.Set(cacheKey, response, DefaultTTL)
		c.currentSize += entrySize
		slog.Debug("LLM response cached", "key", c.shortKey(cacheKey), "size", entrySize, "totalSize", c.currentSize)
	} else {
		slog.Debug("LLM cache full, skipping cache", "currentSize", c.currentSize, "entrySize", entrySize, "maxSize", MaxCacheSize)
	}
	c.mu.Unlock()

	return response, nil
}

// GenerateTemplate with caching
func (c *CachedLLMClient) GenerateTemplate(ctx context.Context, parts []*ai.Part) (providers.TemplateGenerationFlow, error) {
	// Generate cache key
	cacheKey, err := c.generateCacheKey("generateTemplate", parts)
	if err != nil {
		slog.Warn("LLM cache key generation failed, bypassing cache", "err", err)
		return c.client.GenerateTemplate(ctx, parts)
	}

	// Try to get from cache
	c.mu.RLock()
	if cachedValue, found := c.cache.Get(cacheKey); found {
		c.mu.RUnlock()
		if response, ok := cachedValue.(providers.TemplateGenerationFlow); ok {
			atomic.AddInt64(&c.hits, 1)
			slog.Info("LLM cache HIT", "operation", "generateTemplate", "key", c.shortKey(cacheKey))
			return response, nil
		}
		c.mu.RLock()
	}
	c.mu.RUnlock()

	// Cache miss - call the underlying client
	atomic.AddInt64(&c.misses, 1)
	slog.Info("LLM cache MISS", "operation", "generateTemplate", "key", c.shortKey(cacheKey))

	response, err := c.client.GenerateTemplate(ctx, parts)
	if err != nil {
		return providers.TemplateGenerationFlow{}, err
	}

	// Cache the result with size tracking
	entrySize := c.estimateSize(response)
	c.mu.Lock()
	if c.currentSize+entrySize <= MaxCacheSize {
		c.cache.Set(cacheKey, response, DefaultTTL)
		c.currentSize += entrySize
		slog.Debug("LLM response cached", "key", c.shortKey(cacheKey), "size", entrySize, "totalSize", c.currentSize)
	} else {
		slog.Debug("LLM cache full, skipping cache", "currentSize", c.currentSize, "entrySize", entrySize, "maxSize", MaxCacheSize)
	}
	c.mu.Unlock()

	return response, nil
}

func (c *CachedLLMClient) GenerateText(ctx context.Context, parts []*ai.Part) (string, error) {
	// Generate cache key
	cacheKey, err := c.generateCacheKey("generateText", parts)
	if err != nil {
		slog.Warn("LLM cache key generation failed, bypassing cache", "err", err)
		return c.client.GenerateText(ctx, parts)
	}

	// Try to get from cache
	c.mu.RLock()
	if cachedValue, found := c.cache.Get(cacheKey); found {
		c.mu.RUnlock()
		if response, ok := cachedValue.(string); ok {
			atomic.AddInt64(&c.hits, 1)
			slog.Info("LLM cache HIT", "operation", "generateText", "key", c.shortKey(cacheKey))
			return response, nil
		}
		c.mu.RLock()
	}
	c.mu.RUnlock()

	// Cache miss - call the underlying client
	atomic.AddInt64(&c.misses, 1)
	slog.Info("LLM cache MISS", "operation", "generateText", "key", c.shortKey(cacheKey))

	response, err := c.client.GenerateText(ctx, parts)
	if err != nil {
		return "", err
	}

	// Cache the result with size tracking
	entrySize := c.estimateSize(response)
	c.mu.Lock()
	if c.currentSize+entrySize <= MaxCacheSize {
		c.cache.Set(cacheKey, response, DefaultTTL)
		c.currentSize += entrySize
		slog.Debug("LLM response cached", "key", c.shortKey(cacheKey), "size", entrySize, "totalSize", c.currentSize)
	} else {
		slog.Debug("LLM cache full, skipping cache", "currentSize", c.currentSize, "entrySize", entrySize, "maxSize", MaxCacheSize)
	}
	c.mu.Unlock()

	return response, nil
}

// GetName returns the underlying client name
func (c *CachedLLMClient) GetName() string {
	return c.client.GetName()
}

// ImageRecognition with caching (not cached - always calls underlying client)
func (c *CachedLLMClient) ImageRecognition(ctx context.Context, parts []*ai.Part) (providers.ImageRecognitionFlow, error) {
	// Image recognition is not cached as images are typically unique
	return c.client.ImageRecognition(ctx, parts)
}

// GetStats returns cache statistics
func (c *CachedLLMClient) GetStats() (hits, misses, size int64) {
	return atomic.LoadInt64(&c.hits), atomic.LoadInt64(&c.misses), atomic.LoadInt64(&c.currentSize)
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

// estimateSize estimates the memory size of a response in bytes
func (c *CachedLLMClient) estimateSize(response any) int64 {
	// Rough estimate: JSON size + overhead
	data, err := json.Marshal(response)
	if err != nil {
		return 1024 // Default 1KB if can't estimate
	}
	// Add 50% overhead for Go object structure
	return int64(len(data)) * 3 / 2
}

// shortKey returns a shortened version of the cache key for logging
func (c *CachedLLMClient) shortKey(key string) string {
	if len(key) > 16 {
		return key[:16] + "..."
	}
	return key
}

func (c *CachedLLMClient) SummarizeWithHistory(ctx context.Context, history []*ai.Message, parts []*ai.Part) (providers.SummarizeFlow, error) {
	return c.client.SummarizeWithHistory(ctx, history, parts)
}

func (c *CachedLLMClient) StructurizeWithHistory(ctx context.Context, history []*ai.Message, parts []*ai.Part) (providers.StructurizeFlow, error) {
	return c.client.StructurizeWithHistory(ctx, history, parts)
}
