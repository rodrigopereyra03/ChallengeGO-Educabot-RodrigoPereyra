package resilience_test

import (
	"testing"
	"time"

	"educabot.com/bookshop/internal/platform/resilience"
	"github.com/stretchr/testify/assert"
)

func TestCircuitBreaker_ClosedInitially(t *testing.T) {
	cb := resilience.NewCircuitBreaker(2, time.Second)

	err := cb.AllowRequest()
	assert.NoError(t, err, "circuit breaker should allow requests initially")
}

func TestCircuitBreaker_OpensAfterFailureLimit(t *testing.T) {
	cb := resilience.NewCircuitBreaker(2, time.Second)

	cb.ReportFailure()
	cb.ReportFailure()

	err := cb.AllowRequest()
	assert.ErrorIs(t, err, resilience.ErrCircuitOpen, "circuit should be open after reaching failure limit")
}

func TestCircuitBreaker_ClosesAfterTimeout(t *testing.T) {
	cb := resilience.NewCircuitBreaker(1, 100*time.Millisecond)

	cb.ReportFailure()

	err := cb.AllowRequest()
	assert.Error(t, err, "circuit should be open right after failure")

	time.Sleep(150 * time.Millisecond)

	err = cb.AllowRequest()
	assert.NoError(t, err, "circuit should be closed after openDuration expires")
}

func TestCircuitBreaker_SuccessResetsFailures(t *testing.T) {
	cb := resilience.NewCircuitBreaker(2, time.Second)

	cb.ReportFailure()

	cb.ReportSuccess()

	err := cb.AllowRequest()
	assert.NoError(t, err, "circuit should allow request after success resets failure count")
}
