package azure

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

type AzureInkRecognizer struct {
	cfg    config.OCRProviderConfig
	client *http.Client
}

func NewAzureInkRecognizer(cfg config.OCRProviderConfig) *AzureInkRecognizer {
	return &AzureInkRecognizer{
		cfg: cfg,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

func (a *AzureInkRecognizer) RecognizeInk(ctx context.Context, input models.InkInput) (models.TranscriptionResult, error) {
	// Convert strokes to Azure format
	body := map[string]any{
		"language": "auto",
		"strokes":  input.Strokes,
	}

	bodyBytes, _ := json.Marshal(body)
	req, _ := http.NewRequestWithContext(ctx, "POST", a.cfg.BaseURL+"/recognizeInk", bytes.NewReader(bodyBytes))
	req.Header.Set("Ocp-Apim-Subscription-Key", a.cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return models.TranscriptionResult{}, fmt.Errorf("azure request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return models.TranscriptionResult{}, fmt.Errorf("azure error %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return models.TranscriptionResult{}, fmt.Errorf("failed to parse azure response: %w", err)
	}

	text := extractString(result, "recognitionResult", "recognitionUnits", "0", "recognizedString")
	return models.TranscriptionResult{
		Text:     text,
		Language: "en",
		Metadata: result,
	}, nil
}

func (a *AzureInkRecognizer) RecognizeImage(ctx context.Context, input models.ImageInput) (models.TranscriptionResult, error) {
	// Use Azure Computer Vision OCR API
	body := map[string]any{
		"url": input.ImageURL,
	}

	bodyBytes, _ := json.Marshal(body)
	req, _ := http.NewRequestWithContext(ctx, "POST", a.cfg.BaseURL+"/recognizeText", bytes.NewReader(bodyBytes))
	req.Header.Set("Ocp-Apim-Subscription-Key", a.cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return models.TranscriptionResult{}, fmt.Errorf("azure ocr request failed: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return models.TranscriptionResult{}, fmt.Errorf("failed to parse azure ocr response: %w", err)
	}

	return models.TranscriptionResult{
		Text:     extractString(result, "recognitionResult", "lines", "0", "text"),
		Language: "en",
		Metadata: result,
	}, nil
}

func extractString(data map[string]interface{}, keys ...string) string {
	current := interface{}(data)
	for _, key := range keys {
		switch v := current.(type) {
		case map[string]interface{}:
			current = v[key]
		case []interface{}:
			idx := 0
			// Try to parse as index
			if i, err := json.Number(key).Int64(); err == nil {
				idx = int(i)
			}
			if idx < len(v) {
				current = v[idx]
			} else {
				return ""
			}
		default:
			return ""
		}
	}
	if s, ok := current.(string); ok {
		return s
	}
	return ""
}

// ===== MyScript Recognizer =====

type MyScriptRecognizer struct {
	cfg    config.OCRProviderConfig
	client *http.Client
}

func NewMyScriptRecognizer(cfg config.OCRProviderConfig) *MyScriptRecognizer {
	return &MyScriptRecognizer{
		cfg: cfg,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

func (m *MyScriptRecognizer) RecognizeInk(ctx context.Context, input models.InkInput) (models.TranscriptionResult, error) {
	// MyScript API for handwriting recognition
	body := map[string]interface{}{
		"inputUnits": m.convertStrokes(input.Strokes),
	}

	bodyBytes, _ := json.Marshal(body)
	req, _ := http.NewRequestWithContext(ctx, "POST", m.cfg.BaseURL+"/api/v4.0/recognizer", bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+m.cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return models.TranscriptionResult{}, fmt.Errorf("myscript request failed: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return models.TranscriptionResult{}, fmt.Errorf("failed to parse myscript response: %w", err)
	}

	// Parse results
	text := ""
	if results, ok := result["results"].([]interface{}); ok && len(results) > 0 {
		if r, ok := results[0].(map[string]interface{}); ok {
			text = fmt.Sprintf("%v", r["text"])
		}
	}

	return models.TranscriptionResult{
		Text:     text,
		Language: "en",
		Metadata: result,
	}, nil
}

func (m *MyScriptRecognizer) RecognizeImage(ctx context.Context, input models.ImageInput) (models.TranscriptionResult, error) {
	// MyScript OCR API for images
	return models.TranscriptionResult{
		Text: "[MyScript] OCR not yet implemented",
	}, nil
}

func (m *MyScriptRecognizer) convertStrokes(strokes [][]models.InkPoint) []interface{} {
	var result []interface{}
	for _, stroke := range strokes {
		unit := map[string]interface{}{
			"type": "stroke",
			"x":    make([]float64, len(stroke)),
			"y":    make([]float64, len(stroke)),
			"t":    make([]int64, len(stroke)),
			"p":    make([]float64, len(stroke)),
		}
		xArr := unit["x"].([]float64)
		yArr := unit["y"].([]float64)
		tArr := unit["t"].([]int64)
		pArr := unit["p"].([]float64)

		for i, pt := range stroke {
			xArr[i] = pt.X
			yArr[i] = pt.Y
			tArr[i] = pt.T
			pArr[i] = pt.Pressure
		}
		result = append(result, unit)
	}
	return result
}
