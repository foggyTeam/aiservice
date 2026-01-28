package storage

import (
	"context"
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

func (s *InMemoryJobStorage) Abort(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	job, ok := s.jobs[id]
	if !ok {
		return fmt.Errorf("job not found")
	}
	job.Status = models.JobStatusAborted
	s.jobs[id] = job
	return nil
}

func (s *InMemoryJobStorage) GetAll() ([]models.Job, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	jobs := make([]models.Job, 0, len(s.jobs))
	for _, j := range s.jobs {
		jobs = append(jobs, j)
	}
	return jobs, nil
}

func (s *InMemoryJobStorage) DeleteJobs(ids ...string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, id := range ids {
		delete(s.jobs, id)
	}
	return nil
}

// func (s *InMemoryJobStorage) Lock() func() {
// 	s.mu.Lock()
// 	return s.mu.Unlock
// }
