package pipeline

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/aiservice/internal/digitalink"
	"github.com/aiservice/internal/models"
	"github.com/aiservice/internal/preprocessing"
	"github.com/aiservice/internal/providers"
	"github.com/aiservice/internal/services/cache"
	"github.com/aiservice/internal/services/graph"
	"github.com/aiservice/internal/services/image"
	"github.com/aiservice/internal/utils"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core/x/session"
)

// Preprocessor for transforming raw data into structured formats
var preprocessor = preprocessing.NewPreprocessor()

// DigitalInkClient for recognizing handwriting
var digitalInkClient = digitalink.NewClient("en", 0)

// InkCache для кэширования результатов распознавания
var inkCache = NewInkCache()

// GraphPreprocessor для преобразования графов
var graphPreprocessor = graph.NewGraphPreprocessor()

// ImageService for downloading and cleaning up images
var imageService *image.Service

// CacheService for incremental analysis caching
var cacheService *cache.AnalysisCacheService

// Cropper for image cropping
var cropper *image.ImageCropper

// LLM client for incremental full rescan
var llmForIncremental providers.LLMClient

// SetImageService sets the image service for the pipeline
func SetImageService(svc *image.Service) {
	imageService = svc
}

// SetCacheService sets the cache service for the pipeline
func SetCacheService(svc *cache.AnalysisCacheService) {
	cacheService = svc
}

// SetCropper sets the cropper for the pipeline
func SetCropper(c *image.ImageCropper) {
	cropper = c
}

// SetLLMForIncremental sets the LLM client for incremental full rescan
func SetLLMForIncremental(llm providers.LLMClient) {
	llmForIncremental = llm
}

func newImageDownloadStep() Step {
	return func(ctx context.Context, state *PipelineState) error {
		// Get image URL from request
		imageURL := ""
		switch state.AnalyzeRequest.RequestType {
		case models.SummarizeType:
			imageURL = state.AnalyzeRequest.SummarizeRequest.Board.ImageURL
		case models.StructurizeType:
			imageURL = state.AnalyzeRequest.StructurizeRequest.Board.ImageURL
		}

		if imageURL == "" || imageService == nil {
			return nil
		}

		// Download image only for Ollama provider (local model)
		// For remote providers (Gemini, Yandex), use URL directly
		if state.Provider == "ollama" {
			downloadResult, err := imageService.DownloadImage(ctx, imageURL)
			if err != nil {
				slog.Warn("failed to download image, continuing without it", "err", err, "url", imageURL)
				return nil
			}
			// For Ollama, read file and create data URL (base64)
			imageData, err := os.ReadFile(downloadResult.LocalPath)
			if err != nil {
				slog.Warn("failed to read image file", "err", err, "path", downloadResult.LocalPath)
				return nil
			}
			base64Data := base64.StdEncoding.EncodeToString(imageData)
			state.ImageURI = fmt.Sprintf("data:image/jpeg;base64,%s", base64Data)
			state.ImageDownloadResult = downloadResult
			slog.Debug("image loaded as data URL", "size", len(imageData))
		} else {
			// For remote providers, use direct URL
			state.ImageURI = imageURL
		}

		return nil
	}
}

func newDigitalInkAnalysisStep() Step {
	return func(ctx context.Context, state *PipelineState) error {
		var lineElements []models.Element
		switch state.AnalyzeRequest.RequestType {
		case models.SummarizeType:
			lineElements = extractLineElements(state.AnalyzeRequest.SummarizeRequest.Board.Elements)
			slog.Debug("starting digital ink analysis for summarize request")
		case models.StructurizeType:
			lineElements = extractLineElements(state.AnalyzeRequest.StructurizeRequest.Board.Elements)
			slog.Debug("starting digital ink analysis for structurize request")
		default:
			slog.Debug("unknown request type, skipping digital ink analysis", "type", state.AnalyzeRequest.RequestType)
			return nil
		}

		// Skip if no line elements found
		if len(lineElements) == 0 {
			slog.Debug("no line elements found for digital ink analysis")
			return nil
		}

		// Generate cache key from line elements
		cacheKey := GenerateInkCacheKey(lineElements)
		slog.Debug("ink cache key", "key", cacheKey)

		// Try to get from cache first
		if recognizedText, found := inkCache.Get(cacheKey); found {
			slog.Info("Digital ink cache HIT", "key", cacheKey[:16]+"...")
			state.DigitalInkText = recognizedText
			return nil
		}

		// Cache miss - recognize handwriting
		slog.Info("Digital ink cache MISS", "key", cacheKey[:16]+"...")
		recognizedText, err := digitalInkClient.RecognizeInk(ctx, lineElements)
		if err != nil {
			slog.Warn("digital ink recognition failed", "err", err)
			// Continue without recognized text - don't fail the entire pipeline
			return nil
		}

		slog.Info("digital ink recognition result:", "", recognizedText)
		// Cache the result
		inkCache.Set(cacheKey, recognizedText)
		slog.Debug("ink recognition cached", "key", cacheKey[:16]+"...", "text_len", len(recognizedText))

		state.DigitalInkText = recognizedText
		return nil
	}
}

