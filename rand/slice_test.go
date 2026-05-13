package rand

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testStruct struct {
	Values []string
}

func TestEntryFromSliceEmptySlice(t *testing.T) {
	value, err := EntryFromSlice(([]string)(nil))

	require.Error(t, err)
	assert.Empty(t, value)
	assert.ErrorContains(t, err, "cannot choose from zero-length slice")
}

func TestEntryFromSliceTypes(t *testing.T) {
	t.Run("strings", func(t *testing.T) {
		values := []string{"alpha", "bravo", "charlie"}

		value, err := EntryFromSlice(values)

		require.NoError(t, err)
		assert.Contains(t, values, value)
	})

	t.Run("maps", func(t *testing.T) {
		values := []map[string]string{
			{"value": "alpha"},
			{"value": "bravo"},
		}

		value, err := EntryFromSlice(values)

		require.NoError(t, err)
		assert.Contains(t, []string{"alpha", "bravo"}, value["value"])
	})

	t.Run("slices", func(t *testing.T) {
		values := [][]string{
			{"alpha"},
			{"bravo"},
		}

		value, err := EntryFromSlice(values)

		require.NoError(t, err)
		assert.Contains(t, []string{"alpha", "bravo"}, value[0])
	})

	t.Run("structs with slices", func(t *testing.T) {
		values := []testStruct{
			{Values: []string{"alpha"}},
			{Values: []string{"bravo"}},
		}

		value, err := EntryFromSlice(values)

		require.NoError(t, err)
		assert.Contains(t, []string{"alpha", "bravo"}, value.Values[0])
	})
}
