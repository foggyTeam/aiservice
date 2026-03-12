package pipeline

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"os"

	"github.com/aiservice/internal/digitalink"
	"github.com/aiservice/internal/models"
	"github.com/aiservice/internal/preprocessing"
	"github.com/aiservice/internal/providers"
	"github.com/aiservice/internal/services/graph"
	"github.com/aiservice/internal/services/image"
	"github.com/aiservice/internal/utils"
	"github.com/firebase/genkit/go/ai"
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

// SetImageService sets the image service for the pipeline
func SetImageService(svc *image.Service) {
	imageService = svc
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

func newImageCleanupStep() Step {
	return func(ctx context.Context, state *PipelineState) error {
		// Cleanup image after processing (only for Ollama)
		if state.ImageDownloadResult != nil && imageService != nil {
			if err := imageService.DeleteImage(state.ImageDownloadResult.LocalPath); err != nil {
				slog.Warn("failed to cleanup image", "err", err)
			}
		}
		return nil
	}
}

func newDigitalInkAnalysisStep() Step {
	return func(ctx context.Context, state *PipelineState) error {
		// Only process digital ink for summarization requests
		if state.AnalyzeRequest.RequestType != models.SummarizeType {
			return nil
		}

		// Extract line elements for handwriting recognition
		lineElements := extractLineElements(state.AnalyzeRequest.SummarizeRequest.Board.Elements)

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

		if state.AnalyzeRequest.RequestType == models.SummarizeType {
			graphNodes = state.AnalyzeRequest.SummarizeRequest.Board.GraphNodes
			graphEdges = state.AnalyzeRequest.SummarizeRequest.Board.GraphEdges
		} else if state.AnalyzeRequest.RequestType == models.StructurizeType {
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

		state.ImageRecognitionFlow = resp
		return nil
	}
}

func newSummarizeStep(llm providers.LLMClient) Step {
	return func(ctx context.Context, state *PipelineState) error {
		// Prepare dynamic data for user message using preprocessor
		userData := preprocessing.PreprocessSummarizeData(
			state.ImageRecognitionFlow.ImageDescription,
			state.DigitalInkText,
			state.SemanticGraph,
		)

		parts := []*ai.Part{
			ai.NewTextPart(userData),
		}

		resp, err := llm.Summarize(ctx, parts)
		if err != nil {
			return err
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
		// Prepare dynamic data for user message using preprocessor
		userData := preprocessing.PreprocessStructurizeData(
			state.ImageRecognitionFlow.ImageDescription,
			state.DigitalInkText,
			state.SemanticGraph,
		)

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

	if template.BoardType == "simple" {
		// Convert template elements to board elements
		for _, elem := range template.Elements {
			board.Elements = append(board.Elements, models.Element{
				Id:           fmt.Sprintf("elem-%s", elem.Type),
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
	} else if template.BoardType == "graph" {
		// Convert template nodes to graph nodes
		for _, node := range template.GraphNodes {
			board.GraphNodes = append(board.GraphNodes, models.GNode{
				ID:   node.ID,
				Type: node.Type,
				Position: models.GPosition{
					X: node.PositionX,
					Y: node.PositionY,
				},
				Data: models.GNodeData{
					Title:       node.Title,
					Description: node.Description,
					Shape:       node.Shape,
					Color:       node.Color,
					URL:         node.URL,
				},
			})
		}

		// Convert template edges to graph edges
		for _, edge := range template.GraphEdges {
			board.GraphEdges = append(board.GraphEdges, models.GEdge{
				ID:     edge.ID,
				Type:   edge.Type,
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