// newGraphPreprocessingStep преобразует граф в семантический формат
func newGraphPreprocessingStep() Step {
	return func(ctx context.Context, state *PipelineState) error {
		// Проверяем, есть ли граф в запросе
		var graphNodes []models.GNode
		var graphEdges []models.GEdge

		switch state.AnalyzeRequest.RequestType {
		case models.SummarizeType:
			graphNodes = state.AnalyzeRequest.SummarizeRequest.Board.GraphNodes
			graphEdges = state.AnalyzeRequest.SummarizeRequest.Board.GraphEdges
		case models.StructurizeType:
			graphNodes = state.AnalyzeRequest.StructurizeRequest.Board.GraphNodes
			graphEdges = state.AnalyzeRequest.StructurizeRequest.Board.GraphEdges
		}

		// Если графа нет - пропускаем шаг
		if len(graphNodes) == 0 {
			slog.Debug("no graph nodes found, skipping graph preprocessing")
			return nil
		}

		// Преобразуем граф в семантический формат
		semanticGraph, err := graphPreprocessor.Process(graphNodes, graphEdges)
		if err != nil {
			slog.Warn("graph preprocessing failed", "err", err)
			return nil // не прерываем пайплайн
		}

		// Сохраняем в state для использования в Summarize/Structurize шаге
		state.SemanticGraph = semanticGraph
		slog.Info("graph preprocessed", "nodes", len(semanticGraph.Nodes), "edges", len(semanticGraph.Edges))

		return nil
	}
}

// newTemplateGenerationStep generates a board template from a prompt
func newTemplateGenerationStep(llm providers.LLMClient) Step {
	return func(ctx context.Context, state *PipelineState) error {
		// Prepare data for LLM
		parts := preprocessing.PreprocessGenerateTemplateRequest(
			state.AnalyzeRequest.TemplateRequest.Prompt,
			state.AnalyzeRequest.TemplateRequest.BoardType,
		)

		// Call LLM
		resp, err := llm.GenerateTemplate(ctx, parts)
		if err != nil {
			return err
		}

		// Store result in state
		state.TemplateGenerationFlow = resp
		return nil
	}
}

func extractLineElements(elements []models.Element) []models.Element {
	var lineElements []models.Element
	for _, elem := range elements {
		if elem.Type == "line" {
			lineElements = append(lineElements, elem)
		}
	}
	return lineElements
}

func newImageRecognitionStep(llm providers.LLMClient) Step {
	return func(ctx context.Context, state *PipelineState) error {
		parts := []*ai.Part{
			ai.NewTextPart(preprocessing.ImageRecognitionPrompt),
		}

		// Add image if available
		if state.ImageURI != "" {
			parts = append(parts, ai.NewMediaPart("image/jpeg", state.ImageURI))
		}

		resp, err := llm.ImageRecognition(ctx, parts)
		if err != nil {
			slog.Warn("image recognition failed", "err", err)
			return nil
		}
		slog.Info("image recognition result:", "", resp.ImageDescription)

		state.ImageRecognitionFlow = resp
		return nil
	}
}

func newBoardTextExtractionStep() Step {
	return func(ctx context.Context, state *PipelineState) error {
		var elements []models.Element
		switch state.AnalyzeRequest.RequestType {
		case models.SummarizeType:
			elements = state.AnalyzeRequest.SummarizeRequest.Board.Elements
		case models.StructurizeType:
			elements = state.AnalyzeRequest.StructurizeRequest.Board.Elements
		default:
			return nil
		}

		state.BoardTextContent = formatTextElementsContent(elements)
		if state.BoardTextContent != "" {
			slog.Debug("extracted board text content", "chars", len(state.BoardTextContent))
		}
		return nil
	}
}

