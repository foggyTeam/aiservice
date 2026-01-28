package handlers

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/aiservice/internal/models"
	analysis "github.com/aiservice/internal/services/analysis"
	"github.com/aiservice/internal/utils"
	"github.com/labstack/echo/v4"
)

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
		if acceptedErr, ok := utils.MapErr[analysis.ErrAccepted](err); ok {
			slog.Info("enque job:", "jobID", acceptedErr.JobID)
			return c.JSON(http.StatusAccepted, acceptedErr.JobID)
		}
		return c.JSON(http.StatusInternalServerError, fmt.Errorf("failed to start for analyzing: %w", err))
	}
	return c.JSON(http.StatusOK, resp)

}
