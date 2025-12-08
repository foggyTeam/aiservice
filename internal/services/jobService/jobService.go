package jobservice

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/aiservice/internal/models"
)

type JobQueueService struct {
	queue   chan models.Job
	wg      sync.WaitGroup
	storage JobStorage
	log     *log.Logger
	request Processor
}

type JobStorage interface {
	Save(job models.Job) error
	Get(id string) (models.Job, error)
	Update(job models.Job) error
}

type Processor interface {
	Process(ctx context.Context, req models.AnalyzeRequest) (models.AnalyzeResponse, error)
}

func NewJob(req models.AnalyzeRequest) models.Job {
	return models.Job{
		ID:        generateJobID(),
		Request:   req,
		CreatedAt: time.Now().Unix(),
		Status:    models.JobStatusPending,
	}
}

func NewJobQueueService(bufSize, workers int, storage JobStorage, logger *log.Logger, p Processor) *JobQueueService {
	q := &JobQueueService{
		queue:   make(chan models.Job, bufSize),
		storage: storage,
		log:     logger,
		request: p,
	}
	for range workers {
		q.wg.Add(1)
		go q.worker()
	}
	return q
}

func (q *JobQueueService) Enqueue(job models.Job) error {
	if err := q.storage.Save(job); err != nil {
		return fmt.Errorf("failed to save job: %w", err)
	}
	select {
	case q.queue <- job:
		return nil
	default:
		return fmt.Errorf("job queue full")
	}
}

func (q *JobQueueService) GetJob(ctx context.Context, jobID string) (models.Job, error) {
	return q.storage.Get(jobID)
}

func (q *JobQueueService) worker() {
	defer q.wg.Done()
	for job := range q.queue {
		q.processJob(job)
	}
}

func (q *JobQueueService) processJob(job models.Job) {
	q.log.Printf("[job %s] starting processing", job.ID)

	job.Status = models.JobStatusRunning
	_ = q.storage.Update(job)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	resp, err := q.request.Process(ctx, job.Request)

	if err != nil {
		q.log.Printf("[job %s] error: %v", job.ID, err)
		job.Status = models.JobStatusFailed
		_ = q.storage.Update(job)
		q.deliverCallback(job, map[string]any{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	job.Status = models.JobStatusCompleted
	_ = q.storage.Update(job)
	q.deliverCallback(job, map[string]any{
		"status": "success",
		"result": resp,
	})

	q.log.Printf("[job %s] completed", job.ID)
}

func (q *JobQueueService) deliverCallback(job models.Job, payload map[string]any) {
	if job.Request.CallbackURL == "" {
		return
	}
	// Implementation similar to original postJSON
	// TODO: implement callback delivery
}

func (q *JobQueueService) Shutdown() {
	close(q.queue)
	q.wg.Wait()
}

func generateJobID() string {
	return fmt.Sprintf("job_%d", time.Now().UnixNano())
}
