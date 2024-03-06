// Package io contains helpers for I/O tasks
package io

import (
	"errors"
	"fmt"
	"io"
	"time"
)

// ThrottledReader implements a reader imposing a rate limit to the
// reading side to i.e. limit downloads, limit I/O on a filesystem, …
// The reads will burst and then wait until the rate "calmed" to the
// desired rate.
type ThrottledReader struct {
	startRead      time.Time
	totalReadBytes uint64
	readRateBpns   float64

	next io.Reader
}

// NewThrottledReader creates a reader with next as its underlying reader and
// rate as its throttle rate in Bytes / Second
func NewThrottledReader(next io.Reader, rate float64) *ThrottledReader {
	return &ThrottledReader{next: next, readRateBpns: rate / float64(time.Second)}
}

// Read implements the io.Reader interface
func (t *ThrottledReader) Read(p []byte) (n int, err error) {
	if t.startRead.IsZero() {
		t.startRead = time.Now()
	}

	// First read is for free
	n, err = t.next.Read(p)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return n, io.EOF
		}
		return n, fmt.Errorf("reading from next: %w", err)
	}

	// Count the data
	t.totalReadBytes += uint64(n)

	// Now lets see how long we need to wait
	var (
		currentRate  float64
		timePassedNS = int64(time.Since(t.startRead))
	)

	if timePassedNS > 0 {
		currentRate = float64(t.totalReadBytes) / float64(timePassedNS)
	}

	if currentRate > t.readRateBpns {
		timeToWait := int64(float64(t.totalReadBytes)/t.readRateBpns - float64(timePassedNS))
		time.Sleep(time.Duration(timeToWait))
	}

	// Waited long enough, rate is fine again, return
	return n, nil
}
