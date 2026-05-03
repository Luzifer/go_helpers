package fieldcollection

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBool(t *testing.T) {
	fc := FieldCollectionFromData(map[string]any{
		"int":                  12,
		"invalidBoolString":    "I'm a string!",
		"validBool":            true,
		"validBoolString":      "true",
		"validBoolStringFalse": "false",
	})

	_, err := fc.Bool("_")
	require.ErrorIs(t, err, ErrValueNotSet)

	_, err = fc.Bool("int")
	require.ErrorIs(t, err, ErrValueMismatch)

	_, err = fc.Bool("invalidBoolString")
	require.Error(t, err)

	v, err := fc.Bool("validBool")
	require.NoError(t, err)
	assert.True(t, v)

	v, err = fc.Bool("validBoolString")
	require.NoError(t, err)
	assert.True(t, v)

	v, err = fc.Bool("validBoolStringFalse")
	require.NoError(t, err)
	assert.False(t, v)

	assert.True(t, fc.CanBool("validBool"))
	assert.False(t, fc.CanBool("int"))

	assert.NotPanics(t, func() { fc.MustBool("validBool", nil) })
	assert.Panics(t, func() { fc.MustBool("int", nil) })

	assert.True(t, fc.MustBool("_", func(v bool) *bool { return &v }(true)))
}