func formatTextElementsContent(elements []models.Element) string {
	var builder strings.Builder
	for _, elem := range elements {
		if elem.Type != "text" {
			continue
		}
		content := strings.TrimSpace(elem.Content)
		if content == "" {
			continue
		}
		if builder.Len() > 0 {
			builder.WriteString("\n")
		}
		builder.WriteString("- ")
		builder.WriteString(content)
	}
	return builder.String()
}

func newSummarizeStep(llm providers.LLMClient) Step {
	return func(ctx context.Context, state *PipelineState) error {
		userData := preprocessing.PreprocessSummarizeData(
			state.ImageRecognitionFlow.ImageDescription,
			state.DigitalInkText,
			state.SemanticGraph,
		)

		if state.BoardTextContent != "" {
			userData = fmt.Sprintf("%s\n\nBoard text elements detected on the board:\n%s", userData, state.BoardTextContent)
		}

		parts := []*ai.Part{
			ai.NewTextPart(userData),
		}

		// Check for Session History
		sess := session.FromContext[models.BoardSessionState](ctx)
		if sess == nil {
			// No session, call standard
			resp, err := llm.Summarize(ctx, parts)
			if err != nil {
				return err
			}
			state.SummarizeFlow = resp
			return nil
		}
		history := sess.State().Messages
		genkitHistory := models.ToGenkitMessages(history)

		// Call with history
		resp, err := llm.SummarizeWithHistory(ctx, genkitHistory, parts)
		if err != nil {
			return err
		}

		// Update Session
		sessState := sess.State()
		sessState.Messages = append(sessState.Messages,
			models.MessageEntry{Role: "user", Content: userData},
			models.MessageEntry{Role: "model", Content: resp.Summarization},
		)
		if err := sess.UpdateState(ctx, sessState); err != nil {
			slog.Warn("Failed to update session state", "err", err)
		}

		state.SummarizeFlow = resp
		return nil
	}
}

func newFillSummarizeResponseStep() Step {
	return func(ctx context.Context, state *PipelineState) error {
		state.AnalyzeResponse.SummarizeResponse = fillSumRespWithMeta(state.SummarizeFlow, state)
		return nil
	}
}

func fillSumRespWithMeta(summarizeFlow providers.SummarizeFlow, state *PipelineState) models.SummarizeResponse {
	return models.SummarizeResponse{
		RequestID:   state.AnalyzeRequest.SummarizeRequest.RequestID,
		UserID:      state.AnalyzeRequest.SummarizeRequest.UserID,
		RequestType: models.SummarizeType,
		Element:     createTextElementFromSummarization(summarizeFlow.Summarization, state),
	}
}

func createTextElementFromSummarization(summarization string, state *PipelineState) models.Text {
	// Calculate text size based on content
	width, height := utils.CalculateTextSize(summarization)

	// Find free space on the board
	var position utils.TextPosition
	if state.AnalyzeRequest.RequestType == models.SummarizeType {
		position = utils.FindFreeSpace(state.AnalyzeRequest.SummarizeRequest.Board.Elements, width, height)
	} else {
		position = utils.FindFreeSpace(state.AnalyzeRequest.StructurizeRequest.Board.Elements, width, height)
	}

	return models.Text{
		BaseElement: models.BaseElement{
			Id:          "generated-summary",
			Type:        "text",
			X:           position.X,
			Y:           position.Y,
			Width:       position.Width,
			Height:      position.Height,
			Rotation:    0,
			Fill:        "#000000",
			Stroke:      "#000000",
			StrokeWidth: 1,
		},
		Content: summarization,
	}
}

func newStructurizeStep(llm providers.LLMClient) Step {
	return func(ctx context.Context, state *PipelineState) error {
		userData := preprocessing.PreprocessStructurizeData(
			state.ImageRecognitionFlow.ImageDescription,
			state.DigitalInkText,
			state.SemanticGraph,
		)

		if state.BoardTextContent != "" {
			userData = fmt.Sprintf("%s\n\nBoard text elements detected on the board:\n%s", userData, state.BoardTextContent)
		}

		parts := []*ai.Part{
			ai.NewTextPart(userData),
		}

		resp, err := llm.Structurize(ctx, parts)
		if err != nil {
			return err
		}

		state.StructurizeFlow = resp
		return nil
	}
}

