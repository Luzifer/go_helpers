package sentrus

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	callerSkip     = 3
	maxStackFrames = 64
)

type (
	event struct {
		Environment string         `json:"environment,omitempty"`
		EventID     string         `json:"event_id"`
		Exception   eventException `json:"exception"`
		Extra       map[string]any `json:"extra,omitempty"`
		Level       string         `json:"level"`
		Message     string         `json:"message"`
		Platform    string         `json:"platform"`
		Release     string         `json:"release,omitempty"`
		Timestamp   time.Time      `json:"timestamp"`
	}

	eventException struct {
		Values []exceptionValue `json:"values"`
	}

	exceptionValue struct {
		Stacktrace stacktrace `json:"stacktrace"`
		Type       string     `json:"type"`
		Value      string     `json:"value"`
	}

	frame struct {
		Filename string `json:"filename"`
		Function string `json:"function"`
		InApp    bool   `json:"in_app"`
		LineNo   int    `json:"lineno"`
	}

	stacktrace struct {
		Frames []frame `json:"frames"`
	}
)

func buildEvent(entry *logrus.Entry, cfg config) (event, error) {
	eventID, err := newEventID()
	if err != nil {
		return event{}, err
	}

	timestamp := entry.Time
	if timestamp.IsZero() {
		timestamp = time.Now().UTC()
	}

	exceptionType, exceptionMessage := exceptionFromEntry(entry)
	return event{
		Environment: cfg.environment,
		EventID:     eventID,
		Exception: eventException{Values: []exceptionValue{{
			Stacktrace: stacktrace{Frames: captureFrames()},
			Type:       exceptionType,
			Value:      exceptionMessage,
		}}},
		Extra:     extrasFromEntry(entry),
		Level:     sentryLevel(entry.Level),
		Message:   entry.Message,
		Platform:  "go",
		Release:   cfg.release,
		Timestamp: timestamp,
	}, nil
}

func captureFrames() []frame {
	callers := make([]uintptr, maxStackFrames)
	count := runtime.Callers(callerSkip, callers)
	iterator := runtime.CallersFrames(callers[:count])
	frames := make([]frame, 0, count)

	for {
		runtimeFrame, more := iterator.Next()
		frames = append(frames, frame{
			Filename: runtimeFrame.File,
			Function: runtimeFrame.Function,
			InApp:    isInApp(runtimeFrame.Function),
			LineNo:   runtimeFrame.Line,
		})
		if !more {
			break
		}
	}

	slices.Reverse(frames)
	return frames
}

func exceptionFromEntry(entry *logrus.Entry) (exceptionType, exceptionMessage string) {
	err, ok := entry.Data[logrus.ErrorKey].(error)
	if !ok {
		return "logrus.Error", entry.Message
	}

	return reflect.TypeOf(err).String(), err.Error()
}

func extrasFromEntry(entry *logrus.Entry) map[string]any {
	if len(entry.Data) == 0 {
		return nil
	}

	extra := make(map[string]any, len(entry.Data))
	for key, value := range entry.Data {
		if key != logrus.ErrorKey {
			extra[key] = value
		}
	}

	return extra
}

func isInApp(function string) bool {
	return !strings.HasPrefix(function, "github.com/Luzifer/go_helpers/sentrus.") &&
		!strings.HasPrefix(function, "github.com/sirupsen/logrus.") &&
		!strings.HasPrefix(function, "runtime.")
}

func newEventID() (string, error) {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generating event ID: %w", err)
	}

	return hex.EncodeToString(raw), nil
}

func sentryLevel(level logrus.Level) string {
	switch level {
	case logrus.PanicLevel, logrus.FatalLevel:
		return "fatal"
	case logrus.WarnLevel:
		return "warning"
	case logrus.InfoLevel:
		return "info"
	case logrus.DebugLevel, logrus.TraceLevel:
		return "debug"
	default:
		return "error"
	}
}
