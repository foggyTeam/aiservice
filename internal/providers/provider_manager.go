package providers

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/aiservice/internal/models"
	"github.com/firebase/genkit/go/ai"
)

// ProviderError represents an error from a specific provider
type ProviderError struct {
	Type         ProviderErrorType
	Message      string
	StatusCode   int
	ProviderName string
	OriginalErr  error
}

func (e *ProviderError) Error() string {
	return fmt.Sprintf("provider %s error (%s): %s", e.ProviderName, e.Type, e.Message)
}

// Unwrap returns the original error
func (e *ProviderError) Unwrap() error {
	return e.OriginalErr
}

// ProviderErrorType represents different types of provider errors
type ProviderErrorType string

const (
	AccessDeniedError  ProviderErrorType = "access_denied"  // 403 errors
	InternalError      ProviderErrorType = "internal_error" // 500 errors
	RateLimitError     ProviderErrorType = "rate_limit"     // 429 errors
	TimeoutError       ProviderErrorType = "timeout"
	AuthError          ProviderErrorType = "auth_failure"
	ConnectionError    ProviderErrorType = "connection_error"
	UnknownError       ProviderErrorType = "unknown"
)

// ProviderStatus represents the operational status of a provider
type ProviderStatus string

const (
	StatusHealthy     ProviderStatus = "healthy"
	StatusUnhealthy   ProviderStatus = "unhealthy"
	StatusRestricted  ProviderStatus = "restricted" // Regional restrictions
	StatusRateLimited ProviderStatus = "rate_limited"
)

// ProviderInfo contains information about a provider
type ProviderInfo struct {
	Name        string
	Status      ProviderStatus
	LastCheck   time.Time
	ErrorCount  int
	LastError   *ProviderError
	Region      string // Target region for this provider
	Priority    int    // Lower number = higher priority
	Enabled     bool
}

// MultiProviderConfig holds configuration for multiple providers
type MultiProviderConfig struct {
	Providers []ProviderConfig `json:"providers"`
}

// ProviderConfig holds configuration for a single provider
type ProviderConfig struct {
	Name     string        `json:"name"`
	APIKey   string        `json:"api_key"`
	BaseURL  string        `json:"base_url"`
	Model    string        `json:"model"`
	Timeout  time.Duration `json:"timeout"`
	Regions  []string      `json:"regions"`  // Supported regions
	Priority int           `json:"priority"` // Lower number = higher priority
	Enabled  bool          `json:"enabled"`
}

// ProviderManager manages multiple AI providers with automatic failover
type ProviderManager struct {
	providers      map[string]LLMClient
	providerInfos  map[string]*ProviderInfo
	mutex          sync.RWMutex
	circuitBreaker *CircuitBreaker
}

// NewProviderManager creates a new provider manager with the given configuration
func NewProviderManager(config *MultiProviderConfig) *ProviderManager {
	pm := &ProviderManager{
		providers:     make(map[string]LLMClient),
		providerInfos: make(map[string]*ProviderInfo),
		circuitBreaker: NewCircuitBreaker(),
	}

	// Initialize provider info from config
	for _, providerCfg := range config.Providers {
		if providerCfg.Enabled {
			pm.providerInfos[providerCfg.Name] = &ProviderInfo{
				Name:     providerCfg.Name,
				Status:   StatusHealthy,
				Priority: providerCfg.Priority,
				Region:   "", // Will be set based on request
				Enabled:  true,
			}
		}
	}

	return pm
}

// RegisterProvider registers a new provider with the manager
func (pm *ProviderManager) RegisterProvider(name string, client LLMClient) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	pm.providers[name] = client
	if _, exists := pm.providerInfos[name]; !exists {
		pm.providerInfos[name] = &ProviderInfo{
			Name:     name,
			Status:   StatusHealthy,
			Priority: 10, // Default priority
			Enabled:  true,
		}
	}
}



// classifyError categorizes an error from a provider
func (pm *ProviderManager) classifyError(err error, providerName string) *ProviderError {
	// This is a simplified classification - in practice, you'd need more sophisticated error parsing
	errStr := err.Error()
	
	// Check for common error patterns
	if containsAny(errStr, []string{"403", "Forbidden", "access denied", "unauthorized"}) {
		return &ProviderError{
			Type:         AccessDeniedError,
			Message:      errStr,
			StatusCode:   403,
			ProviderName: providerName,
			OriginalErr:  err,
		}
	}
	
	if containsAny(errStr, []string{"500", "Internal Server Error", "internal error", "server error"}) {
		return &ProviderError{
			Type:         InternalError,
			Message:      errStr,
			StatusCode:   500,
			ProviderName: providerName,
			OriginalErr:  err,
		}
	}
	
	if containsAny(errStr, []string{"429", "Too Many Requests", "rate limit", "quota"}) {
		return &ProviderError{
			Type:         RateLimitError,
			Message:      errStr,
			StatusCode:   429,
			ProviderName: providerName,
			OriginalErr:  err,
		}
	}
	
	if containsAny(errStr, []string{"timeout", "deadline exceeded", "connection refused", "connection reset"}) {
		return &ProviderError{
			Type:         TimeoutError,
			Message:      errStr,
			ProviderName: providerName,
			OriginalErr:  err,
		}
	}

	// Default to unknown error
	return &ProviderError{
		Type:         UnknownError,
		Message:      errStr,
		ProviderName: providerName,
		OriginalErr:  err,
	}
}

// isCriticalError determines if an error type should trigger a provider switch
func (pm *ProviderManager) isCriticalError(errorType ProviderErrorType) bool {
	switch errorType {
	case AccessDeniedError:  // 403 - regional restrictions
		return true
	case InternalError:      // 500 - provider infrastructure failure
		return true
	case RateLimitError:     // 429 - rate limiting (might be temporary)
		return true
	case ConnectionError:    // Connection issues
		return true
	default:
		return false
	}
}

