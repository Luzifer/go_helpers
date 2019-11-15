package backoff

import (
	"errors"
	"testing"
	"time"
)

var testError = errors.New("Test-Error")

func TestMaxExecutionTime(t *testing.T) {
	b := NewBackoff()
	// Define these values even if they match the defaults as
	// the defaults might change and should not break this test
	b.MaxIterationTime = 60 * time.Second
	b.MaxTotalTime = 2500 * time.Millisecond
	b.MinIterationTime = 100 * time.Millisecond
	b.Multiplier = 1.5

	var start = time.Now()

	err := b.Retry(func() error { return testError })

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
		return testError
	})

	if counter != 5 {
		t.Errorf("Function was not executed 5 times: counter=%d", counter)
	}

	if err == nil {
		t.Error("Retry function had successful exit")
	}
}

func TestWrappedError(t *testing.T) {
	b := NewBackoff()
	b.MaxIterations = 5

	err := b.Retry(func() error { return testError })

	if errors.Unwrap(err) != testError {
		t.Errorf("Error unwrapping did not yield test error: %v", err)
	}
}
