package gemini

import (
	"context"
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

func (g *GeminiClient) Summarize(ctx context.Context, parts []*ai.Part) (models.SummarizeResponse, error) {
	flow := providers.DefineSummarizeFlow(g.gkit, parts)
	aiResp, err := flow.Run(ctx, &providers.SummarizeFlow{})
	if err != nil {
		slog.Error("could not generate response:", "err", err)
		return models.SummarizeResponse{}, err
	}
	return models.SummarizeResponse{
		Element: aiResp.Element,
	}, nil
}

func (g *GeminiClient) Structurize(ctx context.Context, parts []*ai.Part) (models.StructurizeResponse, error) {
	flow := providers.DefineStructurizeFlow(g.gkit, parts)
	aiResp, err := flow.Run(ctx, &providers.StructurizeFlow{})
	if err != nil {
		slog.Error("could not generate response:", "err", err)
		return models.StructurizeResponse{}, err
	}
	return models.StructurizeResponse{
		AiTreeResponse: aiResp.AiTreeResponse,
		File:           aiResp.File,
	}, nil
}
