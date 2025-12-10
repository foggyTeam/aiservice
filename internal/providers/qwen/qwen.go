package qwen

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/aiservice/internal/config"
	"github.com/aiservice/internal/models"
)

type QwenClient struct {
	cfg    config.LLMProviderConfig
	client *http.Client
}

func NewQwenClient(cfg config.LLMProviderConfig) *QwenClient {
	return &QwenClient{
		cfg: cfg,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

func (c *QwenClient) Analyze(ctx context.Context, transcription, contextData string) (models.AnalyzeResponse, error) {
	// Similar to OpenAI, but with Alibaba Qwen API structure
	prompt := fmt.Sprintf("Analyze: %s\nContext: %s", transcription, contextData)

	body := map[string]any{
		"model": c.cfg.Model,
		"input": map[string][]map[string]string{
			"messages": {
				{"role": "user", "content": prompt},
			},
		},
		"parameters": map[string]any{
			"temperature": 0.7,
		},
	}

	bodyBytes, _ := json.Marshal(body)
	req, _ := http.NewRequestWithContext(ctx, "POST", c.cfg.BaseURL+"/api/v1/services/aigc/text-generation/generation", bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return models.AnalyzeResponse{}, fmt.Errorf("qwen request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return models.AnalyzeResponse{}, fmt.Errorf("qwen error %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return models.AnalyzeResponse{}, fmt.Errorf("failed to parse qwen response: %w", err)
	}

	return c.parseResponse(result)
}

func (c *QwenClient) parseResponse(data map[string]any) (models.AnalyzeResponse, error) {
	// Implement Qwen-specific parsing
	return models.AnalyzeResponse{
		Intent:     "create_sticky",
		Confidence: 0.8,
	}, nil
}
