package pipeline

import (
	"context"

	"github.com/aiservice/internal/models"
	"github.com/aiservice/internal/providers"
	"github.com/firebase/genkit/go/ai"
)

func newLlmSummarizeParts(req models.SummarizeRequest) ([]*ai.Part, error) {
	parts := make([]*ai.Part, 0, 3)
	if url := req.Board.ImageURL; url != "" {
		parts = append(parts, ai.NewMediaPart("image/jpeg", url))
	}
	parts = append(parts, ai.NewTextPart("summarize"))
	return parts, nil
}

func newLlmStructurizeParts(req models.StucturizeRequest) ([]*ai.Part, error) {
	parts := make([]*ai.Part, 0, 3)
	if url := req.Board.ImageURL; url != "" {
		parts = append(parts, ai.NewMediaPart("image/jpeg", url))
	}
	parts = append(parts, ai.NewTextPart("structurize"))
	return parts, nil
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
		state.AnalyzeResponse.SummarizeResponse = resp
		return nil
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
		state.AnalyzeResponse.StructurizeResponse = resp
		return nil
	}
}
