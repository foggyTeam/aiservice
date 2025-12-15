package pipeline

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aiservice/internal/models"

	"github.com/aiservice/internal/providers"
)

type PipelineState struct {
	Request       models.AnalyzeRequest
	Transcription models.TranscriptionResult
	LLMResp       models.AnalyzeResponse
	ContextData   string
}

type Step func(ctx context.Context, state *PipelineState) error

type Pipeline struct {
	steps []Step
}

func NewPipeline(steps ...Step) *Pipeline {
	return &Pipeline{steps: steps}
}

func (p *Pipeline) Execute(ctx context.Context, state *PipelineState) error {
	for _, step := range p.steps {
		if err := step(ctx, state); err != nil {
			return err
		}
	}
	return nil
}

func BuildPipeline(t string, ink providers.InkRecognizer, llm providers.LLMClient) (*Pipeline, error) {
	switch t {
	case "ink":
		return NewPipeline(
			validateInkInput,
			recognizeInkStep(ink),
		), nil
	case "image":
		return NewPipeline(
			validateImageInputStep,
			recognizeImageStep(ink),
		), nil
	case "text":
		return NewPipeline(
			validateTextInputStep,
			llmAnalyzeTextStep(llm),
		), nil
	case "textWithImage":
		return NewPipeline(
			validateTextInputStep,
			validateTextInputStep,
			llmAnalyzeTextImageStep(ink, llm),
		), nil
	default:
		return nil, fmt.Errorf("unsupported input type: %s", t)
	}
}

func llmAnalyzeTextImageStep(ink providers.InkRecognizer, llm providers.LLMClient) func(context.Context, *PipelineState) error {
	return func(ctx context.Context, ps *PipelineState) error {
		tr, err := ink.RecognizeImage(ctx, ps.Request.ImageInput)
		if err != nil {
			return fmt.Errorf("image recognition failed: %w", err)
		}
		ps.Transcription = tr
		ps.LLMResp.Metadata = tr.Metadata
		// Какой то промежуточный шаг
		// TODO что делать с проанализированным текстом и входным текстом?
		resp, err := llm.Analyze(ctx, ps.Transcription.Text, ps.ContextData)
		if err != nil {
			return fmt.Errorf("llm analysis failed: %w", err)
		}
		ps.LLMResp = resp
		return nil
	}
}

func validateImageInputStep(_ context.Context, state *PipelineState) error {
	if state.Request.ImageInput.ImageURL == "" {
		return fmt.Errorf("image URL is required")
	}
	return nil
}

func validateInkInput(_ context.Context, state *PipelineState) error {
	pointCount := 0
	for _, stroke := range state.Request.InkInput.Strokes {
		pointCount += len(stroke)
	}
	if pointCount > MaxStrokesPoints {
		return fmt.Errorf("too many stroke points: %d (max %d)", pointCount, MaxStrokesPoints)
	}
	if pointCount == 0 {
		return fmt.Errorf("empty strokes")
	}
	return nil
}

func validateTextInputStep(_ context.Context, state *PipelineState) error {
	if state.Transcription.Text == "" {
		return fmt.Errorf("text input is empty")
	}
	return nil
}
func complexImageRecognitionStep(ink providers.InkRecognizer, llm providers.LLMClient) func(context.Context, *PipelineState) error {
	return func(ctx context.Context, state *PipelineState) error {
		tr, err := ink.RecognizeImage(ctx, state.Request.ImageInput)
		if err != nil {
			return fmt.Errorf("image recognition failed: %w", err)
		}
		state.LLMResp = models.AnalyzeResponse{
			Metadata: tr.Metadata,
		}
		state.Transcription = tr
		return nil
	}
}

func recognizeInkStep(ink providers.InkRecognizer) func(context.Context, *PipelineState) error {
	return func(ctx context.Context, state *PipelineState) error {
		tr, err := ink.RecognizeInk(ctx, state.Request.InkInput)
		if err != nil {
			return fmt.Errorf("ink recognition failed: %w", err)
		}
		state.Transcription = tr
		return nil

	}
}

func recognizeImageStep(ink providers.InkRecognizer) func(context.Context, *PipelineState) error {
	return func(ctx context.Context, state *PipelineState) error {
		tr, err := ink.RecognizeImage(ctx, state.Request.ImageInput)
		if err != nil {
			return fmt.Errorf("image recognition failed: %w", err)
		}
		state.Transcription = tr
		return nil
	}
}

func llmAnalyzeTextStep(llm providers.LLMClient) func(context.Context, *PipelineState) error {
	return func(ctx context.Context, state *PipelineState) error {
		if state.Transcription.Text == "" {
			return fmt.Errorf("empty transcription for llm analysis")
		}
		resp, err := llm.Analyze(ctx, state.Transcription.Text, state.ContextData)
		if err != nil {
			return fmt.Errorf("llm analysis failed: %w", err)
		}
		state.LLMResp = resp
		return nil
	}
}

func BuildContextData(ctxMap map[string]any) string {
	if ctxMap == nil {
		return ""
	}
	data, _ := json.Marshal(ctxMap)
	return string(data)
}

const (
	MaxStrokesPoints = 20000
)
