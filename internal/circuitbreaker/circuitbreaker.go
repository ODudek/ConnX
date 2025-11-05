package circuitbreaker

import (
	"sync"
	"time"
)

// State represents the circuit breaker state
type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	mu              sync.RWMutex
	state           State
	failureCount    uint64
	successCount    uint64
	lastFailureTime time.Time
	lastStateChange time.Time

	// Configuration
	maxFailures     uint64        // Number of failures before opening
	timeout         time.Duration // Time to wait before trying half-open
	resetTimeout    time.Duration // Time to wait in half-open before resetting
	halfOpenSuccess uint64        // Consecutive successes needed in half-open to close
}

// Config holds circuit breaker configuration
type Config struct {
	MaxFailures     uint64
	Timeout         time.Duration
	ResetTimeout    time.Duration
	HalfOpenSuccess uint64
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config Config) *CircuitBreaker {
	if config.MaxFailures == 0 {
		config.MaxFailures = 5
	}
	if config.Timeout == 0 {
		config.Timeout = 60 * time.Second
	}
	if config.ResetTimeout == 0 {
		config.ResetTimeout = 30 * time.Second
	}
	if config.HalfOpenSuccess == 0 {
		config.HalfOpenSuccess = 2
	}

	return &CircuitBreaker{
		state:           StateClosed,
		maxFailures:     config.MaxFailures,
		timeout:         config.Timeout,
		resetTimeout:    config.ResetTimeout,
		halfOpenSuccess: config.HalfOpenSuccess,
		lastStateChange: time.Now(),
	}
}

// IsAllowed checks if a request is allowed through the circuit breaker
func (cb *CircuitBreaker) IsAllowed() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		// Check if timeout has elapsed to move to half-open
		if time.Since(cb.lastStateChange) > cb.timeout {
			cb.state = StateHalfOpen
			cb.successCount = 0
			cb.lastStateChange = time.Now()
			return true
		}
		return false
	case StateHalfOpen:
		return true
	}
	return false
}

// RecordSuccess records a successful request
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount = 0

	if cb.state == StateHalfOpen {
		cb.successCount++
		if cb.successCount >= cb.halfOpenSuccess {
			cb.state = StateClosed
			cb.lastStateChange = time.Now()
		}
	}
}

// RecordFailure records a failed request
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.lastFailureTime = time.Now()
	cb.failureCount++

	switch cb.state {
	case StateClosed:
		if cb.failureCount >= cb.maxFailures {
			cb.state = StateOpen
			cb.lastStateChange = time.Now()
		}
	case StateHalfOpen:
		// Any failure in half-open state immediately opens the circuit
		cb.state = StateOpen
		cb.successCount = 0
		cb.lastStateChange = time.Now()
	}
}

// GetState returns the current state
func (cb *CircuitBreaker) GetState() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetStats returns current statistics
func (cb *CircuitBreaker) GetStats() (state State, failures uint64, successes uint64) {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state, cb.failureCount, cb.successCount
}

// Reset manually resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.failureCount = 0
	cb.successCount = 0
	cb.lastStateChange = time.Now()
}

// String returns string representation of state
func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}
