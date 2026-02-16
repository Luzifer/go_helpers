package str_test

import (
	"testing"

	. "github.com/Luzifer/go_helpers/str"
	"github.com/stretchr/testify/assert"
)

func TestAppendIfMissing(t *testing.T) {
	sl := []string{
		"test1",
		"test2",
		"test3",
	}

	// should not append existing elements
	assert.Len(t, AppendIfMissing(sl, "test1"), 3)
	assert.Len(t, AppendIfMissing(sl, "test2"), 3)
	assert.Len(t, AppendIfMissing(sl, "test3"), 3)

	// should append not existing elements
	assert.Len(t, AppendIfMissing(sl, "test4"), 4)
	assert.Len(t, AppendIfMissing(sl, "test5"), 4)
	assert.Len(t, AppendIfMissing(sl, "test6"), 4)
}

func TestStringInSlice(t *testing.T) {
	sl := []string{
		"test1",
		"test2",
		"test3",
	}

	// should find elements of slice
	assert.True(t, StringInSlice("test1", sl))
	assert.True(t, StringInSlice("test2", sl))
	assert.True(t, StringInSlice("test3", sl))

	// should not find elements not in slice
	assert.False(t, StringInSlice("test4", sl))
	assert.False(t, StringInSlice("test5", sl))
	assert.False(t, StringInSlice("test6", sl))
}