func newConvertTemplateToBoardStep() Step {
	return func(ctx context.Context, state *PipelineState) error {
		// Convert TemplateGenerationFlow to models.Board
		board := convertTemplateToBoard(state.TemplateGenerationFlow, state.AnalyzeRequest.TemplateRequest.BoardID)
		state.GeneratedBoard = board
		return nil
	}
}

func convertTemplateToBoard(template providers.TemplateGenerationFlow, boardID string) *models.Board {
	board := &models.Board{
		BoardID:    boardID,
		ImageURL:   "",
		Elements:   make([]models.Element, 0),
		GraphNodes: make([]models.GNode, 0),
		GraphEdges: make([]models.GEdge, 0),
	}

	switch template.BoardType {
	case "simple":
		// Convert template elements to board elements
		for _, elem := range template.Elements {
			board.Elements = append(board.Elements, models.Element{
				Id:           fmt.Sprintf("elem-%d", len(board.Elements)),
				Type:         elem.Type,
				X:            elem.X,
				Y:            elem.Y,
				Width:        elem.Width,
				Height:       elem.Height,
				Rotation:     elem.Rotation,
				Fill:         elem.Fill,
				Stroke:       elem.Stroke,
				StrokeWidth:  elem.StrokeWidth,
				Content:      elem.Content,
				CornerRadius: elem.CornerRadius,
			})
		}
	case "graph":
		// Convert template nodes to graph nodes
		for _, node := range template.GraphNodes {
			board.GraphNodes = append(board.GraphNodes, models.GNode{
				ID:   node.ID,
				Type: node.Type,
				Data: models.GNodeData{
					Title:       node.Title,
					Description: node.Description,
					URL:         node.URL,
				},
			})
		}

		// Convert template edges to graph edges
		for _, edge := range template.GraphEdges {
			board.GraphEdges = append(board.GraphEdges, models.GEdge{
				ID:     edge.ID,
				Source: edge.Source,
				Target: edge.Target,
				Label:  edge.Label,
			})
		}
	}

	return board
}

func newFillTemplateResponseStep() Step {
	return func(ctx context.Context, state *PipelineState) error {
		if state.GeneratedBoard != nil {
			state.AnalyzeResponse.TemplateResponse = models.GenerateTemplateResponse{
				RequestID:   state.AnalyzeRequest.TemplateRequest.RequestID,
				UserID:      state.AnalyzeRequest.TemplateRequest.UserID,
				RequestType: state.AnalyzeRequest.TemplateRequest.RequestType,
				Board:       *state.GeneratedBoard,
			}
		}
		return nil
	}
}

func newFillStructurizeResponseStep() Step {
	return func(ctx context.Context, state *PipelineState) error {
		state.AnalyzeResponse.StructurizeResponse = fillStructRespWithMeta(state.StructurizeFlow, state)
		return nil
	}
}

func fillStructRespWithMeta(flow providers.StructurizeFlow, state *PipelineState) models.StructurizeResponse {
	// Parse ASCII tree into FileHierarchy
	fileHierarchy := utils.ParseASCIITree(flow.AiTreeResponse)

	// Convert FileHierarchy to models.File
	modelFile := utils.ToModelFile(fileHierarchy)

	return models.StructurizeResponse{
		RequestID:      state.AnalyzeRequest.StructurizeRequest.RequestID,
		UserID:         state.AnalyzeRequest.StructurizeRequest.UserID,
		RequestType:    models.StructurizeType,
		AiTreeResponse: flow.AiTreeResponse,
		File:           modelFile,
	}
}

