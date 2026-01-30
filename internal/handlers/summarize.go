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

	// Validate board elements
	if len(req.Board.Elements) > 1000 { // Prevent too many elements
		return fmt.Errorf("too many elements in board, maximum allowed is 1000")
	}

	// Validate individual elements
	for _, elem := range req.Board.Elements {
		if elem.Id == "" {
			return fmt.Errorf("element ID cannot be empty")
		}

		// Validate coordinates and dimensions are reasonable
		if elem.Width < 0 || elem.Height < 0 {
			return fmt.Errorf("element width and height must be non-negative")
		}

		// Validate content length if it's a text element
		if elem.Type == "text" && len(elem.Content) > 10000 {
			return fmt.Errorf("text content too long, maximum allowed is 10000 characters")
		}
	}

	return nil
}

// Summarize processes a board and returns a summary
// @Summary Summarize a board
// @Description Process a board and return a summary of the content
// @Tags Processing
// @Accept json
// @Produce json
// @Param request body models.SummarizeRequest true "Summarize Request"
// @Success 200 {object} models.SummarizeResponse
// @Success 202 {string} string "Job ID"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /summarize [post]
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
