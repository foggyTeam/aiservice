package providers

import (
	"sync"
	"time"
)

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState string

const (
	ClosedState   CircuitBreakerState = "closed"   // Normal operation
	OpenState     CircuitBreakerState = "open"     // Tripped, requests blocked
	HalfOpenState CircuitBreakerState = "half_open" // Testing recovery
)

// CircuitBreaker implements the circuit breaker pattern for providers
type CircuitBreaker struct {
	breakers map[string]*singleCircuitBreaker
	mutex    sync.RWMutex
}

// singleCircuitBreaker represents a circuit breaker for a single provider
type singleCircuitBreaker struct {
	state          CircuitBreakerState
	failureCount   int
	lastFailure    time.Time
	openUntil      time.Time
	maxFailures    int
	resetTimeout   time.Duration
	halfOpenTryAt  time.Time
	halfOpenTryNum int
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker() *CircuitBreaker {
	return &CircuitBreaker{
		breakers: make(map[string]*singleCircuitBreaker),
	}
}

// IsOpen returns true if the circuit breaker for the given provider is open
func (cb *CircuitBreaker) IsOpen(providerName string) bool {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	breaker, exists := cb.breakers[providerName]
	if !exists {
		return false
	}

	now := time.Now()
	
	// Check if we're past the open timeout
	if breaker.state == OpenState && now.After(breaker.openUntil) {
		// Transition to half-open state
		breaker.state = HalfOpenState
		breaker.halfOpenTryAt = now.Add(1 * time.Second) // Allow one test request after 1 second
		breaker.halfOpenTryNum = 0
	}

	// In half-open state, allow one request after the timeout
	if breaker.state == HalfOpenState {
		if now.After(breaker.halfOpenTryAt) {
			breaker.halfOpenTryNum++
			// Allow only one trial request in half-open state
			if breaker.halfOpenTryNum == 1 {
				// Reset timer for next trial
				breaker.halfOpenTryAt = now.Add(5 * time.Second)
				return false // Allow this request
			}
			// Block other requests while testing
			return true
		}
		return true // Still waiting for trial timeout
	}

	return breaker.state == OpenState
}

// Trip trips the circuit breaker for the given provider
func (cb *CircuitBreaker) Trip(providerName string) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	breaker, exists := cb.breakers[providerName]
	if !exists {
		breaker = &singleCircuitBreaker{
			maxFailures:  3,              // Trip after 3 failures
			resetTimeout: 30 * time.Second, // Wait 30 seconds before retrying
		}
		cb.breakers[providerName] = breaker
	}

	breaker.failureCount++
	breaker.lastFailure = time.Now()

	// Trip the circuit if we've exceeded the failure threshold
	if breaker.failureCount >= breaker.maxFailures {
		breaker.state = OpenState
		breaker.openUntil = time.Now().Add(breaker.resetTimeout)
	}
}

// Reset resets the circuit breaker for the given provider
func (cb *CircuitBreaker) Reset(providerName string) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	breaker, exists := cb.breakers[providerName]
	if !exists {
		breaker = &singleCircuitBreaker{
			maxFailures:  3,
			resetTimeout: 30 * time.Second,
		}
		cb.breakers[providerName] = breaker
	}

	breaker.state = ClosedState
	breaker.failureCount = 0
	breaker.lastFailure = time.Time{}
	breaker.openUntil = time.Time{}
	breaker.halfOpenTryAt = time.Time{}
	breaker.halfOpenTryNum = 0
}

// Success should be called when a request succeeds
func (cb *CircuitBreaker) Success(providerName string) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	breaker, exists := cb.breakers[providerName]
	if !exists {
		// Create a new breaker if it doesn't exist
		breaker = &singleCircuitBreaker{
			maxFailures:  3,
			resetTimeout: 30 * time.Second,
		}
		cb.breakers[providerName] = breaker
		return
	}

	// If we're in half-open state and succeeded, reset the breaker
	if breaker.state == HalfOpenState {
		breaker.state = ClosedState
		breaker.failureCount = 0
		breaker.lastFailure = time.Time{}
		breaker.openUntil = time.Time{}
		breaker.halfOpenTryAt = time.Time{}
		breaker.halfOpenTryNum = 0
	}
}