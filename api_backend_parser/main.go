package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aiservice/internal/models"
)

// FoggyBoard represents the board structure from foggy backend
type FoggyBoard struct {
	ID         string           `json:"id"`
	ProjectID  string           `json:"projectId"`
	SectionIDs []string         `json:"sectionIds"`
	Name       string           `json:"name"`
	Type       string           `json:"type"`
	Layers     [][]FoggyElement `json:"layers"`
	UpdatedAt  string           `json:"updatedAt"`
}

// FoggyElement represents an element from foggy backend
type FoggyElement struct {
	ID           string    `json:"id"`
	Draggable    bool      `json:"draggable"`
	DragDistance float64   `json:"dragDistance"`
	X            float64   `json:"x"`
	Y            float64   `json:"y"`
	Rotation     float64   `json:"rotation"`
	Fill         string    `json:"fill"`
	Stroke       string    `json:"stroke"`
	StrokeWidth  int       `json:"strokeWidth"`
	Width        float64   `json:"width"`
	Height       float64   `json:"height"`
	Type         string    `json:"type"`
	Points       []float64 `json:"points,omitempty"`
	Tension      float64   `json:"tension,omitempty"`
	LineJoin     string    `json:"lineJoin,omitempty"`
	LineCap      string    `json:"lineCap,omitempty"`
	Content      string    `json:"content,omitempty"`
}

// GoogleHandwritingRequest represents the request to Google Handwriting API
type GoogleHandwritingRequest struct {
	Options  string                   `json:"options"`
	Requests []GoogleHandwritingInput `json:"requests"`
}

// GoogleHandwritingInput represents a single handwriting input
type GoogleHandwritingInput struct {
	WritingGuide GoogleWritingGuide `json:"writing_guide,omitempty"`
	Ink          [][]interface{}    `json:"ink"`
	Language     string             `json:"language"`
}

// GoogleWritingGuide defines the writing area dimensions
type GoogleWritingGuide struct {
	Width  int `json:"writing_area_width,omitempty"`
	Height int `json:"writing_area_height,omitempty"`
}

// GoogleHandwritingResponse represents the response from Google Handwriting API
type GoogleHandwritingResponse struct {
	Results [][]GoogleCandidate `json:"1"`
}

// GoogleCandidate represents a single recognition candidate
type GoogleCandidate struct {
	Text string `json:"1"`
}

