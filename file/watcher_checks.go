package file

import (
	"crypto/sha512"
	"fmt"
	"hash"
	"io"
	"io/fs"
	"os"

	"github.com/pkg/errors"
)

const (
	keyWatcherCheckHash     = "WatcherCheckHash"
	keyWatcherCheckMtime    = "WatcherCheckMtime"
	keyWatcherCheckPresence = "WatcherCheckPresence"
	keyWatcherCheckSize     = "WatcherCheckSize"
)

// WatcherCheckHash returns a WatcherCheck configured with the given
// hash method (i.e. provide md5.New, sha1.New, ...). If the file is
// not present at the time of the check the check is skipped and will
// NOT cause an error.
func WatcherCheckHash(hcf func() hash.Hash) WatcherCheck {
	return func(w *Watcher) (WatcherEvent, error) {
		var lastHash string
		if v, ok := w.GetState(keyWatcherCheckHash).(string); ok {
			lastHash = v
		}

		if _, err := os.Stat(w.FilePath); errors.Is(err, fs.ErrNotExist) {
			return WatcherEventInvalid, nil
		}

		f, err := os.Open(w.FilePath)
		if err != nil {
			return WatcherEventInvalid, errors.Wrap(err, "opening file")
		}
		defer f.Close()

		h := hcf()
		if _, err = io.Copy(h, f); err != nil {
			return WatcherEventInvalid, errors.Wrap(err, "reading file")
		}

		currentHash := fmt.Sprintf("%x", h.Sum(nil))
		if lastHash == currentHash {
			return WatcherEventNoChange, nil
		}

		w.SetState(keyWatcherCheckHash, currentHash)
		return WatcherEventFileModified, nil
	}
}

var _ WatcherCheck = WatcherCheckHash(sha512.New)

// WatcherCheckMtime checks whether the mtime attribute of the file
// has changed. If the file is not present at the time of the check
// the check is skipped and will NOT cause an error.
func WatcherCheckMtime(w *Watcher) (WatcherEvent, error) {
	var lastChange int64
	if v, ok := w.GetState(keyWatcherCheckMtime).(int64); ok {
		lastChange = v
	}

	s, err := os.Stat(w.FilePath)
	switch {
	case err == nil:
		// handle size change
	case errors.Is(err, fs.ErrNotExist):
		return WatcherEventInvalid, nil
	default:
		return WatcherEventInvalid, errors.Wrap(err, "getting file stat")
	}

	if s.ModTime().UnixNano() == lastChange {
		return WatcherEventNoChange, nil
	}

	w.SetState(keyWatcherCheckMtime, s.ModTime().UnixNano())
	return WatcherEventFileModified, nil
}

var _ WatcherCheck = WatcherCheckMtime

// WatcherCheckPresence simply checks whether the file is present and
// allows to emit WatcherEventFileAppeared / WatcherEventFileVanished
// events when the file existence state changes.
func WatcherCheckPresence(w *Watcher) (WatcherEvent, error) {
	var wasPresent bool
	if v, ok := w.GetState(keyWatcherCheckPresence).(bool); ok {
		wasPresent = v
	}

	_, err := os.Stat(w.FilePath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		// Some weird error occurred
		return WatcherEventInvalid, errors.Wrap(err, "getting file stat")
	}

	isPresent := err == nil
	w.SetState(keyWatcherCheckPresence, isPresent)

	switch {
	case !wasPresent && isPresent:
		return WatcherEventFileAppeared, nil
	case wasPresent && !isPresent:
		return WatcherEventFileVanished, nil
	default:
		return WatcherEventNoChange, nil
	}
}

var _ WatcherCheck = WatcherCheckPresence

// WatcherCheckSize checks whether the size of the file has changed.
// If the file is not present at the time of the check the check is
// skipped and will NOT cause an error.
func WatcherCheckSize(w *Watcher) (WatcherEvent, error) {
	var knownSize int64
	if v, ok := w.GetState(keyWatcherCheckSize).(int64); ok {
		knownSize = v
	}

	s, err := os.Stat(w.FilePath)
	switch {
	case err == nil:
		// handle size change
	case errors.Is(err, fs.ErrNotExist):
		return WatcherEventInvalid, nil
	default:
		return WatcherEventInvalid, errors.Wrap(err, "getting file stat")
	}

	if s.Size() == knownSize {
		return WatcherEventNoChange, nil
	}

	w.SetState(keyWatcherCheckSize, s.Size())
	return WatcherEventFileModified, nil
}

var _ WatcherCheck = WatcherCheckSize
