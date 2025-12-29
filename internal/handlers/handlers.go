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
		return c.JSON(http.StatusBadRequest, fmt.Errorf("failed to parse request: %w", err))
	}

	if req.BoardID == "" || req.UserID == "" {
		return c.JSON(http.StatusBadRequest, fmt.Errorf("invalid request data, boardID: %s, userID: %s", req.BoardID, req.RequestID))
	}

	resp, err := h.service.StartJob(c.Request().Context(), req)
	if err != nil {
		slog.Error("analysis error:", "err", err)
		return c.JSON(http.StatusInternalServerError, fmt.Errorf("failed to analyze request: %w", err))
	}
	return c.JSON(http.StatusOK, resp)

}

func (h *AnalyzeHandler) GetJobStatus(c echo.Context) error {
	jobID := c.Param("id")
	job, err := h.service.GetJob(c.Request().Context(), jobID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fmt.Errorf("failed to get job: %w", err))
	}
	return c.JSON(http.StatusOK, job)
}

func HealthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}
