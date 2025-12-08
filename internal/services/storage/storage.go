package storage

import (
	"fmt"
	"sync"

	"github.com/aiservice/internal/models"
)

type InMemoryJobStorage struct {
	jobs map[string]models.Job
	mu   sync.RWMutex
}

func NewInMemoryJobStorage() *InMemoryJobStorage {
	return &InMemoryJobStorage{
		jobs: make(map[string]models.Job),
	}
}

func (s *InMemoryJobStorage) Save(job models.Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[job.ID] = job
	return nil
}

func (s *InMemoryJobStorage) Get(id string) (models.Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if job, ok := s.jobs[id]; ok {
		return job, nil
	}
	return models.Job{}, fmt.Errorf("job not found")
}

func (s *InMemoryJobStorage) Update(job models.Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[job.ID] = job
	return nil
}
