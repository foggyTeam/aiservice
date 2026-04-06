package cache

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/aiservice/internal/models"
)

// AnalysisCacheService manages board analysis caching with bounded memory
type AnalysisCacheService struct {
	cache           map[string]*models.BoardAnalysisCache
	mu              sync.RWMutex
	maxAge          time.Duration
	rescanThreshold int
	maxSize         int
	accessOrder     []string // Track LRU for eviction
}

// NewAnalysisCacheService creates a new analysis cache service with bounded memory
func NewAnalysisCacheService(maxAge time.Duration, rescanThreshold int) *AnalysisCacheService {
	return &AnalysisCacheService{
		cache:           make(map[string]*models.BoardAnalysisCache),
		maxAge:          maxAge,
		rescanThreshold: rescanThreshold,
		maxSize:         100, // Max 100 boards in cache
		accessOrder:     []string{},
	}
}

// GetCache retrieves cached analysis for a board
func (s *AnalysisCacheService) GetCache(boardID string) (*models.BoardAnalysisCache, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cache, exists := s.cache[boardID]
	if !exists {
		slog.Info("[AnalysisCacheService.GetCache] Cache miss", "boardId", boardID)
		return nil, false
	}

	// Check if cache is expired
	if time.Since(cache.LastFullScan) > s.maxAge {
		slog.Info("[AnalysisCacheService.GetCache] Cache expired",
			"boardId", boardID,
			"age", time.Since(cache.LastFullScan))
		return nil, false
	}

	slog.Info("[AnalysisCacheService.GetCache] Cache hit",
		"boardId", boardID,
		"changeCount", cache.ChangeCount,
		"lastFullScan", cache.LastFullScan)

	return cache, true
}

// UpdateCache updates the cached analysis for a board
func (s *AnalysisCacheService) UpdateCache(cache *models.BoardAnalysisCache) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cache[cache.BoardID] = cache
	return nil
}

// InvalidateCache removes cached analysis for a board
func (s *AnalysisCacheService) InvalidateCache(boardID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.cache, boardID)
}

// evictLRU removes the least recently used cache entry
func (s *AnalysisCacheService) evictLRU() {
	if len(s.cache) == 0 || len(s.accessOrder) == 0 {
		return
	}

	// Remove oldest entry
	oldestID := s.accessOrder[0]
	delete(s.cache, oldestID)
	s.accessOrder = s.accessOrder[1:]

	slog.Info("[AnalysisCacheService.evictLRU] Evicted oldest cache entry", "boardId", oldestID)
}

// CalculateElementHashes calculates hashes for all elements in a board
func (s *AnalysisCacheService) CalculateElementHashes(elements []models.Element) map[string]string {
	hashes := make(map[string]string)

	for _, elem := range elements {
		// Create a hash of the element's content and position
		data := fmt.Sprintf("%s:%s:%f:%f:%f:%f",
			elem.Id, elem.Type, elem.X, elem.Y, elem.Width, elem.Height)
		if elem.Content != "" {
			data += ":" + elem.Content
		}

		hash := sha256.Sum256([]byte(data))
		hashes[elem.Id] = fmt.Sprintf("%x", hash[:8])
	}

	return hashes
}

// DetectChanges detects changes between old and new element hashes
func (s *AnalysisCacheService) DetectChanges(
	oldHashes map[string]string,
	newHashes map[string]string,
	oldElements []models.Element,
	newElements []models.Element,
) []models.ElementChange {
	slog.Info("[AnalysisCacheService.DetectChanges] Starting change detection",
		"oldElementCount", len(oldHashes),
		"newElementCount", len(newHashes))

	var changes []models.ElementChange

	// Create maps for quick lookup
	oldElemMap := make(map[string]models.Element)
	for _, elem := range oldElements {
		oldElemMap[elem.Id] = elem
	}

	newElemMap := make(map[string]models.Element)
	for _, elem := range newElements {
		newElemMap[elem.Id] = elem
	}

	// Detect added and modified elements
	for id, newHash := range newHashes {
		oldHash, exists := oldHashes[id]
		if !exists {
			// Element was added
			elem := newElemMap[id]
			changes = append(changes, models.ElementChange{
				ElementID:  id,
				ChangeType: models.ChangeAdded,
				NewElement: &elem,
			})
			slog.Info("[AnalysisCacheService.DetectChanges] Element added", "elementID", id)
		} else if oldHash != newHash {
			// Element was modified
			oldElem := oldElemMap[id]
			newElem := newElemMap[id]
			changes = append(changes, models.ElementChange{
				ElementID:  id,
				ChangeType: models.ChangeModified,
				OldElement: &oldElem,
				NewElement: &newElem,
			})
			slog.Info("[AnalysisCacheService.DetectChanges] Element modified", "elementID", id)
		}
	}

	// Detect deleted elements
	for id := range oldHashes {
		if _, exists := newHashes[id]; !exists {
			elem := oldElemMap[id]
			changes = append(changes, models.ElementChange{
				ElementID:  id,
				ChangeType: models.ChangeDeleted,
				OldElement: &elem,
			})
			slog.Info("[AnalysisCacheService.DetectChanges] Element deleted", "elementID", id)
		}
	}

	slog.Info("[AnalysisCacheService.DetectChanges] Change detection completed",
		"totalChanges", len(changes),
		"added", countByType(changes, models.ChangeAdded),
		"modified", countByType(changes, models.ChangeModified),
		"deleted", countByType(changes, models.ChangeDeleted))

	return changes
}

