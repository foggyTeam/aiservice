package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
		setup        func()
		cleanup      func()
	}{
		{
			name:         "returns environment variable when set",
			key:          "TEST_KEY_1",
			defaultValue: "default",
			envValue:     "from_env",
			expected:     "from_env",
			setup: func() {
				os.Setenv("TEST_KEY_1", "from_env")
			},
			cleanup: func() {
				os.Unsetenv("TEST_KEY_1")
			},
		},
		{
			name:         "returns default value when env var not set",
			key:          "TEST_KEY_NONEXISTENT",
			defaultValue: "default_value",
			expected:     "default_value",
			setup:        func() {},
			cleanup:      func() {},
		},
		{
			name:         "returns default when env var is empty string",
			key:          "TEST_KEY_EMPTY",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
			setup: func() {
				os.Setenv("TEST_KEY_EMPTY", "")
			},
			cleanup: func() {
				os.Unsetenv("TEST_KEY_EMPTY")
			},
		},
		{
			name:         "handles special characters in value",
			key:          "TEST_KEY_SPECIAL",
			defaultValue: "default",
			envValue:     "value!@#$%^&*()",
			expected:     "value!@#$%^&*()",
			setup: func() {
				os.Setenv("TEST_KEY_SPECIAL", "value!@#$%^&*()")
			},
			cleanup: func() {
				os.Unsetenv("TEST_KEY_SPECIAL")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			defer tt.cleanup()

			result := getEnv(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetIntEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue int
		envValue     string
		expected     int
		setup        func()
		cleanup      func()
	}{
		{
			name:         "returns integer from environment variable",
			key:          "TEST_INT_KEY_1",
			defaultValue: 10,
			envValue:     "42",
			expected:     42,
			setup: func() {
				os.Setenv("TEST_INT_KEY_1", "42")
			},
			cleanup: func() {
				os.Unsetenv("TEST_INT_KEY_1")
			},
		},
		{
			name:         "returns default when env var not set",
			key:          "TEST_INT_KEY_NONEXISTENT",
			defaultValue: 99,
			expected:     99,
			setup:        func() {},
			cleanup:      func() {},
		},
		{
			name:         "returns default for invalid integer",
			key:          "TEST_INT_KEY_INVALID",
			defaultValue: 50,
			envValue:     "not_a_number",
			expected:     50,
			setup: func() {
				os.Setenv("TEST_INT_KEY_INVALID", "not_a_number")
			},
			cleanup: func() {
				os.Unsetenv("TEST_INT_KEY_INVALID")
			},
		},
		{
			name:         "handles negative integers",
			key:          "TEST_INT_KEY_NEGATIVE",
			defaultValue: 10,
			envValue:     "-100",
			expected:     -100,
			setup: func() {
				os.Setenv("TEST_INT_KEY_NEGATIVE", "-100")
			},
			cleanup: func() {
				os.Unsetenv("TEST_INT_KEY_NEGATIVE")
			},
		},
		{
			name:         "returns default for empty string",
			key:          "TEST_INT_KEY_EMPTY",
			defaultValue: 25,
			envValue:     "",
			expected:     25,
			setup: func() {
				os.Setenv("TEST_INT_KEY_EMPTY", "")
			},
			cleanup: func() {
				os.Unsetenv("TEST_INT_KEY_EMPTY")
			},
		},
		{
			name:         "handles zero value",
			key:          "TEST_INT_KEY_ZERO",
			defaultValue: 10,
			envValue:     "0",
			expected:     0,
			setup: func() {
				os.Setenv("TEST_INT_KEY_ZERO", "0")
			},
			cleanup: func() {
				os.Unsetenv("TEST_INT_KEY_ZERO")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			defer tt.cleanup()

			result := getIntEnv(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetDurationEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue time.Duration
		envValue     string
		expected     time.Duration
		setup        func()
		cleanup      func()
	}{
		{
			name:         "returns duration from environment variable",
			key:          "TEST_DURATION_KEY_1",
			defaultValue: 10 * time.Second,
			envValue:     "5s",
			expected:     5 * time.Second,
			setup: func() {
				os.Setenv("TEST_DURATION_KEY_1", "5s")
			},
			cleanup: func() {
				os.Unsetenv("TEST_DURATION_KEY_1")
			},
		},
		{
			name:         "returns default when env var not set",
			key:          "TEST_DURATION_KEY_NONEXISTENT",
			defaultValue: 2 * time.Minute,
			expected:     2 * time.Minute,
			setup:        func() {},
			cleanup:      func() {},
		},
		{
			name:         "returns default for invalid duration",
			key:          "TEST_DURATION_KEY_INVALID",
			defaultValue: 30 * time.Second,
			envValue:     "invalid_duration",
			expected:     30 * time.Second,
			setup: func() {
				os.Setenv("TEST_DURATION_KEY_INVALID", "invalid_duration")
			},
			cleanup: func() {
				os.Unsetenv("TEST_DURATION_KEY_INVALID")
			},
		},
		{
			name:         "handles minutes",
			key:          "TEST_DURATION_KEY_MINUTES",
			defaultValue: 10 * time.Second,
			envValue:     "3m",
			expected:     3 * time.Minute,
			setup: func() {
				os.Setenv("TEST_DURATION_KEY_MINUTES", "3m")
			},
			cleanup: func() {
				os.Unsetenv("TEST_DURATION_KEY_MINUTES")
			},
		},
		{
			name:         "handles hours",
			key:          "TEST_DURATION_KEY_HOURS",
			defaultValue: 10 * time.Second,
			envValue:     "1h",
			expected:     1 * time.Hour,
			setup: func() {
				os.Setenv("TEST_DURATION_KEY_HOURS", "1h")
			},
			cleanup: func() {
				os.Unsetenv("TEST_DURATION_KEY_HOURS")
			},
		},
		{
			name:         "handles combined duration",
			key:          "TEST_DURATION_KEY_COMBINED",
			defaultValue: 10 * time.Second,
			envValue:     "1h30m45s",
			expected:     1*time.Hour + 30*time.Minute + 45*time.Second,
			setup: func() {
				os.Setenv("TEST_DURATION_KEY_COMBINED", "1h30m45s")
			},
			cleanup: func() {
				os.Unsetenv("TEST_DURATION_KEY_COMBINED")
			},
		},
		{
			name:         "returns default for empty string",
			key:          "TEST_DURATION_KEY_EMPTY",
			defaultValue: 15 * time.Second,
			envValue:     "",
			expected:     15 * time.Second,
			setup: func() {
				os.Setenv("TEST_DURATION_KEY_EMPTY", "")
			},
			cleanup: func() {
				os.Unsetenv("TEST_DURATION_KEY_EMPTY")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			defer tt.cleanup()

			result := getDurationEnv(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLoadFromEnv(t *testing.T) {
	// Store original env vars
	originalEnv := make(map[string]string)
	envVars := []string{
		"LLM_PROVIDER", "OLLAMA_BASE_URL", "LLM_TEXT_MODEL", "LLM_VISION_MODEL",
		"LLM_API_KEY", "LLM_TIMEOUT", "PORT", "ENV", "VERIFICATION_KEY",
		"JOB_QUEUE_SIZE", "JOB_WORKERS", "DB_JOB_WORKERS", "JOB_MAX_RETRIES", "JOB_RETRY_BACKOFF",
		"TIMEOUT_SYNC_PROCESS", "TIMEOUT_INK_RECOGNIZE", "TIMEOUT_LLM_REQUEST",
		"DB_TYPE", "DB_HOST", "DB_PORT", "DB_NAME", "DB_USER", "DB_PASSWORD", "DB_SSL_MODE",
		"SQLITE_FILE_PATH", "DB_DEBUG", "S3_BUCKET_NAME", "S3_ENDPOINT", "S3_REGION",
	}

	for _, v := range envVars {
		originalEnv[v] = os.Getenv(v)
		os.Unsetenv(v)
	}

	defer func() {
		for k, v := range originalEnv {
			if v != "" {
				os.Setenv(k, v)
			}
		}
	}()

	t.Run("loads default ollama configuration", func(t *testing.T) {
		os.Unsetenv("LLM_PROVIDER")

		cfg := LoadFromEnv()

		assert.NotNil(t, cfg)
		assert.Equal(t, "ollama", cfg.LLM.Provider)
		assert.Equal(t, "http://localhost:11434", cfg.LLM.BaseURL)
		assert.Equal(t, "gemma3:4b", cfg.LLM.TextModel)
		assert.Equal(t, "gemma3:12b", cfg.LLM.VisionModel)
		assert.Equal(t, "8080", cfg.Server.Port)
		assert.Equal(t, "dev", cfg.Server.Env)
		assert.Equal(t, 100, cfg.Job.QueueSize)
		assert.Equal(t, 10, cfg.Job.WorkerCount)
	})

	t.Run("loads custom environment values", func(t *testing.T) {
		os.Setenv("PORT", "3000")
		os.Setenv("ENV", "prod")
		os.Setenv("VERIFICATION_KEY", "test_key_123")
		os.Setenv("LLM_PROVIDER", "gemini")
		os.Setenv("JOB_QUEUE_SIZE", "200")
		os.Setenv("JOB_WORKERS", "4")

		cfg := LoadFromEnv()

		assert.Equal(t, "3000", cfg.Server.Port)
		assert.Equal(t, "prod", cfg.Server.Env)
		assert.Equal(t, "test_key_123", cfg.Server.VerificationKey)
		assert.Equal(t, "gemini", cfg.LLM.Provider)
		assert.Equal(t, 200, cfg.Job.QueueSize)
		assert.Equal(t, 4, cfg.Job.WorkerCount)
	})

	t.Run("loads gemini provider config", func(t *testing.T) {
		os.Setenv("LLM_PROVIDER", "gemini")
		os.Setenv("GEMINI_BASE_URL", "https://api.gemini.com")

		cfg := LoadFromEnv()

		assert.Equal(t, "gemini", cfg.LLM.Provider)
		assert.Equal(t, "https://api.gemini.com", cfg.LLM.BaseURL)
		assert.Equal(t, "googleai/gemini-2.5-flash", cfg.LLM.TextModel)
	})

	t.Run("loads database configuration", func(t *testing.T) {
		os.Setenv("DB_TYPE", "sqlite")
		os.Setenv("DB_HOST", "db.example.com")
		os.Setenv("DB_PORT", "5432")
		os.Setenv("DB_NAME", "mydb")
		os.Setenv("SQLITE_FILE_PATH", "/data/mydb.db")

		cfg := LoadFromEnv()

		assert.Equal(t, "sqlite", cfg.Database.Type)
		assert.Equal(t, "db.example.com", cfg.Database.Host)
		assert.Equal(t, "5432", cfg.Database.Port)
		assert.Equal(t, "mydb", cfg.Database.Name)
		assert.Equal(t, "/data/mydb.db", cfg.Database.FilePath)
	})

	t.Run("loads timeouts configuration", func(t *testing.T) {
		os.Setenv("TIMEOUT_SYNC_PROCESS", "10m")
		os.Setenv("TIMEOUT_INK_RECOGNIZE", "3m")
		os.Setenv("TIMEOUT_LLM_REQUEST", "5m")

		cfg := LoadFromEnv()

		assert.Equal(t, 10*time.Minute, cfg.Timeouts.SyncProcess)
		assert.Equal(t, 3*time.Minute, cfg.Timeouts.InkRecognize)
		assert.Equal(t, 5*time.Minute, cfg.Timeouts.LLMRequest)
	})

	t.Run("loads s3 configuration", func(t *testing.T) {
		os.Setenv("S3_BUCKET_NAME", "my-bucket")
		os.Setenv("S3_ENDPOINT", "https://s3.custom.com")
		os.Setenv("S3_REGION", "us-west-2")

		cfg := LoadFromEnv()

		assert.Equal(t, "my-bucket", cfg.S3.BucketName)
		assert.Equal(t, "https://s3.custom.com", cfg.S3.Endpoint)
		assert.Equal(t, "us-west-2", cfg.S3.Region)
	})

	t.Run("loads LLM timeouts", func(t *testing.T) {
		os.Setenv("LLM_TIMEOUT", "1m30s")

		cfg := LoadFromEnv()

		assert.Equal(t, 1*time.Minute+30*time.Second, cfg.LLM.Timeout)
	})

	t.Run("handles invalid timeout values gracefully", func(t *testing.T) {
		os.Setenv("LLM_TIMEOUT", "invalid_duration")

		cfg := LoadFromEnv()

		// Should use default value (2 minutes)
		assert.Equal(t, 2*time.Minute, cfg.LLM.Timeout)
	})

	t.Run("loads all job configuration values", func(t *testing.T) {
		os.Setenv("JOB_QUEUE_SIZE", "150")
		os.Setenv("JOB_WORKERS", "5")
		os.Setenv("DB_JOB_WORKERS", "2")
		os.Setenv("JOB_MAX_RETRIES", "5")
		os.Setenv("JOB_RETRY_BACKOFF", "5s")

		cfg := LoadFromEnv()

		assert.Equal(t, 150, cfg.Job.QueueSize)
		assert.Equal(t, 5, cfg.Job.WorkerCount)
		assert.Equal(t, 2, cfg.Job.DbWorkerCount)
		assert.Equal(t, 5, cfg.Job.MaxRetries)
		assert.Equal(t, 5*time.Second, cfg.Job.RetryBackoff)
	})

	t.Run("loads db debug flag", func(t *testing.T) {
		os.Setenv("DB_DEBUG", "true")

		cfg := LoadFromEnv()

		assert.True(t, cfg.Database.Debug)
	})

	t.Run("db debug flag defaults to false", func(t *testing.T) {
		os.Unsetenv("DB_DEBUG")

		cfg := LoadFromEnv()

		assert.False(t, cfg.Database.Debug)
	})
}

func TestLoadFromEnvAllDefaults(t *testing.T) {
	// Clear all relevant env vars
	envVars := []string{
		"LLM_PROVIDER", "OLLAMA_BASE_URL", "LLM_TEXT_MODEL", "LLM_VISION_MODEL",
		"LLM_API_KEY", "LLM_TIMEOUT", "PORT", "ENV", "VERIFICATION_KEY",
		"JOB_QUEUE_SIZE", "JOB_WORKERS", "DB_JOB_WORKERS", "JOB_MAX_RETRIES", "JOB_RETRY_BACKOFF",
		"TIMEOUT_SYNC_PROCESS", "TIMEOUT_INK_RECOGNIZE", "TIMEOUT_LLM_REQUEST",
		"DB_TYPE", "DB_HOST", "DB_PORT", "DB_NAME", "DB_USER", "DB_PASSWORD", "DB_SSL_MODE",
		"SQLITE_FILE_PATH", "DB_DEBUG", "S3_BUCKET_NAME", "S3_ENDPOINT", "S3_REGION",
	}

	originalEnv := make(map[string]string)
	for _, v := range envVars {
		originalEnv[v] = os.Getenv(v)
		os.Unsetenv(v)
	}

	defer func() {
		for k, v := range originalEnv {
			if v != "" {
				os.Setenv(k, v)
			}
		}
	}()

	cfg := LoadFromEnv()

	require.NotNil(t, cfg)

	// Verify all defaults are set
	assert.Equal(t, "8080", cfg.Server.Port)
	assert.Equal(t, "dev", cfg.Server.Env)
	assert.Equal(t, "", cfg.Server.VerificationKey)

	assert.Equal(t, "ollama", cfg.LLM.Provider)
	assert.Equal(t, "http://localhost:11434", cfg.LLM.BaseURL)
	assert.Equal(t, "gemma3:4b", cfg.LLM.TextModel)
	assert.Equal(t, "gemma3:12b", cfg.LLM.VisionModel)
	assert.Equal(t, 2*time.Minute, cfg.LLM.Timeout)

	assert.Equal(t, 100, cfg.Job.QueueSize)
	assert.Equal(t, 10, cfg.Job.WorkerCount)
	assert.Equal(t, 2, cfg.Job.DbWorkerCount)
	assert.Equal(t, 3, cfg.Job.MaxRetries)

	assert.Equal(t, 5*time.Minute, cfg.Timeouts.SyncProcess)
	assert.Equal(t, 2*time.Minute, cfg.Timeouts.InkRecognize)
	assert.Equal(t, 2*time.Minute, cfg.Timeouts.LLMRequest)

	assert.Equal(t, "memory", cfg.Database.Type)
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, "5432", cfg.Database.Port)
	assert.Equal(t, "aiservice", cfg.Database.Name)

	assert.Equal(t, "foggy", cfg.S3.BucketName)
	assert.Equal(t, "https://storage.yandexcloud.net", cfg.S3.Endpoint)
	assert.Equal(t, "ru-central1", cfg.S3.Region)
}
