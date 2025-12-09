package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/aiservice/internal/models"
	jobservice "github.com/aiservice/internal/services/jobService"
)

const (
	MaxStrokesPoints = 20000
)

type AnalysisService struct {
	ink     InkRecognizer
	llm     LLMClient
	timeout time.Duration
	db      jobservice.JobQueueService
}

type InkRecognizer interface {
	RecognizeInk(ctx context.Context, input models.InkInput) (models.TranscriptionResult, error)
	RecognizeImage(ctx context.Context, input models.ImageInput) (models.TranscriptionResult, error)
}

type LLMClient interface {
	Analyze(ctx context.Context, transcription, contextData string) (models.AnalyzeResponse, error)
}

func NewAnalysisService(timeout time.Duration, ink InkRecognizer, llm LLMClient) *AnalysisService {
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
		if req.CallbackURL == "" {
			return models.AnalyzeResponse{}, fmt.Errorf("processing timed out and no callback_url provided")
		}

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
	transcription, err := s.transcribe(ctx, req.Type, req.Input)
	if err != nil {
		return models.AnalyzeResponse{}, fmt.Errorf("transcription failed: %w", err)
	}

	contextData := s.buildContextData(req.Context)
	resp, err := s.llm.Analyze(ctx, transcription.Text, contextData)
	if err != nil {
		return models.AnalyzeResponse{}, fmt.Errorf("llm analysis failed: %w", err)
	}

	if resp.Metadata == nil {
		resp.Metadata = make(map[string]any)
	}
	resp.Metadata["transcription_meta"] = transcription.Metadata

	return resp, nil
}

func (s *AnalysisService) transcribe(ctx context.Context, inputType string, raw json.RawMessage) (models.TranscriptionResult, error) {
	switch inputType {
	// TODO enum, не зыбать и в модели использовать enum
	case "ink":
		var ink models.InkInput
		if err := json.Unmarshal(raw, &ink); err != nil {
			return models.TranscriptionResult{}, fmt.Errorf("invalid ink input: %w", err)
		}
		if err := s.validateInkInput(ink); err != nil {
			return models.TranscriptionResult{}, err
		}
		return s.ink.RecognizeInk(ctx, ink)

	case "image":
		var img models.ImageInput
		if err := json.Unmarshal(raw, &img); err != nil {
			return models.TranscriptionResult{}, fmt.Errorf("invalid image input: %w", err)
		}
		return s.ink.RecognizeImage(ctx, img)

	case "text":
		var txt models.TextInput
		if err := json.Unmarshal(raw, &txt); err != nil {
			return models.TranscriptionResult{}, fmt.Errorf("invalid text input: %w", err)
		}
		return models.TranscriptionResult{
			Text:     txt.Text,
			Language: "en",
		}, nil

	default:
		return models.TranscriptionResult{}, fmt.Errorf("unsupported input type: %s", inputType)
	}
}

func (s *AnalysisService) validateInkInput(input models.InkInput) error {
	pointCount := 0
	for _, stroke := range input.Strokes {
		pointCount += len(stroke)
	}
	if pointCount > MaxStrokesPoints {
		return fmt.Errorf("too many stroke points: %d (max %d)", pointCount, MaxStrokesPoints)
	}
	if pointCount == 0 {
		return fmt.Errorf("empty strokes")
	}
	return nil
}

func (s *AnalysisService) buildContextData(ctxMap map[string]any) string {
	if ctxMap == nil {
		return ""
	}
	data, _ := json.Marshal(ctxMap)
	return string(data)
}
