package pipeline

import (
	"context"

	"github.com/aiservice/internal/models"
	"github.com/aiservice/internal/preprocessing"
	"github.com/aiservice/internal/providers"
	"github.com/firebase/genkit/go/ai"
)

// Preprocessor for transforming raw data into structured formats
var preprocessor = preprocessing.NewPreprocessor()

func newLlmSummarizeParts(req models.SummarizeRequest) ([]*ai.Part, error) {
	return preprocessor.PreprocessSummarizeRequest(req)
}

func newLlmStructurizeParts(req models.StructurizeRequest) ([]*ai.Part, error) {
	return preprocessor.PreprocessStructurizeRequest(req)
}

func newSummarizeStep(llm providers.LLMClient) Step {
	return func(ctx context.Context, state *PipelineState) error {
		parts, err := newLlmSummarizeParts(state.AnalyzeRequest.SummarizeRequest)
		if err != nil {
			return err
		}
		resp, err := llm.Summarize(ctx, parts)
		if err != nil {
			return err
		}
		state.AnalyzeResponse.SummarizeResponse = fillSumRespWithMeta(resp, state)
		return nil
	}
}

func fillSumRespWithMeta(aiResp models.SummarizeResponse, state *PipelineState) models.SummarizeResponse {
	return models.SummarizeResponse{
		RequestID:   state.AnalyzeRequest.SummarizeRequest.RequestID,
		UserID:      state.AnalyzeRequest.SummarizeRequest.UserID,
		RequestType: models.SummarizeType,
		Element:     aiResp.Element,
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
		state.AnalyzeResponse.StructurizeResponse = fillStructRespWithMeta(resp, state)
		return nil
	}
}

func fillStructRespWithMeta(aiResp models.StructurizeResponse, state *PipelineState) models.StructurizeResponse {
	return models.StructurizeResponse{
		RequestID:      state.AnalyzeRequest.StructurizeRequest.RequestID,
		UserID:         state.AnalyzeRequest.StructurizeRequest.UserID,
		RequestType:    models.StructurizeType,
		AiTreeResponse: aiResp.AiTreeResponse,
		File:           aiResp.File,
	}
}
