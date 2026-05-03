package fieldcollection

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDuration(t *testing.T) {
	fc := FieldCollectionFromData(map[string]any{
		"int":           12,
		"bool":          true,
		"invalidString": "I'm a string!",
		"valid":         time.Second,
		"validString":   "12m",
	})

	_, err := fc.Duration("_")
	require.ErrorIs(t, err, ErrValueNotSet)

	_, err = fc.Duration("bool")
	require.ErrorIs(t, err, ErrValueMismatch)

	_, err = fc.Duration("invalidString")
	require.Error(t, err)

	v, err := fc.Duration("valid")
	require.NoError(t, err)
	assert.Equal(t, time.Second, v)

	v, err = fc.Duration("validString")
	require.NoError(t, err)
	assert.Equal(t, 12*time.Minute, v)

	v, err = fc.Duration("int")
	require.NoError(t, err)
	assert.Equal(t, 12*time.Nanosecond, v)

	assert.True(t, fc.CanDuration("valid"))
	assert.False(t, fc.CanDuration("bool"))

	assert.NotPanics(t, func() { fc.MustDuration("valid", nil) })
	assert.Panics(t, func() { fc.MustDuration("bool", nil) })

	assert.Equal(t, time.Second, fc.MustDuration("_", func(v time.Duration) *time.Duration { return &v }(time.Second)))
}
