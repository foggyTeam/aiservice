package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aiservice/internal/config"
	authmiddleware "github.com/aiservice/internal/middleware"
	"github.com/aiservice/internal/services/analysis"
	"github.com/aiservice/internal/services/database"
	"github.com/aiservice/internal/services/image"
	jobservice "github.com/aiservice/internal/services/jobService"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddlewareIntegration(t *testing.T) {
	// Create a minimal config for testing
	cfg := &config.Config{
		Server: config.ServerConfig{
			Env:            "prod",
			VerificationKey: "test-verification-key",
		},
	}

	// Create image service with temp directory (not used in this test, but required for constructor)
	_ = image.NewService("/tmp/aiservice-test-images", 5*time.Minute)

	// Create a minimal analysis service
	jobStorage, _ := database.NewStorage(cfg.Database)
	analysisService := analysis.NewAnalysisServiceWithoutJobQueue(cfg.Timeouts.SyncProcess, nil, nil, "ollama")
	jobQueueService := jobservice.NewJobQueueService(10, 1, 1, jobStorage, analysisService)
	analysisService.SetJobQueueService(jobQueueService)

	// Create Echo instance with middleware
	e := echo.New()
	
	// Apply auth middleware
	e.Use(authmiddleware.APIKeyAuth(cfg))

	// Define a test route
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	// Test 1: Valid API key should succeed
	t.Run("Valid API key succeeds", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("x-api-key", "test-verification-key")
		rec := httptest.NewRecorder()
		
		e.ServeHTTP(rec, req)
		
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "success", rec.Body.String())
	})

	// Test 2: Invalid API key should fail
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

	// Test 3: Missing API key should fail
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
	// Create a config with dev environment
	cfg := &config.Config{
		Server: config.ServerConfig{
			Env:            "dev",
			VerificationKey: "test-verification-key",
		},
	}

	// Create image service with temp directory (not used in this test, but required for constructor)
	_ = image.NewService("/tmp/aiservice-test-images", 5*time.Minute)

	// Create Echo instance with middleware
	e := echo.New()

	// Apply auth middleware
	e.Use(authmiddleware.APIKeyAuth(cfg))

	// Define a test route
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	// In dev mode, even with wrong key, it should succeed
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("x-api-key", "wrong-key")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "success", rec.Body.String())
}