package providers

import (
	"context"
	"fmt"
	"sync"

	"github.com/aiservice/internal/models"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
)

type LLMClient interface {
	Structurize(ctx context.Context, parts []*ai.Part) (models.StructurizeResponse, error)
	Summarize(ctx context.Context, parts []*ai.Part) (models.SummarizeResponse, error)
	GetName() string // Added for provider identification
}

type SummarizeFlow struct {
	Prompt  string      `json:"userPrompt"`
	Element models.Text `json:"element"`
}

// For the recursive structure, we'll use a different approach that doesn't trigger schema generation
// We'll use a direct call to the LLM instead of Genkit's flow system for the recursive structure

var summarizeFlowInstance *core.Flow[*SummarizeFlow, *SummarizeFlow, struct{}]
var summarizeFlowOnce sync.Once

func GetSummarizeFlow(gkit *genkit.Genkit) *core.Flow[*SummarizeFlow, *SummarizeFlow, struct{}] {
	summarizeFlowOnce.Do(func() {
		summarizeFlowInstance = genkit.DefineFlow(gkit, "summarize flow", func(ctx context.Context, input *SummarizeFlow) (*SummarizeFlow, error) {
			// Note: This flow is not meant to be run directly, use GenerateData instead
			return nil, fmt.Errorf("this flow is not meant to be run directly, use GenerateData instead")
		})
	})
	return summarizeFlowInstance
}

func RunSummarizeGeneration(ctx context.Context, gkit *genkit.Genkit, parts []*ai.Part) (*SummarizeFlow, error) {
	prompt := ai.NewUserMessage(parts...)
	resp, _, err := genkit.GenerateData[SummarizeFlow](ctx, gkit, ai.WithMessages(prompt))
	if err != nil {
		return nil, fmt.Errorf("failed to generate llm request: %w", err)
	}
	return resp, nil
}

// For structurize, we'll define a flow that doesn't use the recursive File structure in its definition
// to avoid schema generation issues, but the LLM will be instructed to return the proper structure
type SimpleStructurizeFlow struct {
	Prompt         string `json:"userPrompt"`
	Answer         string `json:"answer"`
	AiTreeResponse string `json:"aiTreeResponse"`
	FileJSON       string `json:"file"` // JSON string representation to avoid recursive schema
}

var structurizeFlowInstance *core.Flow[*SimpleStructurizeFlow, *SimpleStructurizeFlow, struct{}]
var structurizeFlowOnce sync.Once

func GetStructurizeFlow(gkit *genkit.Genkit) *core.Flow[*SimpleStructurizeFlow, *SimpleStructurizeFlow, struct{}] {
	structurizeFlowOnce.Do(func() {
		structurizeFlowInstance = genkit.DefineFlow(gkit, "structurize flow", func(ctx context.Context, input *SimpleStructurizeFlow) (*SimpleStructurizeFlow, error) {
			// Note: This flow is not meant to be run directly, use GenerateData instead
			return nil, fmt.Errorf("this flow is not meant to be run directly, use GenerateData instead")
		})
	})
	return structurizeFlowInstance
}

func RunStructurizeGeneration(ctx context.Context, gkit *genkit.Genkit, parts []*ai.Part) (*SimpleStructurizeFlow, error) {
	prompt := ai.NewUserMessage(parts...)
	resp, _, err := genkit.GenerateData[SimpleStructurizeFlow](ctx, gkit, ai.WithMessages(prompt))
	if err != nil {
		return nil, fmt.Errorf("failed to generate llm request: %w", err)
	}
	return resp, nil
}
