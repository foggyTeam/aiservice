package cache

import (
	"context"
	"testing"

	"github.com/aiservice/internal/providers"
	"github.com/firebase/genkit/go/ai"
	"github.com/stretchr/testify/assert"
)

// mockLLMClient is a mock LLM client for testing
type mockLLMClient struct {
	summarizeFunc   func(ctx context.Context, parts []*ai.Part) (providers.SummarizeFlow, error)
	structurizeFunc func(ctx context.Context, parts []*ai.Part) (providers.StructurizeFlow, error)
	callCount       int
}

func (m *mockLLMClient) Summarize(ctx context.Context, parts []*ai.Part) (providers.SummarizeFlow, error) {
	m.callCount++
	if m.summarizeFunc != nil {
		return m.summarizeFunc(ctx, parts)
	}
	return providers.SummarizeFlow{Summarization: "mock summary"}, nil
}

func (m *mockLLMClient) Structurize(ctx context.Context, parts []*ai.Part) (providers.StructurizeFlow, error) {
	m.callCount++
	if m.structurizeFunc != nil {
		return m.structurizeFunc(ctx, parts)
	}
	return providers.StructurizeFlow{AiTreeResponse: "mock tree"}, nil
}

func (m *mockLLMClient) ImageRecognition(ctx context.Context, parts []*ai.Part) (providers.ImageRecognitionFlow, error) {
	return providers.ImageRecognitionFlow{ImageDescription: "mock image description"}, nil
}

func (m *mockLLMClient) GenerateTemplate(ctx context.Context, parts []*ai.Part) (providers.TemplateGenerationFlow, error) {
	return providers.TemplateGenerationFlow{BoardType: "simple", Title: "mock template"}, nil
}

func (m *mockLLMClient) GenerateText(ctx context.Context, parts []*ai.Part) (string, error) {
	return "mock generated text", nil
}

func (m *mockLLMClient) GetName() string {
	return "mock"
}

func TestCachedLLMClient_Summarize_CacheMiss(t *testing.T) {
	mockLLM := &mockLLMClient{}
	cache := NewInMemoryCache()
	cached := NewCachedLLMClient(mockLLM, cache)

	ctx := context.Background()
	parts := []*ai.Part{ai.NewTextPart("test prompt")}

	// First call - should be a cache miss
	resp, err := cached.Summarize(ctx, parts)
	assert.NoError(t, err)
	assert.Equal(t, "mock summary", resp.Summarization)
	assert.Equal(t, 1, mockLLM.callCount)
}

func TestCachedLLMClient_Summarize_CacheHit(t *testing.T) {
	mockLLM := &mockLLMClient{}
	cache := NewInMemoryCache()
	cached := NewCachedLLMClient(mockLLM, cache)

	ctx := context.Background()
	parts := []*ai.Part{ai.NewTextPart("test prompt")}

	// First call - cache miss
	_, err := cached.Summarize(ctx, parts)
	assert.NoError(t, err)
	assert.Equal(t, 1, mockLLM.callCount)

	// Second call - cache hit
	_, err = cached.Summarize(ctx, parts)
	assert.NoError(t, err)
	assert.Equal(t, 1, mockLLM.callCount) // Should not increment
}

func TestCachedLLMClient_Structurize_CacheMiss(t *testing.T) {
	mockLLM := &mockLLMClient{}
	cache := NewInMemoryCache()
	cached := NewCachedLLMClient(mockLLM, cache)

	ctx := context.Background()
	parts := []*ai.Part{ai.NewTextPart("test prompt")}

	// First call - should be a cache miss
	resp, err := cached.Structurize(ctx, parts)
	assert.NoError(t, err)
	assert.Equal(t, "mock tree", resp.AiTreeResponse)
	assert.Equal(t, 1, mockLLM.callCount)
}

func TestCachedLLMClient_Structurize_CacheHit(t *testing.T) {
	mockLLM := &mockLLMClient{}
	cache := NewInMemoryCache()
	cached := NewCachedLLMClient(mockLLM, cache)

	ctx := context.Background()
	parts := []*ai.Part{ai.NewTextPart("test prompt")}

	// First call - cache miss
	_, err := cached.Structurize(ctx, parts)
	assert.NoError(t, err)
	assert.Equal(t, 1, mockLLM.callCount)

	// Second call - cache hit
	_, err = cached.Structurize(ctx, parts)
	assert.NoError(t, err)
	assert.Equal(t, 1, mockLLM.callCount) // Should not increment
}

func TestCachedLLMClient_DifferentPrompts_DifferentCacheKeys(t *testing.T) {
	mockLLM := &mockLLMClient{}
	cache := NewInMemoryCache()
	cached := NewCachedLLMClient(mockLLM, cache)

	ctx := context.Background()

	// First prompt
	parts1 := []*ai.Part{ai.NewTextPart("prompt 1")}
	_, err := cached.Summarize(ctx, parts1)
	assert.NoError(t, err)
	assert.Equal(t, 1, mockLLM.callCount)

	// Second prompt - different content, should be cache miss
	parts2 := []*ai.Part{ai.NewTextPart("prompt 2")}
	_, err = cached.Summarize(ctx, parts2)
	assert.NoError(t, err)
	assert.Equal(t, 2, mockLLM.callCount) // Should increment
}

func TestCachedLLMClient_GetStats(t *testing.T) {
	mockLLM := &mockLLMClient{}
	cache := NewInMemoryCache()
	cached := NewCachedLLMClient(mockLLM, cache)

	ctx := context.Background()
	parts := []*ai.Part{ai.NewTextPart("test prompt")}

	// First call - miss
	_, err := cached.Summarize(ctx, parts)
	assert.NoError(t, err)

	// Second call - hit
	_, err = cached.Summarize(ctx, parts)
	assert.NoError(t, err)

	hits, misses, size := cached.GetStats()
	assert.Equal(t, int64(1), hits)
	assert.Equal(t, int64(1), misses)
	assert.Greater(t, size, int64(0))
}

func (m *mockLLMClient) SummarizeWithHistory(ctx context.Context, history []*ai.Message, parts []*ai.Part) (providers.SummarizeFlow, error) {
	return providers.SummarizeFlow{Summarization: "mock summary with history"}, nil
}

func (m *mockLLMClient) StructurizeWithHistory(ctx context.Context, history []*ai.Message, parts []*ai.Part) (providers.StructurizeFlow, error) {
	return providers.StructurizeFlow{}, nil
}
