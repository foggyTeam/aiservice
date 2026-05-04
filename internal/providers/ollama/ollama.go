package ollama

import (
	"context"
	"log/slog"

	"github.com/aiservice/internal/config"
	"github.com/aiservice/internal/preprocessing"
	"github.com/aiservice/internal/providers"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/ollama"
)

type OllamaClient struct {
	cfg         config.LLMProviderConfig
	gkit        *genkit.Genkit
	textModel   ai.Model // для суммаризации (gemma3:4b)
	visionModel ai.Model // для распознавания изображений (gemma3:12b)
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

	if cfg.TextModel == cfg.VisionModel {
		slog.Info("Using same model for text and vision tasks", "model", cfg.TextModel)
		model := ollamaPlugin.DefineModel(
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
			textModel:   model,
			visionModel: model,
		}
	}
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
	// System message - fixed instructions
	systemMsg := ai.NewSystemMessage(
		ai.NewTextPart(preprocessing.SummarizeSystemPrompt),
	)

	// User message - dynamic data
	userMsg := ai.NewUserMessage(parts...)

	resp, err := genkit.Generate(ctx, o.gkit,
		ai.WithModel(o.textModel),
		ai.WithMessages(systemMsg, userMsg),
		ai.WithOutputType(&providers.SummarizeFlow{}),
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
	// System message - fixed instructions
	systemMsg := ai.NewSystemMessage(
		ai.NewTextPart(preprocessing.StructurizeSystemPrompt),
	)

	// User message - dynamic data
	userMsg := ai.NewUserMessage(parts...)

	resp, err := genkit.Generate(ctx, o.gkit,
		ai.WithModel(o.textModel),
		ai.WithMessages(systemMsg, userMsg),
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

func (o *OllamaClient) GetName() string {
	return "ollama"
}

func (o *OllamaClient) GenerateText(ctx context.Context, parts []*ai.Part) (string, error) {
	systemMsg := ai.NewSystemMessage(
		ai.NewTextPart(preprocessing.GenererateTextPrompt),
	)
	userMsg := ai.NewUserMessage(parts...)
	resp, err := genkit.Generate(ctx, o.gkit,
		ai.WithMessages(systemMsg, userMsg),
		ai.WithOutputType(&providers.TextGenerationFlow{}),
	)
	if err != nil {
		slog.Error("could not generate text:", "err", err)
		return "", err
	}
	var textFlow providers.TextGenerationFlow
	if err := resp.Output(&textFlow); err != nil {
		slog.Error("could not parse text output:", "err", err)
		return "", err
	}
	return textFlow.Content, nil
}

func (o *OllamaClient) GenerateTemplate(ctx context.Context, parts []*ai.Part) (providers.TemplateGenerationFlow, error) {
	// System message - fixed instructions
	systemMsg := ai.NewSystemMessage(
		ai.NewTextPart(preprocessing.GenerateTemplatePrompt),
	)

	// User message - dynamic data
	userMsg := ai.NewUserMessage(parts...)

	resp, err := genkit.Generate(ctx, o.gkit,
		ai.WithModel(o.textModel),
		ai.WithMessages(systemMsg, userMsg),
		ai.WithOutputType(&providers.TemplateGenerationFlow{}),
		// ai.WithConfig(ai.WithOutputFormat("json")),
	)
	if err != nil {
		slog.Error("could not generate template:", "err", err)
		return providers.TemplateGenerationFlow{}, err
	}

	var templateFlow providers.TemplateGenerationFlow
	if err := resp.Output(&templateFlow); err != nil {
		slog.Error("could not parse template output:", "err", err)
		return providers.TemplateGenerationFlow{}, err
	}

	return templateFlow, nil
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

func (o *OllamaClient) SummarizeWithHistory(ctx context.Context, history []*ai.Message, parts []*ai.Part) (providers.SummarizeFlow, error) {
	systemMsg := ai.NewSystemMessage(ai.NewTextPart(preprocessing.SummarizeSystemPrompt))
	userMsg := ai.NewUserMessage(parts...)

	// Combine System + History + User
	messages := []*ai.Message{systemMsg}
	messages = append(messages, history...)
	messages = append(messages, userMsg)

	if len(history) > 0 {
		slog.Info("Using session history", "messageCount", len(history))
	}

	resp, err := genkit.Generate(ctx, o.gkit,
		ai.WithModel(o.textModel),
		ai.WithMessages(messages...),
		ai.WithOutputType(&providers.SummarizeFlow{}),
	)
	if err != nil {
		slog.Error("could not generate summary with history:", "err", err)
		return providers.SummarizeFlow{}, err
	}

	var flow providers.SummarizeFlow
	if err := resp.Output(&flow); err != nil {
		slog.Error("could not parse summary output with history:", "err", err)
		return providers.SummarizeFlow{}, err
	}

	return flow, nil
}

func (o *OllamaClient) StructurizeWithHistory(ctx context.Context, history []*ai.Message, parts []*ai.Part) (providers.StructurizeFlow, error) {
	systemMsg := ai.NewSystemMessage(ai.NewTextPart(preprocessing.StructurizeSystemPrompt))
	userMsg := ai.NewUserMessage(parts...)

	// Combine System + History + User
	messages := []*ai.Message{systemMsg}
	messages = append(messages, history...)
	messages = append(messages, userMsg)

	if len(history) > 0 {
		slog.Info("Using session history", "messageCount", len(history))
	}

	resp, err := genkit.Generate(ctx, o.gkit,
		ai.WithModel(o.textModel),
		ai.WithMessages(messages...),
		ai.WithOutputType(&providers.StructurizeFlow{}),
	)
	if err != nil {
		slog.Error("could not generate structurize with history:", "err", err)
		return providers.StructurizeFlow{}, err
	}

	var flow providers.StructurizeFlow
	if err := resp.Output(&flow); err != nil {
		slog.Error("could not parse structurize output with history:", "err", err)
		return providers.StructurizeFlow{}, err
	}

	return flow, nil
}
