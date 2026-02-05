package gemini

import (
	"context"
	"encoding/json"
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
	aiResp, err := providers.RunSummarizeGeneration(ctx, g.gkit, parts)
	if err != nil {
		slog.Error("could not generate response:", "err", err)
		return models.SummarizeResponse{}, err
	}
	return models.SummarizeResponse{Element: aiResp.Element}, nil
}

func (g *GeminiClient) Structurize(ctx context.Context, parts []*ai.Part) (models.StructurizeResponse, error) {
	aiResp, err := providers.RunStructurizeGeneration(ctx, g.gkit, parts)
	if err != nil {
		slog.Error("could not generate response:", "err", err)
		return models.StructurizeResponse{}, err
	}

	// Convert the file structure JSON string back to models.File
	var fileStructure models.File
	if err := json.Unmarshal([]byte(aiResp.FileJSON), &fileStructure); err != nil {
		slog.Warn("could not unmarshal file structure as JSON, treating as plain text:", "err", err, "response", aiResp.FileJSON)

		// If JSON parsing fails, treat the response as a plain text description
		// and create a basic file structure based on the text
		fileStructure = models.File{
			Name:     "parsed-from-text",
			Type:     "section",
			Children: []*models.File{},
		}
	}

	return models.StructurizeResponse{
		AiTreeResponse: aiResp.AiTreeResponse,
		File:           fileStructure,
	}, nil
}

func (g *GeminiClient) GetName() string {
	return "gemini"
}
