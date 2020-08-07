package backoff

import (
	"fmt"
	"time"
)

const (
	// Default value to use for number of iterations: infinite
	DefaultMaxIterations uint64 = 0
	// Default value to use for maximum iteration time
	DefaultMaxIterationTime = 60 * time.Second
	// Default value to use for maximum execution time: infinite
	DefaultMaxTotalTime time.Duration = 0
	// Default value to use for initial iteration time
	DefaultMinIterationTime = 100 * time.Millisecond
	// Default multiplier to apply to iteration time after each iteration
	DefaultMultipler float64 = 1.5
)

// Backoff holds the configuration for backoff function retries
type Backoff struct {
	MaxIterations    uint64
	MaxIterationTime time.Duration
	MaxTotalTime     time.Duration
	MinIterationTime time.Duration
	Multiplier       float64
}

// NewBackoff creates a new Backoff configuration with default values (see constants)
func NewBackoff() *Backoff {
	return &Backoff{
		MaxIterations:    DefaultMaxIterations,
		MaxIterationTime: DefaultMaxIterationTime,
		MaxTotalTime:     DefaultMaxTotalTime,
		MinIterationTime: DefaultMinIterationTime,
		Multiplier:       DefaultMultipler,
	}
}

// Retry executes the function and waits for it to end successul or for the
// given limites to be reached. The returned error uses Go1.13 wrapping of
// errors and can be unwrapped into the error of the function itself.
func (b Backoff) Retry(f Retryable) error {
	var (
		iterations uint64
		sleepTime  = b.MinIterationTime
		start      = time.Now()
	)

	for {
		err := f()

		if err == nil {
			return nil
		}

		iterations++
		if b.MaxIterations > 0 && iterations == b.MaxIterations {
			return fmt.Errorf("Maximum iterations reached: %w", err)
		}

		if b.MaxTotalTime > 0 && time.Since(start) >= b.MaxTotalTime {
			return fmt.Errorf("Maximum execution time reached: %w", err)
		}

		time.Sleep(sleepTime)
		sleepTime = b.nextIterationSleep(sleepTime)
	}
}

// WithMaxIterations is a wrapper around setting the MaxIterations
// and then returning the Backoff object to use in chained creation
func (b *Backoff) WithMaxIterations(v uint64) *Backoff {
	b.MaxIterations = v
	return b
}

// WithMaxIterationTime is a wrapper around setting the MaxIterationTime
// and then returning the Backoff object to use in chained creation
func (b *Backoff) WithMaxIterationTime(v time.Duration) *Backoff {
	b.MaxIterationTime = v
	return b
}

// WithMaxTotalTime is a wrapper around setting the MaxTotalTime
// and then returning the Backoff object to use in chained creation
func (b *Backoff) WithMaxTotalTime(v time.Duration) *Backoff {
	b.MaxTotalTime = v
	return b
}

// WithMinIterationTime is a wrapper around setting the MinIterationTime
// and then returning the Backoff object to use in chained creation
func (b *Backoff) WithMinIterationTime(v time.Duration) *Backoff {
	b.MinIterationTime = v
	return b
}

// WithMultiplier is a wrapper around setting the Multiplier
// and then returning the Backoff object to use in chained creation
func (b *Backoff) WithMultiplier(v float64) *Backoff {
	b.Multiplier = v
	return b
}

func (b Backoff) nextIterationSleep(currentSleep time.Duration) time.Duration {
	next := time.Duration(float64(currentSleep) * b.Multiplier)
	if next > b.MaxIterationTime {
		next = b.MaxIterationTime
	}
	return next
}

// Retryable is a function which takes no parameters and yields an error
// when it should be retried and nil when it was successful
type Retryable func() error

// Retry is a convenience wrapper to execute the retry with default values
// (see exported constants)
func Retry(f Retryable) error { return NewBackoff().Retry(f) }
