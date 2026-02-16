package which_test

import (
	"testing"

	. "github.com/Luzifer/go_helpers/which"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindInDirectory(t *testing.T) {
	found, err := FindInDirectory("bash", "/bin")
	require.NoError(t, err)
	assert.True(t, found)
}

func TestFindInPath(t *testing.T) {
	// Searching bash on the system
	result, err := FindInPath("bash")
	assert.NoError(t, err)
	assert.Greater(t, len(result), 0)

	// Searching a non existent file
	_, err = FindInPath("dfqoiwurgtqi3uegrds")
	assert.ErrorIs(t, err, ErrBinaryNotFound)

	// Searching an empty file
	_, err = FindInPath("")
	assert.ErrorIs(t, err, ErrNoSearchSpecified)
}
