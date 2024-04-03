package fieldcollection

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
	assert.ErrorIs(t, err, ErrValueNotSet)

	_, err = fc.String("int")
	assert.ErrorIs(t, err, ErrValueMismatch)

	_, err = fc.String("invalidString")
	assert.Error(t, err)

	v, err := fc.String("validString")
	assert.NoError(t, err)
	assert.Equal(t, "Ello!", v)

	v, err = fc.String("stringer")
	assert.NoError(t, err)
	assert.Equal(t, "ohai", v)

	assert.True(t, fc.CanString("validString"))
	assert.False(t, fc.CanString("bool"))

	assert.NotPanics(t, func() { fc.MustString("validString", nil) })
	assert.Panics(t, func() { fc.MustString("bool", nil) })

	assert.Equal(t, "a", fc.MustString("_", func(v string) *string { return &v }("a")))
}
