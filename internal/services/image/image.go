package image

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
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
	filename := s.generateFilename(imageURL)
	localPath := filepath.Join(s.imagesDir, filename)

	// Check if file already exists
	if info, err := os.Stat(localPath); err == nil {
		slog.Info("Image already exists, reusing", "path", localPath, "size", info.Size())
		slog.Info("Image exists, NEED FIX SAME URL")

		// return &DownloadResult{
		// 	LocalPath: localPath,
		// 	URL:       imageURL,
		// 	Size:      info.Size(),
		// }, nil
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

	slog.Info("Image downloaded", "path", localPath, "size", size)

	return &DownloadResult{
		LocalPath: localPath,
		URL:       imageURL,
		Size:      size,
	}, nil
}

// StartCleaner starts a background process to clean up old images
func (s *Service) StartCleaner(ctx context.Context, ttl time.Duration) {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		slog.Info("Image cleaner started", "ttl", ttl)

		for {
			select {
			case <-ctx.Done():
				slog.Info("Image cleaner stopped")
				return
			case <-ticker.C:
				s.cleanOldImages(ttl)
			}
		}
	}()
}

// cleanOldImages removes images older than ttl
func (s *Service) cleanOldImages(ttl time.Duration) {
	entries, err := os.ReadDir(s.imagesDir)
	if err != nil {
		slog.Warn("Failed to read images directory", "err", err)
		return
	}

	now := time.Now()
	cutoff := now.Add(-ttl)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			path := filepath.Join(s.imagesDir, entry.Name())
			if err := os.Remove(path); err != nil {
				slog.Warn("Failed to remove old image", "path", path, "err", err)
			} else {
				slog.Debug("Removed old image", "path", path)
			}
		}
	}
}

// generateFilename creates a unique filename from URL
func (s *Service) generateFilename(url string) string {
	hash := sha256.Sum256([]byte(url))
	return fmt.Sprintf("%x.jpeg", hash)
}
