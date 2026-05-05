package pipeline

import (
	"context"
	"testing"

	"github.com/aiservice/internal/models"
	"github.com/aiservice/internal/providers"
	"github.com/aiservice/internal/services/image"
	"github.com/firebase/genkit/go/ai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock implementations for testing

type mockLLMClient struct {
	imageRecognitionResp       providers.ImageRecognitionFlow
	imageRecognitionErr        error
	summarizeResp              providers.SummarizeFlow
	summarizeErr               error
	summarizeWithHistoryResp   providers.SummarizeFlow
	summarizeWithHistoryErr    error
	structurizeResp            providers.StructurizeFlow
	structurizeErr             error
	structurizeWithHistoryResp providers.StructurizeFlow
	structurizeWithHistoryErr  error
	generateTemplateResp       providers.TemplateGenerationFlow
	generateTemplateErr        error
	generateTextResp           string
	generateTextErr            error
}

func (m *mockLLMClient) ImageRecognition(ctx context.Context, parts []*ai.Part) (providers.ImageRecognitionFlow, error) {
	return m.imageRecognitionResp, m.imageRecognitionErr
}

func (m *mockLLMClient) Summarize(ctx context.Context, parts []*ai.Part) (providers.SummarizeFlow, error) {
	return m.summarizeResp, m.summarizeErr
}

func (m *mockLLMClient) SummarizeWithHistory(ctx context.Context, history []*ai.Message, parts []*ai.Part) (providers.SummarizeFlow, error) {
	return m.summarizeWithHistoryResp, m.summarizeWithHistoryErr
}

func (m *mockLLMClient) Structurize(ctx context.Context, parts []*ai.Part) (providers.StructurizeFlow, error) {
	return m.structurizeResp, m.structurizeErr
}

func (m *mockLLMClient) StructurizeWithHistory(ctx context.Context, history []*ai.Message, parts []*ai.Part) (providers.StructurizeFlow, error) {
	return m.structurizeWithHistoryResp, m.structurizeWithHistoryErr
}

func (m *mockLLMClient) GenerateTemplate(ctx context.Context, parts []*ai.Part) (providers.TemplateGenerationFlow, error) {
	return m.generateTemplateResp, m.generateTemplateErr
}

func (m *mockLLMClient) GenerateText(ctx context.Context, parts []*ai.Part) (string, error) {
	return m.generateTextResp, m.generateTextErr
}

func (m *mockLLMClient) GetName() string {
	return "mock"
}

type mockDigitalInkRecognizer struct {
	recognizeResp string
	recognizeErr  error
}

func (m *mockDigitalInkRecognizer) RecognizeInk(ctx context.Context, elements []models.Element) (string, error) {
	return m.recognizeResp, m.recognizeErr
}

type mockImageService struct {
	downloadResult *image.DownloadResult
	downloadErr    error
}

func (m *mockImageService) DownloadImage(ctx context.Context, url string) (*image.DownloadResult, error) {
	return m.downloadResult, m.downloadErr
}

func TestExtractLineElements(t *testing.T) {
	tests := []struct {
		name     string
		elements []models.Element
		expected []models.Element
	}{
		{
			name:     "no elements",
			elements: []models.Element{},
			expected: nil,
		},
		{
			name: "only line elements",
			elements: []models.Element{
				{Type: "line", Content: "line1"},
				{Type: "line", Content: "line2"},
			},
			expected: []models.Element{
				{Type: "line", Content: "line1"},
				{Type: "line", Content: "line2"},
			},
		},
		{
			name: "mixed elements",
			elements: []models.Element{
				{Type: "text", Content: "text1"},
				{Type: "line", Content: "line1"},
				{Type: "rectangle", Content: "rect1"},
				{Type: "line", Content: "line2"},
			},
			expected: []models.Element{
				{Type: "line", Content: "line1"},
				{Type: "line", Content: "line2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractLineElements(tt.elements)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatTextElementsContent(t *testing.T) {
	tests := []struct {
		name     string
		elements []models.Element
		expected string
	}{
		{
			name:     "no elements",
			elements: []models.Element{},
			expected: "",
		},
		{
			name: "no text elements",
			elements: []models.Element{
				{Type: "line", Content: "line1"},
				{Type: "rectangle", Content: "rect1"},
			},
			expected: "",
		},
		{
			name: "single text element",
			elements: []models.Element{
				{Type: "text", Content: "Hello World"},
			},
			expected: "- Hello World",
		},
		{
			name: "multiple text elements",
			elements: []models.Element{
				{Type: "text", Content: "First"},
				{Type: "text", Content: "Second"},
				{Type: "line", Content: "line1"},
				{Type: "text", Content: "Third"},
			},
			expected: "- First\n- Second\n- Third",
		},
		{
			name: "text elements with empty content",
			elements: []models.Element{
				{Type: "text", Content: "Valid"},
				{Type: "text", Content: ""},
				{Type: "text", Content: "   "},
				{Type: "text", Content: "Another"},
			},
			expected: "- Valid\n- Another",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTextElementsContent(tt.elements)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewBoardTextExtractionStep(t *testing.T) {
	tests := []struct {
		name                string
		requestType         string
		boardElements       []models.Element
		expectedTextContent string
	}{
		{
			name:        "summarize request with text elements",
			requestType: models.SummarizeType,
			boardElements: []models.Element{
				{Type: "text", Content: "Note 1"},
				{Type: "line", Content: "line1"},
				{Type: "text", Content: "Note 2"},
			},
			expectedTextContent: "- Note 1\n- Note 2",
		},
		{
			name:        "structurize request with text elements",
			requestType: models.StructurizeType,
			boardElements: []models.Element{
				{Type: "text", Content: "Item 1"},
				{Type: "text", Content: "Item 2"},
			},
			expectedTextContent: "- Item 1\n- Item 2",
		},
		{
			name:        "unknown request type",
			requestType: models.GenerateTemplateType,
			boardElements: []models.Element{
				{Type: "text", Content: "Should not be extracted"},
			},
			expectedTextContent: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			step := newBoardTextExtractionStep()

			state := &PipelineState{
				AnalyzeRequest: models.AnalyzeRequest{
					RequestType: tt.requestType,
				},
			}

			if tt.requestType == models.SummarizeType {
				state.AnalyzeRequest.SummarizeRequest = models.SummarizeRequest{
					Board: models.Board{
						Elements: tt.boardElements,
					},
				}
			} else if tt.requestType == models.StructurizeType {
				state.AnalyzeRequest.StructurizeRequest = models.StructurizeRequest{
					Board: models.Board{
						Elements: tt.boardElements,
					},
				}
			}

			err := step(context.Background(), state)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedTextContent, state.BoardTextContent)
		})
	}
}

func TestNewImageRecognitionStep(t *testing.T) {
	tests := []struct {
		name        string
		imageURI    string
		mockResp    providers.ImageRecognitionFlow
		mockErr     error
		expectError bool
	}{
		{
			name:     "successful recognition with image",
			imageURI: "data:image/jpeg;base64,test",
			mockResp: providers.ImageRecognitionFlow{
				ImageDescription: "Test description",
			},
			mockErr:     nil,
			expectError: false,
		},
		{
			name:     "successful recognition without image",
			imageURI: "",
			mockResp: providers.ImageRecognitionFlow{
				ImageDescription: "No image description",
			},
			mockErr:     nil,
			expectError: false,
		},
		{
			name:        "recognition error",
			imageURI:    "",
			mockResp:    providers.ImageRecognitionFlow{},
			mockErr:     assert.AnError,
			expectError: false, // Step should not fail, just warn
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llm := &mockLLMClient{
				imageRecognitionResp: tt.mockResp,
				imageRecognitionErr:  tt.mockErr,
			}

			step := newImageRecognitionStep(llm)

			state := &PipelineState{
				ImageURI: tt.imageURI,
			}

			err := step(context.Background(), state)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.mockErr == nil {
					assert.Equal(t, tt.mockResp, state.ImageRecognitionFlow)
				}
			}
		})
	}
}

func TestNewTextGenerationStep(t *testing.T) {
	tests := []struct {
		name        string
		mockResp    string
		mockErr     error
		expectError bool
	}{
		{
			name:        "successful generation",
			mockResp:    "Generated text",
			mockErr:     nil,
			expectError: false,
		},
		{
			name:        "generation error",
			mockResp:    "",
			mockErr:     assert.AnError,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llm := &mockLLMClient{
				generateTextResp: tt.mockResp,
				generateTextErr:  tt.mockErr,
			}

			step := newTextGenerationStep(llm)

			state := &PipelineState{
				AnalyzeRequest: models.AnalyzeRequest{
					TemplateRequest: models.GenerateTemplateRequest{
						Prompt: "Test prompt",
					},
				},
			}

			err := step(context.Background(), state)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mockResp, state.BoardTextContent)
			}
		})
	}
}

func TestNewTemplateGenerationStep(t *testing.T) {
	mockResp := providers.TemplateGenerationFlow{
		BoardType: "simple",
		Elements: []providers.TemplateElement{
			{Type: "text", Content: "Test element"},
		},
	}

	llm := &mockLLMClient{
		generateTemplateResp: mockResp,
	}

	step := newTemplateGenerationStep(llm)

	state := &PipelineState{
		AnalyzeRequest: models.AnalyzeRequest{
			TemplateRequest: models.GenerateTemplateRequest{
				Prompt:    "Test prompt",
				BoardType: "simple",
			},
		},
	}

	err := step(context.Background(), state)
	require.NoError(t, err)
	assert.Equal(t, mockResp, state.TemplateGenerationFlow)
}

func TestSessionMsgs(t *testing.T) {
	userData := "User input"
	resp := "Model response"

	result := sessionMsgs(userData, resp)

	expected := []models.MessageEntry{
		{Role: "user", Content: userData},
		{Role: "model", Content: resp},
	}

	assert.Equal(t, expected, result)
}

func TestConvertTemplateToBoard(t *testing.T) {
	tests := []struct {
		name     string
		template providers.TemplateGenerationFlow
		boardID  string
		expected *models.Board
	}{
		{
			name: "simple board",
			template: providers.TemplateGenerationFlow{
				BoardType: "simple",
				Elements: []providers.TemplateElement{
					{
						Type:    "text",
						X:       10,
						Y:       20,
						Width:   100,
						Height:  50,
						Content: "Test text",
					},
				},
			},
			boardID: "test-board",
			expected: &models.Board{
				BoardID: "test-board",
				Elements: []models.Element{
					{
						Id:      "elem-0",
						Type:    "text",
						X:       10,
						Y:       20,
						Width:   100,
						Height:  50,
						Content: "Test text",
					},
				},
				GraphNodes: []models.GNode{},
				GraphEdges: []models.GEdge{},
			},
		},
		{
			name: "graph board",
			template: providers.TemplateGenerationFlow{
				BoardType: "graph",
				GraphNodes: []providers.TemplateNode{
					{
						ID:    "node1",
						Type:  "customNode",
						Title: "Node 1",
					},
				},
				GraphEdges: []providers.TemplateEdge{
					{
						ID:     "edge1",
						Source: "node1",
						Target: "node2",
						Label:  "connects to",
					},
				},
			},
			boardID: "graph-board",
			expected: &models.Board{
				BoardID:  "graph-board",
				Elements: []models.Element{},
				GraphNodes: []models.GNode{
					{
						ID:   "node1",
						Type: "customNode",
						Data: models.GNodeData{
							Title: "Node 1",
						},
					},
				},
				GraphEdges: []models.GEdge{
					{
						ID:     "edge1",
						Source: "node1",
						Target: "node2",
						Label:  "connects to",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertTemplateToBoard(tt.template, tt.boardID)
			assert.Equal(t, tt.expected, result)
		})
	}
}
