package digitalink

import (
	"context"
	"time"

	"github.com/aiservice/internal/models"
)

// RealDigitalInkClient wraps the existing digitalink.Client to implement DigitalInkRecognizer
type RealDigitalInkClient struct {
	client *Client
}

// NewRealDigitalInkClient creates a new RealDigitalInkClient
func NewRealDigitalInkClient(language string, timeout time.Duration) *RealDigitalInkClient {
	return &RealDigitalInkClient{
		client: NewClient(language, timeout),
	}
}

// RecognizeInk implements DigitalInkRecognizer interface
func (r *RealDigitalInkClient) RecognizeInk(ctx context.Context, elements []models.Element) (string, error) {
	return r.client.RecognizeInk(ctx, elements)
}
