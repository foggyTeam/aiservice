package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/aiservice/internal/models"
	jobservice "github.com/aiservice/internal/services/jobService"
)

// CachedJobStorage wraps a JobStorage with caching capabilities
type CachedJobStorage struct {
	storage jobservice.JobStorage
	cache   Cache
}

// NewCachedJobStorage creates a new cached storage wrapper
func NewCachedJobStorage(storage jobservice.JobStorage, cache Cache) *CachedJobStorage {
	return &CachedJobStorage{
		storage: storage,
		cache:   cache,
	}
}

func (c *CachedJobStorage) Save(job models.Job) error {
	// Save to underlying storage
	err := c.storage.Save(job)
	if err != nil {
		return err
	}
	
	// Update cache
	cacheKey := fmt.Sprintf("job:%s", job.ID)
	c.cache.Set(cacheKey, job, 10*time.Minute) // Cache for 10 minutes
	
	return nil
}

func (c *CachedJobStorage) Get(id string) (models.Job, error) {
	cacheKey := fmt.Sprintf("job:%s", id)
	
	// Try to get from cache first
	if cachedValue, found := c.cache.Get(cacheKey); found {
		if job, ok := cachedValue.(models.Job); ok {
			return job, nil
		}
	}
	
	// Get from underlying storage
	job, err := c.storage.Get(id)
	if err != nil {
		return models.Job{}, err
	}
	
	// Cache the result
	c.cache.Set(cacheKey, job, 10*time.Minute) // Cache for 10 minutes
	
	return job, nil
}

func (c *CachedJobStorage) Update(job models.Job) error {
	// Update underlying storage
	err := c.storage.Update(job)
	if err != nil {
		return err
	}
	
	// Update cache
	cacheKey := fmt.Sprintf("job:%s", job.ID)
	c.cache.Set(cacheKey, job, 10*time.Minute) // Cache for 10 minutes
	
	return nil
}

func (c *CachedJobStorage) Abort(ctx context.Context, id string) error {
	// Perform abort on underlying storage
	err := c.storage.Abort(ctx, id)
	if err != nil {
		return err
	}
	
	// Invalidate cache entry
	cacheKey := fmt.Sprintf("job:%s", id)
	c.cache.Delete(cacheKey)
	
	return nil
}

func (c *CachedJobStorage) GetAll() ([]models.Job, error) {
	// For now, bypass cache for GetAll as it's complex to cache
	// In a production system, we might want to implement a more sophisticated approach
	return c.storage.GetAll()
}

func (c *CachedJobStorage) DeleteJobs(ids ...string) error {
	// Delete from underlying storage
	err := c.storage.DeleteJobs(ids...)
	if err != nil {
		return err
	}
	
	// Invalidate cache entries
	for _, id := range ids {
		cacheKey := fmt.Sprintf("job:%s", id)
		c.cache.Delete(cacheKey)
	}
	
	return nil
}