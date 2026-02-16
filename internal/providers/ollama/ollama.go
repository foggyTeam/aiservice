package ollama

import (
	"context"
	"log/slog"

	"github.com/aiservice/internal/config"
	"github.com/aiservice/internal/providers"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/ollama"
)

type OllamaClient struct {
	cfg   config.LLMProviderConfig
	gkit  *genkit.Genkit
	model ai.Model
}

func NewOllamaClient(ctx context.Context, cfg config.LLMProviderConfig) *OllamaClient {
	ollamaPlugin := &ollama.Ollama{
		ServerAddress: cfg.BaseURL,
		Timeout:       int(cfg.Timeout.Seconds()),
	}

	gkit := genkit.Init(ctx,
		genkit.WithPlugins(ollamaPlugin),
		genkit.WithDefaultModel("ollama/"+cfg.Model),
	)

	model := ollamaPlugin.DefineModel(
		gkit,
		ollama.ModelDefinition{
			Name: cfg.Model,
			Type: "generate",
		},
		&ai.ModelOptions{
			Supports: &ai.ModelSupports{
				Multiturn:  true,
				SystemRole: true,
				Media:      true,
			},
		},
	)

	return &OllamaClient{
		cfg:   cfg,
		gkit:  gkit,
		model: model,
	}
}

func (o *OllamaClient) Summarize(ctx context.Context, parts []*ai.Part) (providers.SummarizeFlow, error) {
	resp, err := genkit.Generate(ctx, o.gkit,
		ai.WithModel(o.model),
		ai.WithMessages(ai.NewUserMessage(parts...)),
		// ai.WithOutputType(providers.SummarizeFlow{}),
	)
	slog.Info("info", "", resp.Text())
	if err != nil {
		slog.Error("could not generate response:", "err", err)
		return providers.SummarizeFlow{}, err
	}
	slog.Info("info", "", resp.Text())
	var summarizeFlow providers.SummarizeFlow
	if err := resp.Output(&summarizeFlow); err != nil {
		slog.Error("could not parse response:", "err", err)
		return providers.SummarizeFlow{}, err
	}
	return summarizeFlow, nil
}

func (o *OllamaClient) Structurize(ctx context.Context, parts []*ai.Part) (providers.StructurizeFlow, error) {
	resp, err := genkit.Generate(ctx, o.gkit,
		ai.WithModel(o.model),
		ai.WithMessages(ai.NewUserMessage(parts...)),
		ai.WithOutputType(providers.StructurizeFlow{}),
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

func (o *OllamaClient) GetName() string {
	return "ollama"
}
