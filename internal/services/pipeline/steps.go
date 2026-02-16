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
	"github.com/aiservice/internal/services/image"
	"github.com/aiservice/internal/utils"
	"github.com/firebase/genkit/go/ai"
)

// Preprocessor for transforming raw data into structured formats
var preprocessor = preprocessing.NewPreprocessor()

// DigitalInkClient for recognizing handwriting
var digitalInkClient = digitalink.NewClient("en", 0)

// ImageService for downloading and cleaning up images
var imageService *image.Service

// SetImageService sets the image service for the pipeline
func SetImageService(svc *image.Service) {
	imageService = svc
}

func newLlmSummarizeParts(req models.SummarizeRequest, recognizedText string, imageURI string) ([]*ai.Part, error) {
	return preprocessor.PreprocessSummarizeRequest(req, recognizedText, imageURI)
}

func newLlmStructurizeParts(req models.StructurizeRequest) ([]*ai.Part, error) {
	return preprocessor.PreprocessStructurizeRequest(req)
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

		// Recognize handwriting
		recognizedText, err := digitalInkClient.RecognizeInk(ctx, lineElements)
		if err != nil {
			slog.Warn("digital ink recognition failed", "err", err)
			// Continue without recognized text - don't fail the entire pipeline
			return nil
		}

		// Store recognized text in state for later use
		if recognizedText != "" {
			state.DigitalInkText = recognizedText
			slog.Debug("digital ink recognized", "text_length", len(recognizedText))
		}

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

func newSummarizeStep(llm providers.LLMClient) Step {
	return func(ctx context.Context, state *PipelineState) error {
		parts, err := newLlmSummarizeParts(
			state.AnalyzeRequest.SummarizeRequest,
			state.DigitalInkText,
			state.ImageURI,
		)
		if err != nil {
			return err
		}
		resp, err := llm.Summarize(ctx, parts)
		if err != nil {
			return err
		}
		// Store the summarization result in state
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
		parts, err := newLlmStructurizeParts(state.AnalyzeRequest.StructurizeRequest)
		if err != nil {
			return err
		}
		resp, err := llm.Structurize(ctx, parts)
		if err != nil {
			return err
		}
		// Store the structurization result in state
		state.StructurizeFlow = resp
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
	// Convert FileHierarchy to models.File using existing ToModelFile method
	modelFile := flow.File.ToModelFile()

	return models.StructurizeResponse{
		RequestID:      state.AnalyzeRequest.StructurizeRequest.RequestID,
		UserID:         state.AnalyzeRequest.StructurizeRequest.UserID,
		RequestType:    models.StructurizeType,
		AiTreeResponse: flow.AiTreeResponse,
		File:           modelFile,
	}
}
