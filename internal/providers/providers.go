package providers

import (
	"context"
	"fmt"

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

type StructurizeFlow struct {
	Prompt         string      `json:"userPrompt"`
	Answer         string      `json:"answer"`
	AiTreeResponse string      `json:"aiTreeResponse"`
	File           models.File `json:"file"`
}

func DefineSummarizeFlow(gkit *genkit.Genkit, parts []*ai.Part) *core.Flow[*SummarizeFlow, *SummarizeFlow, struct{}] {
	return genkit.DefineFlow(gkit, "summarize flow", func(ctx context.Context, input *SummarizeFlow) (*SummarizeFlow, error) {
		prompt := ai.NewUserMessage(parts...)
		resp, _, err := genkit.GenerateData[SummarizeFlow](ctx, gkit, ai.WithMessages(prompt))
		if err != nil {
			return nil, fmt.Errorf("failed to generate llm request flow: %w", err)
		}
		return resp, nil
	})
}

func DefineStructurizeFlow(gkit *genkit.Genkit, parts []*ai.Part) *core.Flow[*StructurizeFlow, *StructurizeFlow, struct{}] {
	return genkit.DefineFlow(gkit, "structurize flow", func(ctx context.Context, input *StructurizeFlow) (*StructurizeFlow, error) {
		prompt := ai.NewUserMessage(parts...)
		resp, _, err := genkit.GenerateData[StructurizeFlow](ctx, gkit, ai.WithMessages(prompt))
		if err != nil {
			return nil, fmt.Errorf("failed to generate llm request flow: %w", err)
		}
		return resp, nil
	})
}
