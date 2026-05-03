package fieldcollection

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInt64(t *testing.T) {
	fc := FieldCollectionFromData(map[string]any{
		"int":           int(12),
		"int16":         int16(12),
		"int32":         int32(12),
		"int64":         int64(12),
		"bool":          true,
		"invalidString": "I'm a string!",
		"validString":   "12",
	})

	_, err := fc.Int64("_")
	require.ErrorIs(t, err, ErrValueNotSet)

	_, err = fc.Int64("bool")
	require.ErrorIs(t, err, ErrValueMismatch)

	_, err = fc.Int64("invalidString")
	require.Error(t, err)

	v, err := fc.Int64("int")
	require.NoError(t, err)
	assert.Equal(t, int64(12), v)

	v, err = fc.Int64("int16")
	require.NoError(t, err)
	assert.Equal(t, int64(12), v)

	v, err = fc.Int64("int32")
	require.NoError(t, err)
	assert.Equal(t, int64(12), v)

	v, err = fc.Int64("int64")
	require.NoError(t, err)
	assert.Equal(t, int64(12), v)

	v, err = fc.Int64("validString")
	require.NoError(t, err)
	assert.Equal(t, int64(12), v)

	assert.True(t, fc.CanInt64("int"))
	assert.False(t, fc.CanInt64("bool"))

	assert.NotPanics(t, func() { fc.MustInt64("int32", nil) })
	assert.Panics(t, func() { fc.MustInt64("bool", nil) })

	assert.Equal(t, int64(5), fc.MustInt64("_", func(v int64) *int64 { return &v }(5)))
}