func main() {
	// Parse command line flags
	boardID := flag.String("board", "69a7f58476e8c3b1fb9705c1", "Board ID to fetch from foggy backend")
	mode := flag.String("mode", "summarize", "Mode: summarize, structurize, or template")
	prompt := flag.String("prompt", "", "Prompt for template generation (required for template mode)")
	boardType := flag.String("board-type", "simple", "Board type for template: simple or graph (default: simple)")
	foggyURL := flag.String("foggy-url", "http://localhost:3001", "Foggy backend URL")
	aiServiceURL := flag.String("ai-url", "http://localhost:8080", "AIService URL")
	userID := flag.String("user", "parser-user", "User ID for the request")
	output := flag.String("output", "", "Output file to save the response (optional)")
	google := flag.Bool("google", false, "Use Google Handwriting API instead of AIService")
	googleLang := flag.String("google-lang", "en", "Google Handwriting API language code")
	flag.Parse()

	// Validate template mode parameters
	if *mode == "template" && *prompt == "" {
		fmt.Fprintf(os.Stderr, "Error: -prompt is required for template mode\n")
		os.Exit(1)
	}

	if *mode == "template" && *boardType != "simple" && *boardType != "graph" {
		fmt.Fprintf(os.Stderr, "Error: -board-type must be 'simple' or 'graph'\n")
		os.Exit(1)
	}

	fmt.Printf("Fetching board %s from %s...\n", *boardID, *foggyURL)

	// Fetch board from foggy backend
	boardURL := fmt.Sprintf("%s/boards/%s", *foggyURL, *boardID)
	resp, err := http.Get(boardURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching board: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading response: %v\n", err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Error: foggy backend returned status %d\n", resp.StatusCode)
		fmt.Fprintf(os.Stderr, "Response: %s\n", string(body))
		os.Exit(1)
	}

	// Parse foggy board
	var foggyBoard FoggyBoard
	if err := json.Unmarshal(body, &foggyBoard); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing foggy board: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Board fetched: %s (type: %s)\n", foggyBoard.Name, foggyBoard.Type)
	fmt.Printf("Elements found: %d layers\n", len(foggyBoard.Layers))

	// Transform to our format
	elements := transformElements(foggyBoard)
	fmt.Printf("Transformed elements: %d\n", len(elements))

	var responseBody []byte

	// Route to Google API or AIService based on flag
	if *google {
		fmt.Printf("Sending handwriting recognition request to Google API...\n")

		// Transform elements to Google handwriting trace format
		trace := transformToGoogleTrace(elements)
		fmt.Printf("Created trace with %d strokes\n", len(trace))

		// Create Google API request
		googleReq := GoogleHandwritingRequest{
			Options: "enable_pre_space",
			Requests: []GoogleHandwritingInput{
				{
					Ink:      trace,
					Language: *googleLang,
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

		googleReqBody, _ := json.Marshal(googleReq)
		fmt.Println(string(googleReqBody))

		responseBody = sendToGoogleAPI(googleReqBody, *googleLang)
	} else {
		// Create request based on mode
		var aiReqBody []byte
		var requestType string

		switch *mode {
		case "summarize":
			sumReq := models.SummarizeRequest{
				RequestID:   fmt.Sprintf("foggy-%s-%d", foggyBoard.ID, time.Now().Unix()),
				UserID:      *userID,
				RequestType: "summarize",
				Board: models.Board{
					BoardID:  foggyBoard.ID,
					ImageURL: "https://storage.yandexcloud.net/foggy/board_images/board_temp_image_E0bwLdwqSAxYMnZnVzUGcPnNVqWKH6BGQQxw0N44db7.jpeg",
					Elements: elements,
				},
			}
			aiReqBody, _ = json.Marshal(sumReq)
			requestType = "summarize"
		case "structurize":
			structReq := models.StructurizeRequest{
				RequestID:   fmt.Sprintf("foggy-%s-%d", foggyBoard.ID, time.Now().Unix()),
				UserID:      *userID,
				RequestType: "structurize",
				Board: models.Board{
					BoardID:  foggyBoard.ID,
					ImageURL: "https://storage.yandexcloud.net/foggy/board_images/board_temp_image_E0bwLdwqSAxYMnZnVzUGcPnNVqWKH6BGQQxw0N44db7.jpeg",
					Elements: elements,
				},
				File: models.File{
					Name:     sanitizeName(foggyBoard.Name),
					Type:     "doc",
					Children: []models.File{},
				},
			}
			aiReqBody, _ = json.Marshal(structReq)
			requestType = "structurize"
		case "template":
			templateReq := models.GenerateTemplateRequest{
				RequestID:   fmt.Sprintf("foggy-%s-%d", foggyBoard.ID, time.Now().Unix()),
				UserID:      *userID,
				RequestType: "generateTemplate",
				BoardID:     foggyBoard.ID,
				Prompt:      *prompt,
				BoardType:   models.BoardType(*boardType),
			}
			aiReqBody, _ = json.Marshal(templateReq)
			requestType = "template"
		}

		// Send to AIService
		aiEndpoint := fmt.Sprintf("%s/%s", *aiServiceURL, requestType)
		fmt.Printf("Sending %s request to %s...\n", requestType, aiEndpoint)

		fmt.Println(string(aiReqBody))

		aiResp, err := http.Post(aiEndpoint, "application/json", bytes.NewBuffer(aiReqBody))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error sending to AIService: %v\n", err)
			os.Exit(1)
		}
		defer aiResp.Body.Close()

		responseBody, err = io.ReadAll(aiResp.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading AIService response: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("AIService response status: %d\n", aiResp.StatusCode)
	}

	// Print response
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, responseBody, "", "  "); err != nil {
		fmt.Printf("Response: %s\n", string(responseBody))
	} else {
		fmt.Printf("Response:\n%s\n", prettyJSON.String())
	}

	// Save to file if requested
	if *output != "" {
		if err := os.WriteFile(*output, responseBody, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving to file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Response saved to %s\n", *output)
	}
}

// transformElements converts foggy elements to our format
func transformElements(foggyBoard FoggyBoard) []models.Element {
	var elements []models.Element

	for layerIdx, layer := range foggyBoard.Layers {
		for _, elem := range layer {
			ourElem := models.Element{
				Id:          fmt.Sprintf("layer%d_%s", layerIdx, elem.ID),
				Type:        transformType(elem.Type),
				X:           float32(elem.X),
				Y:           float32(elem.Y),
				Width:       float32(elem.Width),
				Height:      float32(elem.Height),
				Rotation:    float32(elem.Rotation),
				Fill:        elem.Fill,
				Stroke:      elem.Stroke,
				StrokeWidth: elem.StrokeWidth,
			}

			// Transform points if present
			if len(elem.Points) > 0 {
				ourElem.Points = make([]float32, len(elem.Points))
				for i, p := range elem.Points {
					ourElem.Points[i] = float32(p)
				}
			}

			// Transform tension if present
			if elem.Tension > 0 {
				ourElem.Tension = float32(elem.Tension)
			}

			// Add content for text-like elements
			if elem.Type == "text" || elem.Type == "textbox" {
				ourElem.Content = elem.Content
			}

			elements = append(elements, ourElem)
		}
	}

	return elements
}

// transformType converts foggy element types to our types
func transformType(foggyType string) string {
	switch strings.ToLower(foggyType) {
	case "line":
		return "line"
	case "text", "textbox", "text-element":
		return "text"
	case "rectangle", "rect", "box":
		return "rectangle"
	case "ellipse", "circle", "oval":
		return "ellipse"
	case "pencil", "freehand", "drawing":
		return "line"
	default:
		return "rectangle"
	}
}

// sanitizeName creates a valid name for file structure
func sanitizeName(name string) string {
	// Replace spaces and special characters with underscores
	result := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return '_'
	}, name)

	// Limit length
	if len(result) > 50 {
		result = result[:50]
	}

	return result
}

