package jobservice

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/aiservice/internal/models"
	"github.com/aiservice/internal/utils"
)

type JobQueueService struct {
	queue       chan models.Job
	oldJobQueue chan models.Job
	wg          sync.WaitGroup
	storage     JobStorage
	request     Processor
}

type JobStorage interface {
	Save(job models.Job) error
	Get(id string) (models.Job, error)
	GetAll() ([]models.Job, error)
	Update(job models.Job) error
	Abort(ctx context.Context, id string) error
	DeleteJobs(ids ...string) error
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

func NewJobQueueService(bufSize, workers, dbWorkers int, storage JobStorage, p Processor) *JobQueueService {
	q := &JobQueueService{
		queue:   make(chan models.Job, bufSize),
		storage: storage,
		request: p,
	}
	for range workers {
		q.wg.Add(1)
		go q.worker()
	}
	for range dbWorkers {
		q.wg.Add(1)
		go q.dbWorker()
	}
	go q.processOldJobs()
	return q
}

func (q *JobQueueService) processOldJobs() {
	ticker := time.NewTicker(time.Minute * 2)
	for {
		for range ticker.C {
			jobs, err := q.storage.GetAll()
			if err != nil {
				slog.Error("failed to get all jobs:", "err", err)
				break
			}
			if err := q.cleanJobs(jobs...); err != nil {
				slog.Error("failed to clean aborted jobs:", "err", err)
			}
			for _, j := range jobs {
				q.oldJobQueue <- j
			}
		}
	}
}

func (q *JobQueueService) cleanJobs(jobs ...models.Job) error {
	abortedJobs := utils.Filter(jobs, func(j models.Job) bool { return j.Status == models.JobStatusAborted })
	abortedJobsIds := utils.Map(abortedJobs, func(j models.Job) string { return j.ID })
	slog.Info("clean abored jobs:", "job ids:", abortedJobsIds)
	return q.storage.DeleteJobs(abortedJobsIds...)
}

type QueueFullErr struct {
}

func (q QueueFullErr) Error() string {
	return "queue is full"
}

func (q *JobQueueService) Enqueue(job models.Job) error {
	if err := q.storage.Save(job); err != nil {
		return fmt.Errorf("failed to save job: %w", err)
	}
	select {
	case q.queue <- job:
		return nil
	default:
		return QueueFullErr{}
	}
}

func (q *JobQueueService) GetJob(ctx context.Context, jobID string) (models.Job, error) {
	return q.storage.Get(jobID)
}

// TODO
// Тут нужен какой-то локер, чтобы не получился такой сценарий:
// GetJobId() -> aborted
// but already in processingJob
// RESULT: processing aborted job

// TODO
// Джобы, лежащие в БД не процессятся дальше

// TODO нужно добавить чистку неактивных джоб

func (q *JobQueueService) Status(jobID string) (models.JobStatus, error) {
	job, err := q.storage.Get(jobID)
	if err != nil {
		return "", err
	}
	return job.Status, nil
}

func (q *JobQueueService) worker() {
	defer q.wg.Done()
	for job := range q.queue {
		q.processJob(job)
	}
}

func (q *JobQueueService) dbWorker() {
	defer q.wg.Done()
	for job := range q.oldJobQueue {
		q.processJob(job)
	}
}

func (q *JobQueueService) processJob(job models.Job) {
	slog.Info("job starting processing", "id", job.ID)

	status, err := q.Status(job.ID)
	if err != nil {
		slog.Error("failed to get job status", "id", job.ID, "err", err)
		return
	}

	if status == models.JobStatusAborted {
		slog.Info("stop processing aborted job", "id", job.ID)
		return
	}

	job.Status = models.JobStatusRunning
	_ = q.storage.Update(job)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	resp, err := q.request.Process(ctx, job.Request)

	if err != nil {
		slog.Info("[job %s] error: %v", job.ID, err)
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

	slog.Info("job completed", "id", job.ID)
}

func (q *JobQueueService) Abort(ctx context.Context, jobID string) error {
	return q.storage.Abort(ctx, jobID)
}

func (q *JobQueueService) deliverCallback(job models.Job, payload map[string]any) {
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
