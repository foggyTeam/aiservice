package providers

import (
	"context"

	"github.com/aiservice/internal/models"
)

type InkRecognizer interface {
	RecognizeInk(ctx context.Context, input models.InkInput) (models.TranscriptionResult, error)
	RecognizeImage(ctx context.Context, input models.ImageInput) (models.TranscriptionResult, error)
}

type LLMClient interface {
	Analyze(ctx context.Context, transcription, contextData string) (models.AnalyzeResponse, error)
}
