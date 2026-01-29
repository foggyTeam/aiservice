package cache

import (
	"context"
	"testing"
	"time"

	"github.com/aiservice/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestInMemoryCache(t *testing.T) {
	cache := NewInMemoryCache()

	// Test setting and getting a value
	testValue := "test-value"
	cache.Set("test-key", testValue, 5*time.Minute)

	value, found := cache.Get("test-key")
	assert.True(t, found)
	assert.Equal(t, testValue, value)

	// Test getting a non-existent key
	_, found = cache.Get("non-existent-key")
	assert.False(t, found)

	// Test expiration
	cache.Set("expiring-key", "exp-value", 1*time.Millisecond)
	time.Sleep(2 * time.Millisecond) // Wait for expiration

	_, found = cache.Get("expiring-key")
	assert.False(t, found)

	// Test deletion
	cache.Set("to-delete", "del-value", 5*time.Minute)
	cache.Delete("to-delete")

	_, found = cache.Get("to-delete")
	assert.False(t, found)
}

func TestGenerateKey(t *testing.T) {
	// Test that the same inputs produce the same key
	key1 := GenerateKey("input1", "input2")
	key2 := GenerateKey("input1", "input2")
	assert.Equal(t, key1, key2)

	// Test that different inputs produce different keys
	key3 := GenerateKey("input1", "input3")
	assert.NotEqual(t, key1, key3)
}

func TestCachedJobStorage(t *testing.T) {
	// Create a mock storage implementation for testing
	mockStorage := &MockJobStorage{
		jobs: make(map[string]models.Job),
	}
	cache := NewInMemoryCache()
	cachedStorage := NewCachedJobStorage(mockStorage, cache)

	// Test saving and getting a job
	job := models.Job{
		ID:        "cached-test-job",
		CreatedAt: 1234567890,
		Status:    models.JobStatusPending,
		Request: models.AnalyzeRequest{
			RequestType: "test",
		},
	}

	err := cachedStorage.Save(job)
	assert.NoError(t, err)

	// First get should hit storage and populate cache
	retrievedJob, err := cachedStorage.Get("cached-test-job")
	assert.NoError(t, err)
	assert.Equal(t, job.ID, retrievedJob.ID)
	assert.Equal(t, job.Status, retrievedJob.Status)

	// Second get should hit cache
	retrievedJob2, err := cachedStorage.Get("cached-test-job")
	assert.NoError(t, err)
	assert.Equal(t, job.ID, retrievedJob2.ID)
	assert.Equal(t, job.Status, retrievedJob2.Status)

	// Test update
	job.Status = models.JobStatusRunning
	err = cachedStorage.Update(job)
	assert.NoError(t, err)

	updatedJob, err := cachedStorage.Get("cached-test-job")
	assert.NoError(t, err)
	assert.Equal(t, models.JobStatusRunning, updatedJob.Status)

	// Test abort
	err = cachedStorage.Abort(nil, "cached-test-job")
	assert.NoError(t, err)

	abortedJob, err := cachedStorage.Get("cached-test-job")
	assert.NoError(t, err)
	assert.Equal(t, models.JobStatusAborted, abortedJob.Status)
}

// MockJobStorage is a mock implementation of the JobStorage interface for testing
type MockJobStorage struct {
	jobs map[string]models.Job
}

func (m *MockJobStorage) Save(job models.Job) error {
	m.jobs[job.ID] = job
	return nil
}

func (m *MockJobStorage) Get(id string) (models.Job, error) {
	job, exists := m.jobs[id]
	if !exists {
		return models.Job{}, nil // Return empty job instead of error for this test
	}
	return job, nil
}

func (m *MockJobStorage) Update(job models.Job) error {
	m.jobs[job.ID] = job
	return nil
}

func (m *MockJobStorage) Abort(_ context.Context, id string) error {
	if job, exists := m.jobs[id]; exists {
		job.Status = models.JobStatusAborted
		m.jobs[id] = job
	}
	return nil
}

func (m *MockJobStorage) GetAll() ([]models.Job, error) {
	var jobs []models.Job
	for _, job := range m.jobs {
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func (m *MockJobStorage) DeleteJobs(ids ...string) error {
	for _, id := range ids {
		delete(m.jobs, id)
	}
	return nil
}

func (m *MockJobStorage) Close() error {
	// For testing purposes, no resources to close
	return nil
}