package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/aiservice/internal/models"

	analysis "github.com/aiservice/internal/services/analysis"
	jobservice "github.com/aiservice/internal/services/jobService"
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

func (h *AnalyzeHandler) Handle(c echo.Context) error {
	var req models.AnalyzeRequest

	if err := c.Bind(&req); err != nil {
		slog.Error("bind error:", "err", err)
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid JSON",
			Details: err.Error(),
		})
	}

	if req.BoardID == "" || req.UserID == "" || len(req.Input) == 0 {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "MISSING_FIELDS",
			Message: "board_id, user_id, and input are required",
		})
	}

	resp, err := h.service.StartJob(c.Request().Context(), req)
	if err != nil {
		slog.Error("analysis error:", "err", err)
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "ANALYSIS_ERROR",
			Message: "Failed to analyze input",
			Details: err.Error(),
		})
	}
	return c.JSON(http.StatusOK, resp)

}

func (h *AnalyzeHandler) GetJobStatus(c echo.Context) error {
	jobID := c.Param("id")
	job, err := h.service.GetJob(c.Request().Context(), jobID)
	if err != nil {
		return c.JSON(http.StatusNotFound, models.ErrorResponse{
			Code:    "JOB_NOT_FOUND",
			Message: fmt.Sprintf("Job %s not found", jobID),
		})
	}
	return c.JSON(http.StatusOK, job)
}

func HealthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}
