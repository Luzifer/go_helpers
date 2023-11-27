package backoff

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var errTestError = errors.New("Test-Error")

func TestBreakFree(t *testing.T) {
	var seen int
	err := NewBackoff().WithMaxIterations(5).Retry(func() error {
		seen++
		return NewErrCannotRetry(errTestError)
	})
	assert.Error(t, err)
	assert.Equal(t, 1, seen)
	assert.Equal(t, errTestError, err)
}

func TestMaxExecutionTime(t *testing.T) {
	b := NewBackoff()
	// Define these values even if they match the defaults as
	// the defaults might change and should not break this test
	b.MaxIterationTime = 60 * time.Second
	b.MaxTotalTime = 2500 * time.Millisecond
	b.MinIterationTime = 100 * time.Millisecond
	b.Multiplier = 1.5

	start := time.Now()

	err := b.Retry(func() error { return errTestError })

	// After 6 iterations the time of 2078.125ms and after 7 iterations
	// the time of 3217.1875ms should be reached and therefore no further
	// iteration should be done.
	if d := time.Since(start); d < 3000*time.Millisecond || d > 3400*time.Millisecond {
		t.Errorf("Function did not end within expected time: duration=%s", d)
	}

	if err == nil {
		t.Error("Retry function had successful exit")
	}
}

func TestMaxIterations(t *testing.T) {
	b := NewBackoff()
	b.MaxIterations = 5

	var counter int

	err := b.Retry(func() error {
		counter++
		return errTestError
	})

	if counter != 5 {
		t.Errorf("Function was not executed 5 times: counter=%d", counter)
	}

	if err == nil {
		t.Error("Retry function had successful exit")
	}
}

func TestSuccessfulExecution(t *testing.T) {
	b := NewBackoff()
	b.MaxIterations = 5

	err := b.Retry(func() error { return nil })
	if err != nil {
		t.Errorf("An error was thrown: %s", err)
	}
}

func TestWrappedError(t *testing.T) {
	b := NewBackoff()
	b.MaxIterations = 5

	err := b.Retry(func() error { return errTestError })

	if errors.Unwrap(err) != errTestError {
		t.Errorf("Error unwrapping did not yield test error: %v", err)
	}
}
