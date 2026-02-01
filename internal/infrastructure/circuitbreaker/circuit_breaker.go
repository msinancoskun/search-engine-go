package circuitbreaker

import (
	"context"
	"errors"
	"sync"
	"time"
)

type CircuitState int

const (
	CircuitStateClosed CircuitState = iota
	CircuitStateOpen
	CircuitStateHalfOpen
)

type CircuitBreaker struct {
	maxFailures     int
	resetTimeout    time.Duration
	halfOpenTimeout time.Duration

	mu              sync.RWMutex
	state           CircuitState
	failureCount    int
	lastFailureTime time.Time
	successCount    int
}

func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:     maxFailures,
		resetTimeout:    resetTimeout,
		halfOpenTimeout: resetTimeout / 2,
		state:           CircuitStateClosed,
	}
}

func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	state := cb.getState()

	switch state {
	case CircuitStateOpen:
		cb.mu.RLock()
		timeSinceLastFailure := time.Since(cb.lastFailureTime)
		cb.mu.RUnlock()

		if timeSinceLastFailure >= cb.resetTimeout {
			cb.transitionToHalfOpen()
		} else {
			return errors.New("circuit breaker is open")
		}

	case CircuitStateHalfOpen:
		break

	case CircuitStateClosed:
		break
	}

	err := fn()

	if err != nil {
		cb.recordFailure()
		return err
	}

	cb.recordSuccess()
	return nil
}

func (cb *CircuitBreaker) getState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

func (cb *CircuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	cb.lastFailureTime = time.Now()

	if cb.state == CircuitStateHalfOpen {
		cb.state = CircuitStateOpen
		cb.failureCount = 1
		cb.successCount = 0
	} else if cb.failureCount >= cb.maxFailures {
		cb.state = CircuitStateOpen
	}
}

func (cb *CircuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount = 0

	switch cb.state {
	case CircuitStateHalfOpen:
		cb.successCount++
		if cb.successCount >= 2 {
			cb.state = CircuitStateClosed
			cb.successCount = 0
		}
	case CircuitStateOpen:
		cb.state = CircuitStateClosed
	}
}

func (cb *CircuitBreaker) transitionToHalfOpen() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitStateOpen:
		cb.state = CircuitStateHalfOpen
		cb.successCount = 0
		cb.failureCount = 0
	}
}

func (cb *CircuitBreaker) GetState() CircuitState {
	return cb.getState()
}

func (s CircuitState) String() string {
	switch s {
	case CircuitStateClosed:
		return "closed"
	case CircuitStateOpen:
		return "open"
	case CircuitStateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = CircuitStateClosed
	cb.failureCount = 0
	cb.successCount = 0
}
