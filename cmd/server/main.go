package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/aiservice/internal/config"
	"github.com/aiservice/internal/handlers"
	"github.com/aiservice/internal/log"
	"github.com/aiservice/internal/providers"
	"github.com/aiservice/internal/providers/azure"
	"github.com/aiservice/internal/providers/gemini"
	"github.com/aiservice/internal/providers/openai"
	"github.com/aiservice/internal/providers/qwen"
	analysis "github.com/aiservice/internal/services/analysis"
	jobservice "github.com/aiservice/internal/services/jobService"
	"github.com/aiservice/internal/services/storage"
)

func main() {
	cfg := config.LoadFromEnv()

	_ = log.SetupJsonLogger()

	inkRecognizer := initINCRecognizers(cfg)
	llmClient := initLLMProviders(cfg)

	analysisService := analysis.NewAnalysisService(cfg.Timeouts.SyncProcess, inkRecognizer, llmClient)
	jobStorage := storage.NewInMemoryJobStorage()
	jobQueueService := jobservice.NewJobQueueService(
		cfg.Job.QueueSize,
		cfg.Job.WorkerCount,
		jobStorage,
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
	)

	e.GET("/health", handlers.HealthHandler)
	e.POST("/analyze", analyzeHandler.Handle)
	e.GET("/jobs/:id", analyzeHandler.GetJobStatus)

	startServer(cfg, jobQueueService, e)
}

func startServer(cfg *config.Config, jobQueueService *jobservice.JobQueueService, e *echo.Echo) {
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		slog.Info("shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		jobQueueService.Shutdown()
		if err := e.Shutdown(ctx); err != nil {
			slog.Error("shutdown error:", "err", err)
		}
	}()

	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	slog.Info("starting server", "addr", addr, "env", cfg.Server.Env)

	if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
		slog.Error("server error:", "err", err)
	}
}

func initINCRecognizers(cfg *config.Config) providers.InkRecognizer {
	switch cfg.OCR.Provider {
	case "azure":
		slog.Info("Using Azure Ink Recognizer")
		return azure.NewAzureInkRecognizer(cfg.OCR)
	case "myscript":
		slog.Info("Using MyScript Recognizer")
		return azure.NewMyScriptRecognizer(cfg.OCR)
	default:
		panic("no providers")
	}
}

func initLLMProviders(cfg *config.Config) providers.LLMClient {
	switch cfg.LLM.Provider {
	case "openai":
		// TODO вынести в отдельный файл
		slog.Info("Using OpenAI LLM")
		return openai.NewOpenAIClient(cfg.LLM)
	case "qwen":
		slog.Info("Using Qwen LLM")
		return qwen.NewQwenClient(cfg.LLM)
	case "gemini":
		slog.Info("Using Gemini LLM")
		return gemini.NewGeminiClient(cfg.LLM)
	default:
		panic("no providers")
	}
}
