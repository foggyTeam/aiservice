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
	"github.com/aiservice/internal/providers/gemini"
	"github.com/aiservice/internal/services/analysis"
	jobservice "github.com/aiservice/internal/services/jobService"
	"github.com/aiservice/internal/services/storage"
)

func main() {
	cfg := config.LoadFromEnv()

	_ = log.SetupJsonLogger()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	llmClient := initLLMProviders(ctx, cfg)

	analysisService := analysis.NewAnalysisService(cfg.Timeouts.SyncProcess, llmClient)
	jobStorage := storage.NewInMemoryJobStorage()
	jobQueueService := jobservice.NewJobQueueService(
		cfg.Job.QueueSize,
		cfg.Job.WorkerCount,
		cfg.Job.DbWorkerCount,
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

	AnalyzeHandler := handlers.NewAnalyzeHandler(
		analysisService,
		jobQueueService,
		cfg.Timeouts.SyncProcess,
	)

	e.GET("/health", handlers.HealthHandler)
	e.GET("/jobs/:id", AnalyzeHandler.GetJobStatus)
	e.PUT("/jobs/:id/abort", AnalyzeHandler.Abort)
	e.POST("/summarize", AnalyzeHandler.Summarize)
	e.POST("/structurize", AnalyzeHandler.Structurize)

	startServer(ctx, cancel, cfg, jobQueueService, e)
}

func startServer(ctx context.Context, cancelAiServices context.CancelFunc, cfg *config.Config, jobQueueService *jobservice.JobQueueService, e *echo.Echo) {
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		slog.Info("shutting down...")

		cancelAiServices()

		ctx, cancelWithTimeout := context.WithTimeout(ctx, 10*time.Second)
		defer cancelWithTimeout()

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

func initLLMProviders(ctx context.Context, cfg *config.Config) providers.LLMClient {
	switch cfg.LLM.Provider {
	case "openai":
		// slog.Info("Using OpenAI LLM")
		// return openai.NewOpenAIClient(cfg.LLM)
		return nil
	case "gemini":
		slog.Info("Using Gemini LLM")
		return gemini.NewGeminiClient(ctx, cfg.LLM)
	default:
		panic("no providers")
	}
}
