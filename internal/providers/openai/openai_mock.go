package openai

import (
	"context"
	"log/slog"
	"time"

	"github.com/aiservice/internal/models"
	"github.com/firebase/genkit/go/ai"
)

// OpenAIClient represents a mock OpenAI client
type OpenAIClient struct {
	name string
}

// NewOpenAIClient creates a new mock OpenAI client
func NewOpenAIClient() *OpenAIClient {
	return &OpenAIClient{
		name: "openai-mock",
	}
}

// Summarize implements the LLMClient interface
func (o *OpenAIClient) Summarize(ctx context.Context, parts []*ai.Part) (models.SummarizeResponse, error) {
	// Simulate API call delay
	time.Sleep(100 * time.Millisecond)

	// For demo purposes, return a mock response
	response := models.SummarizeResponse{
		Element: models.Text{
			BaseElement: models.BaseElement{
				Id:     "mock-element-id",
				Type:   "text",
				X:      100,
				Y:      100,
				Width:  200,
				Height: 100,
			},
			Content: "This is a mock summary from OpenAI provider",
		},
	}

	slog.Info("OpenAI mock client processed summarize request")
	return response, nil
}

// Structurize implements the LLMClient interface
func (o *OpenAIClient) Structurize(ctx context.Context, parts []*ai.Part) (models.StructurizeResponse, error) {
	// Simulate API call delay
	time.Sleep(150 * time.Millisecond)

	// For demo purposes, return a mock response
	fileStructure := models.File{
		Name: "mock-project",
		Type: "section",
		Children: []models.File{
			{Name: "main.go", Type: "doc"},
			{Name: "utils", Type: "section"},
		},
	}

	response := models.StructurizeResponse{
		AiTreeResponse: "mock-tree-response-from-openai",
		File:           fileStructure,
	}

	slog.Info("OpenAI mock client processed structurize request")
	return response, nil
}

// GetName returns the provider name
func (o *OpenAIClient) GetName() string {
	return o.name
}
