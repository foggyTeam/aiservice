package yandex

import (
	"context"
	"log/slog"
	"time"

	"github.com/aiservice/internal/models"
	"github.com/firebase/genkit/go/ai"
)

// YandexGPTClient represents a mock Yandex GPT client
type YandexGPTClient struct {
	name string
}

// NewYandexGPTClient creates a new mock Yandex GPT client
func NewYandexGPTClient() *YandexGPTClient {
	return &YandexGPTClient{
		name: "yandex-gpt-mock",
	}
}

// Summarize implements the LLMClient interface
func (y *YandexGPTClient) Summarize(ctx context.Context, parts []*ai.Part) (models.SummarizeResponse, error) {
	// Simulate API call delay
	time.Sleep(80 * time.Millisecond)

	// For demo purposes, return a mock response
	response := models.SummarizeResponse{
		Element: models.Text{
			BaseElement: models.BaseElement{
				Id:     "yandex-mock-element-id",
				Type:   "text",
				X:      150,
				Y:      150,
				Width:  250,
				Height: 120,
			},
			Content: "Это фиктивная сводка от Yandex GPT провайдера",
		},
	}

	slog.Info("Yandex GPT mock client processed summarize request")
	return response, nil
}

// Structurize implements the LLMClient interface
func (y *YandexGPTClient) Structurize(ctx context.Context, parts []*ai.Part) (models.StructurizeResponse, error) {
	// Simulate API call delay
	time.Sleep(120 * time.Millisecond)

	// For demo purposes, return a mock response
	fileStructure := models.File{
		Name: "mock-yandex-project",
		Type: "section",
		Children: []models.File{
			{Name: "main.py", Type: "doc"},
			{Name: "helpers", Type: "section"},
		},
	}

	response := models.StructurizeResponse{
		AiTreeResponse: "mock-tree-response-from-yandex",
		File:           fileStructure,
	}

	slog.Info("Yandex GPT mock client processed structurize request")
	return response, nil
}

// GetName returns the provider name
func (y *YandexGPTClient) GetName() string {
	return y.name
}
