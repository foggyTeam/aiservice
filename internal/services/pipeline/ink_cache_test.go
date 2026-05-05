package pipeline

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/aiservice/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestGenerateInkCacheKey(t *testing.T) {
	t.Parallel()

	emptySha := "ink_74234e98afe7498"
	tests := []struct {
		name     string
		elements []models.Element
		wantLen  int
		prefix   string
	}{
		{
			name: "один line с точками",
			elements: []models.Element{
				{
					Type:   "line",
					Points: []float32{10, 20, 30, 40},
				},
			},
			wantLen: 20,
			prefix:  "ink_",
		},
		{
			name: "несколько line элементов",
			elements: []models.Element{
				{Type: "line", Points: []float32{0, 0, 1, 1}},
				{Type: "text"},
				{Type: "line", Points: []float32{2, 2, 3, 3}},
			},
			wantLen: 20,
			prefix:  "ink_",
		},
		{
			name:     "пустой вход - error-fallback ключ",
			elements: []models.Element{},
			wantLen:  0,
			prefix:   emptySha,
		},
		{
			name: "line без points - пропускается",
			elements: []models.Element{
				{Type: "line", X: 10, Y: 20},
			},
			wantLen: 0,
			prefix:  emptySha,
		},
		{
			name: "неполные пары точек",
			elements: []models.Element{
				{Type: "line", Points: []float32{1, 2, 3}},
			},
			wantLen: 20,
			prefix:  "ink_",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			key := GenerateInkCacheKey(tt.elements)

			if tt.prefix != "" {
				assert.Contains(t, key, tt.prefix)
			}
			if tt.wantLen > 0 {
				assert.Len(t, key, tt.wantLen)
			}
		})
	}
}

func TestGenerateInkCacheKey_Deterministic(t *testing.T) {
	t.Parallel()

	elements := []models.Element{
		{Type: "line", Points: []float32{1.5, 2.7, 3.14, 4.2}},
	}

	key1 := GenerateInkCacheKey(elements)
	key2 := GenerateInkCacheKey(elements)

	assert.Equal(t, key1, key2, "одинаковые элементы должны давать одинаковый ключ")
}

func TestGenerateInkCacheKey_OrderMatters(t *testing.T) {
	t.Parallel()

	elems1 := []models.Element{
		{Type: "line", Points: []float32{0, 0, 1, 1}},
		{Type: "line", Points: []float32{2, 2, 3, 3}},
	}
	elems2 := []models.Element{
		{Type: "line", Points: []float32{2, 2, 3, 3}},
		{Type: "line", Points: []float32{0, 0, 1, 1}},
	}

	key1 := GenerateInkCacheKey(elems1)
	key2 := GenerateInkCacheKey(elems2)

	assert.NotEqual(t, key1, key2, "")
}

func TestInkCache_SetAndGet(t *testing.T) {
	t.Parallel()

	cache := NewInkCache()
	defer stopCacheCleanup(cache)

	cache.Set("key-1", "распознанный текст")
	text, found := cache.Get("key-1")

	assert.True(t, found)
	assert.Equal(t, "распознанный текст", text)
}

func TestInkCache_GetMiss(t *testing.T) {
	t.Parallel()

	cache := NewInkCache()
	defer stopCacheCleanup(cache)

	text, found := cache.Get("non-existent-key")

	assert.False(t, found)
	assert.Empty(t, text)
}

func TestInkCache_Overwrite(t *testing.T) {
	t.Parallel()

	cache := NewInkCache()
	defer stopCacheCleanup(cache)

	cache.Set("key", "version-1")
	cache.Set("key", "version-2")

	text, found := cache.Get("key")

	assert.True(t, found)
	assert.Equal(t, "version-2", text)
}

func TestInkCache_TTL_Expired(t *testing.T) {
	t.Parallel()

	cache := &InkCache{
		cache: make(map[string]inkCacheEntry),
	}

	key := "expiring-key"
	cache.cache[key] = inkCacheEntry{
		Text:       "old text",
		Expiration: time.Now().Add(-1 * time.Second),
	}

	text, found := cache.Get(key)

	assert.False(t, found)
	assert.Empty(t, text)
}

