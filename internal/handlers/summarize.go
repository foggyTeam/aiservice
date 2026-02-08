package handlers

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

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

	// If ImageURL is not empty, download the image from S3 and update the request
	if req.Board.ImageURL != "" && h.s3Client != nil {
		imageData, err := h.downloadImageFromS3(c.Request().Context(), req.Board.ImageURL)
		if err != nil {
			slog.Error("failed to download image from S3:", "err", err, "url", req.Board.ImageURL)
			// Continue without the image if download fails
		} else {
			// Convert the image to a data URL and update the ImageURL field in the request
			// This will be picked up by the preprocessing layer
			req.Board.ImageURL = "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(imageData)
			slog.Info("Image downloaded from S3 and converted to data URL", "size", len(imageData))
		}
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

// downloadImageFromS3 downloads an image from S3 using the provided URL
func (h *AnalyzeHandler) downloadImageFromS3(ctx context.Context, imageURL string) ([]byte, error) {
	// Extract the key from the S3 URL
	// Assuming the URL format is like: https://storage.yandexcloud.net/bucket/key
	if h.s3Client == nil {
		return nil, fmt.Errorf("S3 client not initialized")
	}

	// Extract the key from the URL by removing the base S3 endpoint
	baseURL := "https://storage.yandexcloud.net/"
	if after, ok := strings.CutPrefix(imageURL, baseURL); ok {
		key := after
		// The key might contain the bucket name, so we need to extract just the actual key
		parts := strings.SplitN(key, "/", 2)
		if len(parts) == 2 {
			actualKey := parts[1]
			return h.s3Client.DownloadFile(ctx, actualKey)
		}
	}

	// If the URL doesn't match the expected format, try to use it as a direct key
	// This assumes the URL is just the key part
	return h.s3Client.DownloadFile(ctx, imageURL)
}
