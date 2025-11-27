package resilience

import (
	"errors"
	"sync"
	"time"
)

type CircuitBreaker struct {
	failures     int
	failureLimit int
	openUntil    time.Time
	openDuration time.Duration
	mu           sync.Mutex
}

var ErrCircuitOpen = errors.New("circuit breaker abierto")

func NewCircuitBreaker(limit int, openDuration time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		failureLimit: limit,
		openDuration: openDuration,
	}
}

func (cb *CircuitBreaker) AllowRequest() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if time.Now().Before(cb.openUntil) {
		return ErrCircuitOpen
	}

	return nil
}

func (cb *CircuitBreaker) ReportFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	if cb.failures >= cb.failureLimit {
		cb.openUntil = time.Now().Add(cb.openDuration)
		cb.failures = 0
	}
}

func (cb *CircuitBreaker) ReportSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
}
