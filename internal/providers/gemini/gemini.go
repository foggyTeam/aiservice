package gemini

import (
	"context"
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

func (g *GeminiClient) Analyze(ctx context.Context, parts []*ai.Part) (models.AnalyzeResponse, error) {
	flow := providers.DefineAnalyzeFlow(g.gkit, parts)
	aiResp, err := flow.Run(ctx, &providers.AnalyzeFlow{})
	if err != nil {
		slog.Error("could not generate response:", "err", err)
		return models.AnalyzeResponse{}, err
	}
	return models.AnalyzeResponse{
		ResponseMessage: aiResp.Answer,
		GraphResponse:   aiResp.Graph,
		FileStructure:   aiResp.FileStructure,
	}, nil
}

func (g *GeminiClient) RecognizeInk(ctx context.Context, input models.InkInput) (models.TranscriptionResult, error) {
	return models.TranscriptionResult{}, fmt.Errorf("gemini does not support ink recognition")
}
