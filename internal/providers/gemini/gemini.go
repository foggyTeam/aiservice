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
	cfg         config.LLMProviderConfig
	gkit        *genkit.Genkit
	textModel   ai.Model
	visionModel ai.Model
}

func NewGeminiClient(ctx context.Context, cfg config.LLMProviderConfig) *GeminiClient {
	gkit := genkit.Init(ctx,
		genkit.WithPlugins(&googlegenai.GoogleAI{APIKey: cfg.APIKey}),
		genkit.WithDefaultModel(cfg.TextModel),
	)

	return &GeminiClient{
		cfg:         cfg,
		gkit:        gkit,
		textModel:   nil,  // Gemini uses model name in Generate call
		visionModel: nil,
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

func (g *GeminiClient) ImageRecognition(ctx context.Context, parts []*ai.Part) (providers.ImageRecognitionFlow, error) {
	prompt := ai.NewUserMessage(parts...)

	resp, err := genkit.Generate(ctx, g.gkit,
		ai.WithMessages(prompt),
		ai.WithOutputType(&providers.ImageRecognitionFlow{}),
	)
	if err != nil {
		slog.Error("could not generate image recognition:", "err", err)
		return providers.ImageRecognitionFlow{}, err
	}

	var flow providers.ImageRecognitionFlow
	if err := resp.Output(&flow); err != nil {
		slog.Error("could not parse image recognition output:", "err", err)
		return providers.ImageRecognitionFlow{}, err
	}

	return flow, nil
}
