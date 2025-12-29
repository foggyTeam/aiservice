package pipeline

import (
	"context"
	"fmt"

	"github.com/aiservice/internal/models"
	"github.com/aiservice/internal/providers"
	"github.com/firebase/genkit/go/ai"
)

func newLlmParts(req models.AnalyzeRequest) ([]*ai.Part, error) {
	parts := make([]*ai.Part, 0, 3)
	if url := req.ImageInput.ImageURL; url != "" {
		parts = append(parts, ai.NewMediaPart("image/jpeg", url))
	}
	if text := req.TextInput.Text; text != "" {
		parts = append(parts, ai.NewTextPart(text))
	}
	if req.InkInput.Strokes != nil {
		// TODO convert ink to image or other format
	}
	if len(parts) == 0 {
		return nil, fmt.Errorf("no valid input parts for LLM")
	}
	return parts, nil
}

func fileStructureStep(ink providers.InkRecognizer, llm providers.LLMClient) Step {
	return func(ctx context.Context, ps *PipelineState) error {
		return nil
	}
}

func complexAnalyzeStep(_ providers.InkRecognizer, llm providers.LLMClient) Step {
	return func(ctx context.Context, ps *PipelineState) error {
		parts, err := newLlmParts(ps.Request)
		if err != nil {
			return err
		}
		resp, err := llm.Analyze(ctx, parts)
		if err != nil {
			return err
		}
		ps.Response.ResponseMessage = resp.ResponseMessage
		ps.Response.FileStructure = resp.FileStructure
		ps.Response.GraphResponse = resp.GraphResponse
		return nil
	}
}
