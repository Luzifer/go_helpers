package io

import (
	"crypto/rand"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestThrottledReader(t *testing.T) {
	var (
		testSize  = 10 * 1024 * 1024 // 10 Mi
		testLimit = 20 * 1024 * 1024 // 20 Mi/s

		// 20Mi/s on 10M = 500ms exec time
		expectedTimeMillisecs = float64(testSize) / float64(testLimit) * 1000
		tolerance             = 50 // Millisecs
	)

	lr := io.LimitReader(rand.Reader, int64(testSize))
	var tr io.Reader = NewThrottledReader(lr, float64(testLimit))

	start := time.Now()
	n, err := io.Copy(io.Discard, tr)
	require.NoError(t, err)

	assert.Equal(t, int64(testSize), n)
	assert.Greater(t, time.Since(start)/time.Millisecond, time.Duration(expectedTimeMillisecs-float64(tolerance)))
	assert.Less(t, time.Since(start)/time.Millisecond, time.Duration(expectedTimeMillisecs+float64(tolerance)))
}
