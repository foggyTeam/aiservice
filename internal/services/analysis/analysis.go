package analysis

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/aiservice/internal/models"
	"github.com/aiservice/internal/providers"
	"github.com/aiservice/internal/services/image"
	jobservice "github.com/aiservice/internal/services/jobService"
	"github.com/aiservice/internal/services/pipeline"
	"github.com/aiservice/internal/utils"
	"github.com/firebase/genkit/go/core/x/session"
)

type ErrAccepted struct {
	JobID string
}

func (e ErrAccepted) Error() string {
	return fmt.Sprintf("job: %s in processing", e.JobID)
}

type AnalysisService struct {
	llm              providers.LLMClient
	digitalInkClient providers.DigitalInkRecognizer
	timeout          time.Duration
	jobQueue         *jobservice.JobQueueService
	imageService     *image.Service
	sessionStore     session.Store[models.BoardSessionState]
	provider         string
}

func NewAnalysisService(timeout time.Duration, llm providers.LLMClient, digitalInkClient providers.DigitalInkRecognizer, jobQueue *jobservice.JobQueueService, imageService *image.Service, provider string) *AnalysisService {
	return &AnalysisService{
		timeout:          timeout,
		llm:              llm,
		digitalInkClient: digitalInkClient,
		jobQueue:         jobQueue,
		imageService:     imageService,
		provider:         provider,
	}
}

// Alternative constructor for when job queue is set later
func NewAnalysisServiceWithoutJobQueue(timeout time.Duration, llm providers.LLMClient, digitalInkClient providers.DigitalInkRecognizer, imageService *image.Service, provider string) *AnalysisService {
	return &AnalysisService{
		timeout:          timeout,
		llm:              llm,
		digitalInkClient: digitalInkClient,
		imageService:     imageService,
		provider:         provider,
	}
}

// SetJobQueueService sets the job queue service
func (s *AnalysisService) SetJobQueueService(jobQueueService *jobservice.JobQueueService) {
	s.jobQueue = jobQueueService
}

// SetSessionStore sets the session store for chat history
func (s *AnalysisService) SetSessionStore(store session.Store[models.BoardSessionState]) {
	s.sessionStore = store
	slog.Info("Session store configured for analysis service")
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

func (s *AnalysisService) GetJobResponse(ctx context.Context, jobID string) (models.AnalyzeResponse, bool, error) {
	if s.jobQueue == nil {
		return models.AnalyzeResponse{}, false, fmt.Errorf("job queue service not initialized")
	}
	if _, err := s.jobQueue.GetJob(ctx, jobID); err != nil {
		return models.AnalyzeResponse{}, false, err
	}
	return s.jobQueue.GetJobResponse(jobID)
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
	// Set services in pipeline
	pipeline.SetImageService(s.imageService)
	pipeline.SetDigitalInkClient(s.digitalInkClient)

	// Handle Session for Summarize and Structurize
	if (req.RequestType == models.SummarizeType || req.RequestType == models.StructurizeType) && s.sessionStore != nil {
		boardID := s.extractBoardID(req)
		if boardID != "" {
			slog.Info("Initializing session for request on board", "requestType", req.RequestType, "boardID", boardID)
			ctx = s.initSession(ctx, boardID)
		}
	}

	// Build pipeline state
	state := &pipeline.PipelineState{
		AnalyzeRequest: req,
		Provider:       s.provider,
	}

	p, err := pipeline.BuildPipeline(req.RequestType, s.llm, s.provider)
	if err != nil {
		return models.AnalyzeResponse{}, fmt.Errorf("failed to build pipeline: %w", err)
	}

	if err := p.Execute(ctx, state); err != nil {
		return models.AnalyzeResponse{}, fmt.Errorf("processing pipeline failed: %w", err)
	}

	return state.AnalyzeResponse, nil
}

func (s *AnalysisService) extractBoardID(req models.AnalyzeRequest) string {
	switch req.RequestType {
	case models.SummarizeType:
		return req.SummarizeRequest.Board.BoardID
	case models.StructurizeType:
		return req.StructurizeRequest.Board.BoardID
	default:
		return ""
	}
}

func (s *AnalysisService) initSession(ctx context.Context, boardID string) context.Context {
	sess, err := session.Load(ctx, s.sessionStore, boardID)
	if err != nil {
		sess, err = session.New(ctx,
			session.WithID[models.BoardSessionState](boardID),
			session.WithStore(s.sessionStore),
			session.WithInitialState(models.BoardSessionState{}),
		)
		if err != nil {
			slog.Warn("Failed to create session", "err", err)
			return ctx
		}
		slog.Info("Created new session for board", "boardID", boardID)
	} else {
		state := sess.State()
		slog.Info("Retrieved existing session for board", "boardID", boardID, "messageCount", len(state.Messages))
	}
	return session.NewContext(ctx, sess)
}
