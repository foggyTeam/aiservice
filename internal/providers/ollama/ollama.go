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
	cfg         config.LLMProviderConfig
	gkit        *genkit.Genkit
	textModel   ai.Model  // для суммаризации (gemma3:4b)
	visionModel ai.Model  // для распознавания изображений (gemma3:12b)
}

func NewOllamaClient(ctx context.Context, cfg config.LLMProviderConfig) *OllamaClient {
	ollamaPlugin := &ollama.Ollama{
		ServerAddress: cfg.BaseURL,
		Timeout:       int(cfg.Timeout.Seconds()),
	}

	gkit := genkit.Init(ctx,
		genkit.WithPlugins(ollamaPlugin),
		genkit.WithDefaultModel("ollama/"+cfg.TextModel),
	)

	// Инициализируем текстовую модель (для суммаризации)
	textModel := ollamaPlugin.DefineModel(
		gkit,
		ollama.ModelDefinition{
			Name: cfg.TextModel,
			Type: "chat",
		},
		&ai.ModelOptions{
			Supports: &ai.ModelSupports{
				Multiturn:  true,
				SystemRole: true,
			},
		},
	)

	// Инициализируем vision модель (для распознавания изображений)
	visionModel := ollamaPlugin.DefineModel(
		gkit,
		ollama.ModelDefinition{
			Name: cfg.VisionModel,
			Type: "chat",
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
		cfg:         cfg,
		gkit:        gkit,
		textModel:   textModel,
		visionModel: visionModel,
	}
}

func (o *OllamaClient) Summarize(ctx context.Context, parts []*ai.Part) (providers.SummarizeFlow, error) {
	resp, err := genkit.Generate(ctx, o.gkit,
		ai.WithModel(o.textModel),
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

func (o *OllamaClient) Structurize(ctx context.Context, parts []*ai.Part) (providers.StructurizeFlow, error) {
	resp, err := genkit.Generate(ctx, o.gkit,
		ai.WithModel(o.textModel),
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

func (o *OllamaClient) ImageRecognition(ctx context.Context, parts []*ai.Part) (providers.ImageRecognitionFlow, error) {
	resp, err := genkit.Generate(ctx, o.gkit,
		ai.WithModel(o.visionModel),
		ai.WithMessages(ai.NewUserMessage(parts...)),
		ai.WithOutputType(providers.ImageRecognitionFlow{}),
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
