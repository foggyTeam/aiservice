package pipeline

import (
	"context"
	"errors"
	"testing"

	"github.com/aiservice/internal/models"
	"github.com/firebase/genkit/go/ai"
)

// simple nop implementations for BuildPipeline calls
type nopInk struct{}

func (n *nopInk) RecognizeInk(ctx context.Context, input models.InkInput) (models.TranscriptionResult, error) {
	return models.TranscriptionResult{Text: "ink-ok"}, nil
}

type nopLLM struct{}

func (n *nopLLM) Analyze(ctx context.Context, parts []*ai.Part) (models.AnalyzeResponse, error) {
	return models.AnalyzeResponse{ResponseMessage: "llm-ok"}, nil
}

func TestNewPipeline_Execute_OrderAndErrorPropagation(t *testing.T) {
	// step1 sets transcription text
	step1 := func(ctx context.Context, s *PipelineState) error {
		s.Transcription.Text = "first"
		return nil
	}
	// step2 appends to response
	step2 := func(ctx context.Context, s *PipelineState) error {
		s.Response.ResponseMessage = s.Transcription.Text + "-second"
		return nil
	}
	// stepErr returns error and should stop pipeline
	stepErr := func(ctx context.Context, s *PipelineState) error {
		return errors.New("step failed")
	}
	stepAfterErr := func(ctx context.Context, s *PipelineState) error {
		s.Response.ResponseMessage = "after"
		return nil
	}

	p := NewPipeline(step1, step2)
	state := &PipelineState{}
	if err := p.Execute(context.Background(), state); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state.Response.ResponseMessage != "first-second" {
		t.Fatalf("unexpected response: %q", state.Response.ResponseMessage)
	}

	// test error propagation and stop-on-error
	p2 := NewPipeline(step1, stepErr, stepAfterErr)
	state2 := &PipelineState{}
	err := p2.Execute(context.Background(), state2)
	if err == nil {
		t.Fatalf("expected error from pipeline, got nil")
	}
	if state2.Response.ResponseMessage != "" {
		t.Fatalf("expected no changes after error, got %q", state2.Response.ResponseMessage)
	}
}

func TestBuildPipeline_SupportedAndUnsupportedTypes(t *testing.T) {
	ink := &nopInk{}
	llm := &nopLLM{}

	t.Run("supported userQuestion", func(t *testing.T) {
		p, err := BuildPipeline("userQuestion", ink, llm)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if p == nil {
			t.Fatalf("expected pipeline, got nil")
		}
	})

	t.Run("supported fileStructure", func(t *testing.T) {
		p, err := BuildPipeline("fileStructure", ink, llm)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if p == nil {
			t.Fatalf("expected pipeline, got nil")
		}
	})

	t.Run("unsupported", func(t *testing.T) {
		p, err := BuildPipeline("unknown-type", ink, llm)
		if err == nil {
			t.Fatalf("expected error for unsupported type, got nil")
		}
		if p != nil {
			t.Fatalf("expected nil pipeline on error, got non-nil")
		}
	})
}

func TestBuildContextData(t *testing.T) {
	m := map[string]any{"k": "v"}
	d := BuildContextData(m)
	if d == "" {
		t.Fatalf("expected non-empty context data for map")
	}
	if BuildContextData(nil) != "" {
		t.Fatalf("expected empty string for nil context")
	}
}
