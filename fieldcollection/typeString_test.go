package fieldcollection

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type (
	testStringer struct{}
)

func (testStringer) String() string { return "ohai" }

func TestString(t *testing.T) {
	fc := FieldCollectionFromData(map[string]any{
		"int":         12,
		"validString": "Ello!",
		"stringer":    testStringer{},
	})

	_, err := fc.String("_")
	require.ErrorIs(t, err, ErrValueNotSet)

	_, err = fc.String("int")
	require.ErrorIs(t, err, ErrValueMismatch)

	_, err = fc.String("invalidString")
	require.Error(t, err)

	v, err := fc.String("validString")
	require.NoError(t, err)
	assert.Equal(t, "Ello!", v)

	v, err = fc.String("stringer")
	require.NoError(t, err)
	assert.Equal(t, "ohai", v)

	assert.True(t, fc.CanString("validString"))
	assert.False(t, fc.CanString("bool"))

	assert.NotPanics(t, func() { fc.MustString("validString", nil) })
	assert.Panics(t, func() { fc.MustString("bool", nil) })

	assert.Equal(t, "a", fc.MustString("_", func(v string) *string { return &v }("a")))
}
