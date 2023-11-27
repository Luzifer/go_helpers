package backoff

import "fmt"

type (
	// ErrCannotRetry wraps the original error and signals the backoff
	// should be stopped now as a retry i.e. would be harmful or would
	// make no sense
	ErrCannotRetry struct{ inner error }
)

// NewErrCannotRetry wraps the given error into an ErrCannotRetry and
// should be used to break from a Retry() function when the retry
// should stop immediately
func NewErrCannotRetry(err error) error {
	return ErrCannotRetry{err}
}

func (e ErrCannotRetry) Error() string {
	return fmt.Sprintf("retry cancelled by error: %s", e.inner.Error())
}

func (e ErrCannotRetry) Unwrap() error {
	return e.inner
}
