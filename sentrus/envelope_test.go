package sentrus

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildEnvelope(t *testing.T) {
	entry := &logrus.Entry{
		Data:    logrus.Fields{"answer": 42, logrus.ErrorKey: errors.New("broken")},
		Level:   logrus.ErrorLevel,
		Message: "request failed",
		Time:    time.Date(2026, time.October, 27, 12, 0, 0, 0, time.UTC),
	}
	builtEvent, err := buildEvent(entry, config{environment: "test", release: "app@1"})
	require.NoError(t, err)
	require.Len(t, builtEvent.EventID, 32)
	assert.Equal(t, "test", builtEvent.Environment)
	assert.Equal(t, "error", builtEvent.Level)
	assert.Equal(t, "request failed", builtEvent.Message)
	assert.Equal(t, "go", builtEvent.Platform)
	assert.Equal(t, "app@1", builtEvent.Release)
	assert.Equal(t, map[string]any{"answer": 42}, builtEvent.Extra)
	require.Len(t, builtEvent.Exception.Values, 1)
	assert.Equal(t, "*errors.errorString", builtEvent.Exception.Values[0].Type)
	assert.Equal(t, "broken", builtEvent.Exception.Values[0].Value)

	payload, err := buildEnvelope(builtEvent)
	require.NoError(t, err)
	require.True(t, strings.HasSuffix(string(payload), "\n"))
	parts := strings.SplitN(string(payload), "\n", 3)
	require.Len(t, parts, 3)

	var header envelopeHeader
	require.NoError(t, json.Unmarshal([]byte(parts[0]), &header))
	assert.Equal(t, builtEvent.EventID, header.EventID)

	var item itemHeader
	require.NoError(t, json.Unmarshal([]byte(parts[1]), &item))
	assert.Equal(t, "application/json", item.ContentType)
	assert.Equal(t, "event", item.Type)
	eventPayload := strings.TrimSuffix(parts[2], "\n")
	assert.Equal(t, len(eventPayload), item.Length)

	var encodedEvent event
	require.NoError(t, json.Unmarshal([]byte(eventPayload), &encodedEvent))
	assert.Equal(t, builtEvent.EventID, encodedEvent.EventID)
}

func TestBuildSyntheticException(t *testing.T) {
	event, err := buildEvent(&logrus.Entry{Level: logrus.WarnLevel, Message: "warning"}, config{})
	require.NoError(t, err)
	assert.Equal(t, "logrus.Error", event.Exception.Values[0].Type)
	assert.Equal(t, "warning", event.Exception.Values[0].Value)
	assert.NotZero(t, event.Timestamp)
	assert.Nil(t, event.Extra)
}

func TestCaptureFrames(t *testing.T) {
	var captured []frame
	handler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		captured = captureFrames()
	})
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil))

	require.NotEmpty(t, captured)
	var foundHandler bool
	for _, capturedFrame := range captured {
		assert.False(t, shouldSkipFrame(capturedFrame.Function))
		assert.True(t, capturedFrame.InApp)
		if capturedFrame.Function == "net/http.HandlerFunc.ServeHTTP" {
			foundHandler = true
			assert.NotEmpty(t, capturedFrame.Filename)
			assert.Positive(t, capturedFrame.LineNo)
		}
	}
	assert.True(t, foundHandler)
}

func TestBuildEnvelopeRejectsUnsupportedExtra(t *testing.T) {
	event, err := buildEvent(&logrus.Entry{
		Data:  logrus.Fields{"callback": func() {}},
		Level: logrus.ErrorLevel,
	}, config{})
	require.NoError(t, err)
	_, err = buildEnvelope(event)
	require.Error(t, err)
}

func TestFrameHelpers(t *testing.T) {
	for _, function := range []string{
		"github.com/Luzifer/go_helpers/sentrus.captureFrames",
		"github.com/sirupsen/logrus.(*Entry).log",
		"runtime.main",
		"testing.tRunner",
	} {
		assert.True(t, shouldSkipFrame(function), function)
	}

	for _, function := range []string{
		"example.com/app.run",
		"github.com/sirupsen/logruswrapper.Log",
		"net/http.(*conn).serve",
		"runtimehelper.Run",
	} {
		assert.False(t, shouldSkipFrame(function), function)
	}
}

func TestSentryLevels(t *testing.T) {
	expected := map[logrus.Level]string{
		logrus.PanicLevel: "fatal",
		logrus.FatalLevel: "fatal",
		logrus.ErrorLevel: "error",
		logrus.WarnLevel:  "warning",
		logrus.InfoLevel:  "info",
		logrus.DebugLevel: "debug",
		logrus.TraceLevel: "debug",
	}

	for level, sentryName := range expected {
		assert.Equal(t, sentryName, sentryLevel(level))
	}
	assert.Equal(t, "error", sentryLevel(logrus.Level(99)))
}
