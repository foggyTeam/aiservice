package gemini

import (
	"context"
	"log/slog"

	"github.com/aiservice/internal/config"
	"github.com/aiservice/internal/providers"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
)

type GeminiClient struct {
	cfg  config.LLMProviderConfig
	gkit *genkit.Genkit
}

func NewGeminiClient(ctx context.Context, cfg config.LLMProviderConfig) *GeminiClient {
	gkit := genkit.Init(ctx,
		genkit.WithPlugins(&googlegenai.GoogleAI{APIKey: cfg.APIKey}),
		genkit.WithDefaultModel(cfg.Model),
	)

	return &GeminiClient{
		cfg:  cfg,
		gkit: gkit,
	}
}

func (g *GeminiClient) Summarize(ctx context.Context, parts []*ai.Part) (providers.SummarizeFlow, error) {
	resp, err := genkit.Generate(ctx, g.gkit,
		ai.WithMessages(ai.NewUserMessage(parts...)),
		ai.WithOutputType(providers.SummarizeFlow{}),
	)
	if err != nil {
		slog.Error("could not generate summarization:", "err", err)
		return providers.SummarizeFlow{}, err
	}

	var summarizeFlow providers.SummarizeFlow
	if err := resp.Output(&summarizeFlow); err != nil {
		slog.Error("could not parse summarization output:", "err", err)
		return providers.SummarizeFlow{}, err
	}

	return summarizeFlow, nil
}

func (g *GeminiClient) Structurize(ctx context.Context, parts []*ai.Part) (providers.StructurizeFlow, error) {
	prompt := ai.NewUserMessage(parts...)

	resp, err := genkit.Generate(ctx, g.gkit,
		ai.WithMessages(prompt),
		ai.WithOutputType(&providers.StructurizeFlow{}),
	)
	if err != nil {
		slog.Error("could not generate structurization:", "err", err)
		return providers.StructurizeFlow{}, err
	}

	var flow providers.StructurizeFlow
	if err := resp.Output(&flow); err != nil {
		slog.Error("could not parse structurization output:", "err", err)
		return providers.StructurizeFlow{}, err
	}

	return flow, nil
}

func (g *GeminiClient) GetName() string {
	return "gemini"
}
