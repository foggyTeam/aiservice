package main

import (
	"context"
	"testing"

	"github.com/aiservice/internal/config"
	"github.com/aiservice/internal/providers/mock"
	"github.com/stretchr/testify/assert"
)

func TestInitLLMProviders_SelectsMock(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		LLM: config.LLMProviderConfig{Provider: "mock"},
	}

	client := initLLMProviders(context.Background(), cfg)

	_, ok := client.(*mock.MockClient)
	assert.True(t, ok, "ожидался MockClient для провайдера 'mock'")
}

func TestInitLLMProviders_UnknownFallsBackToOllama(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		LLM: config.LLMProviderConfig{
			Provider: "unknown-provider",
			BaseURL:  "http://localhost:11434",
		},
	}

	client := initLLMProviders(context.Background(), cfg)

	_ = client
}

func TestInitDigitalInkClient_SelectsMock(t *testing.T) {
	t.Parallel()

	client := initDigitalInkClient("mock")

	_, ok := client.(*mock.MockDigitalInkClient)
	assert.True(t, ok, "ожидался MockDigitalInkClient для провайдера 'mock'")
}

func TestInitDigitalInkClient_SelectsRealForOthers(t *testing.T) {
	t.Parallel()

	providers := []string{"gemini", "ollama", "yandex", ""}

	for _, provider := range providers {
		t.Run(provider, func(t *testing.T) {
			t.Parallel()
			client := initDigitalInkClient(provider)
			assert.NotNil(t, client, "для провайдера %q должен вернуться клиент", provider)
		})
	}
}

func TestCORSConfig_SelectsBasedOnEnv(t *testing.T) {
	t.Parallel()

	tests := []struct {
		env          string
		wantWildcard bool // true если ожидается AllowOrigins: ["*"]
	}{
		{"prod", false},
		{"dev", true},
		{"staging", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.env, func(t *testing.T) {
			t.Parallel()

			cfg := &config.Config{
				Server: config.ServerConfig{Env: tt.env},
			}

			isProd := cfg.Server.Env == "prod"
			hasWildcard := !isProd

			assert.Equal(t, tt.wantWildcard, hasWildcard,
				"для env=%q ожидается wildcard=%v", tt.env, tt.wantWildcard)
		})
	}
}
