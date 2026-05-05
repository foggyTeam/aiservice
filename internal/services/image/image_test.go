package image

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateFilename(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "простой URL",
			url:      "https://example.com/image.png",
			expected: fmt.Sprintf("%x.jpeg", sha256.Sum256([]byte("https://example.com/image.png"))),
		},
		{
			name:     "URL с параметрами",
			url:      "https://cdn.example.com/img.jpg?size=large&v=2",
			expected: fmt.Sprintf("%x.jpeg", sha256.Sum256([]byte("https://cdn.example.com/img.jpg?size=large&v=2"))),
		},
		{
			name:     "одинаковые URL - одинаковый хеш",
			url:      "https://same.com/pic",
			expected: fmt.Sprintf("%x.jpeg", sha256.Sum256([]byte("https://same.com/pic"))),
		},
		{
			name: "разные URL - разные хеши",
		},
	}

	s := NewService("/tmp", time.Second)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := s.generateFilename(tt.url)
			if tt.expected != "" {
				assert.Equal(t, tt.expected, got)
			}
		})
	}

	t.Run("разные URL - разные файлы", func(t *testing.T) {
		t.Parallel()
		name1 := s.generateFilename("https://a.com/1")
		name2 := s.generateFilename("https://a.com/2")
		assert.NotEqual(t, name1, name2, "разные URL должны давать разные имена файлов")
	})
}

func TestNewService(t *testing.T) {
	t.Parallel()

	svc := NewService("/my/images", 30*time.Second)

	assert.Equal(t, "/my/images", svc.imagesDir)
	assert.NotNil(t, svc.httpClient)
	assert.Equal(t, 30*time.Second, svc.httpClient.Timeout)
}

func TestDownloadImage_Success(t *testing.T) {

	tempDir := t.TempDir()

	testImage := []byte("fake-image-data-12345")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "image/jpeg")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(testImage)
	}))
	defer server.Close()

	svc := NewService(tempDir, 5*time.Second)

	ctx := context.Background()
	result, err := svc.DownloadImage(ctx, server.URL)

	require.NoError(t, err)
	assert.Equal(t, server.URL, result.URL)
	assert.Equal(t, int64(len(testImage)), result.Size)
	assert.FileExists(t, result.LocalPath)
	assert.Contains(t, result.LocalPath, tempDir)
	assert.Contains(t, result.LocalPath, ".jpeg")

	content, err := os.ReadFile(result.LocalPath)
	require.NoError(t, err)
	assert.Equal(t, testImage, content)
}

func TestDownloadImage_InvalidURL(t *testing.T) {
	tempDir := t.TempDir()
	svc := NewService(tempDir, 2*time.Second)

	ctx := context.Background()
	_, err := svc.DownloadImage(ctx, "not-a-valid-url")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to download")
}

func TestDownloadImage_HTTPError(t *testing.T) {
	tempDir := t.TempDir()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("Not Found"))
	}))
	defer server.Close()

	svc := NewService(tempDir, 2*time.Second)
	ctx := context.Background()

	_, err := svc.DownloadImage(ctx, server.URL)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "status 404")
}

func TestDownloadImage_EmptyImagesDir(t *testing.T) {
	t.Parallel()

	svc := NewService("", time.Second)
	ctx := context.Background()

	_, err := svc.DownloadImage(ctx, "https://example.com/img.jpg")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "images directory not configured")
}

func TestDownloadImage_ContextCancellation(t *testing.T) {
	tempDir := t.TempDir()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
	}))
	defer server.Close()

	svc := NewService(tempDir, 5*time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := svc.DownloadImage(ctx, server.URL)

	require.Error(t, err)
	assert.True(t, err != nil)
}

func TestDownloadImage_DirectoryCreation(t *testing.T) {
	tempBase := t.TempDir()
	newDir := filepath.Join(tempBase, "new", "nested", "images")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("img"))
	}))
	defer server.Close()

	svc := NewService(newDir, 2*time.Second)
	ctx := context.Background()

	result, err := svc.DownloadImage(ctx, server.URL)

	require.NoError(t, err)
	assert.DirExists(t, newDir)
	assert.FileExists(t, result.LocalPath)
}

