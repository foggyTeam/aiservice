package mock

import (
	"context"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/aiservice/internal/models"
	"github.com/aiservice/internal/providers"
	"github.com/firebase/genkit/go/ai"
)

type MockMode string

const (
	ModeLight   MockMode = "light"
	ModeDefault MockMode = "default"
	ModeHeavy   MockMode = "heavy"
	ModeMix     MockMode = "mix"
)

type MockClient struct {
	mode         MockMode
	responsesDir string
	rng          *rand.Rand
}

func NewMockClient(ctx context.Context) *MockClient {
	return NewMockClientWithMode(ctx, ModeMix)
}

func NewMockClientWithMode(ctx context.Context, mode MockMode) *MockClient {
	seed1, seed2 := uint64(time.Now().Second()), uint64(time.Now().Day())
	return &MockClient{
		mode: mode,
		rng:  rand.New(rand.NewPCG(seed1, seed2)),
	}
}

func (m *MockClient) GetName() string {
	return "mock"
}

func (m *MockClient) resolveMode() MockMode {
	if m.mode != ModeMix {
		return m.mode
	}

	modes := []MockMode{
		ModeLight,
		ModeDefault,
		ModeHeavy,
	}

	return modes[m.rng.IntN(len(modes))]
}

func (m *MockClient) applyDelay(mode MockMode, baseMin, baseMax time.Duration) {
	var min, max time.Duration
	switch mode {
	case ModeLight:
		min, max = 0, 500*time.Millisecond
	case ModeDefault:
		min, max = baseMin, baseMax
	case ModeHeavy:
		min, max = baseMin*3, baseMax*4
	case ModeMix:
		min, max = 0, baseMax*2
	default:
		min, max = baseMin, baseMax
	}

	delay := min
	if max > min {
		delay = min + time.Duration(m.rng.Int64N(int64(max-min)))
	}

	time.Sleep(delay)
}

func (m *MockClient) ImageRecognition(ctx context.Context, parts []*ai.Part) (providers.ImageRecognitionFlow, error) {
	m.applyDelay(m.mode, 1*time.Second, 3*time.Second)
	return providers.ImageRecognitionFlow{ImageDescription: "Mock image description: This is a mocked response for image recognition."}, nil
}

func (m *MockClient) Summarize(ctx context.Context, parts []*ai.Part) (providers.SummarizeFlow, error) {
	m.applyDelay(m.mode, 500*time.Millisecond, 2*time.Second)
	mode := m.resolveMode()
	fixtures := summarizeFixtures[mode]
	if len(fixtures) == 0 {
		return providers.SummarizeFlow{
			Summarization: "Fallback mock summarization",
		}, nil
	}
	return randomItem(m.rng, fixtures), nil
}

func (m *MockClient) SummarizeWithHistory(ctx context.Context, history []*ai.Message, parts []*ai.Part) (providers.SummarizeFlow, error) {
	m.applyDelay(m.mode, 500*time.Millisecond, 2*time.Second)
	mode := m.resolveMode()
	fixtures := summarizeFixtures[mode]
	if len(fixtures) == 0 {
		return providers.SummarizeFlow{
			Summarization: "Fallback mock summarization",
		}, nil
	}
	return randomItem(m.rng, fixtures), nil
}

func (m *MockClient) Structurize(ctx context.Context, parts []*ai.Part) (providers.StructurizeFlow, error) {
	m.applyDelay(m.mode, 700*time.Millisecond, 3*time.Second)
	mode := m.resolveMode()
	fixtures := structurizeFixtures[mode]
	if len(fixtures) == 0 {
		return providers.StructurizeFlow{
			AiTreeResponse: "└── Empty",
		}, nil
	}

	return randomItem(m.rng, fixtures), nil
}

func (m *MockClient) StructurizeWithHistory(ctx context.Context, history []*ai.Message, parts []*ai.Part) (providers.StructurizeFlow, error) {
	m.applyDelay(m.mode, 700*time.Millisecond, 3*time.Second)
	mode := m.resolveMode()
	fixtures := structurizeFixtures[mode]
	if len(fixtures) == 0 {
		return providers.StructurizeFlow{
			AiTreeResponse: "└── Empty",
		}, nil
	}
	return randomItem(m.rng, fixtures), nil
}

func (m *MockClient) GenerateTemplate(ctx context.Context, parts []*ai.Part) (providers.TemplateGenerationFlow, error) {
	m.applyDelay(m.mode, 500*time.Millisecond, 2*time.Second)
	mode := m.resolveMode()
	fixtures := templateFixtures[mode]
	if len(fixtures) == 0 {
		return providers.TemplateGenerationFlow{
			BoardType: "simple",
			Title:     "Fallback Template",
		}, nil
	}
	return randomItem(m.rng, fixtures), nil
}

func (m *MockClient) GenerateText(ctx context.Context, parts []*ai.Part) (string, error) {
	m.applyDelay(m.mode, 500*time.Millisecond, 2*time.Second)
	return strings.Repeat("a", 10*len(parts)), nil
}

type MockDigitalInkClient struct {
	mode         MockMode
	responsesDir string
	rng          *rand.Rand
}

func NewMockDigitalInkClient() *MockDigitalInkClient {
	return NewMockDigitalInkClientWithMode(ModeMix, "../../api_tests/response_fixtures")
}

func NewMockDigitalInkClientWithMode(mode MockMode, responsesDir string) *MockDigitalInkClient {
	seed1, seed2 := uint64(time.Now().Second()), uint64(time.Now().Day())
	return &MockDigitalInkClient{
		mode:         mode,
		responsesDir: responsesDir,
		rng:          rand.New(rand.NewPCG(seed1, seed2)),
	}
}

func (m *MockDigitalInkClient) RecognizeInk(ctx context.Context, elements []models.Element) (string, error) {
	m.applyDelay(m.mode, 200*time.Millisecond, 1*time.Second)
	return "Mock digital ink recognition: Handwritten text extracted from line elements.", nil
}

func (m *MockDigitalInkClient) applyDelay(mode MockMode, baseMin, baseMax time.Duration) {
	var min, max time.Duration
	switch mode {
	case ModeLight:
		min, max = 0, 200*time.Millisecond
	case ModeDefault:
		min, max = baseMin, baseMax
	case ModeHeavy:
		min, max = baseMin*2, baseMax*2
	case ModeMix:
		min, max = 0, baseMax*2
	default:
		min, max = baseMin, baseMax
	}

	delay := min
	if max > min {
		delay = min + time.Duration(m.rng.Int64N(int64(max-min)))
	}
	time.Sleep(delay)
}

func randomItem[T any](rng *rand.Rand, items []T) T {
	var zero T
	if len(items) == 0 {
		return zero
	}
	return items[rng.IntN(len(items))]
}
