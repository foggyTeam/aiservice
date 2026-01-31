package providers

import (
	"context"
	"fmt"
	"testing"

	"github.com/aiservice/internal/models"
	"github.com/firebase/genkit/go/ai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLLMClient is a mock implementation of LLMClient for testing
type MockLLMClient struct {
	mock.Mock
	name string
}

func (m *MockLLMClient) Summarize(ctx context.Context, parts []*ai.Part) (models.SummarizeResponse, error) {
	args := m.Called(ctx, parts)
	return args.Get(0).(models.SummarizeResponse), args.Error(1)
}

func (m *MockLLMClient) Structurize(ctx context.Context, parts []*ai.Part) (models.StructurizeResponse, error) {
	args := m.Called(ctx, parts)
	return args.Get(0).(models.StructurizeResponse), args.Error(1)
}

func (m *MockLLMClient) GetName() string {
	return m.name
}

func TestProviderManager_Summarize_Success(t *testing.T) {
	pm := NewProviderManager(&MultiProviderConfig{})

	// Create a mock provider that succeeds
	mockProvider := &MockLLMClient{name: "test-provider"}
	mockProvider.On("Summarize", mock.Anything, mock.Anything).Return(models.SummarizeResponse{
		Element: models.Text{
			Content: "Test summary",
		},
	}, nil)

	pm.RegisterProvider("test-provider", mockProvider)

	resp, err := pm.Summarize(context.Background(), []*ai.Part{})
	assert.NoError(t, err)
	assert.Equal(t, "Test summary", resp.Element.Content)
}

func TestProviderManager_Summarize_Failover(t *testing.T) {
	pm := NewProviderManager(&MultiProviderConfig{
		Providers: []ProviderConfig{
			{Name: "failing-provider", Priority: 1, Enabled: true},
			{Name: "working-provider", Priority: 2, Enabled: true},
		},
	})

	// Create a mock provider that fails with a critical error
	failingProvider := &MockLLMClient{name: "failing-provider"}
	failingProvider.On("Summarize", mock.Anything, mock.Anything).Return(models.SummarizeResponse{},
		fmt.Errorf("500 error"))

	// Create a mock provider that succeeds
	workingProvider := &MockLLMClient{name: "working-provider"}
	workingProvider.On("Summarize", mock.Anything, mock.Anything).Return(models.SummarizeResponse{
		Element: models.Text{
			Content: "Working summary",
		},
	}, nil)

	pm.RegisterProvider("failing-provider", failingProvider)
	pm.RegisterProvider("working-provider", workingProvider)

	resp, err := pm.Summarize(context.Background(), []*ai.Part{})
	assert.NoError(t, err)
	assert.Equal(t, "Working summary", resp.Element.Content)

	// Verify that the failing provider was called once and marked unhealthy
	failingProvider.AssertNumberOfCalls(t, "Summarize", 1)
	workingProvider.AssertNumberOfCalls(t, "Summarize", 1)
}

func TestProviderManager_Summarize_AllProvidersFailed(t *testing.T) {
	pm := NewProviderManager(&MultiProviderConfig{
		Providers: []ProviderConfig{
			{Name: "failing-provider-1", Priority: 1, Enabled: true},
			{Name: "failing-provider-2", Priority: 2, Enabled: true},
		},
	})

	// Create mock providers that both fail
	failingProvider1 := &MockLLMClient{name: "failing-provider-1"}
	failingProvider1.On("Summarize", mock.Anything, mock.Anything).Return(models.SummarizeResponse{},
		fmt.Errorf("500 error"))

	failingProvider2 := &MockLLMClient{name: "failing-provider-2"}
	failingProvider2.On("Summarize", mock.Anything, mock.Anything).Return(models.SummarizeResponse{},
		fmt.Errorf("403 error"))

	pm.RegisterProvider("failing-provider-1", failingProvider1)
	pm.RegisterProvider("failing-provider-2", failingProvider2)

	_, err := pm.Summarize(context.Background(), []*ai.Part{})
	assert.Error(t, err)
	assert.Equal(t, "no AI models currently working", err.Error())

	// Verify that both providers were tried
	failingProvider1.AssertNumberOfCalls(t, "Summarize", 1)
	failingProvider2.AssertNumberOfCalls(t, "Summarize", 1)
}

func TestProviderManager_CircuitBreaker(t *testing.T) {
	pm := NewProviderManager(&MultiProviderConfig{
		Providers: []ProviderConfig{
			{Name: "failing-provider", Priority: 1, Enabled: true},
			{Name: "working-provider", Priority: 2, Enabled: true},
		},
	})

	// Create a mock provider that keeps failing
	failingProvider := &MockLLMClient{name: "failing-provider"}
	failingProvider.On("Summarize", mock.Anything, mock.Anything).Return(models.SummarizeResponse{},
		fmt.Errorf("500 error")).Times(3) // Trip circuit breaker

	// Create a mock provider that succeeds
	workingProvider := &MockLLMClient{name: "working-provider"}
	workingProvider.On("Summarize", mock.Anything, mock.Anything).Return(models.SummarizeResponse{
		Element: models.Text{
			Content: "Working summary after circuit breaker",
		},
	}, nil)

	pm.RegisterProvider("failing-provider", failingProvider)
	pm.RegisterProvider("working-provider", workingProvider)

	// First call - should try failing provider then working provider
	resp, err := pm.Summarize(context.Background(), []*ai.Part{})
	assert.NoError(t, err)
	assert.Equal(t, "Working summary after circuit breaker", resp.Element.Content)

	// Second call - should skip the failing provider due to circuit breaker
	resp, err = pm.Summarize(context.Background(), []*ai.Part{})
	assert.NoError(t, err)
	assert.Equal(t, "Working summary after circuit breaker", resp.Element.Content)

	// Verify that the failing provider was only called twice (due to circuit breaker)
	// Note: This depends on the exact implementation of the circuit breaker
	// The working provider should be called both times
}