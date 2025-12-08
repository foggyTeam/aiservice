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

type StubInkRecognizer struct{}

func (s *StubInkRecognizer) RecognizeInk(ctx context.Context, input models.InkInput) (models.TranscriptionResult, error) {
	pointCount := 0
	for _, stroke := range input.Strokes {
		pointCount += len(stroke)
	}
	return models.TranscriptionResult{
		Text:     "[Stub] Recognized handwriting",
		Language: "en",
		Metadata: map[string]interface{}{
			"stroke_count": len(input.Strokes),
			"point_count":  pointCount,
		},
	}, nil
}

func (s *StubInkRecognizer) RecognizeImage(ctx context.Context, input models.ImageInput) (models.TranscriptionResult, error) {
	return models.TranscriptionResult{
		Text:     "[Stub] Recognized image text",
		Language: "en",
		Metadata: map[string]interface{}{
			"image_url": input.ImageURL,
		},
	}, nil
}

type StubLLMClient struct{}

func (s *StubLLMClient) Analyze(ctx context.Context, transcription, contextData string) (models.AnalyzeResponse, error) {
	return models.AnalyzeResponse{
		Intent:     "create_sticky",
		Confidence: 0.87,
		Actions: []models.Action{
			{
				Type: "create_sticky",
				Payload: map[string]interface{}{
					"text":     transcription,
					"x":        200,
					"y":        300,
					"color":    "#FFEEAA",
					"fontSize": 14,
				},
			},
		},
		Explanation: "Stub: Created sticky note from recognized text",
	}, nil
}