// newIncrementalFullRescanCheckStep handles the case when full rescan is needed
func newIncrementalFullRescanCheckStep() Step {
	return func(ctx context.Context, state *PipelineState) error {
		req := state.AnalyzeRequest.IncrementalRequest

		// If we have cache and don't need full rescan, continue with incremental
		if state.IncrementalCache != nil && !req.IsFullRescan {
			slog.Info("[Pipeline] Using cached analysis, continuing with incremental")
			return nil
		}

		// No cache or full rescan requested - need to do full analysis first
		slog.Info("[Pipeline] No cache or full rescan requested, performing full analysis first")

		if req.FullBoard == nil {
			return fmt.Errorf("full board data required for full rescan")
		}

		// Create a summarize request to do full analysis
		summarizeReq := models.AnalyzeRequest{
			RequestType: models.SummarizeType,
			SummarizeRequest: models.SummarizeRequest{
				RequestID:   req.RequestID,
				UserID:      req.UserID,
				RequestType: models.SummarizeType,
				Board:       *req.FullBoard,
			},
		}

		// Build and execute summarize pipeline
		sumPipeline, err := BuildPipeline(models.SummarizeType, llmForIncremental, state.Provider)
		if err != nil {
			return fmt.Errorf("failed to build summarize pipeline: %w", err)
		}

		sumState := &PipelineState{
			AnalyzeRequest: summarizeReq,
			Provider:       state.Provider,
		}

		if err := sumPipeline.Execute(ctx, sumState); err != nil {
			return fmt.Errorf("full analysis failed: %w", err)
		}

		// Create cache from full analysis result
		regions := createRegionsFromElementsForPipeline(req.FullBoard.Elements)
		boardSize := calculateBoardSize(req.FullBoard.Elements)

		if cacheService != nil {
			cacheService.UpdateCacheAfterFullScan(
				req.BoardID,
				sumState.SummarizeFlow.Summarization,
				[]string{"board-analysis"},
				regions,
				req.FullBoard.Elements,
				req.FullBoard.ImageURL,
				boardSize,
			)

			// Get the cache we just created
			cached, _ := cacheService.GetCache(req.BoardID)
			state.IncrementalCache = cached
		}

		// Set incremental response from summarize result
		state.IncrementalResponse = models.IncrementalAnalysisResponse{
			RequestID:      req.RequestID,
			UserID:         req.UserID,
			RequestType:    models.IncrementalType,
			GlobalSummary:  sumState.SummarizeFlow.Summarization,
			KeyConcepts:    []string{"board-analysis"},
			UpdatedRegions: []string{},
			IsFullRescan:   true,
		}

		slog.Info("[Pipeline] Full rescan completed, returning result")

		// Return special error to short-circuit remaining steps
		return &ErrIncrementalFullRescan{Response: state.IncrementalResponse}
	}
}

// ErrIncrementalFullRescan is returned when full rescan is performed
type ErrIncrementalFullRescan struct {
	Response models.IncrementalAnalysisResponse
}

func (e *ErrIncrementalFullRescan) Error() string {
	return "full rescan performed"
}

// newIncrementalCacheCheckStep checks if we have cached analysis and if full rescan is needed
func newIncrementalCacheCheckStep() Step {
	return func(ctx context.Context, state *PipelineState) error {
		slog.Info("[Pipeline] Incremental cache check started", "boardId", state.AnalyzeRequest.IncrementalRequest.BoardID)

		req := state.AnalyzeRequest.IncrementalRequest

		// Check if full rescan is requested
		if req.IsFullRescan {
			slog.Info("[Pipeline] Full rescan requested, skipping cache check")
			return nil
		}

		// Try to get from cache
		if cacheService == nil {
			slog.Warn("[Pipeline] Cache service not set, skipping cache check")
			return nil
		}

		cached, hasCache := cacheService.GetCache(req.BoardID)
		if !hasCache {
			slog.Info("[Pipeline] Cache miss, will perform full analysis", "boardId", req.BoardID)
			return nil
		}

		// Check if cache needs full rescan
		if cacheService.NeedsFullRescan(cached) {
			slog.Info("[Pipeline] Cache needs full rescan", "boardId", req.BoardID)
			return nil
		}

		state.IncrementalCache = cached
		slog.Info("[Pipeline] Cache hit", "boardId", req.BoardID, "changeCount", cached.ChangeCount)

		return nil
	}
}

// newIncrementalChangeDetectionStep detects changes between cached and current elements
func newIncrementalChangeDetectionStep() Step {
	return func(ctx context.Context, state *PipelineState) error {
		req := state.AnalyzeRequest.IncrementalRequest

		// If no cache or full rescan, skip change detection
		if state.IncrementalCache == nil || req.IsFullRescan {
			slog.Info("[Pipeline] No cache or full rescan, skipping change detection")
			return nil
		}

		if cacheService == nil {
			slog.Warn("[Pipeline] Cache service not set, skipping change detection")
			return nil
		}

		if req.FullBoard == nil {
			return fmt.Errorf("full board data required for incremental analysis")
		}

		slog.Info("[Pipeline] Detecting changes", "boardId", req.BoardID)

		// Detect changes
		changes := cacheService.DetectChanges(
			state.IncrementalCache.ElementHashes,
			cacheService.CalculateElementHashes(req.FullBoard.Elements),
			state.IncrementalCache.Elements,
			req.FullBoard.Elements,
		)

		state.IncrementalChanges = changes
		slog.Info("[Pipeline] Changes detected", "count", len(changes))

		return nil
	}
}

