package sentrus

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHookDeliversEnvelope(t *testing.T) {
	received := make(chan []byte, 1)
	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/prefix/api/17/envelope/", req.URL.Path)
		assert.Equal(t, "application/x-sentry-envelope", req.Header.Get("Content-Type"))
		assert.Equal(t, "Sentry sentry_version=7, sentry_key=key", req.Header.Get("X-Sentry-Auth"))
		body, err := io.ReadAll(req.Body)
		assert.NoError(t, err)
		received <- body
		res.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	hook, err := NewHook(
		WithDSN(testDSN(server.URL, "prefix/17")),
		WithEnvironment("production"),
		WithLevels(logrus.InfoLevel),
		WithQueueLength(2),
		WithRelease("app@version"),
		WithTimeout(time.Second),
	)
	require.NoError(t, err)
	assert.Equal(t, []logrus.Level{logrus.InfoLevel}, hook.Levels())

	require.NoError(t, hook.Fire(&logrus.Entry{Level: logrus.InfoLevel, Message: "hello"}))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	require.NoError(t, hook.Flush(ctx))
	assert.NotEmpty(t, <-received)
	assert.Zero(t, hook.Dropped())
	assert.Zero(t, hook.Failed())
}

func TestHookConfigurationErrors(t *testing.T) {
	_, err := NewHook()
	require.Error(t, err)

	_, err = NewHook(nil)
	require.Error(t, err)

	_, err = NewHook(WithDSN("invalid"))
	require.Error(t, err)

	_, err = NewHook(WithDSN("https://key@example.com/1"), WithQueueLength(0))
	require.Error(t, err)
}

func TestHookCountsEncodingFailure(t *testing.T) {
	hook, err := NewHook(WithDSN("https://key@example.com/1"))
	require.NoError(t, err)
	err = hook.Fire(&logrus.Entry{
		Data:  logrus.Fields{"unsupported": make(chan struct{})},
		Level: logrus.ErrorLevel,
	})
	require.Error(t, err)
	assert.EqualValues(t, 1, hook.Failed())
}

func TestHookDropsWhenQueueIsFull(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, _ *http.Request) {
		select {
		case <-started:
		default:
			close(started)
		}
		<-release
		res.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	hook, err := NewHook(WithDSN(testDSN(server.URL, "1")), WithQueueLength(1), WithTimeout(time.Second))
	require.NoError(t, err)
	require.NoError(t, hook.Fire(&logrus.Entry{Level: logrus.ErrorLevel, Message: "first"}))
	<-started
	require.NoError(t, hook.Fire(&logrus.Entry{Level: logrus.ErrorLevel, Message: "second"}))
	require.NoError(t, hook.Fire(&logrus.Entry{Level: logrus.ErrorLevel, Message: "third"}))
	assert.EqualValues(t, 1, hook.Dropped())

	close(release)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	require.NoError(t, hook.Flush(ctx))
}

func TestHookFireIgnoresUnconfiguredLevel(t *testing.T) {
	hook, err := NewHook(WithDSN("https://key@example.com/1"), WithLevels(logrus.ErrorLevel))
	require.NoError(t, err)
	require.NoError(t, hook.Fire(&logrus.Entry{Level: logrus.InfoLevel, Message: "ignored"}))
	assert.Empty(t, hook.queue)
}

func TestHookFlushTimeout(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, _ *http.Request) {
		close(started)
		<-release
		res.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	hook, err := NewHook(WithDSN(testDSN(server.URL, "1")), WithTimeout(time.Second))
	require.NoError(t, err)
	require.NoError(t, hook.Fire(&logrus.Entry{Level: logrus.ErrorLevel}))
	<-started

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	require.ErrorIs(t, hook.Flush(ctx), context.DeadlineExceeded)
	close(release)
}

func TestTransportRetries(t *testing.T) {
	for _, test := range []struct {
		name     string
		status   int
		attempts int
	}{
		{name: "bad request", status: http.StatusBadRequest, attempts: 1},
		{name: "rate limited", status: http.StatusTooManyRequests, attempts: maxDeliveryAttempts},
		{name: "server error", status: http.StatusInternalServerError, attempts: maxDeliveryAttempts},
	} {
		t.Run(test.name, func(t *testing.T) {
			var requests atomic.Int64
			server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, _ *http.Request) {
				requests.Add(1)
				res.WriteHeader(test.status)
			}))
			defer server.Close()

			hook, err := NewHook(WithDSN(testDSN(server.URL, "1")), WithTimeout(time.Second))
			require.NoError(t, err)
			require.NoError(t, hook.Fire(&logrus.Entry{Level: logrus.ErrorLevel}))
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			require.NoError(t, hook.Flush(ctx))
			assert.EqualValues(t, test.attempts, requests.Load())
			assert.EqualValues(t, 1, hook.Failed())
		})
	}
}

func TestTransportRetriesThenSucceeds(t *testing.T) {
	var requests atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, _ *http.Request) {
		if requests.Add(1) < maxDeliveryAttempts {
			res.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		res.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	hook, err := NewHook(WithDSN(testDSN(server.URL, "1")), WithTimeout(time.Second))
	require.NoError(t, err)
	require.NoError(t, hook.Fire(&logrus.Entry{Level: logrus.ErrorLevel}))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, hook.Flush(ctx))
	assert.EqualValues(t, maxDeliveryAttempts, requests.Load())
	assert.Zero(t, hook.Failed())
}

func TestDefaultLevelsAreIsolated(t *testing.T) {
	hook, err := NewHook(WithDSN("https://key@example.com/1"))
	require.NoError(t, err)
	levels := hook.Levels()
	levels[0] = logrus.InfoLevel
	assert.Equal(t, []logrus.Level{logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel}, hook.Levels())
}

func testDSN(serverURL, path string) string {
	return fmt.Sprintf("%s/%s", strings.Replace(serverURL, "://", "://key@", 1), path)
}
