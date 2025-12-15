package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/aiservice/internal/models"
	"github.com/aiservice/internal/providers"
	jobservice "github.com/aiservice/internal/services/jobService"
	"github.com/aiservice/internal/services/pipeline"
)

type AnalysisService struct {
	ink     providers.InkRecognizer
	llm     providers.LLMClient
	timeout time.Duration
	db      jobservice.JobQueueService
}

func NewAnalysisService(timeout time.Duration, ink providers.InkRecognizer, llm providers.LLMClient) *AnalysisService {
	return &AnalysisService{
		timeout: timeout,
		ink:     ink,
		llm:     llm,
	}
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
			slog.Warn("enqueue error: %s", slog.Any("err", err))
			return models.AnalyzeResponse{}, fmt.Errorf("job queue is full, try again later")
		}

		return models.AnalyzeResponse{}, models.AcceptedResponse{
			JobID:     job.ID,
			Status:    string(models.JobStatusPending),
			CreatedAt: job.CreatedAt,
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
		}

	case err := <-errCh:
		slog.Warn("process error: %s", slog.Any("err", err))
		return models.AnalyzeResponse{}, fmt.Errorf("failed to process request: %w", err)

	case resp := <-resultCh:
		return resp, nil
	}
}

func (s *AnalysisService) Process(ctx context.Context, req models.AnalyzeRequest) (models.AnalyzeResponse, error) {
	state := &pipeline.PipelineState{
		Request:     req,
		ContextData: pipeline.BuildContextData(req.Context),
	}

	p, err := pipeline.BuildPipeline(req.Type, s.ink, s.llm)
	if err != nil {
		return models.AnalyzeResponse{}, err
	}

	if err := p.Execute(ctx, state); err != nil {
		return models.AnalyzeResponse{}, fmt.Errorf("processing pipeline failed: %w", err)
	}

	// attach transcription metadata
	if state.LLMResp.Metadata == nil {
		state.LLMResp.Metadata = make(map[string]any)
	}
	state.LLMResp.Metadata["transcription_meta"] = state.Transcription.Metadata

	return state.LLMResp, nil
}
