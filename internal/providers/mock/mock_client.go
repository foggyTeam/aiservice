package mock

import (
	"context"
	"log/slog"

	"github.com/aiservice/internal/models"
	"github.com/firebase/genkit/go/ai"
)

// MockClient implements the LLMClient interface with mock responses
type MockClient struct{}

// NewMockClient creates a new instance of MockClient
func NewMockClient() *MockClient {
	slog.Info("Using Mock LLM Client - Gemini API key not configured")
	return &MockClient{}
}

// Summarize returns a mock response when the actual LLM service is not available
func (m *MockClient) Summarize(ctx context.Context, parts []*ai.Part) (models.SummarizeResponse, error) {
	slog.Warn("Returning mock summarize response - Gemini API key not configured")

	return models.SummarizeResponse{
		Element: models.Text{
			BaseElement: models.BaseElement{},
			Content:     "Mock Summary: Gemini API key not configured. This is a mocked response for testing purposes.",
		},
	}, nil
}

// Structurize returns a mock response when the actual LLM service is not available
func (m *MockClient) Structurize(ctx context.Context, parts []*ai.Part) (models.StructurizeResponse, error) {
	slog.Warn("Returning mock structurize response - Gemini API key not configured")

	return models.StructurizeResponse{
		AiTreeResponse: "Mock Structured Response: Gemini API key not configured. This is a mocked response for testing purposes.",
		File: models.File{
			Name:     "mock_file",
			Type:     "mock",
			Children: []models.File{},
		},
	}, nil
}

// GetName returns the provider name
func (m *MockClient) GetName() string {
	return "mock"
}