// transformToGoogleTrace converts elements to Google handwriting trace format
// Google expects an array of strokes, where each stroke is [x_coords, y_coords]
func transformToGoogleTrace(elements []models.Element) [][]interface{} {
	var trace [][]interface{}

	for _, elem := range elements {
		// Only process line/pencil elements (handwriting strokes)
		if elem.Type != "line" {
			continue
		}

		// Extract points from the element
		if len(elem.Points) > 0 {
			// Points are in format [x1, y1, x2, y2, ...]
			var xCoords []interface{}
			var yCoords []interface{}

			for i := 0; i < len(elem.Points); i += 2 {
				xCoords = append(xCoords, elem.Points[i])
				if i+1 < len(elem.Points) {
					yCoords = append(yCoords, elem.Points[i+1])
				}
			}

			if len(xCoords) > 0 && len(yCoords) > 0 {
				stroke := []interface{}{xCoords, yCoords}
				trace = append(trace, stroke)
			}
		} else {
			// If no points, create a simple stroke from bounding box
			// This handles elements that don't have detailed point data
			xCoords := []interface{}{elem.X, elem.X + elem.Width/2, elem.X + elem.Width}
			yCoords := []interface{}{elem.Y, elem.Y + elem.Height/2, elem.Y + elem.Height}
			stroke := []interface{}{xCoords, yCoords}
			trace = append(trace, stroke)
		}
	}

	// If no strokes were created, add a default empty stroke
	if len(trace) == 0 {
		trace = append(trace, []interface{}{[]interface{}{}, []interface{}{}})
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

// sendToGoogleAPI sends the handwriting recognition request to Google's API
func sendToGoogleAPI(requestBody []byte, language string) []byte {
	// Google Handwriting API endpoint
	googleURL := "https://www.google.com.tw/inputtools/request?ime=handwriting&app=mobilesearch&cs=1&oe=UTF-8"

	fmt.Printf("Sending request to Google API: %s\n", googleURL)

	// Create HTTP POST request
	req, err := http.NewRequest("POST", googleURL, bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating request: %v\n", err)
		os.Exit(1)
	}

	// Set required headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	// Send request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error sending to Google API: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading Google API response: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Google API response status: %d\n", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Google API returned status %d\n", resp.StatusCode)
		fmt.Fprintf(os.Stderr, "Response: %s\n", string(body))
	}

	return body
}