// newIncrementalNoChangesCheckStep checks if there are no changes and returns cached result
func newIncrementalNoChangesCheckStep() Step {
	return func(ctx context.Context, state *PipelineState) error {
		// If no cache or full rescan, skip
		if state.IncrementalCache == nil || state.AnalyzeRequest.IncrementalRequest.IsFullRescan {
			return nil
		}

		// If no changes, return cached result
		if len(state.IncrementalChanges) == 0 {
			slog.Info("[Pipeline] No changes detected, returning cached result")

			state.IncrementalResponse = models.IncrementalAnalysisResponse{
				RequestID:      state.AnalyzeRequest.IncrementalRequest.RequestID,
				UserID:         state.AnalyzeRequest.IncrementalRequest.UserID,
				RequestType:    models.IncrementalType,
				GlobalSummary:  state.IncrementalCache.GlobalSummary,
				KeyConcepts:    state.IncrementalCache.KeyConcepts,
				UpdatedRegions: []string{},
				IsFullRescan:   false,
			}

			// Return special error to short-circuit pipeline
			return &ErrIncrementalNoChanges{Response: state.IncrementalResponse}
		}

		return nil
	}
}

// ErrIncrementalNoChanges is returned when no changes detected and cached result is returned
type ErrIncrementalNoChanges struct {
	Response models.IncrementalAnalysisResponse
}

func (e *ErrIncrementalNoChanges) Error() string {
	return "no changes detected, returning cached result"
}

// newIncrementalRegionDetectionStep finds regions affected by changes
func newIncrementalRegionDetectionStep() Step {
	return func(ctx context.Context, state *PipelineState) error {
		if state.IncrementalCache == nil {
			slog.Info("[Pipeline] No cache, skipping region detection")
			return nil
		}

		slog.Info("[Pipeline] Finding affected regions", "changeCount", len(state.IncrementalChanges))

		affected := make(map[string]models.RegionSummary)

		for _, region := range state.IncrementalCache.Regions {
			for _, change := range state.IncrementalChanges {
				if slices.Contains(region.ElementIDs, change.ElementID) {
					affected[region.ID] = region
					slog.Info("[Pipeline] Region affected", "regionID", region.ID, "elementID", change.ElementID)
				}
			}
		}

		result := make([]models.RegionSummary, 0, len(affected))
		for _, region := range affected {
			result = append(result, region)
		}

		state.IncrementalRegions = result
		slog.Info("[Pipeline] Affected regions found", "count", len(result))

		return nil
	}
}

// newIncrementalImageDownloadStep downloads the board image
func newIncrementalImageDownloadStep() Step {
	return func(ctx context.Context, state *PipelineState) error {
		req := state.AnalyzeRequest.IncrementalRequest

		if req.FullBoard == nil || req.FullBoard.ImageURL == "" {
			slog.Info("[Pipeline] No image URL, skipping download")
			return nil
		}

		slog.Info("[Pipeline] Downloading image", "url", req.FullBoard.ImageURL)

		httpClient := &http.Client{Timeout: 30 * time.Second}
		resp, err := httpClient.Get(req.FullBoard.ImageURL)
		if err != nil {
			slog.Warn("[Pipeline] Failed to download image", "err", err)
			return nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			slog.Warn("[Pipeline] Image download failed", "status", resp.StatusCode)
			return nil
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			slog.Warn("[Pipeline] Failed to read image", "err", err)
			return nil
		}

		state.IncrementalFullImage = data
		slog.Info("[Pipeline] Image downloaded", "size", len(data))

		return nil
	}
}

