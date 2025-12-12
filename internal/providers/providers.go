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
	RecognizeImage(ctx context.Context, input models.ImageInput) (models.TranscriptionResult, error)
}

type LLMClient interface {
	Analyze(ctx context.Context, transcription, contextData string) (models.AnalyzeResponse, error)
}

type LlmRequestFlow struct {
}

func DefineFlow(gkit *genkit.Genkit) *core.Flow[*LlmRequestFlow, *LlmRequestFlow, struct{}] {
	return genkit.DefineFlow(gkit, "text analyzing flow", func(ctx context.Context, input *LlmRequestFlow) (*LlmRequestFlow, error) {
		prompt := ""
		result, _, err := genkit.GenerateData[LlmRequestFlow](ctx, gkit, ai.WithPrompt(prompt))
		if err != nil {
			return nil, fmt.Errorf("failed to generate recipe: %w", err)
		}
		return result, nil
	})
}
