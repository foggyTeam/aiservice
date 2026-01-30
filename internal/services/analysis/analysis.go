package analysis

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/aiservice/internal/models"
	"github.com/aiservice/internal/providers"
	jobservice "github.com/aiservice/internal/services/jobService"
	"github.com/aiservice/internal/services/pipeline"
	"github.com/aiservice/internal/utils"
)

type ErrAccepted struct {
	JobID string
}

func (e ErrAccepted) Error() string {
	return fmt.Sprintf("job: %s in processing", e.JobID)
}

type AnalysisService struct {
	llm       providers.LLMClient
	timeout   time.Duration
	jobQueue  *jobservice.JobQueueService
}

func NewAnalysisService(timeout time.Duration, llm providers.LLMClient, jobQueue *jobservice.JobQueueService) *AnalysisService {
	return &AnalysisService{
		timeout:  timeout,
		llm:      llm,
		jobQueue: jobQueue,
	}
}

// Alternative constructor for when job queue is set later
func NewAnalysisServiceWithoutJobQueue(timeout time.Duration, llm providers.LLMClient) *AnalysisService {
	return &AnalysisService{
		timeout: timeout,
		llm:     llm,
	}
}

func (s *AnalysisService) SetJobQueueService(jobQueueService *jobservice.JobQueueService) {
	s.jobQueue = jobQueueService
}

func (s *AnalysisService) Abort(ctx context.Context, jobID string) error {
	if s.jobQueue == nil {
		return fmt.Errorf("job queue service not initialized")
	}
	job, err := s.jobQueue.GetJob(ctx, jobID)
	if err != nil {
		return err
	}
	if job.Status == models.JobStatusPending {
		return s.jobQueue.Abort(ctx, jobID)
	}
	return nil
}

func (s *AnalysisService) GetJob(ctx context.Context, jobID string) (models.Job, error) {
	if s.jobQueue == nil {
		return models.Job{}, fmt.Errorf("job queue service not initialized")
	}
	return s.jobQueue.GetJob(ctx, jobID)
}

func (s *AnalysisService) StartJob(ctx context.Context, req models.AnalyzeRequest) (models.AnalyzeResponse, error) {
	syncCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	resultCh := make(chan models.AnalyzeResponse, 1)
	errCh := make(chan error, 1)

	go func() {
		resp, err := s.Process(syncCtx, req)
		if err != nil {
			errCh <- err
			return
		}
		resultCh <- resp
	}()

	select {
	case <-syncCtx.Done():
		job := jobservice.NewJob(req)
		if err := s.jobQueue.Enqueue(job); err != nil {
			if _, ok := utils.MapErr[jobservice.QueueFullErr](err); ok {
				slog.Warn("job queue is full")
				return models.AnalyzeResponse{}, ErrAccepted{JobID: job.ID}
			}
			slog.Warn("enqueue error: %s", slog.Any("err", err))
			return models.AnalyzeResponse{}, err
		}
		return models.AnalyzeResponse{}, ErrAccepted{JobID: job.ID}
	case err := <-errCh:
		slog.Warn("process error: %s", slog.Any("err", err))
		return models.AnalyzeResponse{}, fmt.Errorf("failed to process request: %w", err)

	case resp := <-resultCh:
		return resp, nil
	}
}

func (s *AnalysisService) Process(ctx context.Context, req models.AnalyzeRequest) (models.AnalyzeResponse, error) {
	p, err := pipeline.BuildPipeline(req.RequestType, s.llm)
	if err != nil {
		return models.AnalyzeResponse{}, fmt.Errorf("failed to build pipeline: %w", err)
	}
	state := &pipeline.PipelineState{AnalyzeRequest: req}
	if err := p.Execute(ctx, state); err != nil {
		return models.AnalyzeResponse{}, fmt.Errorf("processing pipeline failed: %w", err)
	}
	return state.AnalyzeResponse, nil
}
