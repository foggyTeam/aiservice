package mock

import (
	"context"
	"log/slog"

	"github.com/aiservice/internal/providers"
	"github.com/firebase/genkit/go/ai"
)

type MockClient struct{}

func NewMockClient(ctx context.Context) *MockClient {
	slog.Info("Initializing Mock provider")
	return &MockClient{}
}

func (m *MockClient) GetName() string {
	return "mock"
}

func (m *MockClient) ImageRecognition(ctx context.Context, parts []*ai.Part) (providers.ImageRecognitionFlow, error) {
	return providers.ImageRecognitionFlow{
		ImageDescription: "Mock image description: This is a mocked response for image recognition. The image contains various elements and shapes.",
	}, nil
}

func (m *MockClient) Summarize(ctx context.Context, parts []*ai.Part) (providers.SummarizeFlow, error) {
	return providers.SummarizeFlow{
		Summarization: "Mock summarization: This is a mocked response for board summarization. The board contains important information that has been summarized into key points.",
	}, nil
}

func (m *MockClient) SummarizeWithHistory(ctx context.Context, history []*ai.Message, parts []*ai.Part) (providers.SummarizeFlow, error) {
	return providers.SummarizeFlow{
		Summarization: "Mock summarization with history: This is a mocked response considering previous conversation history. Based on the context provided, here is the summary.",
	}, nil
}

func (m *MockClient) Structurize(ctx context.Context, parts []*ai.Part) (providers.StructurizeFlow, error) {
	return providers.StructurizeFlow{
		AiTreeResponse: `
├── Root Section
│   ├── Subsection 1
│   │   ├── Item 1.1
│   │   └── Item 1.2
│   ├── Subsection 2
│   │   ├── Item 2.1
│   │   └── Item 2.2
│   └── Subsection 3
└── Another Section
    └── Item 3.1
`,
	}, nil
}

func (m *MockClient) StructurizeWithHistory(ctx context.Context, history []*ai.Message, parts []*ai.Part) (providers.StructurizeFlow, error) {
	return providers.StructurizeFlow{
		AiTreeResponse: `
├── Previous Context
├── Current Section
│   ├── New Item 1
│   ├── New Item 2
│   └── New Item 3
└── Related Items
`,
	}, nil
}

func (m *MockClient) GenerateTemplate(ctx context.Context, parts []*ai.Part) (providers.TemplateGenerationFlow, error) {
	return providers.TemplateGenerationFlow{
		BoardType:   "simple",
		Title:       "Mock Template Board",
		Description: "This is a mocked template generation response with sample elements",
		Elements: []providers.TemplateElement{
			{
				Type:      "rectangle",
				X:         10,
				Y:         10,
				Width:     100,
				Height:    50,
				Fill:      "#3498db",
				Content:   "Title",
				FontSize:  16,
				FontColor: "#ffffff",
			},
			{
				Type:      "rectangle",
				X:         10,
				Y:         70,
				Width:     200,
				Height:    100,
				Fill:      "#ecf0f1",
				Content:   "Content area",
				FontSize:  12,
				FontColor: "#2c3e50",
			},
		},
	}, nil
}

func (m *MockClient) GenerateText(ctx context.Context, parts []*ai.Part) (string, error) {
	return "Mock text generation: This is a mocked response for text generation. It contains sample generated content that would normally come from an LLM.", nil
}
