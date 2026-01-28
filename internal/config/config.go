package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	LLM      LLMProviderConfig
	OCR      OCRProviderConfig
	Job      JobConfig
	Timeouts TimeoutsConfig
}

type ServerConfig struct {
	Port string
	Env  string // "dev", "prod"
}

type LLMProviderConfig struct {
	Provider string // "openai", "qwen", "anthropic"
	APIKey   string
	BaseURL  string
	Model    string
	Timeout  time.Duration
}

type OCRProviderConfig struct {
	Provider string // "azure", "myscript", "google"
	APIKey   string
	BaseURL  string
	Timeout  time.Duration
}

type JobConfig struct {
	QueueSize     int
	WorkerCount   int
	DbWorkerCount int
	MaxRetries    int
	RetryBackoff  time.Duration
}

type TimeoutsConfig struct {
	SyncProcess  time.Duration
	InkRecognize time.Duration
	LLMRequest   time.Duration
}

func LoadFromEnv() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Env:  getEnv("ENV", "dev"),
		},
		LLM: LLMProviderConfig{
			// Provider: getEnv("LLM_PROVIDER", "openai"),
			// APIKey:   getEnv("LLM_API_KEY", ""),
			// BaseURL:  getEnv("LLM_BASE_URL", "https://api.openai.com/v1"),
			Timeout:  getDurationEnv("LLM_TIMEOUT", 20*time.Second),
			Model:    getEnv("LLM_MODEL", "googleai/gemini-2.5-flash"),
			Provider: getEnv("LLM_PROVIDER", "gemini"),
			APIKey:   getEnv("GEMINI_API_KEY", ""),
		},
		OCR: OCRProviderConfig{
			Provider: getEnv("LLM_PROVIDER", "gemini"),
			APIKey:   getEnv("GEMINI_API_KEY", ""),
			// Provider: getEnv("OCR_PROVIDER", "azure"),
			// APIKey:   getEnv("OCR_API_KEY", ""),
			// BaseURL:  getEnv("OCR_BASE_URL", ""),
			// Timeout:  getDurationEnv("OCR_TIMEOUT", 8*time.Second),
		},
		Job: JobConfig{
			QueueSize:     getIntEnv("JOB_QUEUE_SIZE", 100),
			WorkerCount:   getIntEnv("JOB_WORKERS", 2),
			DbWorkerCount: getIntEnv("DB_JOB_WORKERS", 1),
			MaxRetries:    getIntEnv("JOB_MAX_RETRIES", 3),
			RetryBackoff:  getDurationEnv("JOB_RETRY_BACKOFF", 2*time.Second),
		},
		Timeouts: TimeoutsConfig{
			SyncProcess:  getDurationEnv("TIMEOUT_SYNC_PROCESS", 20*time.Minute),
			InkRecognize: getDurationEnv("TIMEOUT_INK_RECOGNIZE", 20*time.Minute),
			LLMRequest:   getDurationEnv("TIMEOUT_LLM_REQUEST", 20*time.Minute),
		},
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getIntEnv(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

func getDurationEnv(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
