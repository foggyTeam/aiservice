package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/aiservice/internal/models"
	"github.com/aiservice/internal/services/analysis"
	jobservice "github.com/aiservice/internal/services/jobService"
	"github.com/aiservice/internal/services/storage"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

type dummyProcessor struct{}

func (p *dummyProcessor) Process(ctx context.Context, req models.AnalyzeRequest) (models.AnalyzeResponse, error) {
	return models.AnalyzeResponse{
		TextResponse: models.TextResponse{
			RequestID:   req.SummarizeRequest.RequestID,
			UserID:      req.SummarizeRequest.UserID,
			RequestType: req.RequestType,
			Content:     "processed",
		},
	}, nil
}

func newTestAnalyzeHandler(t *testing.T) (*AnalyzeHandler, *jobservice.JobQueueService, *storage.InMemoryJobStorage) {
	t.Helper()

	store := storage.NewInMemoryJobStorage()
	queue := &jobservice.JobQueueService{}
	setJobQueueStorage(queue, store)
	setCompletedResponses(queue, map[string]models.AnalyzeResponse{})

	analysisService := analysis.NewAnalysisServiceWithoutJobQueue(time.Second, nil, nil, nil, "test")
	analysisService.SetJobQueueService(queue)

	return NewAnalyzeHandler(analysisService, queue, time.Second), queue, store
}

func setJobQueueStorage(q *jobservice.JobQueueService, storage jobservice.JobStorage) {
	rv := reflect.ValueOf(q).Elem().FieldByName("storage")
	reflect.NewAt(rv.Type(), unsafe.Pointer(rv.UnsafeAddr())).Elem().Set(reflect.ValueOf(storage))
}

func setCompletedResponses(q *jobservice.JobQueueService, responses map[string]models.AnalyzeResponse) {
	rv := reflect.ValueOf(q).Elem().FieldByName("completedResponses")
	reflect.NewAt(rv.Type(), unsafe.Pointer(rv.UnsafeAddr())).Elem().Set(reflect.ValueOf(responses))
}

func newEchoContext(t *testing.T, method, target string, body []byte) (echo.Context, *httptest.ResponseRecorder) {
	t.Helper()
	e := echo.New()
	req := httptest.NewRequest(method, target, nil)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func TestGetJobStatus_Success(t *testing.T) {
	t.Parallel()

	handler, _, store := newTestAnalyzeHandler(t)

	job := models.Job{ID: "job-123", CreatedAt: time.Now().Unix(), Status: models.JobStatusPending}
	assert.NoError(t, store.Save(job))

	c, rec := newEchoContext(t, http.MethodGet, "/jobs/job-123", nil)
	c.SetParamNames("id")
	c.SetParamValues(job.ID)

	err := handler.GetJobStatus(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var actual models.Job
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &actual))
	assert.Equal(t, job, actual)
}

func TestGetJobStatus_NotFound(t *testing.T) {
	t.Parallel()

	handler, _, _ := newTestAnalyzeHandler(t)

	c, rec := newEchoContext(t, http.MethodGet, "/jobs/missing", nil)
	c.SetParamNames("id")
	c.SetParamValues("missing")

	err := handler.GetJobStatus(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetJobResponse_Processing(t *testing.T) {
	t.Parallel()

	handler, _, store := newTestAnalyzeHandler(t)

	job := models.Job{ID: "job-456", CreatedAt: time.Now().Unix(), Status: models.JobStatusPending}
	assert.NoError(t, store.Save(job))

	c, rec := newEchoContext(t, http.MethodGet, "/jobresponse/job-456", nil)
	c.SetParamNames("id")
	c.SetParamValues(job.ID)

	err := handler.GetJobResponse(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusProcessing, rec.Code)
	assert.Empty(t, rec.Body.Bytes())
}

func TestGetJobResponse_Success(t *testing.T) {
	t.Parallel()

	handler, queue, store := newTestAnalyzeHandler(t)

	job := models.Job{ID: "job-789", CreatedAt: time.Now().Unix(), Status: models.JobStatusPending}
	assert.NoError(t, store.Save(job))

	response := models.AnalyzeResponse{
		TextResponse: models.TextResponse{RequestID: "req-1", UserID: "user-1", RequestType: "summarize", Content: "done"},
	}
	setCompletedResponses(queue, map[string]models.AnalyzeResponse{job.ID: response})

	c, rec := newEchoContext(t, http.MethodGet, "/jobresponse/job-789", nil)
	c.SetParamNames("id")
	c.SetParamValues(job.ID)

	err := handler.GetJobResponse(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var actual models.AnalyzeResponse
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &actual))
	assert.Equal(t, response, actual)
}

func TestGetJobResponse_NotFound(t *testing.T) {
	t.Parallel()

	handler, _, _ := newTestAnalyzeHandler(t)

	c, rec := newEchoContext(t, http.MethodGet, "/jobresponse/missing", nil)
	c.SetParamNames("id")
	c.SetParamValues("missing")

	err := handler.GetJobResponse(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestAbort_Success(t *testing.T) {
	t.Parallel()

	handler, _, store := newTestAnalyzeHandler(t)

	job := models.Job{ID: "job-abort", CreatedAt: time.Now().Unix(), Status: models.JobStatusPending}
	assert.NoError(t, store.Save(job))

	c, rec := newEchoContext(t, http.MethodPut, "/jobs/job-abort/abort", nil)
	c.SetParamNames("id")
	c.SetParamValues(job.ID)

	err := handler.Abort(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	updated, err := store.Get(job.ID)
	assert.NoError(t, err)
	assert.Equal(t, models.JobStatusAborted, updated.Status)
}

func TestAbort_NotFound(t *testing.T) {
	t.Parallel()

	handler, _, _ := newTestAnalyzeHandler(t)

	c, rec := newEchoContext(t, http.MethodPut, "/jobs/missing/abort", nil)
	c.SetParamNames("id")
	c.SetParamValues("missing")

	err := handler.Abort(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHealthHandler(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := HealthHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string]string
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "ok", body["status"])
	_, parseErr := time.Parse(time.RFC3339, body["time"])
	assert.NoError(t, parseErr)
}
