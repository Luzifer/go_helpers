package fieldcollection

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestFieldCollectionYAMLMarshal(t *testing.T) {
	var (
		buf = new(bytes.Buffer)
		raw = "key1: test1\nkey2: test2"
		f   = NewFieldCollection()
	)

	require.NoError(t, yaml.NewDecoder(strings.NewReader(raw)).Decode(f))
	require.NoError(t, yaml.NewEncoder(buf).Encode(f))
	assert.Equal(t, raw, strings.TrimSpace(buf.String()))
}
