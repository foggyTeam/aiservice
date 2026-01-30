package jobservice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
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
	Close() error
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

	// Check if job was aborted atomically
	currentJob, err := q.storage.Get(job.ID)
	if err != nil {
		slog.Error("failed to get job from storage", "id", job.ID, "err", err)
		return
	}

	if currentJob.Status == models.JobStatusAborted {
		slog.Info("stop processing aborted job", "id", job.ID)
		return
	}

	// Attempt to update status to running atomically
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
		// Double-check if job was aborted during processing before updating status
		finalJob, getStatusErr := q.storage.Get(job.ID)
		if getStatusErr != nil || finalJob.Status != models.JobStatusAborted {
			job.Status = models.JobStatusFailed
			_ = q.storage.Update(job)
			q.deliverCallback(job, map[string]any{
				"status": "error",
				"error":  err.Error(),
			})
		}
		return
	}

	// Double-check if job was aborted during processing before updating status
	finalJob, getStatusErr := q.storage.Get(job.ID)
	if getStatusErr != nil || finalJob.Status != models.JobStatusAborted {
		job.Status = models.JobStatusCompleted
		_ = q.storage.Update(job)
		q.deliverCallback(job, map[string]any{
			"status": "success",
			"result": resp,
		})
	} else {
		slog.Info("job was aborted during processing", "id", job.ID)
	}

	slog.Info("job completed", "id", job.ID)
}

func (q *JobQueueService) Abort(ctx context.Context, jobID string) error {
	return q.storage.Abort(ctx, jobID)
}

func (q *JobQueueService) deliverCallback(job models.Job, payload map[string]any) {
	// Extract callback URL from the original request if it exists
	callbackURL := q.extractCallbackURL(job.Request)
	if callbackURL == "" {
		// No callback URL provided, nothing to do
		return
	}

	// Send the callback asynchronously to avoid blocking job processing
	go q.sendCallbackRequest(job, callbackURL, payload)
}

func (q *JobQueueService) extractCallbackURL(req models.AnalyzeRequest) string {
	// This would need to be implemented based on how callback URLs are stored in requests
	// For now, returning empty string as the functionality isn't fully implemented in the models
	// In a real implementation, the callback URL would be stored in the request object

	// Placeholder implementation - in a real system, the callback URL would be part of the original request
	return "" // Return empty for now since models don't include callback URL field
}

func (q *JobQueueService) sendCallbackRequest(job models.Job, callbackURL string, payload map[string]any) {
	// Convert payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		slog.Error("failed to marshal callback payload", "job_id", job.ID, "err", err)
		return
	}

	// Create HTTP request
	httpClient := &http.Client{Timeout: 30 * time.Second} // Reasonable timeout for callbacks
	req, err := http.NewRequestWithContext(context.Background(), "POST", callbackURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		slog.Error("failed to create callback request", "job_id", job.ID, "err", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Callback-Type", "job-status-update")
	req.Header.Set("X-Request-ID", job.ID)

	// Send the request
	resp, err := httpClient.Do(req)
	if err != nil {
		slog.Error("failed to send callback request", "job_id", job.ID, "err", err)
		return
	}
	defer resp.Body.Close()

	// Log the response
	slog.Info("callback sent", "job_id", job.ID, "status_code", resp.StatusCode, "callback_url", callbackURL)

	// In a production system, you would want to implement retry logic here
	// for cases where the callback fails
}

func (q *JobQueueService) Shutdown() {
	close(q.queue)
	close(q.oldJobQueue)  // Also close the old job queue
	q.wg.Wait()
}

func generateJobID() string {
	return fmt.Sprintf("job_%d", time.Now().UnixNano())
}
