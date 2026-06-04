package pipeline

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/aiservice/internal/models"
	"github.com/aiservice/internal/preprocessing"
	"github.com/aiservice/internal/providers"
	"github.com/aiservice/internal/services/graph"
	"github.com/aiservice/internal/services/image"
	"github.com/aiservice/internal/utils"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core/x/session"
)

// DigitalInkClient for recognizing handwriting
var digitalInkClient providers.DigitalInkRecognizer

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

// SetDigitalInkClient sets the digital ink client for the pipeline
func SetDigitalInkClient(client providers.DigitalInkRecognizer) {
	digitalInkClient = client
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
		// For remote providers use URL directly
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
			slog.Info("no line elements found for digital ink analysis")
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

// newTextGenerationStep generates text from a prompt
func newTextGenerationStep(llm providers.LLMClient) Step {
	return func(ctx context.Context, state *PipelineState) error {
		parts := preprocessing.PreprocessGenerateTextRequest(state.AnalyzeRequest.TemplateRequest.Prompt)
		slog.Info("generateText", "start", "")
		resp, err := llm.GenerateText(ctx, parts)
		if err != nil {
			return err
		}
		slog.Info("generateText", "end. response", resp)
		state.BoardTextContent = resp
		return nil
	}
}

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

		if err := updateSession(ctx, sess, sessionMsgs(userData, resp.Summarization), 6); err != nil {
			slog.Warn("Failed to update session state", "err", err)
		}

		state.SummarizeFlow = resp
		return nil
	}
}

func updateSession(ctx context.Context, sess *session.Session[models.BoardSessionState], messages []models.MessageEntry, maxChatSize int) error {
	sessState := sess.State()
	sessState.Messages = models.ShiftAndAppendLimited(sessState.Messages, messages, maxChatSize)
	if err := sess.UpdateState(ctx, sessState); err != nil {
		return fmt.Errorf("failed to update session state: %w", err)
	}
	return nil
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
			state.AnalyzeRequest.StructurizeRequest.File,
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
			resp, err := llm.Structurize(ctx, parts)
			if err != nil {
				return err
			}
			state.StructurizeFlow = resp
			return nil
		}

		history := sess.State().Messages
		genkitHistory := models.ToGenkitMessages(history)
		resp, err := llm.StructurizeWithHistory(ctx, genkitHistory, parts)
		if err != nil {
			return err
		}

		if err := updateSession(ctx, sess, sessionMsgs(userData, resp.AiTreeResponse), 6); err != nil {
			slog.Warn("Failed to update session state", "err", err)
		}

		state.StructurizeFlow = resp
		return nil
	}
}

func sessionMsgs(userData string, resp string) []models.MessageEntry {
	return []models.MessageEntry{
		{Role: "user", Content: userData},
		{Role: "model", Content: resp},
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

func newFillTextResponseStep() Step {
	return func(ctx context.Context, state *PipelineState) error {
		state.AnalyzeResponse.TextResponse = models.TextResponse{
			RequestID:   state.AnalyzeRequest.TemplateRequest.RequestID,
			UserID:      state.AnalyzeRequest.TemplateRequest.UserID,
			RequestType: state.AnalyzeRequest.TemplateRequest.RequestType,
			Content:     state.BoardTextContent,
		}
		return nil
	}
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
