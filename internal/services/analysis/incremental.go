package analysis

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"slices"
	"time"

	"github.com/aiservice/internal/models"
	"github.com/aiservice/internal/providers"
	"github.com/aiservice/internal/services/cache"
	"github.com/aiservice/internal/services/image"
	"github.com/firebase/genkit/go/ai"
)

// IncrementalAnalyzer handles incremental board analysis
type IncrementalAnalyzer struct {
	cacheService *cache.AnalysisCacheService
	cropper      *image.ImageCropper
	llm          providers.LLMClient
	fullAnalyzer *AnalysisService
	httpClient   *http.Client
}

// NewIncrementalAnalyzer creates a new incremental analyzer
func NewIncrementalAnalyzer(
	cacheService *cache.AnalysisCacheService,
	cropper *image.ImageCropper,
	llm providers.LLMClient,
	fullAnalyzer *AnalysisService,
) *IncrementalAnalyzer {
	return &IncrementalAnalyzer{
		cacheService: cacheService,
		cropper:      cropper,
		llm:          llm,
		fullAnalyzer: fullAnalyzer,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}
}

// Analyze analyzes a board, choosing between full and incremental analysis
func (a *IncrementalAnalyzer) Analyze(ctx context.Context, req models.IncrementalAnalysisRequest) (*models.IncrementalAnalysisResponse, error) {
	slog.Info("[IncrementalAnalyzer] Analyze started", "boardId", req.BoardID, "isFullRescan", req.IsFullRescan)

	// Check if we need a full rescan
	cached, hasCache := a.cacheService.GetCache(req.BoardID)

	slog.Info("[IncrementalAnalyzer] Cache check",
		"hasCache", hasCache,
		"isFullRescan", req.IsFullRescan,
		"needsFullRescan", !hasCache || req.IsFullRescan || a.cacheService.NeedsFullRescan(cached))

	needsFullRescan := !hasCache || req.IsFullRescan || a.cacheService.NeedsFullRescan(cached)

	if needsFullRescan {
		slog.Info("[IncrementalAnalyzer] Performing full board analysis", "boardId", req.BoardID)
		return a.analyzeFull(ctx, req)
	}

	slog.Info("[IncrementalAnalyzer] Performing incremental board analysis",
		"boardId", req.BoardID,
		"cachedChangeCount", cached.ChangeCount,
		"cachedLastFullScan", cached.LastFullScan)
	return a.analyzeIncremental(ctx, req, cached)
}

// analyzeFull performs a full board analysis
func (a *IncrementalAnalyzer) analyzeFull(ctx context.Context, req models.IncrementalAnalysisRequest) (*models.IncrementalAnalysisResponse, error) {
	slog.Info("[IncrementalAnalyzer.analyzeFull] Starting full analysis", "boardId", req.BoardID)

	if req.FullBoard == nil {
		return nil, fmt.Errorf("full board data required for full analysis")
	}

	slog.Info("[IncrementalAnalyzer.analyzeFull] Board data",
		"elementCount", len(req.FullBoard.Elements),
		"hasImageURL", req.FullBoard.ImageURL != "")

	// Use the full analyzer
	result, err := a.fullAnalyzer.Process(ctx, models.AnalyzeRequest{
		RequestType: models.SummarizeType,
		SummarizeRequest: models.SummarizeRequest{
			RequestID:   "full-scan-" + time.Now().Format(time.RFC3339),
			UserID:      "system",
			RequestType: models.SummarizeType,
			Board:       *req.FullBoard,
		},
	})
	if err != nil {
		slog.Error("[IncrementalAnalyzer.analyzeFull] Full analysis failed", "err", err)
		return nil, fmt.Errorf("full analysis failed: %w", err)
	}

	slog.Info("[IncrementalAnalyzer.analyzeFull] Full analysis completed successfully")

	// Create regions from elements
	regions := a.createRegionsFromElements(req.FullBoard.Elements)
	slog.Info("[IncrementalAnalyzer.analyzeFull] Created regions", "regionCount", len(regions))

	// Calculate board size from elements
	boardSize := a.calculateBoardSize(req.FullBoard.Elements)
	slog.Info("[IncrementalAnalyzer.analyzeFull] Calculated board size", "width", boardSize.W, "height", boardSize.H)

	// Update cache
	a.cacheService.UpdateCacheAfterFullScan(
		req.BoardID,
		result.SummarizeResponse.Element.Content,
		[]string{"board-analysis"},
		regions,
		req.FullBoard.Elements,
		req.FullBoard.ImageURL,
		boardSize,
	)

	slog.Info("[IncrementalAnalyzer.analyzeFull] Cache updated successfully")

	return &models.IncrementalAnalysisResponse{
		GlobalSummary:  result.SummarizeResponse.Element.Content,
		KeyConcepts:    []string{"board-analysis"},
		UpdatedRegions: []string{},
		IsFullRescan:   true,
	}, nil
}

