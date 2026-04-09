package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/aiservice/internal/models"
	analysis "github.com/aiservice/internal/services/analysis"
	jobservice "github.com/aiservice/internal/services/jobService"
	"github.com/aiservice/internal/utils"
	"github.com/labstack/echo/v4"
)

type AnalyzeHandler struct {
	service     *analysis.AnalysisService
	jobQueue    *jobservice.JobQueueService
	syncTimeout time.Duration
}

func NewAnalyzeHandler(
	service *analysis.AnalysisService,
	jobQueue *jobservice.JobQueueService,
	syncTimeout time.Duration,
) *AnalyzeHandler {
	return &AnalyzeHandler{
		service:     service,
		jobQueue:    jobQueue,
		syncTimeout: syncTimeout,
	}
}

// GetJobStatus retrieves the status of a specific job
// @Summary Get job status
// @Description Get the status of a job by ID
// @Tags Jobs
// @Accept json
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} models.Job
// @Failure 404 {object} map[string]string
// @Router /jobs/{id} [get]
func (h *AnalyzeHandler) GetJobStatus(c echo.Context) error {
	jobID := c.Param("id")
	job, err := h.service.GetJob(c.Request().Context(), jobID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fmt.Errorf("failed to get job: %w", err))
	}
	return c.JSON(http.StatusOK, job)
}

// Abort aborts a specific job
// @Summary Abort a job
// @Description Abort a job by ID
// @Tags Jobs
// @Accept json
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /jobs/{id}/abort [put]
func (h *AnalyzeHandler) Abort(c echo.Context) error {
	jobID := c.Param("id")
	if err := h.service.Abort(c.Request().Context(), jobID); err != nil {
		return c.JSON(http.StatusNotFound, fmt.Errorf("failed to abort job: %w", err))
	}
	return c.JSON(http.StatusOK, nil)
}

// HealthHandler returns the health status of the service
// @Summary Health check
// @Description Check if the service is running
// @Tags Health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func HealthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// SummarizeIncremental handles incremental summarization requests
// @Summary Incremental board analysis
// @Description Analyze board changes incrementally
// @Tags Analysis
// @Accept json
// @Produce json
// @Param request body models.IncrementalAnalysisRequest true "Incremental Analysis Request"
// @Success 200 {object} models.IncrementalAnalysisResponse
// @Success 202 {string} string "Job ID"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /summarize/incremental [post]
func (h *AnalyzeHandler) SummarizeIncremental(c echo.Context) error {
	var req models.IncrementalAnalysisRequest

	if err := c.Bind(&req); err != nil {
		slog.Error("bind error:", "err", err)
		return c.JSON(http.StatusBadRequest, fmt.Errorf("failed to parse request: %w", err))
	}

	if req.BoardID == "" {
		return c.JSON(http.StatusBadRequest, fmt.Errorf("boardId is required"))
	}

	// Set request type if not set
	if req.RequestType == "" {
		req.RequestType = models.IncrementalType
	}

	// Use standard job processing through pipeline
	resp, err := h.service.StartJob(c.Request().Context(), models.NewIncrementalAnalyzeReq(req))
	if err != nil {
		if acceptedErr, ok := utils.MapErr[analysis.ErrAccepted](err); ok {
			slog.Info("enqueue job:", "jobID", acceptedErr.JobID)
			return c.JSON(http.StatusAccepted, acceptedErr.JobID)
		}
		slog.Error("incremental analysis failed:", "err", err)
		return c.JSON(http.StatusInternalServerError, fmt.Errorf("analysis failed: %w", err))
	}

	return c.JSON(http.StatusOK, resp.IncrementalResponse)
}
