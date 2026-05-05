package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aiservice/internal/config"
	"github.com/aiservice/internal/handlers"
	authmiddleware "github.com/aiservice/internal/middleware"
	"github.com/aiservice/internal/providers"
	mockprovider "github.com/aiservice/internal/providers/mock"
	"github.com/aiservice/internal/services/analysis"
	"github.com/aiservice/internal/services/database"
	"github.com/aiservice/internal/services/image"
	jobservice "github.com/aiservice/internal/services/jobService"
	"github.com/aiservice/internal/services/storage"
	"github.com/firebase/genkit/go/ai"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddlewareIntegration(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Env:             "prod",
			VerificationKey: "test-verification-key",
		},
	}

	_ = image.NewService("/tmp/aiservice-test-images", 5*time.Minute)

	jobStorage, _ := database.NewStorage(cfg.Database)
	analysisService := analysis.NewAnalysisServiceWithoutJobQueue(cfg.Timeouts.SyncProcess, nil, nil, nil, "ollama")
	jobQueueService := jobservice.NewJobQueueService(10, 1, 1, jobStorage, analysisService)
	analysisService.SetJobQueueService(jobQueueService)

	e := echo.New()

	e.Use(authmiddleware.APIKeyAuth(cfg))

	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	t.Run("Valid API key succeeds", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("x-api-key", "test-verification-key")
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "success", rec.Body.String())
	})

	t.Run("Invalid API key fails", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("x-api-key", "invalid-key")
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)

		var response map[string]string
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"], "Unauthorized")
	})

	t.Run("Missing API key fails", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)

		var response map[string]string
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"], "Unauthorized")
	})
}

func TestAuthMiddlewareDevelopmentBypass(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Env:             "dev",
			VerificationKey: "test-verification-key",
		},
	}

	_ = image.NewService("/tmp/aiservice-test-images", 5*time.Minute)

	e := echo.New()

	e.Use(authmiddleware.APIKeyAuth(cfg))

	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("x-api-key", "wrong-key")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "success", rec.Body.String())
}

func TestSummarizeJobResponsePollingIntegration(t *testing.T) {
	cfg := &config.Config{
		Timeouts: config.TimeoutsConfig{
			SyncProcess: 10 * time.Millisecond,
		},
	}

	storageService := storage.NewInMemoryJobStorage()

	digitalInkClient := mockprovider.NewMockDigitalInkClient()
	llmClient := &slowSummarizeMockClient{}

	analysisService := analysis.NewAnalysisService(cfg.Timeouts.SyncProcess, llmClient, digitalInkClient, nil, nil, "mock")
	jobQueueService := jobservice.NewJobQueueService(0, 1, 1, storageService, analysisService)
	analysisService.SetJobQueueService(jobQueueService)
	defer jobQueueService.Shutdown()

	handler := handlers.NewAnalyzeHandler(analysisService, jobQueueService, cfg.Timeouts.SyncProcess)
	e := echo.New()
	e.POST("/summarize", handler.Summarize)
	e.GET("/jobresponse/:id", handler.GetJobResponse)

	summarizeReq := map[string]any{
		"requestId":   "req-123",
		"userId":      "user-123",
		"requestType": "summarize",
		"board": map[string]any{
			"boardId": "board-123",
			"elements": []map[string]any{
				{
					"id":      "el-1",
					"type":    "text",
					"content": "Test content for the summarize job",
				},
			},
		},
	}

	bodyBytes, err := json.Marshal(summarizeReq)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/summarize", bytes.NewReader(bodyBytes))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusAccepted, rec.Code)

	var jobID string
	err = json.Unmarshal(rec.Body.Bytes(), &jobID)
	assert.NoError(t, err)
	assert.NotEmpty(t, jobID)

	assert.Eventually(t, func() bool {
		pollReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/jobresponse/%s", jobID), nil)
		pollRec := httptest.NewRecorder()
		e.ServeHTTP(pollRec, pollReq)
		t.Logf("polling jobresponse: status=%d body=%s", pollRec.Code, pollRec.Body.String())
		return pollRec.Code == http.StatusOK
	}, 5*time.Second, 1*time.Second, "expected job response to become available")
}

