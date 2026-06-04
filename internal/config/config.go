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
	Port            string
	Env             string // "dev", "prod"
	VerificationKey string // API key for authentication
}

type LLMProviderConfig struct {
	Provider    string // "ollama", "gemini", "yandex"
	APIKey      string
	BaseURL     string
	TextModel   string // модель для суммаризации (текстовая)
	VisionModel string // модель для распознавания изображений
	Timeout     time.Duration
}

type MultiProviderConfig struct {
	Providers []ProviderConfig `json:"providers"`
}

type ProviderConfig struct {
	Name     string        `json:"name"`
	APIKey   string        `json:"api_key"`
	BaseURL  string        `json:"base_url"`
	Model    string        `json:"model"`
	Timeout  time.Duration `json:"timeout"`
	Regions  []string      `json:"regions"`  // Supported regions
	Priority int           `json:"priority"` // Lower number = higher priority
	Enabled  bool          `json:"enabled"`
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
	Type string // "memory", "sqlite"
}

func LoadFromEnv() *Config {
	provider := getEnv("LLM_PROVIDER", "ollama")

	// Set defaults based on provider
	var baseURL, textModel, visionModel string
	switch provider {
	case "ollama":
		baseURL = getEnv("OLLAMA_BASE_URL", "http://localhost:11434")
		textModel = getEnv("LLM_TEXT_MODEL", "gemma3:4b")
		visionModel = getEnv("LLM_VISION_MODEL", "gemma3:12b")
	case "gemini":
		baseURL = getEnv("GEMINI_BASE_URL", "")
		textModel = getEnv("LLM_TEXT_MODEL", "googleai/gemini-2.5-flash")
		visionModel = getEnv("LLM_VISION_MODEL", "googleai/gemini-2.5-flash")
	default:
		baseURL = getEnv("LLM_BASE_URL", "")
		textModel = getEnv("LLM_TEXT_MODEL", "")
		visionModel = getEnv("LLM_VISION_MODEL", "")
	}

	return &Config{
		Server: ServerConfig{
			Port:            getEnv("PORT", "8080"),
			Env:             getEnv("ENV", "dev"),
			VerificationKey: getEnv("VERIFICATION_KEY", ""),
		},
		LLM: LLMProviderConfig{
			Provider:    provider,
			APIKey:      getEnv("LLM_API_KEY", ""),
			BaseURL:     baseURL,
			TextModel:   textModel,
			VisionModel: visionModel,
			Timeout:     getDurationEnv("LLM_TIMEOUT", time.Minute*2),
		},
		Job: JobConfig{
			QueueSize:     getIntEnv("JOB_QUEUE_SIZE", 100),
			WorkerCount:   getIntEnv("JOB_WORKERS", 10),
			DbWorkerCount: getIntEnv("DB_JOB_WORKERS", 2),
			MaxRetries:    getIntEnv("JOB_MAX_RETRIES", 3),
			RetryBackoff:  getDurationEnv("JOB_RETRY_BACKOFF", 2*time.Second),
		},
		Timeouts: TimeoutsConfig{
			SyncProcess:  getDurationEnv("TIMEOUT_SYNC_PROCESS", 5*time.Minute),
			InkRecognize: getDurationEnv("TIMEOUT_INK_RECOGNIZE", 2*time.Minute),
			LLMRequest:   getDurationEnv("TIMEOUT_LLM_REQUEST", 2*time.Minute),
		},
		Database: DatabaseConfig{
			Type: getEnv("DB_TYPE", "memory"),
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
