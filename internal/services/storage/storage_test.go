package storage

import (
	"strconv"
	"sync"
	"testing"

	"github.com/aiservice/internal/models"
)

func TestInMemoryJobStorage_SaveGet(t *testing.T) {
	s := NewInMemoryJobStorage()
	j := models.Job{ID: "job1"}

	if err := s.Save(j); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	got, err := s.Get("job1")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if got.ID != j.ID {
		t.Fatalf("got.ID = %q; want %q", got.ID, j.ID)
	}
}

func TestInMemoryJobStorage_GetNotFound(t *testing.T) {
	s := NewInMemoryJobStorage()
	if _, err := s.Get("no-such-id"); err == nil {
		t.Fatalf("expected error for missing job, got nil")
	}
}

func TestInMemoryJobStorage_Update(t *testing.T) {
	s := NewInMemoryJobStorage()
	orig := models.Job{ID: "job2"}
	if err := s.Save(orig); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	updated := models.Job{ID: "job2"} // keep same ID; update other fields if present
	if err := s.Update(updated); err != nil {
		t.Fatalf("Update returned error: %v", err)
	}

	got, err := s.Get("job2")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if got.ID != updated.ID {
		t.Fatalf("got.ID = %q; want %q", got.ID, updated.ID)
	}
}

func TestInMemoryJobStorage_Concurrency(t *testing.T) {
	s := NewInMemoryJobStorage()
	const n = 200

	var wg sync.WaitGroup
	wg.Add(n)
	for i := range n {
		go func() {
			defer wg.Done()
			id := "job-" + strconv.Itoa(i)
			if err := s.Save(models.Job{ID: id}); err != nil {
				t.Errorf("Save(%s) error: %v", id, err)
			}
		}()
	}
	wg.Wait()

	// concurrent reads
	wg.Add(n)
	for i := range n {
		go func() {
			defer wg.Done()
			id := "job-" + strconv.Itoa(i)
			got, err := s.Get(id)
			if err != nil {
				t.Errorf("Get(%s) error: %v", id, err)
				return
			}
			if got.ID != id {
				t.Errorf("Get(%s) returned id %s", id, got.ID)
			}
		}()
	}
	wg.Wait()
}
