package sentrus

import (
	"context"
	"fmt"
	"slices"
	"sync/atomic"

	"github.com/sirupsen/logrus"
)

type (
	// Hook submits configured Logrus levels as Sentry-compatible issue events.
	Hook struct {
		config    config
		dropped   atomic.Uint64
		failed    atomic.Uint64
		queue     chan queueItem
		transport *transport
	}

	queueItem struct {
		flushed chan struct{}
		payload []byte
	}
)

var _ logrus.Hook = (*Hook)(nil)

// NewHook creates and starts a Sentry-compatible Logrus hook.
func NewHook(options ...Option) (*Hook, error) {
	cfg := defaultConfig()
	for index, option := range options {
		if option == nil {
			return nil, fmt.Errorf("applying option %d: option is nil", index)
		}
		if err := option(&cfg); err != nil {
			return nil, fmt.Errorf("applying option %d: %w", index, err)
		}
	}

	dsn, err := parseDSN(cfg.dsn)
	if err != nil {
		return nil, err
	}

	hook := &Hook{
		config:    cfg,
		queue:     make(chan queueItem, cfg.queueLength),
		transport: newTransport(dsn, cfg.timeout),
	}
	go hook.run()
	return hook, nil
}

// Dropped returns the number of events rejected because the queue was full.
func (h *Hook) Dropped() uint64 {
	return h.dropped.Load()
}

// Failed returns the number of events which could not be encoded or delivered.
func (h *Hook) Failed() uint64 {
	return h.failed.Load()
}

// Fire queues a Logrus entry for delivery without blocking the caller.
func (h *Hook) Fire(entry *logrus.Entry) error {
	if !h.isEnabled(entry.Level) {
		return nil
	}

	event, err := buildEvent(entry, h.config)
	if err != nil {
		h.failed.Add(1)
		return fmt.Errorf("building event: %w", err)
	}

	payload, err := buildEnvelope(event)
	if err != nil {
		h.failed.Add(1)
		return err
	}

	select {
	case h.queue <- queueItem{payload: payload}:
	default:
		h.dropped.Add(1)
	}

	return nil
}

// Flush waits until all events queued before the call have been processed.
func (h *Hook) Flush(ctx context.Context) error {
	flushed := make(chan struct{})
	select {
	case h.queue <- queueItem{flushed: flushed}:
	case <-ctx.Done():
		return fmt.Errorf("queueing flush: %w", ctx.Err())
	}

	select {
	case <-flushed:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("flushing events: %w", ctx.Err())
	}
}

// Levels returns the exact Logrus levels submitted by the hook.
func (h *Hook) Levels() []logrus.Level {
	return append([]logrus.Level(nil), h.config.levels...)
}

func (h *Hook) isEnabled(level logrus.Level) bool {
	return slices.Contains(h.config.levels, level)
}

func (h *Hook) run() {
	for item := range h.queue {
		if item.flushed != nil {
			close(item.flushed)
			continue
		}
		if err := h.transport.send(item.payload); err != nil {
			h.failed.Add(1)
		}
	}
}
