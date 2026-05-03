package fieldcollection

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpect(t *testing.T) {
	var f *FieldCollection
	require.NoError(t, f.Expect())
	require.Error(t, f.Expect("foo"))
}

func TestFieldCollectionNilClone(*testing.T) {
	var f *FieldCollection
	f.Clone()
}

func TestFieldCollectionNilDataGet(t *testing.T) {
	var f *FieldCollection

	for name, fn := range map[string]func(name string) bool{
		"bool":        f.CanBool,
		"duration":    f.CanDuration,
		"int64":       f.CanInt64,
		"string":      f.CanString,
		"stringSlice": f.CanStringSlice,
	} {
		assert.False(t, fn("foo"), "%s key is available", name)
	}
}

func TestGet(t *testing.T) {
	f := &FieldCollection{}
	_, err := f.Get("foo")
	require.Error(t, err)

	f.Set("foo", "bar")
	_, err = f.Get("bar")
	require.ErrorIs(t, err, ErrValueNotSet)

	v, err := f.Get("foo")
	require.NoError(t, err)
	assert.Equal(t, "bar", v)
}

func TestKeys(t *testing.T) {
	f := FieldCollectionFromData(map[string]any{
		"foo": "bar",
	})
	assert.Equal(t, []string{"foo"}, f.Keys())
}

func TestSetOnNew(*testing.T) {
	f := new(FieldCollection)
	f.Set("foo", "bar")
}
