package fieldcollection

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringSlice(t *testing.T) {
	fc := FieldCollectionFromData(map[string]any{
		"int":          12,
		"valid":        []string{"ohai"},
		"invalidSlice": []int{12},
		"mixed":        []any{"ohai", 12},
		"validAny":     []any{"ohai"},
	})

	_, err := fc.StringSlice("_")
	assert.ErrorIs(t, err, ErrValueNotSet)

	_, err = fc.StringSlice("int")
	assert.ErrorIs(t, err, ErrValueMismatch)

	_, err = fc.StringSlice("invalidSlice")
	assert.Error(t, err)

	_, err = fc.StringSlice("mixed")
	assert.Error(t, err)

	v, err := fc.StringSlice("valid")
	assert.NoError(t, err)
	assert.Equal(t, []string{"ohai"}, v)

	v, err = fc.StringSlice("validAny")
	assert.NoError(t, err)
	assert.Equal(t, []string{"ohai"}, v)

	assert.True(t, fc.CanStringSlice("valid"))
	assert.False(t, fc.CanStringSlice("bool"))

	assert.NotPanics(t, func() { fc.MustStringSlice("valid", nil) })
	assert.Panics(t, func() { fc.MustStringSlice("bool", nil) })

	assert.Equal(t, []string{"a"}, fc.MustStringSlice("_", func(v []string) *[]string { return &v }([]string{"a"})))
}