// newIncrementalImageCropStep crops affected regions from the image
func newIncrementalImageCropStep() Step {
	return func(ctx context.Context, state *PipelineState) error {
		if len(state.IncrementalFullImage) == 0 || len(state.IncrementalRegions) == 0 {
			slog.Info("[Pipeline] No image or regions, skipping crop")
			return nil
		}

		if cropper == nil {
			slog.Warn("[Pipeline] Cropper not set, skipping crop")
			return nil
		}

		slog.Info("[Pipeline] Cropping affected regions", "regionCount", len(state.IncrementalRegions))

		req := state.AnalyzeRequest.IncrementalRequest
		if req.FullBoard == nil {
			return fmt.Errorf("full board data required")
		}

		var crops []string
		var bboxes []models.BoundingBox

		for _, region := range state.IncrementalRegions {
			// Find elements in this region
			var regionElements []models.Element
			for _, elem := range req.FullBoard.Elements {
				if slices.Contains(region.ElementIDs, elem.Id) {
					regionElements = append(regionElements, elem)
				}
			}

			// Crop with element-aware bounding box
			croppedData, croppedBBox, err := cropper.CropWithElements(
				state.IncrementalFullImage,
				region.BBox,
				regionElements,
			)
			if err != nil {
				slog.Warn("[Pipeline] Failed to crop region", "regionID", region.ID, "err", err)
				continue
			}

			crops = append(crops, cropper.ImageToBase64(croppedData))
			bboxes = append(bboxes, croppedBBox)
		}

		state.IncrementalCrops = crops
		state.IncrementalBBoxes = bboxes
		slog.Info("[Pipeline] Cropping completed", "cropCount", len(crops))

		return nil
	}
}

// newIncrementalMergeCropsStep merges overlapping crops
func newIncrementalMergeCropsStep() Step {
	return func(ctx context.Context, state *PipelineState) error {
		if cropper == nil || len(state.IncrementalBBoxes) == 0 {
			return nil
		}

		slog.Info("[Pipeline] Merging overlapping crops", "count", len(state.IncrementalBBoxes))

		merged := cropper.MergeOverlappingCrops(state.IncrementalBBoxes)

		slog.Info("[Pipeline] Crops merged", "original", len(state.IncrementalBBoxes), "merged", len(merged))

		state.IncrementalBBoxes = merged
		return nil
	}
}

// newIncrementalFallbackCheckStep checks if we should fallback to full scan
func newIncrementalFallbackCheckStep() Step {
	return func(ctx context.Context, state *PipelineState) error {
		if cropper == nil || state.IncrementalCache == nil {
			return nil
		}

		slog.Info("[Pipeline] Checking fallback condition")

		if cropper.ShouldFallbackToFullScan(state.IncrementalBBoxes, state.IncrementalCache.BoardSize) {
			slog.Info("[Pipeline] Fallback to full scan triggered")
			return &ErrIncrementalFallback{Message: "crops cover too much area"}
		}

		slog.Info("[Pipeline] No fallback needed")
		return nil
	}
}

// ErrIncrementalFallback is returned when incremental analysis should fallback to full scan
type ErrIncrementalFallback struct {
	Message string
}

func (e *ErrIncrementalFallback) Error() string {
	return e.Message
}

// newIncrementalPromptBuildStep builds the prompt for LLM
func newIncrementalPromptBuildStep() Step {
	return func(ctx context.Context, state *PipelineState) error {
		slog.Info("[Pipeline] Building incremental prompt")
		// Prompt will be built in LLM step with all context
		return nil
	}
}

