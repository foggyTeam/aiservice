package jobservice

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/aiservice/internal/models"
	"github.com/aiservice/internal/services/mocks"
	"github.com/aiservice/internal/services/storage"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type fakeProcessor struct {
	resp models.AnalyzeResponse
	err  error
}

func (f fakeProcessor) Process(ctx context.Context, req models.AnalyzeRequest) (models.AnalyzeResponse, error) {
	return f.resp, f.err
}

func TestGetJobResponseNotFound(t *testing.T) {
	st := storage.NewInMemoryJobStorage()
	svc := NewJobQueueService(1, 0, 0, st, fakeProcessor{})

	resp, found, err := svc.GetJobResponse("missing-job")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if found {
		t.Fatalf("expected not found, but got response: %#v", resp)
	}
}

func TestStoreCompletedResponseAndGetJobResponse(t *testing.T) {
	st := storage.NewInMemoryJobStorage()
	svc := NewJobQueueService(1, 0, 0, st, fakeProcessor{})

	expected := models.AnalyzeResponse{
		SummarizeResponse: models.SummarizeResponse{
			RequestID:   "req-1",
			UserID:      "user-1",
			RequestType: models.SummarizeType,
			Element: models.Text{
				BaseElement: models.BaseElement{Id: "generated-summary"},
				Content:     "test summary",
			},
		},
	}

	svc.storeCompletedResponse("job-123", expected)

	resp, found, err := svc.GetJobResponse("job-123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !found {
		t.Fatal("expected response to be found")
	}
	if resp.SummarizeResponse.RequestID != expected.SummarizeResponse.RequestID {
		t.Fatalf("expected request ID %q, got %q", expected.SummarizeResponse.RequestID, resp.SummarizeResponse.RequestID)
	}
}

func TestProcessJobStoresCompletedResponse(t *testing.T) {
	st := storage.NewInMemoryJobStorage()
	proc := fakeProcessor{
		resp: models.AnalyzeResponse{
			TextResponse: models.TextResponse{
				RequestID:   "req-2",
				UserID:      "user-2",
				RequestType: models.GenerateTextType,
				Content:     "generated content",
			},
		},
	}
	svc := NewJobQueueService(1, 1, 0, st, proc)
	defer svc.Shutdown()

	job := models.Job{ID: "job-456", Request: models.AnalyzeRequest{RequestType: models.GenerateTextType}, Status: models.JobStatusPending}
	if err := svc.Enqueue(job); err != nil {
		t.Fatalf("failed to enqueue job: %v", err)
	}

	var found bool
	for range 20 {
		resp, ok, err := svc.GetJobResponse(job.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ok {
			found = true
			if resp.TextResponse.Content != proc.resp.TextResponse.Content {
				t.Fatalf("expected content %q, got %q", proc.resp.TextResponse.Content, resp.TextResponse.Content)
			}
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if !found {
		t.Fatal("expected completed response to be stored after processing")
	}
}

func makeTestJob(id, requestType string) models.Job {
	return models.Job{
		ID:        id,
		Request:   models.AnalyzeRequest{RequestType: requestType},
		CreatedAt: time.Now().Unix(),
		Status:    models.JobStatusPending,
	}
}

func waitForJobCompletion(t *testing.T, q *JobQueueService, jobID string, timeout time.Duration) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.Fatalf("timeout waiting for job %s to complete", jobID)
		case <-ticker.C:
			resp, found, _ := q.GetJobResponse(jobID)
			if found {
				_ = resp
				return
			}
		}
	}
}

func TestNewJob_GeneratesUniqueID(t *testing.T) {
	t.Parallel()

	req := models.AnalyzeRequest{RequestType: models.SummarizeType}
	job1 := NewJob(req)
	job2 := NewJob(req)

	assert.NotEqual(t, job1.ID, job2.ID, "ID должны быть уникальными")
	assert.Equal(t, models.JobStatusPending, job1.Status)
	assert.Equal(t, models.SummarizeType, job1.Request.RequestType)
}

func TestQueueFullErr_Error(t *testing.T) {
	t.Parallel()

	err := QueueFullErr{}
	assert.Equal(t, "queue is full", err.Error())
}

func TestGetJob_DelegatesToStorage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockJobStorage(ctrl)
	q := &JobQueueService{storage: mockStorage}

	ctx := context.Background()
	jobID := "job-123"
	expectedJob := makeTestJob(jobID, models.SummarizeType)

	mockStorage.EXPECT().Get(jobID).Return(expectedJob, nil)

	job, err := q.GetJob(ctx, jobID)

	assert.NoError(t, err)
	assert.Equal(t, expectedJob.ID, job.ID)
}