// analyzeIncremental performs an incremental analysis
func (a *IncrementalAnalyzer) analyzeIncremental(
	ctx context.Context,
	req models.IncrementalAnalysisRequest,
	cached *models.BoardAnalysisCache,
) (*models.IncrementalAnalysisResponse, error) {
	slog.Info("[IncrementalAnalyzer.analyzeIncremental] Starting incremental analysis", "boardId", req.BoardID)

	if req.FullBoard == nil {
		return nil, fmt.Errorf("full board data required for incremental analysis")
	}

	slog.Info("[IncrementalAnalyzer.analyzeIncremental] Board data",
		"elementCount", len(req.FullBoard.Elements),
		"cachedElementCount", len(cached.Elements))

	// Detect changes using cached elements
	slog.Info("[IncrementalAnalyzer.analyzeIncremental] Detecting changes...")
	changes := a.cacheService.DetectChanges(
		cached.ElementHashes,
		a.cacheService.CalculateElementHashes(req.FullBoard.Elements),
		cached.Elements,
		req.FullBoard.Elements,
	)

	slog.Info("[IncrementalAnalyzer.analyzeIncremental] Change detection completed",
		"changeCount", len(changes),
		"changes", func() []string {
			var result []string
			for _, c := range changes {
				result = append(result, fmt.Sprintf("%s:%s", c.ElementID, c.ChangeType))
			}
			return result
		}())

	if len(changes) == 0 {
		slog.Info("[IncrementalAnalyzer.analyzeIncremental] No changes detected, returning cached result")
		// No changes, return cached result
		return &models.IncrementalAnalysisResponse{
			GlobalSummary:  cached.GlobalSummary,
			KeyConcepts:    cached.KeyConcepts,
			UpdatedRegions: []string{},
			IsFullRescan:   false,
		}, nil
	}

	// Find affected regions
	slog.Info("[IncrementalAnalyzer.analyzeIncremental] Finding affected regions...")
	affectedRegions := a.findAffectedRegions(cached, changes)
	slog.Info("[IncrementalAnalyzer.analyzeIncremental] Found affected regions",
		"regionCount", len(affectedRegions),
		"regionIDs", affectedRegionIDs(affectedRegions))

	// Download image if URL is provided
	var imageData []byte
	if req.FullBoard.ImageURL != "" {
		slog.Info("[IncrementalAnalyzer.analyzeIncremental] Downloading image", "url", req.FullBoard.ImageURL)
		var err error
		imageData, err = a.downloadImage(req.FullBoard.ImageURL)
		if err != nil {
			slog.Warn("[IncrementalAnalyzer.analyzeIncremental] Failed to download image, falling back to full scan",
				"err", err, "url", req.FullBoard.ImageURL)
			return a.analyzeFull(ctx, req)
		}
		slog.Info("[IncrementalAnalyzer.analyzeIncremental] Image downloaded successfully", "size", len(imageData))
	} else {
		slog.Info("[IncrementalAnalyzer.analyzeIncremental] No image URL provided")
	}

	// Crop affected regions
	slog.Info("[IncrementalAnalyzer.analyzeIncremental] Cropping affected regions...")
	crops, regionBBoxes, err := a.cropAffectedRegions(imageData, affectedRegions, req.FullBoard.Elements)
	if err != nil {
		slog.Warn("[IncrementalAnalyzer.analyzeIncremental] Failed to crop regions, falling back to full scan", "err", err)
		return a.analyzeFull(ctx, req)
	}

	slog.Info("[IncrementalAnalyzer.analyzeIncremental] Cropping completed",
		"cropCount", len(crops),
		"cropSizes", func() []int {
			var sizes []int
			for _, c := range crops {
				sizes = append(sizes, len(c))
			}
			return sizes
		}())

	// Merge overlapping crops for efficiency
	slog.Info("[IncrementalAnalyzer.analyzeIncremental] Merging overlapping crops...")
	mergedBBoxes := a.cropper.MergeOverlappingCrops(regionBBoxes)
	slog.Info("[IncrementalAnalyzer.analyzeIncremental] Crops merged",
		"originalCount", len(regionBBoxes), "mergedCount", len(mergedBBoxes))

	// Check if we should fallback to full scan
	slog.Info("[IncrementalAnalyzer.analyzeIncremental] Checking fallback condition...")
	if a.cropper.ShouldFallbackToFullScan(mergedBBoxes, cached.BoardSize) {
		slog.Info("[IncrementalAnalyzer.analyzeIncremental] Crops cover too much area, falling back to full scan")
		return a.analyzeFull(ctx, req)
	}

	// Build incremental prompt with full board image context
	slog.Info("[IncrementalAnalyzer.analyzeIncremental] Building incremental prompt...")
	prompt := a.buildIncrementalPrompt(cached, changes, crops, imageData)
	slog.Info("[IncrementalAnalyzer.analyzeIncremental] Prompt built", "length", len(prompt))

	// Call LLM with both text context and image crops
	slog.Info("[IncrementalAnalyzer.analyzeIncremental] Calling LLM for analysis...")
	llmResponse, err := a.callLLMForIncremental(ctx, prompt, crops, imageData)
	if err != nil {
		slog.Error("[IncrementalAnalyzer.analyzeIncremental] LLM analysis failed", "err", err)
		return nil, fmt.Errorf("LLM analysis failed: %w", err)
	}

	slog.Info("[IncrementalAnalyzer.analyzeIncremental] LLM analysis completed successfully")

	// Calculate updated board size
	newBoardSize := a.calculateBoardSize(req.FullBoard.Elements)

	// Update cache with new elements, hashes, and board size
	slog.Info("[IncrementalAnalyzer.analyzeIncremental] Updating cache...")
	a.cacheService.UpdateCacheAfterIncremental(
		cached,
		llmResponse.GlobalSummary,
		llmResponse.KeyConcepts,
		affectedRegionIDs(affectedRegions),
		req.FullBoard.Elements,
		newBoardSize,
	)

	slog.Info("[IncrementalAnalyzer.analyzeIncremental] Cache updated successfully")

	return &models.IncrementalAnalysisResponse{
		GlobalSummary:  llmResponse.GlobalSummary,
		KeyConcepts:    llmResponse.KeyConcepts,
		UpdatedRegions: affectedRegionIDs(affectedRegions),
		IsFullRescan:   false,
	}, nil
}

