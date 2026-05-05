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

// TemplateHandler handles template generation requests
type TemplateHandler struct {
	service *analysis.AnalysisService
}

// NewTemplateHandler creates a new template handler
func NewTemplateHandler(service *analysis.AnalysisService) *TemplateHandler {
	return &TemplateHandler{
		service: service,
	}
}

// TemplateRequest handles template generation requests
// @Summary Generate board template
// @Description Generate a board template from a text prompt
// @Tags Template
// @Accept json
// @Produce json
// @Param request body models.GenerateTemplateRequest true "Generate Template Request"
// @Success 200 {object} models.GenerateTemplateResponse
// @Success 202 {string} string "Job ID"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /template [post]
func (h *TemplateHandler) TemplateRequest(c echo.Context) error {
	var req models.GenerateTemplateRequest
	if err := c.Bind(&req); err != nil {
		slog.Error("bind error:", "err", err)
		return c.JSON(http.StatusBadRequest, fmt.Errorf("failed to parse request: %w", err))
	}
	// Validate request
	if err := validateTemplate(req); err != nil {
		slog.Error("validation error:", "err", err)
		return c.JSON(http.StatusBadRequest, fmt.Errorf("invalid request data: %w", err))
	}
	slog.Info("templateRequest", "boardType", req.BoardType, "requestType", req.RequestType)
	resp, err := h.service.StartJob(c.Request().Context(), models.NewTemplateAnalyzeReq(req))
	if err != nil {
		if acceptedErr, ok := utils.MapErr[analysis.ErrAccepted](err); ok {
			slog.Info("enqueue job:", "jobID", acceptedErr.JobID)
			return c.JSON(http.StatusAccepted, acceptedErr.JobID)
		}
		return c.JSON(http.StatusInternalServerError, fmt.Errorf("failed to start job for template generation: %w", err))
	}
	if req.RequestType == models.GenerateTextType {
		return c.JSON(http.StatusOK, resp.TextResponse)
	}
	return c.JSON(http.StatusOK, resp.TemplateResponse)
}

// Validate validates the generate template request
func validateTemplate(r models.GenerateTemplateRequest) error {
	if r.RequestID == "" {
		return fmt.Errorf("requestId is required")
	}
	if r.UserID == "" {
		return fmt.Errorf("userId is required")
	}
	if r.RequestType != models.GenerateTemplateType && r.RequestType != models.GenerateTextType {
		return fmt.Errorf("requestType must be 'generateTemplate'")
	}
	if r.BoardID == "" {
		return fmt.Errorf("boardId is required")
	}
	if r.Prompt == "" {
		return fmt.Errorf("prompt is required")
	}
	if len(r.Prompt) < 10 {
		return fmt.Errorf("prompt must be at least 10 characters")
	}
	if len(r.Prompt) > 2000 {
		return fmt.Errorf("prompt must be at most 2000 characters")
	}
	if r.BoardType == "" {
		return fmt.Errorf("boardType is required")
	}
	if r.BoardType != models.BoardTypeSimple && r.BoardType != models.BoardTypeGraph && r.BoardType != models.BoardTypeDOC {
		return fmt.Errorf("boardType must be 'simple', 'graph', or 'doc'")
	}
	return nil
}
