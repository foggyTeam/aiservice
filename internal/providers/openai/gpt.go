package openai

// type OpenAIClient struct {
// 	cfg    config.LLMProviderConfig
// 	client *http.Client
// }

// func NewOpenAIClient(cfg config.LLMProviderConfig) *OpenAIClient {
// 	return &OpenAIClient{
// 		cfg:    cfg,
// 		client: &http.Client{Timeout: cfg.Timeout},
// 	}
// }

// func (c *OpenAIClient) Analyze(ctx context.Context, transcription, contextData string) (models.SummarizeResponse, error) {
// 	prompt := c.buildPrompt(transcription, contextData)

// 	body := map[string]any{
// 		"model":           c.cfg.Model,
// 		"messages":        []map[string]string{{"role": "user", "content": prompt}},
// 		"temperature":     0.7,
// 		"response_format": map[string]string{"type": "json_object"},
// 	}

// 	bodyBytes, _ := json.Marshal(body)
// 	req, _ := http.NewRequestWithContext(ctx, "POST", c.cfg.BaseURL+"/chat/completions", bytes.NewReader(bodyBytes))
// 	req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
// 	req.Header.Set("Content-Type", "application/json")

// 	resp, err := c.client.Do(req)
// 	if err != nil {
// 		return models.SummarizeResponse{}, fmt.Errorf("openai request failed: %w", err)
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		body, _ := io.ReadAll(resp.Body)
// 		return models.SummarizeResponse{}, fmt.Errorf("openai error %d: %s", resp.StatusCode, string(body))
// 	}

// 	var result map[string]any
// 	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
// 		return models.SummarizeResponse{}, fmt.Errorf("failed to parse openai response: %w", err)
// 	}

// 	return c.parseResponse(result)
// }

// func (c *OpenAIClient) RecognizeImage(ctx context.Context, input models.ImageInput) (models.TranscriptionResult, error) {
// 	body := map[string]any{
// 		"url": input.ImageURL,
// 	}

// 	bodyBytes, _ := json.Marshal(body)
// 	req, _ := http.NewRequestWithContext(ctx, "POST", c.cfg.BaseURL+"/recognizeText", bytes.NewReader(bodyBytes))
// 	req.Header.Set("Ocp-Apim-Subscription-Key", c.cfg.APIKey)
// 	req.Header.Set("Content-Type", "application/json")

// 	resp, err := c.client.Do(req)
// 	if err != nil {
// 		return models.TranscriptionResult{}, fmt.Errorf("azure ocr request failed: %w", err)
// 	}
// 	defer resp.Body.Close()

// 	var result map[string]any
// 	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
// 		return models.TranscriptionResult{}, fmt.Errorf("failed to parse azure ocr response: %w", err)
// 	}

// 	return models.TranscriptionResult{
// 		Text:     extractString(result, "recognitionResult", "lines", "0", "text"),
// 		Language: "en",
// 		Metadata: result,
// 	}, nil
// }

// func extractString(data map[string]any, keys ...string) string {
// 	current := any(data)
// 	for _, key := range keys {
// 		switch v := current.(type) {
// 		case map[string]any:
// 			current = v[key]
// 		case []any:
// 			idx := 0
// 			// Try to parse as index
// 			if i, err := json.Number(key).Int64(); err == nil {
// 				idx = int(i)
// 			}
// 			if idx < len(v) {
// 				current = v[idx]
// 			} else {
// 				return ""
// 			}
// 		default:
// 			return ""
// 		}
// 	}
// 	if s, ok := current.(string); ok {
// 		return s
// 	}
// 	return ""
// }

// func (c *OpenAIClient) buildPrompt(transcription, contextData string) string {
// 	return fmt.Sprintf(`Analyze this board input and return a JSON response with:
// {
//   "intent": "action_type",
//   "confidence": 0.0-1.0,
//   "actions": [{
//     "type": "action_name",
//     "payload": {}
//   }],
//   "explanation": "why this action"
// }

// Input text: %s
// Context: %s`, transcription, contextData)
// }

// func (c *OpenAIClient) parseResponse(data map[string]any) (models.SummarizeResponse, error) {
// 	if choices, ok := data["choices"].([]any); ok && len(choices) > 0 {
// 		choice := choices[0].(map[string]any)
// 		content := choice["message"].(map[string]any)["content"].(string)

// 		var resp models.SummarizeResponse
// 		if err := json.Unmarshal([]byte(content), &resp); err != nil {
// 			return models.SummarizeResponse{}, fmt.Errorf("failed to parse openai json response: %w", err)
// 		}
// 		return resp, nil
// 	}
// 	return models.SummarizeResponse{}, fmt.Errorf("unexpected openai response format")
// }
