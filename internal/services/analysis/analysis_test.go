package analysis

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aiservice/internal/mocks"
	"github.com/aiservice/internal/models"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestProcess_UnsupportedType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ink := mocks.NewMockInkRecognizer(ctrl)
	llm := mocks.NewMockLLMClient(ctrl)

	svc := NewAnalysisService(2*time.Second, ink, llm)
	_, err := svc.Process(context.Background(), models.AnalyzeRequest{Type: "unknown-type"})
	require.Error(t, err)
}

func TestProcess_LLMErrorPropagated(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ink := mocks.NewMockInkRecognizer(ctrl)
	llm := mocks.NewMockLLMClient(ctrl)

	llm.EXPECT().Analyze(gomock.Any(), gomock.Any()).Return(models.AnalyzeResponse{}, errors.New("llm failed"))

	svc := NewAnalysisService(2*time.Second, ink, llm)
	req := models.AnalyzeRequest{
		Type:      "userQuestion",
		TextInput: models.TextInput{Type: "test", Text: "Hello"},
	}
	_, err := svc.Process(context.Background(), req)
	require.Error(t, err)
}

func TestProcess_Success_ReturnsLLMResponse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ink := mocks.NewMockInkRecognizer(ctrl)
	llm := mocks.NewMockLLMClient(ctrl)

	expected := models.AnalyzeResponse{ResponseMessage: "ok-from-llm"}
	llm.EXPECT().Analyze(gomock.Any(), gomock.Any()).Return(expected, nil).Times(1)

	svc := NewAnalysisService(2*time.Second, ink, llm)
	req := models.AnalyzeRequest{
		Type:      "userQuestion",
		TextInput: models.TextInput{Type: "test", Text: "Hello"},
	}
	resp, err := svc.Process(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, expected.ResponseMessage, resp.ResponseMessage)
}
