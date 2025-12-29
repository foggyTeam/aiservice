package pipeline

import (
	"context"
	"testing"

	"github.com/aiservice/internal/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

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
		return &errStep{"step failed"}
	}
	stepAfterErr := func(ctx context.Context, s *PipelineState) error {
		s.Response.ResponseMessage = "after"
		return nil
	}

	p := NewPipeline(step1, step2)
	state := &PipelineState{}
	require.NoError(t, p.Execute(context.Background(), state))
	require.Equal(t, "first-second", state.Response.ResponseMessage)

	p2 := NewPipeline(step1, stepErr, stepAfterErr)
	state2 := &PipelineState{}
	err := p2.Execute(context.Background(), state2)
	require.Error(t, err)
	require.Equal(t, "", state2.Response.ResponseMessage)
}

type errStep struct{ msg string }

func (e *errStep) Error() string { return e.msg }

func TestBuildPipeline_SupportedAndUnsupportedTypes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ink := mocks.NewMockInkRecognizer(ctrl)
	llm := mocks.NewMockLLMClient(ctrl)

	t.Run("supported userQuestion", func(t *testing.T) {
		p, err := BuildPipeline("userQuestion", ink, llm)
		require.NoError(t, err)
		require.NotNil(t, p)
	})

	t.Run("supported fileStructure", func(t *testing.T) {
		p, err := BuildPipeline("fileStructure", ink, llm)
		require.NoError(t, err)
		require.NotNil(t, p)
	})

	t.Run("unsupported", func(t *testing.T) {
		p, err := BuildPipeline("unknown-type", ink, llm)
		require.Error(t, err)
		require.Nil(t, p)
	})
}

func TestBuildContextData(t *testing.T) {
	m := map[string]any{"k": "v"}
	d := BuildContextData(m)
	require.NotEmpty(t, d)
	require.Empty(t, BuildContextData(nil))
}
