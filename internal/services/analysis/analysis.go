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
	llm     providers.LLMClient
	timeout time.Duration
	db      jobservice.JobQueueService
}

func NewAnalysisService(timeout time.Duration, llm providers.LLMClient) *AnalysisService {
	return &AnalysisService{
		timeout: timeout,
		llm:     llm,
	}
}

func (s *AnalysisService) Abort(ctx context.Context, jobID string) error {
	job, err := s.db.GetJob(ctx, jobID)
	if err != nil {
		return err
	}
	if job.Status == models.JobStatusPending {
		return s.db.Abort(ctx, jobID)
	}
	return nil
}

func (s *AnalysisService) GetJob(ctx context.Context, jobID string) (models.Job, error) {
	return s.db.GetJob(ctx, jobID)
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
		if err := s.db.Enqueue(job); err != nil {
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
