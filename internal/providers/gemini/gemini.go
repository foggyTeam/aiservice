package gemini

import (
	"context"
	"log/slog"

	"github.com/aiservice/internal/config"
	"github.com/aiservice/internal/preprocessing"
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
		textModel:   nil, // Gemini uses model name in Generate call
		visionModel: nil,
	}
}

func (g *GeminiClient) Summarize(ctx context.Context, parts []*ai.Part) (providers.SummarizeFlow, error) {
	// System message - fixed instructions
	systemMsg := ai.NewSystemMessage(
		ai.NewTextPart(preprocessing.SummarizeSystemPrompt),
	)

	// User message - dynamic data
	userMsg := ai.NewUserMessage(parts...)

	resp, err := genkit.Generate(ctx, g.gkit,
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

func (g *GeminiClient) Structurize(ctx context.Context, parts []*ai.Part) (providers.StructurizeFlow, error) {
	// System message - fixed instructions
	systemMsg := ai.NewSystemMessage(
		ai.NewTextPart(preprocessing.StructurizeSystemPrompt),
	)

	// User message - dynamic data
	userMsg := ai.NewUserMessage(parts...)

	resp, err := genkit.Generate(ctx, g.gkit,
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

func (g *GeminiClient) GetName() string {
	return "gemini"
}

func (g *GeminiClient) GenerateTemplate(ctx context.Context, parts []*ai.Part) (providers.TemplateGenerationFlow, error) {
	// System message - fixed instructions
	systemMsg := ai.NewSystemMessage(
		ai.NewTextPart(preprocessing.GenerateTemplatePrompt),
	)

	// User message - dynamic data
	userMsg := ai.NewUserMessage(parts...)

	resp, err := genkit.Generate(ctx, g.gkit,
		ai.WithMessages(systemMsg, userMsg),
		ai.WithOutputType(&providers.TemplateGenerationFlow{}),
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

func (g *GeminiClient) SummarizeWithHistory(ctx context.Context, history []*ai.Message, parts []*ai.Part) (providers.SummarizeFlow, error) {
	systemMsg := ai.NewSystemMessage(ai.NewTextPart(preprocessing.SummarizeSystemPrompt))
	userMsg := ai.NewUserMessage(parts...)

	// Combine System + History + User
	messages := []*ai.Message{systemMsg}
	messages = append(messages, history...)
	messages = append(messages, userMsg)

	resp, err := genkit.Generate(ctx, g.gkit,
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

func (g *GeminiClient) StructurizeWithHistory(ctx context.Context, history []*ai.Message, parts []*ai.Part) (providers.StructurizeFlow, error) {
	systemMsg := ai.NewSystemMessage(ai.NewTextPart(preprocessing.StructurizeSystemPrompt))
	userMsg := ai.NewUserMessage(parts...)

	// Combine System + History + User
	messages := []*ai.Message{systemMsg}
	messages = append(messages, history...)
	messages = append(messages, userMsg)

	resp, err := genkit.Generate(ctx, g.gkit,
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
