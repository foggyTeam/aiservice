package pipeline

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aiservice/internal/models"

	"github.com/aiservice/internal/providers"
)

type PipelineState struct {
	AnalyzeRequest  models.AnalyzeRequest
	AnalyzeResponse models.AnalyzeResponse
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

func BuildPipeline(t string, llm providers.LLMClient) (*Pipeline, error) {
	switch t {
	case models.SummarizeType:
		return NewPipeline(newSummarizeStep(llm)), nil
	case models.StructurizeType:
		return NewPipeline(newStructurizeStep(llm)), nil
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
