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

func validateStructurizeRequest(req models.StructurizeRequest) error {
	if req.File.IsEmpty() {
		return fmt.Errorf("file data is empty")
	}
	if req.UserID == "" {
		return fmt.Errorf("userID is empty")
	}

	// Validate file structure to prevent deep nesting
	if err := validateFileStructure(req.File, 0); err != nil {
		return err
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

// validateFileStructure validates the file structure to prevent deep nesting and other issues
func validateFileStructure(file models.File, depth int) error {
	if depth > 10 { // Prevent overly deep nesting
		return fmt.Errorf("file structure too deep, maximum allowed depth is 10")
	}

	if len(file.Name) > 255 { // Prevent overly long names
		return fmt.Errorf("file name too long, maximum allowed length is 255 characters")
	}

	for _, child := range file.Children {
		if child != nil {
			if err := validateFileStructure(*child, depth+1); err != nil {
				return err
			}
		}
	}

	return nil
}

// Structurize processes a board and returns a structured file hierarchy
// @Summary Structurize a board
// @Description Process a board and return a structured file hierarchy
// @Tags Processing
// @Accept json
// @Produce json
// @Param request body models.StructurizeRequest true "Structurize Request"
// @Success 200 {object} models.StructurizeResponse
// @Success 202 {string} string "Job ID"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /structurize [post]
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
		if acceptedErr, ok := utils.MapErr[analysis.ErrAccepted](err); ok {
			slog.Info("enque job:", "jobID", acceptedErr.JobID)
			return c.JSON(http.StatusAccepted, acceptedErr.JobID)
		}
		return c.JSON(http.StatusInternalServerError, fmt.Errorf("failed to start job for structurizing: %w", err))
	}
	return c.JSON(http.StatusOK, resp.StructurizeResponse)
}