// newIncrementalLLMAnalysisStep calls LLM for incremental analysis
func newIncrementalLLMAnalysisStep(llm providers.LLMClient) Step {
	return func(ctx context.Context, state *PipelineState) error {
		slog.Info("[Pipeline] Calling LLM for incremental analysis")

		cached := state.IncrementalCache
		if cached == nil {
			return fmt.Errorf("cache is required for incremental analysis")
		}

		// Build text prompt
		var prompt strings.Builder
		prompt.WriteString("Ты анализируешь whiteboard.\n\n")
		prompt.WriteString("Текущий анализ всей доски:\n")
		prompt.WriteString(cached.GlobalSummary)
		prompt.WriteString("\n\nКлючевые концепции: ")
		prompt.WriteString(strings.Join(cached.KeyConcepts, ", "))
		prompt.WriteString("\n\nПроизошли изменения:\n")

		for _, change := range state.IncrementalChanges {
			prompt.WriteString(fmt.Sprintf("- Элемент %s: %s\n", change.ElementID, change.ChangeType))
		}

		if len(state.IncrementalCrops) > 0 {
			prompt.WriteString(fmt.Sprintf("\nИзмененные области: %d кропов изображений\n", len(state.IncrementalCrops)))
		}

		// Build parts for LLM
		parts := []*ai.Part{ai.NewTextPart(prompt.String())}

		// Add image crops
		for _, cropBase64 := range state.IncrementalCrops {
			parts = append(parts, ai.NewMediaPart("image/jpeg", "data:image/jpeg;base64,"+cropBase64))
		}

		// Add full board image if available
		if len(state.IncrementalFullImage) > 0 && cropper != nil {
			fullBase64 := cropper.ImageToBase64(state.IncrementalFullImage)
			parts = append(parts, ai.NewMediaPart("image/jpeg", "data:image/jpeg;base64,"+fullBase64))
		}

		slog.Info("[Pipeline] LLM call with parts", "count", len(parts))

		// Call LLM
		response, err := llm.Summarize(ctx, parts)
		if err != nil {
			return fmt.Errorf("LLM analysis failed: %w", err)
		}

		slog.Info("[Pipeline] LLM response received", "length", len(response.Summarization))

		// Try to parse as JSON
		type LLMResponse struct {
			GlobalSummary string   `json:"globalSummary"`
			KeyConcepts   []string `json:"keyConcepts"`
		}

		var result LLMResponse
		err = json.Unmarshal([]byte(response.Summarization), &result)
		if err != nil {
			slog.Warn("[Pipeline] Failed to parse LLM response as JSON, using as plain text", "err", err)
			result.GlobalSummary = response.Summarization
			result.KeyConcepts = cached.KeyConcepts
		}

		// Calculate new board size
		req := state.AnalyzeRequest.IncrementalRequest
		var newBoardSize models.BoundingBox
		if req.FullBoard != nil {
			newBoardSize = calculateBoardSize(req.FullBoard.Elements)
		}

		// Update cache
		if cacheService != nil {
			cacheService.UpdateCacheAfterIncremental(
				cached,
				result.GlobalSummary,
				result.KeyConcepts,
				affectedRegionIDs(state.IncrementalRegions),
				req.FullBoard.Elements,
				newBoardSize,
			)
		}

		slog.Info("[Pipeline] Cache updated")

		return nil
	}
}

// calculateBoardSize calculates board size from elements
func calculateBoardSize(elements []models.Element) models.BoundingBox {
	if len(elements) == 0 {
		return models.BoundingBox{W: 2000, H: 2000}
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

	return models.BoundingBox{
		X: 0,
		Y: 0,
		W: maxX + 100,
		H: maxY + 100,
	}
}

// affectedRegionIDs extracts region IDs
func affectedRegionIDs(regions []models.RegionSummary) []string {
	ids := make([]string, len(regions))
	for i, region := range regions {
		ids[i] = region.ID
	}
	return ids
}

// createRegionsFromElementsForPipeline creates regions from board elements (for pipeline use)
func createRegionsFromElementsForPipeline(elements []models.Element) []models.RegionSummary {
	boardSize := calculateBoardSize(elements)
	gridW := boardSize.W / 3
	gridH := boardSize.H / 3

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
	}

	return regions
}

// newIncrementalCacheUpdateStep updates cache after LLM analysis
func newIncrementalCacheUpdateStep() Step {
	return func(ctx context.Context, state *PipelineState) error {
		// Cache is already updated in LLM step
		return nil
	}
}

// newIncrementalResponseStep fills the incremental response
func newIncrementalResponseStep() Step {
	return func(ctx context.Context, state *PipelineState) error {
		req := state.AnalyzeRequest.IncrementalRequest
		cached := state.IncrementalCache

		if cached == nil {
			// Full rescan case - response should be filled by full analysis
			return nil
		}

		// Try to get response from LLM result
		// Since we store in cache, we can get it from there
		state.IncrementalResponse = models.IncrementalAnalysisResponse{
			RequestID:      req.RequestID,
			UserID:         req.UserID,
			RequestType:    models.IncrementalType,
			GlobalSummary:  cached.GlobalSummary,
			KeyConcepts:    cached.KeyConcepts,
			UpdatedRegions: affectedRegionIDs(state.IncrementalRegions),
			IsFullRescan:   false,
		}

		slog.Info("[Pipeline] Incremental response filled")
		return nil
	}
}