func TestSummarizeJobQueueOverflowIntegration(t *testing.T) {
	cfg := &config.Config{
		Timeouts: config.TimeoutsConfig{
			SyncProcess: 10 * time.Millisecond,
		},
	}

	storageService := storage.NewInMemoryJobStorage()
	digitalInkClient := mockprovider.NewMockDigitalInkClient()
	llmClient := &slowSummarizeMockClient{}

	analysisService := analysis.NewAnalysisService(cfg.Timeouts.SyncProcess, llmClient, digitalInkClient, nil, nil, "mock")
	jobQueueService := jobservice.NewJobQueueService(5, 0, 0, storageService, analysisService)
	analysisService.SetJobQueueService(jobQueueService)
	defer jobQueueService.Shutdown()

	handler := handlers.NewAnalyzeHandler(analysisService, jobQueueService, cfg.Timeouts.SyncProcess)
	e := echo.New()
	e.POST("/summarize", handler.Summarize)
	e.GET("/jobresponse/:id", handler.GetJobResponse)

	createSummarizeReq := func(requestID string) *http.Request {
		payload := map[string]any{
			"requestId":   requestID,
			"userId":      "user-123",
			"requestType": "summarize",
			"board": map[string]any{
				"boardId": "board-123",
				"elements": []map[string]any{
					{
						"id":      fmt.Sprintf("el-%s", requestID),
						"type":    "text",
						"content": "Overflow test content",
					},
				},
			},
		}
		bodyBytes, err := json.Marshal(payload)
		assert.NoError(t, err)
		req := httptest.NewRequest(http.MethodPost, "/summarize", bytes.NewReader(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		return req
	}

	var err error
	for i := 1; i <= 5; i++ {
		req := createSummarizeReq(fmt.Sprintf("req-%d", i))
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusAccepted, rec.Code)

		var jobID string
		err = json.Unmarshal(rec.Body.Bytes(), &jobID)
		assert.NoError(t, err)
		assert.NotEmpty(t, jobID)
	}

	sixthReq := createSummarizeReq("req-6")
	sixthRec := httptest.NewRecorder()
	e.ServeHTTP(sixthRec, sixthReq)
	assert.Equal(t, http.StatusAccepted, sixthRec.Code)

	var sixthJobID string
	err = json.Unmarshal(sixthRec.Body.Bytes(), &sixthJobID)
	assert.NoError(t, err)
	assert.NotEmpty(t, sixthJobID)

	pollReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/jobresponse/%s", sixthJobID), nil)
	pollRec := httptest.NewRecorder()
	e.ServeHTTP(pollRec, pollReq)
	assert.Equal(t, http.StatusProcessing, pollRec.Code)
}

type slowSummarizeMockClient struct{}

func (m *slowSummarizeMockClient) GetName() string {
	return "slow-mock"
}

func (m *slowSummarizeMockClient) ImageRecognition(ctx context.Context, parts []*ai.Part) (providers.ImageRecognitionFlow, error) {
	return providers.ImageRecognitionFlow{ImageDescription: "mock image description"}, nil
}

func (m *slowSummarizeMockClient) Summarize(ctx context.Context, parts []*ai.Part) (providers.SummarizeFlow, error) {
	time.Sleep(200 * time.Millisecond)
	return providers.SummarizeFlow{Summarization: "mock-summary"}, nil
}

func (m *slowSummarizeMockClient) SummarizeWithHistory(ctx context.Context, history []*ai.Message, parts []*ai.Part) (providers.SummarizeFlow, error) {
	return providers.SummarizeFlow{Summarization: "mock-summary"}, nil
}

func (m *slowSummarizeMockClient) Structurize(ctx context.Context, parts []*ai.Part) (providers.StructurizeFlow, error) {
	return providers.StructurizeFlow{AiTreeResponse: "mock-structurize"}, nil
}

func (m *slowSummarizeMockClient) StructurizeWithHistory(ctx context.Context, history []*ai.Message, parts []*ai.Part) (providers.StructurizeFlow, error) {
	return providers.StructurizeFlow{AiTreeResponse: "mock-structurize"}, nil
}

func (m *slowSummarizeMockClient) GenerateTemplate(ctx context.Context, parts []*ai.Part) (providers.TemplateGenerationFlow, error) {
	return providers.TemplateGenerationFlow{}, nil
}

func (m *slowSummarizeMockClient) GenerateText(ctx context.Context, parts []*ai.Part) (string, error) {
	return "", nil
}