func TestStatus_DelegatesToStorage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockJobStorage(ctrl)
	q := &JobQueueService{storage: mockStorage}

	jobID := "job-123"
	mockStorage.EXPECT().Get(jobID).Return(makeTestJob(jobID, models.StructurizeType), nil)

	status, err := q.Status(jobID)

	assert.NoError(t, err)
	assert.Equal(t, models.JobStatusPending, status)
}

func TestAbort_DelegatesToStorage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockJobStorage(ctrl)
	q := &JobQueueService{storage: mockStorage}

	ctx := context.Background()
	jobID := "job-123"

	mockStorage.EXPECT().Abort(ctx, jobID).Return(nil)

	err := q.Abort(ctx, jobID)
	assert.NoError(t, err)
}

func TestStoreAndGetCompletedResponse(t *testing.T) {
	t.Parallel()

	q := &JobQueueService{
		completedResponses: make(map[string]models.AnalyzeResponse),
	}

	jobID := "job-123"
	expectedResp := models.AnalyzeResponse{
		SummarizeResponse: models.SummarizeResponse{RequestID: "req-1"},
	}

	q.storeCompletedResponse(jobID, expectedResp)

	resp, found, err := q.GetJobResponse(jobID)

	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "req-1", resp.SummarizeResponse.RequestID)
}

func TestGetJobResponse_NotFound(t *testing.T) {
	t.Parallel()

	q := &JobQueueService{
		completedResponses: make(map[string]models.AnalyzeResponse),
	}

	resp, found, err := q.GetJobResponse("non-existent")

	assert.NoError(t, err)
	assert.False(t, found)
	assert.Empty(t, resp)
}

func TestDeleteCompletedResponse(t *testing.T) {
	t.Parallel()

	q := &JobQueueService{
		completedResponses: make(map[string]models.AnalyzeResponse),
	}

	jobID := "job-123"
	q.storeCompletedResponse(jobID, models.AnalyzeResponse{})

	err := q.deleteCompletedResponse(jobID)
	assert.NoError(t, err)

	_, found, _ := q.GetJobResponse(jobID)
	assert.False(t, found)

	err = q.deleteCompletedResponse(jobID)
	assert.ErrorIs(t, err, completedResponseJobNotFoundErr)
}

func TestCompletedResponseMap_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	q := &JobQueueService{
		completedResponses: make(map[string]models.AnalyzeResponse),
	}

	const writers = 10
	const readers = 20
	const ops = 100

	var wg sync.WaitGroup

	for w := range writers {
		wg.Add(1)
		go func(writerID int) {
			defer wg.Done()
			for i := range ops {
				jobID := fmt.Sprintf("job-%d-%d", writerID, i%10)
				q.storeCompletedResponse(jobID, models.AnalyzeResponse{})
			}
		}(w)
	}

	for range readers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range ops {
				jobID := fmt.Sprintf("job-%d-%d", i%writers, i%10)
				_, _, _ = q.GetJobResponse(jobID)
			}
		}()
	}

	wg.Wait()
}

func TestEnqueue_SaveFails_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockJobStorage(ctrl)
	q := &JobQueueService{
		queue:   make(chan models.Job, 1),
		storage: mockStorage,
	}

	job := makeTestJob("job-123", models.SummarizeType)
	mockStorage.EXPECT().Save(job).Return(fmt.Errorf("db error"))

	err := q.Enqueue(job)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save job")
}

func TestEnqueue_QueueHasSpace_Succeeds(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockJobStorage(ctrl)
	q := &JobQueueService{
		queue:   make(chan models.Job, 2),
		storage: mockStorage,
	}

	job := makeTestJob("job-123", models.SummarizeType)
	mockStorage.EXPECT().Save(job).Return(nil)

	err := q.Enqueue(job)
	assert.NoError(t, err)

	select {
	case received := <-q.queue:
		assert.Equal(t, job.ID, received.ID)
	default:
		t.Error("джоб не был отправлен в канал")
	}
}

func TestEnqueue_QueueFull_ReturnsQueueFullErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockJobStorage(ctrl)
	q := &JobQueueService{
		queue:   make(chan models.Job, 1),
		storage: mockStorage,
	}

	job1 := makeTestJob("job-1", models.SummarizeType)
	job2 := makeTestJob("job-2", models.StructurizeType)

	mockStorage.EXPECT().Save(job1).Return(nil)
	mockStorage.EXPECT().Save(job2).Return(nil)

	err := q.Enqueue(job1)
	assert.NoError(t, err)

	err = q.Enqueue(job2)
	assert.ErrorIs(t, err, QueueFullErr{})
}

