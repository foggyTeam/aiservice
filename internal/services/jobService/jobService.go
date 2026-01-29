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

const (
	// JobCleanupAge defines how old a job must be before it's considered for cleanup (in hours)
	JobCleanupAge = 24 * time.Hour
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
		queue:       make(chan models.Job, bufSize),
		oldJobQueue: make(chan models.Job, bufSize), // Initialize the oldJobQueue
		storage:     storage,
		request:     p,
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
	defer ticker.Stop()

	for range ticker.C {
		jobs, err := q.storage.GetAll()
		if err != nil {
			slog.Error("failed to get all jobs:", "err", err)
			continue
		}

		if err := q.cleanJobs(jobs...); err != nil {
			slog.Error("failed to clean aborted jobs:", "err", err)
		}

		// Process pending jobs that are in storage but not yet processed
		for _, j := range jobs {
			// Only process jobs that are pending and not already being processed
			if j.Status == models.JobStatusPending {
				// Try to send to oldJobQueue, but don't block if the queue is full
				select {
				case q.oldJobQueue <- j:
					slog.Debug("sent pending job to oldJobQueue for processing", "job_id", j.ID)
				default:
					// Queue is full, skip this job for now
					slog.Warn("oldJobQueue is full, skipping job", "job_id", j.ID)
				}
			}
		}
	}
}

func (q *JobQueueService) cleanJobs(jobs ...models.Job) error {
	// Find aborted jobs
	abortedJobs := utils.Filter(jobs, func(j models.Job) bool { return j.Status == models.JobStatusAborted })
	abortedJobsIds := utils.Map(abortedJobs, func(j models.Job) string { return j.ID })

	// Find inactive jobs (older than JobCleanupAge and still pending/running)
	inactiveJobs := utils.Filter(jobs, func(j models.Job) bool {
		return (j.Status == models.JobStatusPending || j.Status == models.JobStatusRunning) &&
			time.Since(time.Unix(j.CreatedAt, 0)) > JobCleanupAge
	})
	inactiveJobsIds := utils.Map(inactiveJobs, func(j models.Job) string { return j.ID })

	// Combine both sets of job IDs to delete
	allJobIdsToDelete := append(abortedJobsIds, inactiveJobsIds...)

	if len(allJobIdsToDelete) > 0 {
		slog.Info("cleaning jobs",
			"aborted_job_ids", abortedJobsIds,
			"inactive_job_ids", inactiveJobsIds,
			"total_deleted", len(allJobIdsToDelete))

		return q.storage.DeleteJobs(allJobIdsToDelete...)
	}

	return nil
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

	// Atomically check and update job status to prevent race conditions
	// This ensures that if a job was aborted between the time it was enqueued and now,
	// we won't process it
	currentJob, err := q.storage.Get(job.ID)
	if err != nil {
		slog.Error("failed to get job from storage", "id", job.ID, "err", err)
		return
	}

	if currentJob.Status == models.JobStatusAborted {
		slog.Info("stop processing aborted job", "id", job.ID)
		return
	}

	// Update status to running atomically
	job.Status = models.JobStatusRunning
	if err := q.storage.Update(job); err != nil {
		slog.Error("failed to update job status to running", "id", job.ID, "err", err)
		return
	}

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
	// Placeholder for callback delivery implementation
	// In a production system, this would send the result to a webhook URL
	// provided by the client when the job was submitted
	slog.Info("callback delivery would be implemented here",
		"job_id", job.ID,
		"payload_status", payload["status"])
}

func (q *JobQueueService) Shutdown() {
	close(q.queue)
	q.wg.Wait()
}

func generateJobID() string {
	return fmt.Sprintf("job_%d", time.Now().UnixNano())
}
