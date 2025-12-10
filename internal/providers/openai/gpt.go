package openai

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

type OpenAIClient struct {
	cfg    config.LLMProviderConfig
	client *http.Client
}

func NewOpenAIClient(cfg config.LLMProviderConfig) *OpenAIClient {
	return &OpenAIClient{
		cfg: cfg,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

func (c *OpenAIClient) Analyze(ctx context.Context, transcription, contextData string) (models.AnalyzeResponse, error) {
	prompt := c.buildPrompt(transcription, contextData)

	body := map[string]any{
		"model":           c.cfg.Model,
		"messages":        []map[string]string{{"role": "user", "content": prompt}},
		"temperature":     0.7,
		"response_format": map[string]string{"type": "json_object"},
	}

	bodyBytes, _ := json.Marshal(body)
	req, _ := http.NewRequestWithContext(ctx, "POST", c.cfg.BaseURL+"/chat/completions", bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return models.AnalyzeResponse{}, fmt.Errorf("openai request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return models.AnalyzeResponse{}, fmt.Errorf("openai error %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return models.AnalyzeResponse{}, fmt.Errorf("failed to parse openai response: %w", err)
	}

	return c.parseResponse(result)
}

func (c *OpenAIClient) buildPrompt(transcription, contextData string) string {
	return fmt.Sprintf(`Analyze this board input and return a JSON response with:
{
  "intent": "action_type",
  "confidence": 0.0-1.0,
  "actions": [{
    "type": "action_name",
    "payload": {}
  }],
  "explanation": "why this action"
}

Input text: %s
Context: %s`, transcription, contextData)
}

func (c *OpenAIClient) parseResponse(data map[string]any) (models.AnalyzeResponse, error) {
	if choices, ok := data["choices"].([]any); ok && len(choices) > 0 {
		choice := choices[0].(map[string]any)
		content := choice["message"].(map[string]any)["content"].(string)

		var resp models.AnalyzeResponse
		if err := json.Unmarshal([]byte(content), &resp); err != nil {
			return models.AnalyzeResponse{}, fmt.Errorf("failed to parse openai json response: %w", err)
		}
		return resp, nil
	}
	return models.AnalyzeResponse{}, fmt.Errorf("unexpected openai response format")
}
