package file

import (
	"crypto/sha256"
	"sync"
	"time"

	"github.com/pkg/errors"
)

type (
	// Watcher creates a background routine and emits events when the
	// watched file changes on its C channel. If an error occurs the
	// loop is stopped and the error is exposed on the Err property.
	Watcher struct {
		C             <-chan WatcherEvent
		CheckInterval time.Duration
		Err           error
		FilePath      string

		c          chan WatcherEvent
		checks     []WatcherCheck
		lock       sync.RWMutex
		stateCache map[string]any
	}

	// WatcherCheck is an interface to implement own checks
	WatcherCheck func(*Watcher) (WatcherEvent, error)

	// WatcherEvent is the detected change to be signeld through the
	// channel within the Watcher
	WatcherEvent uint
)

const (
	WatcherEventInvalid WatcherEvent = iota
	WatcherEventNoChange
	WatcherEventFileAppeared
	WatcherEventFileModified
	WatcherEventFileVanished
)

// NewCryptographicWatcher is a wrapper around NewWatcher to configure
// the Watcher with presence and sha256 hash checks.
func NewCryptographicWatcher(filePath string, interval time.Duration) (*Watcher, error) {
	return NewWatcher(filePath, interval, WatcherCheckPresence, WatcherCheckHash(sha256.New))
}

// NewSimpleWatcher is a wrapper around NewWatcher to configure the
// Watcher with presence, size and mtime checks.
func NewSimpleWatcher(filePath string, interval time.Duration) (*Watcher, error) {
	return NewWatcher(filePath, interval, WatcherCheckPresence, WatcherCheckSize, WatcherCheckMtime)
}

// NewWatcher creates a new Watcher configured with the given filePath,
// interval and checks given. The checks are executed once during
// initialization and will not cause an event to be sent. The created
// Watcher will automatically start its periodic check and the C
// channel should immediately be watched for changes. If the channel
// is not listened on the check loop will be paused until events are
// retrieved. If during the initial checks an error is detected the
// loop is NOT started and the watcher needs to be initialized again.
func NewWatcher(filePath string, interval time.Duration, checks ...WatcherCheck) (*Watcher, error) {
	w, err := newWatcher(filePath, interval, checks...)

	if err == nil {
		go w.loop()
	}

	return w, err
}

func newWatcher(filePath string, interval time.Duration, checks ...WatcherCheck) (*Watcher, error) {
	notify := make(chan WatcherEvent, 1)

	w := &Watcher{
		C:             notify,
		CheckInterval: interval,
		FilePath:      filePath,

		c:          notify,
		checks:     checks,
		stateCache: make(map[string]any),
	}
	// Initially run checks once
	_, err := w.runStateChecks()

	return w, errors.Wrap(err, "executing initial checks")
}

// GetState is a helper to retrieve state from the internal store for
// usage in checks to have their state retained.
func (w *Watcher) GetState(key string) any {
	w.lock.RLock()
	defer w.lock.RUnlock()

	return w.stateCache[key]
}

// SetState is a helper to set state into the internal store for
// usage in checks to have their state retained.
func (w *Watcher) SetState(key string, value any) {
	w.lock.Lock()
	defer w.lock.Unlock()

	w.stateCache[key] = value
}

func (w *Watcher) loop() {
	for {
		evt, err := w.runStateChecks()
		if err != nil {
			w.Err = err
			break
		}

		if evt != WatcherEventNoChange && evt != WatcherEventInvalid {
			// On "no change" and "invalid" events sending the new event is skipped
			w.c <- evt
		}

		time.Sleep(w.CheckInterval)
	}
}

func (w *Watcher) runStateChecks() (WatcherEvent, error) {
	for _, c := range w.checks {
		evt, err := c(w)
		if err != nil {
			return WatcherEventInvalid, errors.Wrap(err, "checking file state")
		}

		if evt == WatcherEventNoChange {
			// Watcher noticed no change, ask the next one. If one notices
			// a change we will return that one.
			continue
		}

		return evt, nil
	}

	return WatcherEventNoChange, nil
}