func TestCleanJobs_RemovesInactive(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockJobStorage(ctrl)
	q := &JobQueueService{
		storage:            mockStorage,
		completedResponses: make(map[string]models.AnalyzeResponse),
	}

	oldPending := makeTestJob("old-1", models.SummarizeType)
	oldPending.Status = models.JobStatusPending
	oldPending.CreatedAt = time.Now().Add(-25 * time.Hour).Unix()

	newPending := makeTestJob("new-1", models.SummarizeType)
	newPending.Status = models.JobStatusPending
	newPending.CreatedAt = time.Now().Unix()

	q.storeCompletedResponse(oldPending.ID, models.AnalyzeResponse{})

	mockStorage.EXPECT().DeleteJobs("old-1").Return(nil)

	err := q.cleanJobs(oldPending, newPending)
	assert.NoError(t, err)
}

func TestCleanJobs_RemovesAborted(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockJobStorage(ctrl)
	q := &JobQueueService{
		storage:            mockStorage,
		completedResponses: make(map[string]models.AnalyzeResponse),
	}

	abortedJob := makeTestJob("aborted-1", models.SummarizeType)
	abortedJob.Status = models.JobStatusAborted
	activeJob := makeTestJob("active-1", models.StructurizeType)

	q.storeCompletedResponse(abortedJob.ID, models.AnalyzeResponse{})

	mockStorage.EXPECT().DeleteJobs("aborted-1").Return(nil)

	err := q.cleanJobs(abortedJob, activeJob)
	assert.NoError(t, err)
}

func TestCleanJobs_RemovesInactivePending(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockJobStorage(ctrl)
	q := &JobQueueService{
		storage:            mockStorage,
		completedResponses: make(map[string]models.AnalyzeResponse),
	}

	oldPending := makeTestJob("old-pending-1", models.SummarizeType)
	oldPending.Status = models.JobStatusPending
	oldPending.CreatedAt = time.Now().Add(-25 * time.Hour).Unix()

	newPending := makeTestJob("new-pending-1", models.SummarizeType)
	newPending.Status = models.JobStatusPending
	newPending.CreatedAt = time.Now().Unix()

	q.storeCompletedResponse(oldPending.ID, models.AnalyzeResponse{})

	mockStorage.EXPECT().DeleteJobs("old-pending-1").Return(nil)

	err := q.cleanJobs(oldPending, newPending)
	assert.NoError(t, err)
}

func TestCleanJobs_RemovesOldRunningJobs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockJobStorage(ctrl)
	q := &JobQueueService{
		storage:            mockStorage,
		completedResponses: make(map[string]models.AnalyzeResponse),
	}

	oldRunning := makeTestJob("old-running-1", models.SummarizeType)
	oldRunning.Status = models.JobStatusRunning
	oldRunning.CreatedAt = time.Now().Add(-48 * time.Hour).Unix()

	q.storeCompletedResponse(oldRunning.ID, models.AnalyzeResponse{})

	mockStorage.EXPECT().DeleteJobs("old-running-1").Return(nil)

	err := q.cleanJobs(oldRunning)
	assert.NoError(t, err)
}

func TestCleanJobs_KeepsFreshRunningJobs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockJobStorage(ctrl)
	q := &JobQueueService{
		storage:            mockStorage,
		completedResponses: make(map[string]models.AnalyzeResponse),
	}

	freshRunning := makeTestJob("fresh-running-1", models.SummarizeType)
	freshRunning.Status = models.JobStatusRunning
	freshRunning.CreatedAt = time.Now().Add(-1 * time.Hour).Unix()

	err := q.cleanJobs(freshRunning)
	assert.NoError(t, err)
}

func TestCleanJobs_KeepsFreshCompletedJobs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockJobStorage(ctrl)
	q := &JobQueueService{storage: mockStorage}

	oldCompleted := makeTestJob("old-completed-1", models.SummarizeType)
	oldCompleted.Status = models.JobStatusCompleted
	oldCompleted.CreatedAt = time.Now().Add(-48 * time.Hour).Unix()

	err := q.cleanJobs(oldCompleted)
	assert.NoError(t, err)
}

func TestCleanJobs_EmptyList_NoOp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockJobStorage(ctrl)
	q := &JobQueueService{storage: mockStorage}

	err := q.cleanJobs()
	assert.NoError(t, err)
}

func TestJobQueueService_Shutdown_WaitsForWorkers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockJobStorage(ctrl)
	mockProcessor := mocks.NewMockProcessor(ctrl)

	q := NewJobQueueService(10, 2, 1, mockStorage, mockProcessor)

	done := make(chan struct{})
	go func() {
		q.Shutdown()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("Shutdown timed out - workers may be stuck")
	}
}

func TestGenerateJobID_Format(t *testing.T) {
	t.Parallel()

	id := generateJobID()
	assert.Contains(t, id, "job_")
}

func TestGenerateJobID_Unique(t *testing.T) {
	t.Parallel()

	ids := make(map[string]bool)
	for range 100 {
		id := generateJobID()
		assert.False(t, ids[id], "дубликат ID: %s", id)
		ids[id] = true
	}
}
