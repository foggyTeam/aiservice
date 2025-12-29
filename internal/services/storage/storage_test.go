package storage

import (
	"strconv"
	"sync"
	"testing"

	"github.com/aiservice/internal/models"
	"github.com/stretchr/testify/require"
)

func TestInMemoryJobStorage_SaveGet(t *testing.T) {
	s := NewInMemoryJobStorage()
	j := models.Job{ID: "job1"}

	require.NoError(t, s.Save(j))

	got, err := s.Get("job1")
	require.NoError(t, err)
	require.Equal(t, j.ID, got.ID)
}

func TestInMemoryJobStorage_GetNotFound(t *testing.T) {
	s := NewInMemoryJobStorage()
	_, err := s.Get("no-such-id")
	require.Error(t, err)
}

func TestInMemoryJobStorage_Update(t *testing.T) {
	s := NewInMemoryJobStorage()
	orig := models.Job{ID: "job2"}
	require.NoError(t, s.Save(orig))

	updated := models.Job{ID: "job2"}
	require.NoError(t, s.Update(updated))

	got, err := s.Get("job2")
	require.NoError(t, err)
	require.Equal(t, updated.ID, got.ID)
}

func TestInMemoryJobStorage_Concurrency(t *testing.T) {
	s := NewInMemoryJobStorage()
	const n = 200

	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		i := i
		go func() {
			defer wg.Done()
			id := "job-" + strconv.Itoa(i)
			require.NoError(t, s.Save(models.Job{ID: id}))
		}()
	}
	wg.Wait()

	wg.Add(n)
	for i := 0; i < n; i++ {
		i := i
		go func() {
			defer wg.Done()
			id := "job-" + strconv.Itoa(i)
			got, err := s.Get(id)
			require.NoError(t, err)
			require.Equal(t, id, got.ID)
		}()
	}
	wg.Wait()
}
