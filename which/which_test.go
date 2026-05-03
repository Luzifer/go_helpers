package which

import (
	"testing"

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
	require.NoError(t, err)
	assert.NotEmpty(t, result)

	// Searching a non existent file
	_, err = FindInPath("dfqoiwurgtqi3uegrds")
	require.ErrorIs(t, err, ErrBinaryNotFound)

	// Searching an empty file
	_, err = FindInPath("")
	require.ErrorIs(t, err, ErrNoSearchSpecified)
}
