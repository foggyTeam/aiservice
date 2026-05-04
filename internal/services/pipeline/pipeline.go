package pipeline

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aiservice/internal/models"
	"github.com/aiservice/internal/providers"
	"github.com/aiservice/internal/services/image"
)

type PipelineState struct {
	AnalyzeRequest         models.AnalyzeRequest
	AnalyzeResponse        models.AnalyzeResponse
	DigitalInkText         string
	ImageURI               string
	ImageDownloadResult    *image.DownloadResult
	Provider               string
	ImageRecognitionFlow   providers.ImageRecognitionFlow
	SummarizeFlow          providers.SummarizeFlow
	StructurizeFlow        providers.StructurizeFlow
	SemanticGraph          *models.SemanticGraph
	TemplateGenerationFlow providers.TemplateGenerationFlow
	GeneratedBoard         *models.Board

	BoardTextContent string

	// Incremental analysis fields
	IncrementalCache     *models.BoardAnalysisCache
	IncrementalChanges   []models.ElementChange
	IncrementalRegions   []models.RegionSummary
	IncrementalCrops     []string
	IncrementalBBoxes    []models.BoundingBox
	IncrementalFullImage []byte
	IncrementalResponse  models.IncrementalAnalysisResponse
}

type Step func(ctx context.Context, state *PipelineState) error

type Pipeline struct {
	steps []Step
}

func NewPipeline(steps ...Step) *Pipeline {
	return &Pipeline{steps: steps}
}

func (p *Pipeline) Execute(ctx context.Context, state *PipelineState) error {
	for _, step := range p.steps {
		if err := step(ctx, state); err != nil {
			return err
		}
	}
	return nil
}

func BuildPipeline(t string, llm providers.LLMClient, provider string) (*Pipeline, error) {
	switch t {
	case models.SummarizeType:
		return NewPipeline(
			newImageDownloadStep(),
			newDigitalInkAnalysisStep(),
			newBoardTextExtractionStep(),
			newGraphPreprocessingStep(),
			newImageRecognitionStep(llm),
			newSummarizeStep(llm),
			newFillSummarizeResponseStep(),
		), nil
	case models.StructurizeType:
		return NewPipeline(
			newImageDownloadStep(),
			newDigitalInkAnalysisStep(),
			newBoardTextExtractionStep(),
			newGraphPreprocessingStep(),
			newImageRecognitionStep(llm),
			newStructurizeStep(llm),
			newFillStructurizeResponseStep(),
		), nil
	case models.GenerateTemplateType:
		return NewPipeline(
			newTemplateGenerationStep(llm),
			newConvertTemplateToBoardStep(),
			newFillTemplateResponseStep(),
		), nil
	case models.GenerateTextType:
		return NewPipeline(
			newTextGenerationStep(llm),
			newFillTextResponseStep(),
		), nil
	case models.IncrementalType:
		return NewPipeline(
			newIncrementalCacheCheckStep(),
			newIncrementalFullRescanCheckStep(),
			newIncrementalChangeDetectionStep(),
			newIncrementalNoChangesCheckStep(),
			newIncrementalRegionDetectionStep(),
			newIncrementalImageDownloadStep(),
			newIncrementalImageCropStep(),
			newIncrementalMergeCropsStep(),
			newIncrementalFallbackCheckStep(),
			newIncrementalPromptBuildStep(),
			newIncrementalLLMAnalysisStep(llm),
			newIncrementalCacheUpdateStep(),
			newIncrementalResponseStep(),
		), nil
	default:
		return nil, fmt.Errorf("unsupported input type: %s", t)
	}
}

func BuildContextData(ctxMap map[string]any) string {
	if ctxMap == nil {
		return ""
	}
	data, _ := json.Marshal(ctxMap)
	return string(data)
}
