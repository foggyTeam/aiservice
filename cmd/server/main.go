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
	"github.com/aiservice/internal/providers/gemini"
	analysis "github.com/aiservice/internal/services/analysis"
	jobservice "github.com/aiservice/internal/services/jobService"
	"github.com/aiservice/internal/services/storage"
)

func main() {
	cfg := config.LoadFromEnv()

	logger := log.New(os.Stdout, "[aiservice] ", log.LstdFlags|log.Lshortfile)

	inkRecognizer := initINCRecognizers(cfg, logger)
	llmClient := initLLMProviders(cfg, logger)

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

	startServer(logger, cfg, jobQueueService, e)
}

func startServer(logger *log.Logger, cfg *config.Config, jobQueueService *jobservice.JobQueueService, e *echo.Echo) {
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

func initINCRecognizers(cfg *config.Config, logger *log.Logger) providers.InkRecognizer {
	switch cfg.OCR.Provider {
	case "azure":
		logger.Println("Using Azure Ink Recognizer")
		return providers.NewAzureInkRecognizer(cfg.OCR)
	case "myscript":
		logger.Println("Using MyScript Recognizer")
		return providers.NewMyScriptRecognizer(cfg.OCR)
	default:
		logger.Println("Using Stub Ink Recognizer (dev mode)")
		return &providers.StubInkRecognizer{}
	}
}

func initLLMProviders(cfg *config.Config, logger *log.Logger) providers.LLMClient {
	switch cfg.LLM.Provider {
	case "openai":
		// TODO вынести в отдельный файл
		logger.Println("Using OpenAI LLM")
		return providers.NewOpenAIClient(cfg.LLM)
	case "qwen":
		logger.Println("Using Qwen LLM")
		return providers.NewQwenClient(cfg.LLM)
	case "gemini":
		logger.Println("Using Gemini LLM")
		return gemini.NewGeminiClient(cfg.LLM, logger)
	default:
		logger.Println("Using Stub LLM Client (dev mode)")
		return &providers.StubLLMClient{}
	}
}