// downloadImage downloads an image from URL
func (a *IncrementalAnalyzer) downloadImage(url string) ([]byte, error) {
	slog.Info("[IncrementalAnalyzer.downloadImage] Downloading image", "url", url)

	resp, err := a.httpClient.Get(url)
	if err != nil {
		slog.Error("[IncrementalAnalyzer.downloadImage] HTTP request failed", "err", err, "url", url)
		return nil, fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("[IncrementalAnalyzer.downloadImage] Non-OK status", "status", resp.StatusCode, "url", url)
		return nil, fmt.Errorf("image download failed with status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("[IncrementalAnalyzer.downloadImage] Failed to read body", "err", err)
		return nil, fmt.Errorf("failed to read image body: %w", err)
	}

	slog.Info("[IncrementalAnalyzer.downloadImage] Image downloaded successfully", "size", len(data))
	return data, nil
}

// calculateBoardSize calculates the board size from elements
func (a *IncrementalAnalyzer) calculateBoardSize(elements []models.Element) models.BoundingBox {
	slog.Info("[IncrementalAnalyzer.calculateBoardSize] Calculating board size", "elementCount", len(elements))

	if len(elements) == 0 {
		slog.Info("[IncrementalAnalyzer.calculateBoardSize] No elements, using default size")
		return models.BoundingBox{W: 2000, H: 2000} // Default size
	}

	maxX, maxY := float32(0), float32(0)
	for _, elem := range elements {
		if elem.X+elem.Width > maxX {
			maxX = elem.X + elem.Width
		}
		if elem.Y+elem.Height > maxY {
			maxY = elem.Y + elem.Height
		}
	}

	// Add some padding
	result := models.BoundingBox{
		X: 0,
		Y: 0,
		W: maxX + 100,
		H: maxY + 100,
	}

	slog.Info("[IncrementalAnalyzer.calculateBoardSize] Board size calculated",
		"width", result.W, "height", result.H, "maxElementX", maxX, "maxElementY", maxY)

	return result
}

// createRegionsFromElements creates regions from board elements
func (a *IncrementalAnalyzer) createRegionsFromElements(elements []models.Element) []models.RegionSummary {
	slog.Info("[IncrementalAnalyzer.createRegionsFromElements] Creating regions from elements", "elementCount", len(elements))

	// Simple 3x3 grid approach
	boardSize := a.calculateBoardSize(elements)
	gridW := boardSize.W / 3
	gridH := boardSize.H / 3

	slog.Info("[IncrementalAnalyzer.createRegionsFromElements] Grid calculated",
		"boardWidth", boardSize.W, "boardHeight", boardSize.H,
		"gridWidth", gridW, "gridHeight", gridH)

	regions := []models.RegionSummary{
		{ID: "top-left", BBox: models.BoundingBox{X: 0, Y: 0, W: gridW, H: gridH}},
		{ID: "top-center", BBox: models.BoundingBox{X: gridW, Y: 0, W: gridW, H: gridH}},
		{ID: "top-right", BBox: models.BoundingBox{X: gridW * 2, Y: 0, W: gridW, H: gridH}},
		{ID: "center-left", BBox: models.BoundingBox{X: 0, Y: gridH, W: gridW, H: gridH}},
		{ID: "center", BBox: models.BoundingBox{X: gridW, Y: gridH, W: gridW, H: gridH}},
		{ID: "center-right", BBox: models.BoundingBox{X: gridW * 2, Y: gridH, W: gridW, H: gridH}},
		{ID: "bottom-left", BBox: models.BoundingBox{X: 0, Y: gridH * 2, W: gridW, H: gridH}},
		{ID: "bottom-center", BBox: models.BoundingBox{X: gridW, Y: gridH * 2, W: gridW, H: gridH}},
		{ID: "bottom-right", BBox: models.BoundingBox{X: gridW * 2, Y: gridH * 2, W: gridW, H: gridH}},
	}

	// Assign elements to regions
	for i := range regions {
		for _, elem := range elements {
			elemBBox := models.BoundingBox{
				X: elem.X,
				Y: elem.Y,
				W: elem.Width,
				H: elem.Height,
			}

			if regions[i].BBox.Intersects(elemBBox) {
				regions[i].ElementIDs = append(regions[i].ElementIDs, elem.Id)
			}
		}
		regions[i].LastUpdated = time.Now()

		slog.Info("[IncrementalAnalyzer.createRegionsFromElements] Region populated",
			"regionID", regions[i].ID, "elementCount", len(regions[i].ElementIDs))
	}

	slog.Info("[IncrementalAnalyzer.createRegionsFromElements] Regions created successfully", "regionCount", len(regions))

	return regions
}

// findAffectedRegions finds regions affected by changes
func (a *IncrementalAnalyzer) findAffectedRegions(
	cached *models.BoardAnalysisCache,
	changes []models.ElementChange,
) []models.RegionSummary {
	slog.Info("[IncrementalAnalyzer.findAffectedRegions] Finding affected regions",
		"regionCount", len(cached.Regions), "changeCount", len(changes))

	affected := make(map[string]models.RegionSummary)

	for _, region := range cached.Regions {
		for _, change := range changes {
			// Check if changed element is in this region
			if slices.Contains(region.ElementIDs, change.ElementID) {
				affected[region.ID] = region
				slog.Info("[IncrementalAnalyzer.findAffectedRegions] Region affected",
					"regionID", region.ID, "elementID", change.ElementID, "changeType", change.ChangeType)
			}
		}
	}

	// Convert map to slice
	result := make([]models.RegionSummary, 0, len(affected))
	for _, region := range affected {
		result = append(result, region)
	}

	slog.Info("[IncrementalAnalyzer.findAffectedRegions] Found affected regions",
		"affectedCount", len(result), "regionIDs", affectedRegionIDs(result))

	return result
}

// cropAffectedRegions crops the board image for affected regions
func (a *IncrementalAnalyzer) cropAffectedRegions(
	imageData []byte,
	regions []models.RegionSummary,
	elements []models.Element,
) ([]string, []models.BoundingBox, error) {
	slog.Info("[IncrementalAnalyzer.cropAffectedRegions] Starting cropping",
		"imageSize", len(imageData), "regionCount", len(regions), "elementCount", len(elements))

	if len(imageData) == 0 || len(regions) == 0 {
		slog.Info("[IncrementalAnalyzer.cropAffectedRegions] No image or regions, returning empty")
		return []string{}, []models.BoundingBox{}, nil
	}

	var crops []string
	var bboxes []models.BoundingBox

	for _, region := range regions {
		slog.Info("[IncrementalAnalyzer.cropAffectedRegions] Cropping region",
			"regionID", region.ID, "elementCount", len(region.ElementIDs))

		// Find elements in this region
		var regionElements []models.Element
		for _, elem := range elements {
			if slices.Contains(region.ElementIDs, elem.Id) {
				regionElements = append(regionElements, elem)
			}
		}

		slog.Info("[IncrementalAnalyzer.cropAffectedRegions] Found region elements",
			"regionID", region.ID, "count", len(regionElements))

		// Crop the region with element-aware cropping
		croppedData, croppedBBox, err := a.cropper.CropWithElements(imageData, region.BBox, regionElements)
		if err != nil {
			slog.Warn("[IncrementalAnalyzer.cropAffectedRegions] Failed to crop region",
				"region", region.ID, "err", err)
			continue
		}

		slog.Info("[IncrementalAnalyzer.cropAffectedRegions] Region cropped successfully",
			"regionID", region.ID, "croppedSize", len(croppedData),
			"bbox", fmt.Sprintf("[%f,%f,%f,%f]", croppedBBox.X, croppedBBox.Y, croppedBBox.W, croppedBBox.H))

		// Convert to base64
		crops = append(crops, a.cropper.ImageToBase64(croppedData))
		bboxes = append(bboxes, croppedBBox)
	}

	slog.Info("[IncrementalAnalyzer.cropAffectedRegions] Cropping completed",
		"successCount", len(crops), "totalRegions", len(regions))

	return crops, bboxes, nil
}

// buildIncrementalPrompt builds the prompt for incremental analysis with full board context
func (a *IncrementalAnalyzer) buildIncrementalPrompt(
	cached *models.BoardAnalysisCache,
	changes []models.ElementChange,
	crops []string,
	fullBoardImageData []byte,
) string {
	slog.Info("[IncrementalAnalyzer.buildIncrementalPrompt] Building prompt",
		"cachedSummaryLength", len(cached.GlobalSummary),
		"changeCount", len(changes),
		"cropCount", len(crops),
		"hasFullBoardImage", len(fullBoardImageData) > 0)

	prompt := fmt.Sprintf(`Ты анализируешь whiteboard.

Текущий анализ всей доски:
%s

Ключевые концепции: %v

Произошли изменения в следующих регионах:
`, cached.GlobalSummary, cached.KeyConcepts)

	// Add changes info
	for _, change := range changes {
		prompt += fmt.Sprintf("- Элемент %s: %s\n", change.ElementID, change.ChangeType)
	}

	// Add crops info
	if len(crops) > 0 {
		prompt += fmt.Sprintf("\nИзмененные области: %d кропов изображений (показаны в base64)\n", len(crops))
	}

	// Add full board image context
	if len(fullBoardImageData) > 0 {
		fullBoardBase64 := a.cropper.ImageToBase64(fullBoardImageData)
		prompt += fmt.Sprintf("\nПолное изображение доски для контекста (base64):\n%s\n", fullBoardBase64)
	}

	prompt += `

Обнови анализ доски, учитывая эти изменения и визуальный контекст полной доски.
Верни обновленный globalSummary и обновленные keyConcepts.
`

	slog.Info("[IncrementalAnalyzer.buildIncrementalPrompt] Prompt built", "totalLength", len(prompt))

	return prompt
}

// IncrementalLLMResponse represents the parsed LLM response for incremental analysis
type IncrementalLLMResponse struct {
	GlobalSummary string   `json:"globalSummary"`
	KeyConcepts   []string `json:"keyConcepts"`
}

// callLLMForIncremental calls the LLM with multi-part input (text + images) and parses the response
func (a *IncrementalAnalyzer) callLLMForIncremental(
	ctx context.Context,
	textPrompt string,
	imageCrops []string,
	fullBoardImage []byte,
) (*IncrementalLLMResponse, error) {
	slog.Info("[IncrementalAnalyzer.callLLMForIncremental] Preparing LLM call",
		"hasTextPrompt", len(textPrompt) > 0,
		"cropCount", len(imageCrops),
		"hasFullBoardImage", len(fullBoardImage) > 0)

	// Build multi-part request with text and images
	parts := []*ai.Part{ai.NewTextPart(textPrompt)}

	// Add image crops (base64 encoded)
	for i, cropBase64 := range imageCrops {
		// Convert base64 back to data URL format for Genkit
		parts = append(parts, ai.NewMediaPart("image/jpeg", "data:image/jpeg;base64,"+cropBase64))
		slog.Debug("[IncrementalAnalyzer.callLLMForIncremental] Added crop image", "cropIndex", i)
	}

	// Add full board image if available (for visual context)
	if len(fullBoardImage) > 0 {
		fullBoardBase64 := a.cropper.ImageToBase64(fullBoardImage)
		parts = append(parts, ai.NewMediaPart("image/jpeg", "data:image/jpeg;base64,"+fullBoardBase64))
		slog.Debug("[IncrementalAnalyzer.callLLMForIncremental] Added full board image")
	}

	slog.Info("[IncrementalAnalyzer.callLLMForIncremental] Calling LLM with", "partCount", len(parts))

	// Call LLM
	response, err := a.llm.Summarize(ctx, parts)
	if err != nil {
		slog.Error("[IncrementalAnalyzer.callLLMForIncremental] LLM call failed", "err", err)
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	slog.Info("[IncrementalAnalyzer.callLLMForIncremental] LLM response received",
		"length", len(response.Summarization))

	// Parse JSON response
	var result IncrementalLLMResponse
	err = json.Unmarshal([]byte(response.Summarization), &result)
	if err != nil {
		slog.Error("[IncrementalAnalyzer.callLLMForIncremental] Failed to parse LLM response",
			"err", err, "response", response.Summarization)
		// Fallback: try to extract as plain text if JSON parsing fails
		result.GlobalSummary = response.Summarization
		result.KeyConcepts = []string{}
	}

	slog.Info("[IncrementalAnalyzer.callLLMForIncremental] Response parsed",
		"hasSummary", len(result.GlobalSummary) > 0,
		"conceptCount", len(result.KeyConcepts))

	return &result, nil
}

// affectedRegionIDs extracts region IDs from regions
func affectedRegionIDs(regions []models.RegionSummary) []string {
	ids := make([]string, len(regions))
	for i, region := range regions {
		ids[i] = region.ID
	}
	return ids
}
