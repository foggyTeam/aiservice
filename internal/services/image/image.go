package image

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Service handles image downloading and cleanup
type Service struct {
	imagesDir  string
	httpClient *http.Client
}

// NewService creates a new image service instance
func NewService(imagesDir string, timeout time.Duration) *Service {
	return &Service{
		imagesDir:  imagesDir,
		httpClient: &http.Client{Timeout: timeout},
	}
}

// DownloadResult contains the result of an image download
type DownloadResult struct {
	LocalPath string
	URL       string
	Size      int64
}

// DownloadImage downloads an image from URL to local storage
func (s *Service) DownloadImage(ctx context.Context, imageURL string) (*DownloadResult, error) {
	if s.imagesDir == "" {
		return nil, fmt.Errorf("images directory not configured")
	}

	// Generate unique filename from URL
	filename, err := s.generateFilename(imageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to generate filename: %w", err)
	}

	localPath := filepath.Join(s.imagesDir, filename)

	// Check if file already exists
	if _, err := os.Stat(localPath); err == nil {
		return &DownloadResult{
			LocalPath: localPath,
			URL:       imageURL,
		}, nil
	}

	// Download the image
	req, err := http.NewRequestWithContext(ctx, "GET", imageURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("image download failed with status %d", resp.StatusCode)
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(s.imagesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create images directory: %w", err)
	}

	// Create the file
	file, err := os.Create(localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy the content
	size, err := io.Copy(file, resp.Body)
	if err != nil {
		// Clean up on error
		os.Remove(localPath)
		return nil, fmt.Errorf("failed to save image: %w", err)
	}

	return &DownloadResult{
		LocalPath: localPath,
		URL:       imageURL,
		Size:      size,
	}, nil
}

// DeleteImage removes a downloaded image from local storage
func (s *Service) DeleteImage(localPath string) error {
	if localPath == "" {
		return nil
	}

	// Verify the file is in the images directory
	absPath, err := filepath.Abs(localPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	absImagesDir, err := filepath.Abs(s.imagesDir)
	if err != nil {
		return fmt.Errorf("failed to resolve images directory: %w", err)
	}

	// Security check: ensure file is within images directory
	if !strings.HasPrefix(absPath, absImagesDir) {
		return fmt.Errorf("security violation: attempted to delete file outside images directory")
	}

	return os.Remove(localPath)
}

// generateFilename creates a unique filename from URL
func (s *Service) generateFilename(url string) (string, error) {
	hash := sha256.Sum256([]byte(url))

	// Get file extension from URL
	ext := ".jpeg" // default
	if idx := strings.LastIndex(url, "."); idx != -1 {
		if idx < len(url)-1 {
			potentialExt := url[idx:]
			if len(potentialExt) <= 5 {
				ext = potentialExt
			}
		}
	}

	return fmt.Sprintf("%x%s", hash, ext), nil
}
