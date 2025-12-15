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

func DefineFlow(gkit *genkit.Genkit, transcription, contextData string) *core.Flow[*LlmRequestFlow, *LlmRequestFlow, struct{}] {
	return genkit.DefineFlow(gkit, "text analyzing flow", func(ctx context.Context, input *LlmRequestFlow) (*LlmRequestFlow, error) {
		prompt := fmt.Sprintf(`
		Analyze the following transcription and context data:
			Transcription: %s
			Context Data: %s	
		`, transcription, contextData)

		result, _, err := genkit.GenerateData[LlmRequestFlow](ctx, gkit, ai.WithPrompt(prompt))
		if err != nil {
			return nil, fmt.Errorf("failed to generate llm request flow: %w", err)
		}
		return result, nil
	})
}

type LlmAsnwerFlow struct {
	Question    string
	ContextData string
	Answer      string
}

func DefineSimpleAnswerFlow(gkit *genkit.Genkit, transcription, contextData string) *core.Flow[*LlmAsnwerFlow, *LlmAsnwerFlow, struct{}] {
	return genkit.DefineFlow(gkit, "text answer flow", func(ctx context.Context, input *LlmAsnwerFlow) (*LlmAsnwerFlow, error) {
		prompt := fmt.Sprintf(`
		Analyze the following transcription and context data:
			Transcription: %s
			Context Data: %s	
		Give me a simple answer.
		`, transcription, contextData)

		result, _, err := genkit.GenerateData[LlmAsnwerFlow](ctx, gkit, ai.WithPrompt(prompt))
		if err != nil {
			return nil, fmt.Errorf("failed to generate llm request flow: %w", err)
		}
		return result, nil
	})
}

type LlmGraphFlow struct {
	Question    string
	ContextData string
	GraphData   string
}

func DefineGraphFlow(gkit *genkit.Genkit, transcription, contextData string) *core.Flow[*LlmGraphFlow, *LlmGraphFlow, struct{}] {
	return genkit.DefineFlow(gkit, "graph flow", func(ctx context.Context, input *LlmGraphFlow) (*LlmGraphFlow, error) {
		resp, _, err := genkit.GenerateData[LlmGraphFlow](ctx, gkit,
			ai.WithMessages(
				ai.NewUserMessage(
					ai.NewMediaPart("image/jpeg", "https://example.com/photo.jpg"),
					ai.NewTextPart("Compose a poem about this image."),
				),
			),
		)
		_ = resp
		prompt := fmt.Sprintf(`
		Analyze the following transcription and context data:
			Transcription: %s
			Context Data: %s	
		Give me a simple answer.
		`, transcription, contextData)

		result, _, err := genkit.GenerateData[LlmGraphFlow](ctx, gkit, ai.WithPrompt(prompt))
		if err != nil {
			return nil, fmt.Errorf("failed to generate llm request flow: %w", err)
		}
		return result, nil
	})
}
