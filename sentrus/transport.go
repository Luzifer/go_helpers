package sentrus

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Luzifer/go_helpers/backoff"
)

const (
	maxDeliveryAttempts = 5
	maxResponseBodyRead = 4096
)

type (
	statusError struct {
		code int
	}

	transport struct {
		authHeader string
		client     *http.Client
		endpoint   string
	}
)

func (e statusError) Error() string {
	return fmt.Sprintf("unexpected HTTP status: %d", e.code)
}

func newTransport(dsn parsedDSN, timeout time.Duration) *transport {
	return &transport{
		authHeader: dsn.authHeader,
		client: &http.Client{
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Timeout: timeout,
		},
		endpoint: dsn.endpoint,
	}
}

func (t *transport) send(payload []byte) error {
	retry := backoff.NewBackoff().WithMaxIterations(maxDeliveryAttempts)
	if err := retry.Retry(func() error {
		err := t.sendOnce(payload)
		if err == nil || isRetryable(err) {
			return err
		}

		return backoff.NewErrCannotRetry(err)
	}); err != nil {
		return fmt.Errorf("delivering envelope: %w", err)
	}

	return nil
}

func (t *transport) sendOnce(payload []byte) error {
	request, err := http.NewRequestWithContext(context.Background(), http.MethodPost, t.endpoint, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	request.Header.Set("Content-Type", "application/x-sentry-envelope")
	request.Header.Set("X-Sentry-Auth", t.authHeader)

	response, err := t.client.Do(request)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	_, readErr := io.Copy(io.Discard, io.LimitReader(response.Body, maxResponseBodyRead))
	closeErr := response.Body.Close()
	if readErr != nil {
		return fmt.Errorf("reading response: %w", readErr)
	}
	if closeErr != nil {
		return fmt.Errorf("closing response: %w", closeErr)
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return statusError{code: response.StatusCode}
	}

	return nil
}

func isRetryable(err error) bool {
	var statusErr statusError
	if !errors.As(err, &statusErr) {
		return true
	}

	return statusErr.code == http.StatusTooManyRequests || statusErr.code >= http.StatusInternalServerError
}