func TestDownloadImage_FileOverwrite(t *testing.T) {
	tempDir := t.TempDir()

	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("version-1"))
	}))
	defer server1.Close()

	svc := NewService(tempDir, 2*time.Second)
	ctx := context.Background()

	_, err := svc.DownloadImage(ctx, server1.URL)
	require.NoError(t, err)

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("version-2-new"))
	}))
	defer server2.Close()

	_, _ = svc.DownloadImage(ctx, server2.URL)

	entries, err := os.ReadDir(tempDir)
	require.NoError(t, err)
	assert.Len(t, entries, 2, "два разных URL должны создать два файла")
}

func TestCleanOldImages_RemovesExpired(t *testing.T) {
	tempDir := t.TempDir()
	svc := NewService(tempDir, time.Second)

	oldFile := filepath.Join(tempDir, "old.jpeg")
	require.NoError(t, os.WriteFile(oldFile, []byte("old"), 0644))
	oldTime := time.Now().Add(-2 * time.Hour)
	require.NoError(t, os.Chtimes(oldFile, oldTime, oldTime))

	newFile := filepath.Join(tempDir, "new.jpeg")
	require.NoError(t, os.WriteFile(newFile, []byte("new"), 0644))

	svc.cleanOldImages(1 * time.Hour)

	_, err := os.Stat(oldFile)
	assert.True(t, os.IsNotExist(err), "старый файл должен быть удалён")

	_, err = os.Stat(newFile)
	assert.NoError(t, err, "новый файл должен остаться")
}

func TestCleanOldImages_KeepsWithinTTL(t *testing.T) {
	tempDir := t.TempDir()
	svc := NewService(tempDir, time.Hour)

	file := filepath.Join(tempDir, "recent.jpeg")
	require.NoError(t, os.WriteFile(file, []byte("data"), 0644))
	recentTime := time.Now().Add(-30 * time.Minute)
	require.NoError(t, os.Chtimes(file, recentTime, recentTime))

	svc.cleanOldImages(1 * time.Hour)

	_, err := os.Stat(file)
	assert.NoError(t, err, "файл в пределах TTL не должен удаляться")
}

func TestCleanOldImages_EmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()
	svc := NewService(tempDir, time.Hour)

	assert.NotPanics(t, func() {
		svc.cleanOldImages(1 * time.Hour)
	})
}

func TestCleanOldImages_SkipsSubdirectories(t *testing.T) {
	tempDir := t.TempDir()
	svc := NewService(tempDir, time.Minute)

	subDir := filepath.Join(tempDir, "subdir")
	require.NoError(t, os.Mkdir(subDir, 0755))
	oldFile := filepath.Join(subDir, "old.jpeg")
	require.NoError(t, os.WriteFile(oldFile, []byte("old"), 0644))
	oldTime := time.Now().Add(-1 * time.Hour)
	require.NoError(t, os.Chtimes(oldFile, oldTime, oldTime))

	svc.cleanOldImages(10 * time.Minute)

	_, err := os.Stat(oldFile)
	assert.NoError(t, err, "файлы в поддиректориях не обрабатываются")
}

func TestStartCleaner_StartsAndStops(t *testing.T) {
	tempDir := t.TempDir()
	svc := NewService(tempDir, time.Minute)

	ctx, cancel := context.WithCancel(context.Background())

	svc.StartCleaner(ctx, 10*time.Second)

	time.Sleep(50 * time.Millisecond)

	cancel()
	time.Sleep(100 * time.Millisecond)
}

func TestStartCleaner_MultipleCalls(t *testing.T) {
	tempDir := t.TempDir()
	svc := NewService(tempDir, time.Minute)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	assert.NotPanics(t, func() {
		svc.StartCleaner(ctx, time.Second)
		svc.StartCleaner(ctx, time.Second)
		svc.StartCleaner(ctx, time.Second)
	})
}

func TestService_FullLifecycle(t *testing.T) {
	tempDir := t.TempDir()

	imageData := []byte("integration-test-image")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(imageData)
	}))
	defer server.Close()

	svc := NewService(tempDir, 5*time.Second)
	ctx := context.Background()

	result, err := svc.DownloadImage(ctx, server.URL)
	require.NoError(t, err)
	assert.FileExists(t, result.LocalPath)

	oldTime := time.Now().Add(-2 * time.Hour)
	require.NoError(t, os.Chtimes(result.LocalPath, oldTime, oldTime))

	svc.cleanOldImages(1 * time.Hour)

	_, err = os.Stat(result.LocalPath)
	assert.True(t, os.IsNotExist(err), "файл должен быть удалён после очистки")
}