// getAvailableProviders returns provider names in priority order
func (pm *ProviderManager) getAvailableProviders() []string {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	var providers []string
	for name, info := range pm.providerInfos {
		if info.Enabled && info.Status != StatusUnhealthy {
			providers = append(providers, name)
		}
	}

	// Sort by priority (lower number = higher priority)
	// This is a simple bubble sort for demonstration - could be optimized
	for i := 0; i < len(providers); i++ {
		for j := i + 1; j < len(providers); j++ {
			infoI := pm.providerInfos[providers[i]]
			infoJ := pm.providerInfos[providers[j]]
			
			if infoI.Priority > infoJ.Priority {
				providers[i], providers[j] = providers[j], providers[i]
			}
		}
	}

	return providers
}

// markProviderHealthy marks a provider as healthy
func (pm *ProviderManager) markProviderHealthy(providerName string) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if info, exists := pm.providerInfos[providerName]; exists {
		info.Status = StatusHealthy
		info.LastCheck = time.Now()
		info.ErrorCount = 0
		info.LastError = nil
		
		// Reset circuit breaker for this provider
		pm.circuitBreaker.Reset(providerName)
	}
}

// markProviderUnhealthy marks a provider as unhealthy
func (pm *ProviderManager) markProviderUnhealthy(providerName string, providerErr *ProviderError) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if info, exists := pm.providerInfos[providerName]; exists {
		info.Status = StatusUnhealthy
		info.LastCheck = time.Now()
		info.ErrorCount++
		info.LastError = providerErr
		
		// Trip circuit breaker for this provider
		pm.circuitBreaker.Trip(providerName)
	}
}

// containsAny checks if a string contains any of the substrings
func containsAny(str string, substrs []string) bool {
	for _, substr := range substrs {
		if containsIgnoreCase(str, substr) {
			return true
		}
	}
	return false
}

// containsIgnoreCase checks if a string contains a substring ignoring case
func containsIgnoreCase(str, substr string) bool {
	return len(str) >= len(substr) && 
		   (contains(str[:len(str)-len(substr)+1], substr) ||
		    contains(str[len(str)-len(substr):], substr))
}

// contains is a helper function for case-insensitive substring search
func contains(str, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(substr) > len(str) {
		return false
	}
	
	lowerStr := toLower(str)
	lowerSubstr := toLower(substr)
	
	for i := 0; i <= len(lowerStr)-len(lowerSubstr); i++ {
		if lowerStr[i:i+len(lowerSubstr)] == lowerSubstr {
			return true
		}
	}
	return false
}

// toLower converts a string to lowercase
func toLower(s string) string {
	var result []byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result = append(result, c+'a'-'A')
		} else {
			result = append(result, c)
		}
	}
	return string(result)
}

// Summarize implements the LLMClient interface
func (pm *ProviderManager) Summarize(ctx context.Context, parts []*ai.Part) (models.SummarizeResponse, error) {
	// Get available providers in priority order
	availableProviders := pm.getAvailableProviders()

	if len(availableProviders) == 0 {
		return models.SummarizeResponse{}, errors.New("no AI models currently working")
	}

	for _, providerName := range availableProviders {
		provider, exists := pm.providers[providerName]
		if !exists {
			continue
		}

		// Skip if circuit breaker is open for this provider
		if pm.circuitBreaker.IsOpen(providerName) {
			continue
		}

		// Attempt to process with this provider
		resp, err := provider.Summarize(ctx, parts)

		if err == nil {
			// Success - mark provider as healthy and return
			pm.markProviderHealthy(providerName)
			return resp, nil
		}

		// Handle provider-specific error
		providerErr := pm.classifyError(err, providerName)

		// If it's a critical error (403/500), mark provider as unhealthy and try next
		if pm.isCriticalError(providerErr.Type) {
			pm.markProviderUnhealthy(providerName, providerErr)
			continue
		}

		// For other errors, return immediately
		return models.SummarizeResponse{}, err
	}

	// All providers failed
	return models.SummarizeResponse{}, errors.New("no AI models currently working")
}

// Structurize implements the LLMClient interface
func (pm *ProviderManager) Structurize(ctx context.Context, parts []*ai.Part) (models.StructurizeResponse, error) {
	// Get available providers in priority order
	availableProviders := pm.getAvailableProviders()

	if len(availableProviders) == 0 {
		return models.StructurizeResponse{}, errors.New("no AI models currently working")
	}

	for _, providerName := range availableProviders {
		provider, exists := pm.providers[providerName]
		if !exists {
			continue
		}

		// Skip if circuit breaker is open for this provider
		if pm.circuitBreaker.IsOpen(providerName) {
			continue
		}

		// Attempt to process with this provider
		resp, err := provider.Structurize(ctx, parts)

		if err == nil {
			// Success - mark provider as healthy and return
			pm.markProviderHealthy(providerName)
			return resp, nil
		}

		// Handle provider-specific error
		providerErr := pm.classifyError(err, providerName)

		// If it's a critical error (403/500), mark provider as unhealthy and try next
		if pm.isCriticalError(providerErr.Type) {
			pm.markProviderUnhealthy(providerName, providerErr)
			continue
		}

		// For other errors, return immediately
		return models.StructurizeResponse{}, err
	}

	// All providers failed
	return models.StructurizeResponse{}, errors.New("no AI models currently working")
}

// GetName returns the provider name for the manager
func (pm *ProviderManager) GetName() string {
	return "provider-manager"
}