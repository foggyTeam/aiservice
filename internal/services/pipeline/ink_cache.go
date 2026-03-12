package pipeline

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/aiservice/internal/models"
)

const (
	// InkCacheTTL - время жизни кэша для ink recognition
	InkCacheTTL = 30 * time.Minute
)

// InkCache кэширует результаты распознавания digital ink
type InkCache struct {
	cache map[string]inkCacheEntry
	mu    sync.RWMutex
}

type inkCacheEntry struct {
	Text       string
	Expiration time.Time
}

// NewInkCache создаёт новый кэш для ink recognition
func NewInkCache() *InkCache {
	cache := &InkCache{
		cache: make(map[string]inkCacheEntry),
	}

	// Запускаем фоновую очистку устаревших записей
	go cache.cleanupExpired()

	return cache
}

// Get возвращает распознанный текст из кэша
func (c *InkCache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, found := c.cache[key]
	if !found {
		return "", false
	}

	if time.Now().After(entry.Expiration) {
		return "", false
	}

	return entry.Text, true
}

// Set сохраняет распознанный текст в кэш
func (c *InkCache) Set(key string, text string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[key] = inkCacheEntry{
		Text:       text,
		Expiration: time.Now().Add(InkCacheTTL),
	}
}

// cleanupExpired периодически удаляет устаревшие записи
func (c *InkCache) cleanupExpired() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, entry := range c.cache {
			if now.After(entry.Expiration) {
				delete(c.cache, key)
			}
		}
		c.mu.Unlock()
	}
}

// GenerateInkCacheKey генерирует ключ кэша из line элементов
// Ключ основан на points элементов, поэтому одинаковые линии дают одинаковый ключ
func GenerateInkCacheKey(lineElements []models.Element) string {
	// Создаём упрощённое представление для хэширования
	type point struct {
		X float32 `json:"x"`
		Y float32 `json:"y"`
	}

	type lineSig struct {
		Points []point `json:"p"`
	}

	var signatures []lineSig
	for _, elem := range lineElements {
		if elem.Type != "line" || len(elem.Points) == 0 {
			continue
		}

		// Нормализуем points (группируем по парам x,y)
		var points []point
		for i := 0; i < len(elem.Points); i += 2 {
			if i+1 < len(elem.Points) {
				points = append(points, point{
					X: elem.Points[i],
					Y: elem.Points[i+1],
				})
			}
		}

		if len(points) > 0 {
			signatures = append(signatures, lineSig{Points: points})
		}
	}

	// Сериализуем и хэшируем
	data, err := json.Marshal(signatures)
	if err != nil {
		// Fallback: простой хэш от длины
		return fmt.Sprintf("ink_error_%d", len(signatures))
	}

	hash := sha256.Sum256(data)
	return fmt.Sprintf("ink_%x", hash[:8]) // Первые 8 байт для краткости
}
