package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/aiservice/internal/config"
	"github.com/aiservice/internal/handlers"
	"github.com/aiservice/internal/providers"
	analysis "github.com/aiservice/internal/services/analysis"
	jobservice "github.com/aiservice/internal/services/jobService"
	"github.com/aiservice/internal/services/storage"
)

func main() {
	cfg := config.LoadFromEnv()

	logger := log.New(os.Stdout, "[aiservice] ", log.LstdFlags|log.Lshortfile)

	var inkRecognizer providers.InkRecognizer
	var llmClient providers.LLMClient

	switch cfg.OCR.Provider {
	case "azure":
		inkRecognizer = providers.NewAzureInkRecognizer(cfg.OCR)
		logger.Println("Using Azure Ink Recognizer")
	case "myscript":
		inkRecognizer = providers.NewMyScriptRecognizer(cfg.OCR)
		logger.Println("Using MyScript Recognizer")
	default:
		inkRecognizer = &providers.StubInkRecognizer{}
		logger.Println("Using Stub Ink Recognizer (dev mode)")
	}

	switch cfg.LLM.Provider {
	case "openai":
		llmClient = providers.NewOpenAIClient(cfg.LLM)
		logger.Println("Using OpenAI LLM")
	case "qwen":
		llmClient = providers.NewQwenClient(cfg.LLM)
		logger.Println("Using Qwen LLM")
	default:
		llmClient = &providers.StubLLMClient{}
		logger.Println("Using Stub LLM Client (dev mode)")
	}

	analysisService := analysis.NewAnalysisService(inkRecognizer, llmClient, logger)
	jobStorage := storage.NewInMemoryJobStorage()
	jobQueueService := jobservice.NewJobQueueService(
		cfg.Job.QueueSize,
		cfg.Job.WorkerCount,
		jobStorage,
		logger,
		analysisService,
	)

	e := echo.New()
	e.Use(
		middleware.Logger(),
		middleware.Recover(),
		middleware.RequestID(),
		middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: []string{"*"},
			AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodOptions},
			AllowHeaders: []string{echo.HeaderContentType, echo.HeaderAuthorization},
		}),
	)

	analyzeHandler := handlers.NewAnalyzeHandler(
		analysisService,
		jobQueueService,
		cfg.Timeouts.SyncProcess,
		logger,
	)

	e.GET("/health", handlers.HealthHandler)
	e.POST("/analyze", analyzeHandler.Handle)
	e.GET("/jobs/:id", analyzeHandler.GetJobStatus)

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		logger.Println("Shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		jobQueueService.Shutdown()
		if err := e.Shutdown(ctx); err != nil {
			logger.Printf("Shutdown error: %v", err)
		}
	}()

	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	logger.Printf("Starting server on %s (env: %s)", addr, cfg.Server.Env)

	if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("Server error: %v", err)
	}
}
