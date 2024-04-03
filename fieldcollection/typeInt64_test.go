package fieldcollection

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
	assert.ErrorIs(t, err, ErrValueNotSet)

	_, err = fc.Int64("bool")
	assert.ErrorIs(t, err, ErrValueMismatch)

	_, err = fc.Int64("invalidString")
	assert.Error(t, err)

	v, err := fc.Int64("int")
	assert.NoError(t, err)
	assert.Equal(t, int64(12), v)

	v, err = fc.Int64("int16")
	assert.NoError(t, err)
	assert.Equal(t, int64(12), v)

	v, err = fc.Int64("int32")
	assert.NoError(t, err)
	assert.Equal(t, int64(12), v)

	v, err = fc.Int64("int64")
	assert.NoError(t, err)
	assert.Equal(t, int64(12), v)

	v, err = fc.Int64("validString")
	assert.NoError(t, err)
	assert.Equal(t, int64(12), v)

	assert.True(t, fc.CanInt64("int"))
	assert.False(t, fc.CanInt64("bool"))

	assert.NotPanics(t, func() { fc.MustInt64("int32", nil) })
	assert.Panics(t, func() { fc.MustInt64("bool", nil) })

	assert.Equal(t, int64(5), fc.MustInt64("_", func(v int64) *int64 { return &v }(5)))
}