// countByType counts changes by type
func countByType(changes []models.ElementChange, changeType models.ChangeType) int {
	count := 0
	for _, c := range changes {
		if c.ChangeType == changeType {
			count++
		}
	}
	return count
}

// NeedsFullRescan checks if a full rescan is needed
func (s *AnalysisCacheService) NeedsFullRescan(cache *models.BoardAnalysisCache) bool {
	if cache == nil {
		slog.Info("[AnalysisCacheService.NeedsFullRescan] Cache is nil, needs full rescan")
		return true
	}

	// Check change count threshold
	if cache.ChangeCount >= s.rescanThreshold {
		slog.Info("[AnalysisCacheService.NeedsFullRescan] Change count threshold reached",
			"changeCount", cache.ChangeCount,
			"threshold", s.rescanThreshold)
		return true
	}

	// Check time threshold
	timeSinceLastScan := time.Since(cache.LastFullScan)
	if timeSinceLastScan > s.maxAge {
		slog.Info("[AnalysisCacheService.NeedsFullRescan] Time threshold reached",
			"timeSinceLastScan", timeSinceLastScan,
			"maxAge", s.maxAge)
		return true
	}

	slog.Info("[AnalysisCacheService.NeedsFullRescan] No rescan needed",
		"changeCount", cache.ChangeCount,
		"timeSinceLastScan", timeSinceLastScan)

	return false
}

// UpdateCacheAfterIncremental updates the cache after an incremental analysis
func (s *AnalysisCacheService) UpdateCacheAfterIncremental(
	cache *models.BoardAnalysisCache,
	updatedSummary string,
	updatedConcepts []string,
	updatedRegions []string,
	newElements []models.Element,
	newBoardSize models.BoundingBox,
) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cache.GlobalSummary = updatedSummary
	cache.KeyConcepts = updatedConcepts
	cache.ChangeCount++

	// CRITICAL FIX: Update cached elements and hashes
	cache.Elements = newElements
	cache.ElementHashes = s.CalculateElementHashes(newElements)
	cache.BoardSize = newBoardSize

	// Update region timestamps
	for i, region := range cache.Regions {
		for _, updatedID := range updatedRegions {
			if region.ID == updatedID {
				cache.Regions[i].LastUpdated = time.Now()
			}
		}
	}
}

// UpdateCacheAfterFullScan updates the cache after a full scan
func (s *AnalysisCacheService) UpdateCacheAfterFullScan(
	boardID string,
	summary string,
	concepts []string,
	regions []models.RegionSummary,
	elements []models.Element,
	imageURL string,
	boardSize models.BoundingBox,
) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if we need to evict to stay within maxSize limit
	if len(s.cache) >= s.maxSize && s.cache[boardID] == nil {
		s.evictLRU()
	}

	s.cache[boardID] = &models.BoardAnalysisCache{
		BoardID:       boardID,
		GlobalSummary: summary,
		KeyConcepts:   concepts,
		Regions:       regions,
		LastFullScan:  time.Now(),
		ChangeCount:   0,
		ElementHashes: s.CalculateElementHashes(elements),
		Elements:      elements,
		ImageURL:      imageURL,
		BoardSize:     boardSize,
	}

	// Update access order for LRU
	s.accessOrder = append(s.accessOrder, boardID)
}

// SerializeCache serializes the cache to JSON
func (s *AnalysisCacheService) SerializeCache(cache *models.BoardAnalysisCache) ([]byte, error) {
	return json.Marshal(cache)
}

// DeserializeCache deserializes the cache from JSON
func (s *AnalysisCacheService) DeserializeCache(data []byte) (*models.BoardAnalysisCache, error) {
	var cache models.BoardAnalysisCache
	err := json.Unmarshal(data, &cache)
	return &cache, err
}
