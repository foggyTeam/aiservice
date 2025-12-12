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
	RawInput      any
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
			parseInkInputStep,
			validateInkInputStep,
			recognizeInkStep(ink),
			llmAnalyzeStep(llm),
		), nil
	case "image":
		return NewPipeline(
			parseImageInputStep,
			recognizeImageStep(ink),
			llmAnalyzeStep(llm),
		), nil
	case "text":
		return NewPipeline(
			parseTextInputStep,
			llmAnalyzeStep(llm),
		), nil
	default:
		return nil, fmt.Errorf("unsupported input type: %s", t)
	}
}

func parseInkInputStep(ctx context.Context, state *PipelineState) error {
	var ink models.InkInput
	if err := json.Unmarshal(state.Request.Input, &ink); err != nil {
		return fmt.Errorf("invalid ink input: %w", err)
	}
	state.RawInput = ink
	return nil
}

func validateInkInputStep(ctx context.Context, state *PipelineState) error {
	ink, ok := state.RawInput.(models.InkInput)
	if !ok {
		return fmt.Errorf("validateInkInputStep: raw input missing or wrong type")
	}
	if err := validateInkInput(ink); err != nil {
		return err
	}
	return nil
}

func recognizeInkStep(ink providers.InkRecognizer) func(context.Context, *PipelineState) error {
	return func(ctx context.Context, state *PipelineState) error {
		input, ok := state.RawInput.(models.InkInput)
		if !ok {
			return fmt.Errorf("recognizeInkStep: raw input missing or wrong type")
		}
		tr, err := ink.RecognizeInk(ctx, input)
		if err != nil {
			return fmt.Errorf("ink recognition failed: %w", err)
		}
		state.Transcription = tr
		return nil

	}
}

func parseImageInputStep(ctx context.Context, state *PipelineState) error {
	var img models.ImageInput
	if err := json.Unmarshal(state.Request.Input, &img); err != nil {
		return fmt.Errorf("invalid image input: %w", err)
	}
	state.RawInput = img
	return nil
}

func recognizeImageStep(ink providers.InkRecognizer) func(context.Context, *PipelineState) error {
	return func(ctx context.Context, state *PipelineState) error {
		img, ok := state.RawInput.(models.ImageInput)
		if !ok {
			return fmt.Errorf("recognizeImageStep: raw input missing or wrong type")
		}
		tr, err := ink.RecognizeImage(ctx, img)
		if err != nil {
			return fmt.Errorf("image recognition failed: %w", err)
		}
		state.Transcription = tr
		return nil
	}
}

func parseTextInputStep(ctx context.Context, state *PipelineState) error {
	var txt models.TextInput
	if err := json.Unmarshal(state.Request.Input, &txt); err != nil {
		return fmt.Errorf("invalid text input: %w", err)
	}
	state.Transcription = models.TranscriptionResult{
		Text:     txt.Text,
		Language: "en", // Опредение языка убрать
	}
	return nil
}

func llmAnalyzeStep(llm providers.LLMClient) func(context.Context, *PipelineState) error {
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

func validateInkInput(input models.InkInput) error {
	pointCount := 0
	for _, stroke := range input.Strokes {
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
