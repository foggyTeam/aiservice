package jobservice

import (
	"context"
	"testing"
	"time"

	"github.com/aiservice/internal/models"
	"github.com/aiservice/internal/services/storage"
)

type fakeProcessor struct {
	called chan struct{}
}

func (f *fakeProcessor) Process(ctx context.Context, req models.AnalyzeRequest) (models.AnalyzeResponse, error) {
	// signal that Process was invoked
	select {
	case f.called <- struct{}{}:
	default:
	}
	return models.AnalyzeResponse{ResponseMessage: "ok"}, nil
}

func TestWorkerProcessesJobFromQueue(t *testing.T) {
	st := storage.NewInMemoryJobStorage()
	fp := &fakeProcessor{called: make(chan struct{}, 1)}

	// create queue service with 1 worker
	svc := NewJobQueueService(10, 1, st, fp)
	// svc likely runs workers on creation; push job directly into internal queue
	job := models.Job{ID: "job-test-1"}
	select {
	case svc.queue <- job:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("failed to enqueue job into internal queue")
	}

	// wait for processor to be called
	select {
	case <-fp.called:
		// success
	case <-time.After(2 * time.Second):
		t.Fatal("processor was not invoked for queued job")
	}
}

func TestMultipleJobsProcessedConcurrently(t *testing.T) {
	st := storage.NewInMemoryJobStorage()
	fp := &fakeProcessor{called: make(chan struct{}, 10)}

	svc := NewJobQueueService(50, 4, st, fp)

	const n = 10
	for i := range n {
		j := models.Job{ID: "job-" + string(rune(i+65))}
		select {
		case svc.queue <- j:
		case <-time.After(200 * time.Millisecond):
			t.Fatalf("failed to enqueue job %d", i)
		}
	}

	// expect at least n Process calls (buffered channel)
	timeout := time.After(3 * time.Second)
	count := 0
	for count < n {
		select {
		case <-fp.called:
			count++
		case <-timeout:
			t.Fatalf("only %d/%d jobs processed", count, n)
		}
	}
}
