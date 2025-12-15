package providers

import (
	"context"
	"fmt"

	"github.com/aiservice/internal/models"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
)

type InkRecognizer interface {
	RecognizeInk(ctx context.Context, input models.InkInput) (models.TranscriptionResult, error)
}

type LLMClient interface {
	Analyze(ctx context.Context, parts []*ai.Part) (models.AnalyzeResponse, error)
}

type AnalyzeFlow struct {
	UserPrompt    string               `json:"userPrompt"`
	Answer        string               `json:"answer"`
	Graph         string               `json:"graph,omitempty"`
	FileStructure models.FileStructure `json:"fileStructure"`
}

func DefineAnalyzeFlow(gkit *genkit.Genkit, parts []*ai.Part) *core.Flow[*AnalyzeFlow, *AnalyzeFlow, struct{}] {
	return genkit.DefineFlow(gkit, "analyze flow", func(ctx context.Context, input *AnalyzeFlow) (*AnalyzeFlow, error) {
		prompt := ai.NewUserMessage(parts...)
		resp, _, err := genkit.GenerateData[AnalyzeFlow](ctx, gkit, ai.WithMessages(prompt))
		if err != nil {
			return nil, fmt.Errorf("failed to generate llm request flow: %w", err)
		}
		return resp, nil
	})
}
