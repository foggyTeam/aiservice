package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/aiservice/internal/config"
	"github.com/aiservice/internal/models"
	"github.com/aiservice/internal/providers"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
)

type GeminiClient struct {
	cfg    config.LLMProviderConfig
	client *http.Client
	gkit   *genkit.Genkit
}

func NewGeminiClient(ctx context.Context, cfg config.LLMProviderConfig) *GeminiClient {
	return &GeminiClient{
		cfg:    cfg,
		client: &http.Client{Timeout: cfg.Timeout},
		gkit: genkit.Init(ctx,
			genkit.WithPlugins(&googlegenai.GoogleAI{APIKey: cfg.APIKey}),
			genkit.WithDefaultModel(cfg.Model),
		),
	}
}

func (g *GeminiClient) GenerateGraph(ctx context.Context, transcription, contextData string) (models.AnalyzeResponse, error) {
	flow := providers.DefineGraphFlow(g.gkit, transcription, contextData)
	model := &providers.LlmGraphFlow{
		ContextData: contextData,
		GraphData:   transcription,
	}
	response, err := flow.Run(ctx, model)
	if err != nil {
		slog.Error("could not generate response:", "err", err)
		return models.AnalyzeResponse{}, err
	}
	recipeJSON, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		slog.Error("could not marshal response:", "err", err)
		return models.AnalyzeResponse{}, err
	}
	slog.Info("analyze response", "response", string(recipeJSON))
	return models.AnalyzeResponse{}, nil
}

func (g *GeminiClient) Analyze(ctx context.Context, transcription, contextData string) (models.AnalyzeResponse, error) {
	flow := providers.DefineFlow(g.gkit, transcription, contextData)
	response, err := flow.Run(ctx, &providers.LlmRequestFlow{})
	if err != nil {
		slog.Error("could not generate response:", "err", err)
		return models.AnalyzeResponse{}, err
	}
	recipeJSON, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		slog.Error("could not marshal response:", "err", err)
		return models.AnalyzeResponse{}, err
	}
	slog.Info("analyze response", "response", string(recipeJSON))
	return models.AnalyzeResponse{}, nil
}

func (g *GeminiClient) RecognizeImage(ctx context.Context, input models.ImageInput) (models.TranscriptionResult, error) {
	resp, err := genkit.Generate(ctx, g.gkit,
		ai.WithMessages(
			ai.NewUserMessage(
				ai.NewTextPart("What do you see in this image?"),
				ai.NewMediaPart("image/jpeg", input.ImageURL),
			),
		))
	if err != nil {
		return models.TranscriptionResult{}, fmt.Errorf("gemini image recognition failed: %w", err)
	}
	return models.TranscriptionResult{
		Text:     resp.Message.Text(),
		Metadata: resp.Message.Metadata,
	}, nil
}

func (g *GeminiClient) RecognizeInk(ctx context.Context, input models.InkInput) (models.TranscriptionResult, error) {
	return models.TranscriptionResult{}, fmt.Errorf("gemini does not support ink recognition")
}
