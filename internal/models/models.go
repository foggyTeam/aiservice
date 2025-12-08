package models

import (
	"encoding/json"
)

type AnalyzeRequest struct {
	BoardID     string          `json:"board_id" validate:"required"`
	UserID      string          `json:"user_id" validate:"required"`
	Input       json.RawMessage `json:"input" validate:"required"`
	Context     map[string]any  `json:"context,omitempty"`
	CallbackURL string          `json:"callback_url,omitempty"`
	RequestID   string          `json:"request_id,omitempty"`
	Type        string          `json:"type" validate:"required"`
}

type InkInput struct {
	Type    string         `json:"type"`
	Strokes [][]InkPoint   `json:"strokes" validate:"required"`
	Meta    map[string]any `json:"meta,omitempty"`
}

type InkPoint struct {
	X        float64 `json:"x" validate:"required"`
	Y        float64 `json:"y" validate:"required"`
	T        int64   `json:"t,omitempty"`        // timestamp in ms
	Pressure float64 `json:"pressure,omitempty"` // 0-1
	Tilt     float64 `json:"tilt,omitempty"`     // angle in degrees
}

type ImageInput struct {
	Type     string         `json:"type"`
	ImageURL string         `json:"image_url" validate:"required"`
	Base64   string         `json:"base64,omitempty"` // alternative to URL
	Meta     map[string]any `json:"meta,omitempty"`
}

type TextInput struct {
	Type string `json:"type"`
	Text string `json:"text" validate:"required"`
}

// ===== Response Models =====

type AnalyzeResponse struct {
	Intent      string         `json:"intent"`
	Confidence  float64        `json:"confidence"`
	Actions     []Action       `json:"actions"`
	Explanation string         `json:"explanation,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type Action struct {
	Type    string         `json:"type"`
	Payload map[string]any `json:"payload,omitempty"`
	Params  map[string]any `json:"params,omitempty"`
}

type AcceptedResponse struct {
	JobID     string `json:"job_id"`
	Status    string `json:"status"`
	CreatedAt int64  `json:"created_at"`
	ExpiresAt int64  `json:"expires_at"`
}

func (a AcceptedResponse) Error() string {
	return "job accepted with ID: " + a.JobID
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

type Job struct {
	ID        string
	Request   AnalyzeRequest
	CreatedAt int64
	Retries   int
	Status    JobStatus
}

type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
)

type TranscriptionResult struct {
	Text     string
	Language string
	Metadata map[string]any
}
