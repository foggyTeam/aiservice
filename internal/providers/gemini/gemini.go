package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/aiservice/internal/config"
	"github.com/aiservice/internal/models"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
)

type GeminiClient struct {
	cfg    config.LLMProviderConfig
	client *http.Client
}

func NewGeminiClient(cfg config.LLMProviderConfig) *GeminiClient {
	return &GeminiClient{
		cfg: cfg,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

func (g *GeminiClient) Analyze(ctx context.Context, transcription, contextData string) (models.AnalyzeResponse, error) {
	gkit := genkit.Init(ctx,
		genkit.WithPlugins(&googlegenai.GoogleAI{APIKey: g.cfg.APIKey}),
		genkit.WithDefaultModel("googleai/gemini-2.5-flash"),
	)

	// Define a recipe generator flow
	type Hello struct {
		Name    string `json:"name"`
		Message string `json:"message"`
	}
	recipeGeneratorFlow := genkit.DefineFlow(gkit, "helloFoo", func(ctx context.Context, input *Hello) (*Hello, error) {

		prompt := fmt.Sprintf("Generate a friendly greeting for %s.", input.Name)

		// Generate structured recipe data using the same schema
		recipe, _, err := genkit.GenerateData[Hello](ctx, gkit, ai.WithPrompt(prompt))
		if err != nil {
			return nil, fmt.Errorf("failed to generate recipe: %w", err)
		}

		return recipe, nil
	})

	// Run the flow once to test it
	recipe, err := recipeGeneratorFlow.Run(ctx, &Hello{Name: "Alice"})
	if err != nil {
		slog.Error("could not generate recipe:", "err", err)
		return models.AnalyzeResponse{}, err
	}

	recipeJSON, _ := json.MarshalIndent(recipe, "", "  ")
	slog.Info("Sample recipe generated:")
	slog.Info(string(recipeJSON))
	return models.AnalyzeResponse{}, nil
}
