package handlers

import (
	"errors"
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

func validateSummarizeRequest(req models.SummarizeRequest) error {
	if req.Board.BoardID == "" {
		return fmt.Errorf("boardID is empty")
	}
	if req.UserID == "" {
		return fmt.Errorf("userID is empty")
	}
	return nil
}

func (h *AnalyzeHandler) Summarize(c echo.Context) error {
	var req models.SummarizeRequest

	if err := c.Bind(&req); err != nil {
		slog.Error("bind error:", "err", err)
		return c.JSON(http.StatusBadRequest, fmt.Errorf("failed to parse request: %w", err))
	}

	if err := validateSummarizeRequest(req); err != nil {
		slog.Error("validation error:", "err", err)
		return c.JSON(http.StatusBadRequest, fmt.Errorf("invalid request data: %w", err))
	}

	resp, err := h.service.StartJob(c.Request().Context(), models.NewSumAnalyzeReq(req))
	if err != nil {
		if acceptedErr, ok := mapErr[analysis.ErrAccepted](err); ok {
			slog.Info("enque job:", "jobID", acceptedErr.JobID)
			return c.JSON(http.StatusAccepted, acceptedErr.JobID)
		}
		return c.JSON(http.StatusInternalServerError, fmt.Errorf("failed to start for analyzing: %w", err))
	}
	return c.JSON(http.StatusOK, resp)

}

func validateStructurizeRequest(req models.StructurizeRequest) error {
	if req.File.IsEmpty() {
		return fmt.Errorf("file data is empty")
	}
	if req.UserID == "" {
		return fmt.Errorf("userID is empty")
	}
	return nil
}

func (h *AnalyzeHandler) Structurize(c echo.Context) error {
	var req models.StructurizeRequest

	if err := c.Bind(&req); err != nil {
		slog.Error("bind error:", "err", err)
		return c.JSON(http.StatusBadRequest, fmt.Errorf("failed to parse request: %w", err))
	}

	if err := validateStructurizeRequest(req); err != nil {
		slog.Error("validation error:", "err", err)
		return c.JSON(http.StatusBadRequest, fmt.Errorf("invalid request data: %w", err))
	}

	resp, err := h.service.StartJob(c.Request().Context(), models.NewStructAnalyzeReq(req))
	if err != nil {
		if acceptedErr, ok := mapErr[analysis.ErrAccepted](err); ok {
			slog.Info("enque job:", "jobID", acceptedErr.JobID)
			return c.JSON(http.StatusAccepted, acceptedErr.JobID)
		}
		return c.JSON(http.StatusInternalServerError, fmt.Errorf("failed to start job for structurizing: %w", err))
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

func (h *AnalyzeHandler) Abort(c echo.Context) error {
	jobID := c.Param("id")
	if err := h.service.Abort(c.Request().Context(), jobID); err != nil {
		return c.JSON(http.StatusNotFound, fmt.Errorf("failed to abort job: %w", err))
	}
	return c.JSON(http.StatusOK, nil)
}

func HealthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func mapErr[T error](err error) (T, bool) {
	var target T
	ok := errors.As(err, &target)
	return target, ok
}
