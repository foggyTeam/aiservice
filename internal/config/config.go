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
	Database DatabaseConfig
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

type DatabaseConfig struct {
	Type     string // "memory", "sqlite"
	Host     string
	Port     string
	Name     string
	User     string
	Password string
	SSLMode  string
	FilePath string // For SQLite
	Debug    bool   // Enable SQL logging
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
			Provider: getEnv("OCR_PROVIDER", "gemini"),
			APIKey:   getEnv("OCR_API_KEY", ""),
			BaseURL:  getEnv("OCR_BASE_URL", ""),
			Timeout:  getDurationEnv("OCR_TIMEOUT", 8*time.Second),
		},
		Job: JobConfig{
			QueueSize:     getIntEnv("JOB_QUEUE_SIZE", 100),
			WorkerCount:   getIntEnv("JOB_WORKERS", 2),
			DbWorkerCount: getIntEnv("DB_JOB_WORKERS", 1),
			MaxRetries:    getIntEnv("JOB_MAX_RETRIES", 3),
			RetryBackoff:  getDurationEnv("JOB_RETRY_BACKOFF", 2*time.Second),
		},
		Timeouts: TimeoutsConfig{
			SyncProcess:  getDurationEnv("TIMEOUT_SYNC_PROCESS", 5*time.Minute),
			InkRecognize: getDurationEnv("TIMEOUT_INK_RECOGNIZE", 2*time.Minute),
			LLMRequest:   getDurationEnv("TIMEOUT_LLM_REQUEST", 2*time.Minute),
		},
		Database: DatabaseConfig{
			Type:     getEnv("DB_TYPE", "memory"), // Default to memory for backward compatibility
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			Name:     getEnv("DB_NAME", "aiservice"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
			FilePath: getEnv("SQLITE_FILE_PATH", "./aiservice.db"), // Default SQLite file path
			Debug:    getEnv("DB_DEBUG", "false") == "true",
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
