package analysis

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/aiservice/internal/models"
	"github.com/aiservice/internal/providers"
	"github.com/aiservice/internal/services/cache"
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
	llm          providers.LLMClient
	timeout      time.Duration
	jobQueue     *jobservice.JobQueueService
	imageService *image.Service
	cacheService *cache.AnalysisCacheService
	cropper      *image.ImageCropper
	sessionStore session.Store[models.BoardSessionState]
	provider     string
}

func NewAnalysisService(timeout time.Duration, llm providers.LLMClient, jobQueue *jobservice.JobQueueService, imageService *image.Service, provider string) *AnalysisService {
	return &AnalysisService{
		timeout:      timeout,
		llm:          llm,
		jobQueue:     jobQueue,
		imageService: imageService,
		provider:     provider,
	}
}

// Alternative constructor for when job queue is set later
func NewAnalysisServiceWithoutJobQueue(timeout time.Duration, llm providers.LLMClient, imageService *image.Service, provider string) *AnalysisService {
	return &AnalysisService{
		timeout:      timeout,
		llm:          llm,
		imageService: imageService,
		provider:     provider,
	}
}

// SetJobQueueService sets the job queue service
func (s *AnalysisService) SetJobQueueService(jobQueueService *jobservice.JobQueueService) {
	s.jobQueue = jobQueueService
}

// SetCacheService sets the cache service for incremental analysis
func (s *AnalysisService) SetCacheService(cacheService *cache.AnalysisCacheService) {
	s.cacheService = cacheService
}

// SetCropper sets the cropper for incremental analysis
func (s *AnalysisService) SetCropper(cropper *image.ImageCropper) {
	s.cropper = cropper
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
	pipeline.SetCacheService(s.cacheService)
	pipeline.SetCropper(s.cropper)
	pipeline.SetLLMForIncremental(s.llm)

	// Handle Session for Summarize
	if req.RequestType == models.SummarizeType && s.sessionStore != nil {
		boardID := s.extractBoardID(req)
		if boardID != "" {
			slog.Info("Initializing session for summarize request on board", "boardID", boardID)
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
		// Check for incremental no-changes shortcut
		if noChangesErr, ok := utils.MapErr[*pipeline.ErrIncrementalNoChanges](err); ok {
			slog.Info("Incremental analysis: no changes detected, returning cached result")
			return models.AnalyzeResponse{IncrementalResponse: noChangesErr.Response}, nil
		}

		// Check for incremental full rescan shortcut
		if fullRescanErr, ok := utils.MapErr[*pipeline.ErrIncrementalFullRescan](err); ok {
			slog.Info("Incremental analysis: full rescan performed")
			return models.AnalyzeResponse{IncrementalResponse: fullRescanErr.Response}, nil
		}

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
	case models.IncrementalType:
		return req.IncrementalRequest.BoardID
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
