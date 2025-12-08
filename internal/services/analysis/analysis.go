package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
	log     *log.Logger
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

func NewAnalysisService(ink InkRecognizer, llm LLMClient, logger *log.Logger) *AnalysisService {
	return &AnalysisService{
		ink: ink,
		llm: llm,
		log: logger,
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
			s.log.Printf("enqueue error: %v", err)
			return models.AnalyzeResponse{}, fmt.Errorf("job queue is full, try again later")
		}

		return models.AnalyzeResponse{}, models.AcceptedResponse{
			JobID:     job.ID,
			Status:    string(models.JobStatusPending),
			CreatedAt: job.CreatedAt,
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
		}

	case err := <-errCh:
		s.log.Printf("process error: %v", err)
		return models.AnalyzeResponse{}, fmt.Errorf("failed to process request: %w", err)

	case resp := <-resultCh:
		return resp, nil
	}
}

func (s *AnalysisService) Process(ctx context.Context, req models.AnalyzeRequest) (models.AnalyzeResponse, error) {
	inputType, err := s.detectInputType(req.Input) // TODO переделать получение типа запроса
	if err != nil {
		return models.AnalyzeResponse{}, fmt.Errorf("invalid input: %w", err)
	}

	transcription, err := s.transcribe(ctx, inputType, req.Input)
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

func (s *AnalysisService) detectInputType(raw json.RawMessage) (string, error) {
	var peek map[string]any
	if err := json.Unmarshal(raw, &peek); err != nil {
		return "", err
	}
	if t, ok := peek["type"].(string); ok && t != "" {
		return t, nil
	}
	return "", fmt.Errorf("input type not specified")
}

func (s *AnalysisService) transcribe(ctx context.Context, inputType string, raw json.RawMessage) (models.TranscriptionResult, error) {
	switch inputType {
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