func TestInkCache_TTL_Valid(t *testing.T) {
	t.Parallel()

	cache := &InkCache{
		cache: make(map[string]inkCacheEntry),
	}

	key := "valid-key"
	cache.cache[key] = inkCacheEntry{
		Text:       "fresh text",
		Expiration: time.Now().Add(1 * time.Hour),
	}

	text, found := cache.Get(key)

	assert.True(t, found)
	assert.Equal(t, "fresh text", text)
}

func TestInkCache_TTL_Boundary(t *testing.T) {
	t.Parallel()

	cache := &InkCache{
		cache: make(map[string]inkCacheEntry),
	}

	key := "boundary-key"
	cache.cache[key] = inkCacheEntry{
		Text:       "boundary text",
		Expiration: time.Now().Add(1 * time.Millisecond),
	}

	_, found := cache.Get(key)
	assert.True(t, found)

	time.Sleep(5 * time.Millisecond)

	_, found = cache.Get(key)
	assert.False(t, found)
}

func TestInkCache_cleanupExpired_RemovesExpired(t *testing.T) {
	t.Parallel()

	cache := &InkCache{
		cache: make(map[string]inkCacheEntry),
	}

	cache.cache["old"] = inkCacheEntry{
		Text:       "old",
		Expiration: time.Now().Add(-1 * time.Hour),
	}
	cache.cache["new"] = inkCacheEntry{
		Text:       "new",
		Expiration: time.Now().Add(1 * time.Hour),
	}

	cache.mu.Lock()
	now := time.Now()
	for key, entry := range cache.cache {
		if now.After(entry.Expiration) {
			delete(cache.cache, key)
		}
	}
	cache.mu.Unlock()

	_, oldFound := cache.Get("old")
	_, newFound := cache.Get("new")

	assert.False(t, oldFound, "старая запись должна быть удалена")
	assert.True(t, newFound, "новая запись должна остаться")
}

func TestInkCache_cleanupExpired_EmptyCache(t *testing.T) {
	t.Parallel()

	cache := &InkCache{
		cache: make(map[string]inkCacheEntry),
	}

	assert.NotPanics(t, func() {
		cache.mu.Lock()
		now := time.Now()
		for key, entry := range cache.cache {
			if now.After(entry.Expiration) {
				delete(cache.cache, key)
			}
		}
		cache.mu.Unlock()
	})
}

func TestInkCache_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	cache := NewInkCache()
	defer stopCacheCleanup(cache)

	const (
		writers = 10
		readers = 20
		ops     = 100
	)

	var wg sync.WaitGroup

	for w := 0; w < writers; w++ {
		wg.Add(1)
		go func(writerID int) {
			defer wg.Done()
			for i := 0; i < ops; i++ {
				key := fmt.Sprintf("key-%d-%d", writerID, i%10)
				cache.Set(key, fmt.Sprintf("value-%d", i))
			}
		}(w)
	}

	for r := 0; r < readers; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < ops; i++ {
				key := fmt.Sprintf("key-%d-%d", i%writers, i%10)
				_, _ = cache.Get(key) // результат не важен, важна отсутствие паники
			}
		}()
	}

	wg.Wait()
}

func TestInkCache_FullFlow(t *testing.T) {
	t.Parallel()

	cache := NewInkCache()
	defer stopCacheCleanup(cache)

	elements := []models.Element{
		{Type: "line", Points: []float32{10, 20, 30, 40}},
	}
	key := GenerateInkCacheKey(elements)

	_, found := cache.Get(key)
	assert.False(t, found)

	cache.Set(key, "распознанный текст")

	text, found := cache.Get(key)
	assert.True(t, found)
	assert.Equal(t, "распознанный текст", text)
}

func TestInkCache_DifferentElements_DifferentKeys(t *testing.T) {
	t.Parallel()

	cache := NewInkCache()
	defer stopCacheCleanup(cache)

	elems1 := []models.Element{{Type: "line", Points: []float32{0, 0, 1, 1}}}
	elems2 := []models.Element{{Type: "line", Points: []float32{2, 2, 3, 3}}}

	key1 := GenerateInkCacheKey(elems1)
	key2 := GenerateInkCacheKey(elems2)

	cache.Set(key1, "text-1")
	cache.Set(key2, "text-2")

	text1, found1 := cache.Get(key1)
	text2, found2 := cache.Get(key2)

	assert.True(t, found1 && found2)
	assert.Equal(t, "text-1", text1)
	assert.Equal(t, "text-2", text2)
}

func stopCacheCleanup(_ *InkCache) {
}
