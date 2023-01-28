package file

import (
	"crypto/sha256"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWatcherCheckHash(t *testing.T) {
	testDir, err := os.MkdirTemp("", "")
	require.NoError(t, err, "creating test-tempdir")
	t.Cleanup(func() {
		if err := os.RemoveAll(testDir); err != nil {
			t.Logf("failed to clean tempdir %q: %s", testDir, err)
		}
	})

	testFile := path.Join(testDir, "test.txt")

	w, err := newWatcher(testFile, time.Second, WatcherCheckHash(sha256.New))
	require.NoError(t, err, "initial check should not error on non existing file")

	evt, err := w.runStateChecks(false)
	require.NoError(t, err, "check should not error on non existing file")
	assert.Equal(t, WatcherEventInvalid, evt, "expect invalid as file is still missing")

	err = os.WriteFile(testFile, []byte("test"), 0o644)
	require.NoError(t, err, "creating test file")

	evt, err = w.runStateChecks(false)
	require.NoError(t, err, "check should not error on existing file")
	assert.Equal(t, WatcherEventFileModified, evt, "expect change as file now exists and has hash change")

	evt, err = w.runStateChecks(false)
	require.NoError(t, err, "check should not error on existing file")
	assert.Equal(t, WatcherEventNoChange, evt, "expect no change as the file has the same hash")

	err = os.WriteFile(testFile, []byte("hello world"), 0o644)
	require.NoError(t, err, "updating test file")

	evt, err = w.runStateChecks(false)
	require.NoError(t, err, "check should not error on existing file")
	assert.Equal(t, WatcherEventFileModified, evt, "expect change as file was modified")

	err = os.Remove(testFile)
	require.NoError(t, err, "deleting test file")

	evt, err = w.runStateChecks(false)
	require.NoError(t, err, "check should not error on non existing file")
	assert.Equal(t, WatcherEventInvalid, evt, "expect check to be invalid as file is no longer there")

	err = os.WriteFile(testFile, []byte("hello world"), 0o644)
	require.NoError(t, err, "updating test file")

	evt, err = w.runStateChecks(false)
	require.NoError(t, err, "check should not error on existing file")
	assert.Equal(t, WatcherEventNoChange, evt, "expect change as file has same hash")
}

func TestWatcherCheckMtime(t *testing.T) {
	testDir, err := os.MkdirTemp("", "")
	require.NoError(t, err, "creating test-tempdir")
	t.Cleanup(func() {
		if err := os.RemoveAll(testDir); err != nil {
			t.Logf("failed to clean tempdir %q: %s", testDir, err)
		}
	})

	testFile := path.Join(testDir, "test.txt")

	w, err := newWatcher(testFile, time.Second, WatcherCheckMtime)
	require.NoError(t, err, "initial check should not error on non existing file")

	evt, err := w.runStateChecks(false)
	require.NoError(t, err, "check should not error on non existing file")
	assert.Equal(t, WatcherEventInvalid, evt, "expect invalid as file is still missing")

	err = os.WriteFile(testFile, []byte("test"), 0o644)
	require.NoError(t, err, "creating test file")

	evt, err = w.runStateChecks(false)
	require.NoError(t, err, "check should not error on existing file")
	assert.Equal(t, WatcherEventFileModified, evt, "expect change as file now exists and has mtime change")

	evt, err = w.runStateChecks(false)
	require.NoError(t, err, "check should not error on existing file")
	assert.Equal(t, WatcherEventNoChange, evt, "expect no change as the file has the same mtime")

	time.Sleep(time.Second) // Unix mtime is second-based, wait a moment

	err = os.WriteFile(testFile, []byte("hello world"), 0o644)
	require.NoError(t, err, "updating test file")

	evt, err = w.runStateChecks(false)
	require.NoError(t, err, "check should not error on existing file")
	assert.Equal(t, WatcherEventFileModified, evt, "expect change as file was modified")

	err = os.Remove(testFile)
	require.NoError(t, err, "deleting test file")

	evt, err = w.runStateChecks(false)
	require.NoError(t, err, "check should not error on non existing file")
	assert.Equal(t, WatcherEventInvalid, evt, "expect check to be invalid as file is no longer there")

	time.Sleep(time.Second) // Unix mtime is second-based, wait a moment

	err = os.WriteFile(testFile, []byte("hello world"), 0o644)
	require.NoError(t, err, "updating test file")

	evt, err = w.runStateChecks(false)
	require.NoError(t, err, "check should not error on existing file")
	assert.Equal(t, WatcherEventFileModified, evt, "expect change as file is newer")
}

func TestWatcherCheckPresence(t *testing.T) {
	testDir, err := os.MkdirTemp("", "")
	require.NoError(t, err, "creating test-tempdir")
	t.Cleanup(func() {
		if err := os.RemoveAll(testDir); err != nil {
			t.Logf("failed to clean tempdir %q: %s", testDir, err)
		}
	})

	testFile := path.Join(testDir, "test.txt")

	w, err := newWatcher(testFile, time.Second, WatcherCheckPresence)
	require.NoError(t, err, "initial check should not error on non existing file")

	evt, err := w.runStateChecks(false)
	require.NoError(t, err, "check should not error on non existing file")
	assert.Equal(t, WatcherEventNoChange, evt, "expect no change as file is still missing")

	err = os.WriteFile(testFile, []byte("test"), 0o644)
	require.NoError(t, err, "creating test file")

	evt, err = w.runStateChecks(false)
	require.NoError(t, err, "check should not error on existing file")
	assert.Equal(t, WatcherEventFileAppeared, evt, "expect check to state file is now there")

	evt, err = w.runStateChecks(false)
	require.NoError(t, err, "check should not error on existing file")
	assert.Equal(t, WatcherEventNoChange, evt, "expect check to state nothing changed")

	err = os.Remove(testFile)
	require.NoError(t, err, "deleting test file")

	evt, err = w.runStateChecks(false)
	require.NoError(t, err, "check should not error on non existing file")
	assert.Equal(t, WatcherEventFileVanished, evt, "expect check to state file vanished again")
}

func TestWatcherCheckSize(t *testing.T) {
	testDir, err := os.MkdirTemp("", "")
	require.NoError(t, err, "creating test-tempdir")
	t.Cleanup(func() {
		if err := os.RemoveAll(testDir); err != nil {
			t.Logf("failed to clean tempdir %q: %s", testDir, err)
		}
	})

	testFile := path.Join(testDir, "test.txt")

	w, err := newWatcher(testFile, time.Second, WatcherCheckSize)
	require.NoError(t, err, "initial check should not error on non existing file")

	evt, err := w.runStateChecks(false)
	require.NoError(t, err, "check should not error on non existing file")
	assert.Equal(t, WatcherEventInvalid, evt, "expect invalid as file is still missing")

	err = os.WriteFile(testFile, []byte("test"), 0o644)
	require.NoError(t, err, "creating test file")

	evt, err = w.runStateChecks(false)
	require.NoError(t, err, "check should not error on existing file")
	assert.Equal(t, WatcherEventFileModified, evt, "expect change as file has now 4 instead of 0 bytes")

	err = os.WriteFile(testFile, []byte("tset"), 0o644)
	require.NoError(t, err, "updating test file")

	evt, err = w.runStateChecks(false)
	require.NoError(t, err, "check should not error on existing file")
	assert.Equal(t, WatcherEventNoChange, evt, "expect no change as the file has the same size")

	err = os.WriteFile(testFile, []byte("hello world"), 0o644)
	require.NoError(t, err, "updating test file")

	evt, err = w.runStateChecks(false)
	require.NoError(t, err, "check should not error on existing file")
	assert.Equal(t, WatcherEventFileModified, evt, "expect change as we went from 4 to 11 chars")

	err = os.Remove(testFile)
	require.NoError(t, err, "deleting test file")

	evt, err = w.runStateChecks(false)
	require.NoError(t, err, "check should not error on non existing file")
	assert.Equal(t, WatcherEventInvalid, evt, "expect check to be invalid as file is no longer there")

	err = os.WriteFile(testFile, []byte("hello world"), 0o644)
	require.NoError(t, err, "updating test file")

	evt, err = w.runStateChecks(false)
	require.NoError(t, err, "check should not error on existing file")
	assert.Equal(t, WatcherEventNoChange, evt, "expect no change as we restored the file with same content")
}
