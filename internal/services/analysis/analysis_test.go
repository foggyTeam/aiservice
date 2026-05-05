package analysis

import (
	"context"
	"testing"
	"time"

	"github.com/aiservice/internal/models"
	"github.com/stretchr/testify/assert"
)

func makeSummarizeRequest(boardID, userID string) models.AnalyzeRequest {
	return models.NewSumAnalyzeReq(models.SummarizeRequest{
		RequestID: "req-1",
		UserID:    userID,
		Board: models.Board{
			BoardID: boardID,
			Elements: []models.Element{
				{Id: "e1", Type: "text", Content: "test"},
			},
		},
	})
}

func makeStructurizeRequest(boardID, userID string) models.AnalyzeRequest {
	return models.NewStructAnalyzeReq(models.StructurizeRequest{
		RequestID: "req-2",
		UserID:    userID,
		Board: models.Board{
			BoardID: boardID,
			Elements: []models.Element{
				{Id: "e2", Type: "rectangle"},
			},
		},
		File: models.File{Name: "root", Type: "section"},
	})
}

func TestExtractBoardID_Summarize(t *testing.T) {
	t.Parallel()

	svc := &AnalysisService{}
	req := makeSummarizeRequest("board-sum-42", "user-1")

	got := svc.extractBoardID(req)
	assert.Equal(t, "board-sum-42", got)
}

func TestExtractBoardID_Structurize(t *testing.T) {
	t.Parallel()

	svc := &AnalysisService{}
	req := makeStructurizeRequest("board-struct-99", "user-2")

	got := svc.extractBoardID(req)
	assert.Equal(t, "board-struct-99", got)
}

func TestExtractBoardID_OtherTypes_ReturnsEmpty(t *testing.T) {
	t.Parallel()

	svc := &AnalysisService{}
	types := []string{models.GenerateTemplateType, models.GenerateTextType, "unknown"}

	for _, reqType := range types {
		t.Run(reqType, func(t *testing.T) {
			t.Parallel()
			req := models.AnalyzeRequest{RequestType: reqType}
			got := svc.extractBoardID(req)
			assert.Empty(t, got)
		})
	}
}

func TestNewAnalysisServiceWithoutJobQueue_LeavesQueueNil(t *testing.T) {
	t.Parallel()

	svc := NewAnalysisServiceWithoutJobQueue(15*time.Second, nil, nil, nil, "gemini")

	assert.Equal(t, 15*time.Second, svc.timeout)
	assert.Equal(t, "gemini", svc.provider)
	assert.Nil(t, svc.jobQueue)
}

func TestAbort_WhenJobQueueNil_ReturnsError(t *testing.T) {
	t.Parallel()

	svc := NewAnalysisServiceWithoutJobQueue(time.Second, nil, nil, nil, "mock")
	err := svc.Abort(context.Background(), "job-123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "job queue service not initialized")
}

func TestGetJob_WhenJobQueueNil_ReturnsError(t *testing.T) {
	t.Parallel()

	svc := NewAnalysisServiceWithoutJobQueue(time.Second, nil, nil, nil, "mock")
	_, err := svc.GetJob(context.Background(), "job-123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "job queue service not initialized")
}

func TestGetJobResponse_WhenJobQueueNil_ReturnsError(t *testing.T) {
	t.Parallel()

	svc := NewAnalysisServiceWithoutJobQueue(time.Second, nil, nil, nil, "mock")
	_, _, err := svc.GetJobResponse(context.Background(), "job-123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "job queue service not initialized")
}
