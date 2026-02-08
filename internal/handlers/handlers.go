package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/aiservice/internal/s3"
	analysis "github.com/aiservice/internal/services/analysis"
	jobservice "github.com/aiservice/internal/services/jobService"
	"github.com/labstack/echo/v4"
)

type AnalyzeHandler struct {
	service     *analysis.AnalysisService
	jobQueue    *jobservice.JobQueueService
	syncTimeout time.Duration
	s3Client    *s3.YandexS3Client
}

func NewAnalyzeHandler(
	service *analysis.AnalysisService,
	jobQueue *jobservice.JobQueueService,
	syncTimeout time.Duration,
	s3Client *s3.YandexS3Client,
) *AnalyzeHandler {
	return &AnalyzeHandler{
		service:     service,
		jobQueue:    jobQueue,
		syncTimeout: syncTimeout,
		s3Client:    s3Client,
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
