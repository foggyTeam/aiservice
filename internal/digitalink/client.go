package digitalink

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/aiservice/internal/models"
)

var googleHandwritingEndpoint = "https://www.google.com.tw/inputtools/request?ime=handwriting&app=mobilesearch&cs=1&oe=UTF-8"

// GoogleHandwritingRequest represents the request to Google Handwriting API
type GoogleHandwritingRequest struct {
	Options  string                   `json:"options"`
	Requests []GoogleHandwritingInput `json:"requests"`
}

// GoogleHandwritingInput represents a single handwriting input
type GoogleHandwritingInput struct {
	WritingGuide GoogleWritingGuide `json:"writing_guide,omitempty"`
	Ink          [][]any            `json:"ink"`
	Language     string             `json:"language"`
}

// GoogleWritingGuide defines the writing area dimensions
type GoogleWritingGuide struct {
	Width  int `json:"writing_area_width,omitempty"`
	Height int `json:"writing_area_height,omitempty"`
}

// GoogleHandwritingResponse represents the response from Google Handwriting API
// The API returns an array: [status, [candidates], extra_data]
type GoogleHandwritingResponse struct {
	Results [][]GoogleCandidate `json:"1"`
}

// RawGoogleHandwritingResponse represents the raw array response from Google API
type RawGoogleHandwritingResponse []any

// GoogleCandidate represents a single recognition candidate
type GoogleCandidate struct {
	Text string `json:"1"`
}

// Client provides access to Google Handwriting API
type Client struct {
	httpClient *http.Client
	language   string
}

// NewClient creates a new Google Handwriting API client
func NewClient(language string, timeout time.Duration) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: timeout},
		language:   language,
	}
}

// RecognizeInk sends digital ink data to Google Handwriting API and returns recognized text
func (c *Client) RecognizeInk(ctx context.Context, elements []models.Element) (string, error) {
	// Transform elements to Google handwriting trace format
	trace := transformToGoogleTrace(elements)

	// Create Google API request
	googleReq := GoogleHandwritingRequest{
		Options: "enable_pre_space",
		Requests: []GoogleHandwritingInput{
			{
				Ink:      trace,
				Language: c.language,
			},
		},
	}

	// Add writing guide if we have board dimensions
	if len(elements) > 0 {
		maxX, maxY := findMaxDimensions(elements)
		googleReq.Requests[0].WritingGuide = GoogleWritingGuide{
			Width:  int(maxX),
			Height: int(maxY),
		}
	}

	requestBody, err := json.Marshal(googleReq)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP POST request
	req, err := http.NewRequestWithContext(ctx, "POST", googleHandwritingEndpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set required headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("google API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response - Google returns array: [status, [candidates], extra_data]
	var rawResp RawGoogleHandwritingResponse
	if err := json.Unmarshal(body, &rawResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Extract recognized text from raw response
	recognizedText := extractRecognizedTextFromRaw(rawResp)

	return recognizedText, nil
}

// transformToGoogleTrace converts elements to Google handwriting trace format
// Google expects an array of strokes, where each stroke is [x_coords, y_coords]
func transformToGoogleTrace(elements []models.Element) [][]any {
	var trace [][]any

	for _, elem := range elements {
		// Only process line/pencil elements (handwriting strokes)
		if elem.Type != "line" {
			continue
		}

		// Extract points from the element
		if len(elem.Points) > 0 {
			// Points are in format [x1, y1, x2, y2, ...]
			var xCoords []any
			var yCoords []any

			for i := 0; i < len(elem.Points); i += 2 {
				xCoords = append(xCoords, elem.Points[i])
				if i+1 < len(elem.Points) {
					yCoords = append(yCoords, elem.Points[i+1])
				}
			}

			if len(xCoords) > 0 && len(yCoords) > 0 {
				stroke := []any{xCoords, yCoords}
				trace = append(trace, stroke)
			}
		} else {
			// If no points, create a simple stroke from bounding box
			// This handles elements that don't have detailed point data
			xCoords := []any{elem.X, elem.X + elem.Width/2, elem.X + elem.Width}
			yCoords := []any{elem.Y, elem.Y + elem.Height/2, elem.Y + elem.Height}
			stroke := []any{xCoords, yCoords}
			trace = append(trace, stroke)
		}
	}

	// If no strokes were created, add a default empty stroke
	if len(trace) == 0 {
		trace = append(trace, []any{[]any{}, []any{}})
	}

	return trace
}

// findMaxDimensions finds the maximum X and Y coordinates from elements
func findMaxDimensions(elements []models.Element) (float32, float32) {
	var maxX, maxY float32

	for _, elem := range elements {
		rightEdge := elem.X + elem.Width
		bottomEdge := elem.Y + elem.Height

		if rightEdge > maxX {
			maxX = rightEdge
		}
		if bottomEdge > maxY {
			maxY = bottomEdge
		}
	}

	return maxX, maxY
}

// extractRecognizedTextFromRaw extracts text from Google API raw array response
// Response format: ["SUCCESS", [[id, [text1, text2, ...], [], {...}]], extra_data]
func extractRecognizedTextFromRaw(rawResp RawGoogleHandwritingResponse) string {
	if len(rawResp) < 2 {
		return ""
	}

	// Second element contains candidates array
	candidatesArray, ok := rawResp[1].([]any)
	if !ok {
		return ""
	}

	var recognizedTexts []string

	for _, item := range candidatesArray {
		// Each item is an array: [id, [text1, text2, ...], [], {...}]
		candidates, ok := item.([]any)
		if !ok || len(candidates) < 2 {
			continue
		}

		// candidates[1] contains the array of recognized texts
		texts, ok := candidates[1].([]any)
		if !ok || len(texts) == 0 {
			continue
		}

		// Take first (best) text
		if text, ok := texts[0].(string); ok && text != "" {
			recognizedTexts = append(recognizedTexts, text)
		}
	}

	if len(recognizedTexts) == 0 {
		return ""
	}

	// Join all recognized texts with newlines
	var result strings.Builder
	for i, text := range recognizedTexts {
		if i > 0 {
			result.WriteString("\n")
		}
		result.WriteString(text)
	}

	return result.String()
}
