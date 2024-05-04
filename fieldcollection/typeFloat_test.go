package fieldcollection

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFloat64(t *testing.T) {
	fc := FieldCollectionFromData(map[string]any{
		"int":           int(12),
		"int16":         int16(12),
		"int32":         int32(12),
		"int64":         int64(12),
		"float64":       float64(12),
		"bool":          true,
		"invalidString": "I'm a string!",
		"validString":   "12",
	})

	_, err := fc.Float64("_")
	assert.ErrorIs(t, err, ErrValueNotSet)

	_, err = fc.Float64("bool")
	assert.ErrorIs(t, err, ErrValueMismatch)

	_, err = fc.Float64("invalidString")
	assert.Error(t, err)

	v, err := fc.Float64("int")
	assert.NoError(t, err)
	assert.Equal(t, float64(12), v)

	v, err = fc.Float64("int16")
	assert.NoError(t, err)
	assert.Equal(t, float64(12), v)

	v, err = fc.Float64("int32")
	assert.NoError(t, err)
	assert.Equal(t, float64(12), v)

	v, err = fc.Float64("int64")
	assert.NoError(t, err)
	assert.Equal(t, float64(12), v)

	v, err = fc.Float64("validString")
	assert.NoError(t, err)
	assert.Equal(t, float64(12), v)

	assert.True(t, fc.CanFloat64("int"))
	assert.False(t, fc.CanFloat64("bool"))

	assert.NotPanics(t, func() { fc.MustFloat64("int32", nil) })
	assert.Panics(t, func() { fc.MustFloat64("bool", nil) })

	assert.Equal(t, float64(5), fc.MustFloat64("_", func(v float64) *float64 { return &v }(5)))
}
