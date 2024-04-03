package fieldcollection

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFieldCollectionJSONMarshal(t *testing.T) {
	var (
		buf = new(bytes.Buffer)
		raw = `{"key1":"test1","key2":"test2"}`
		f   = NewFieldCollection()
	)

	require.NoError(t, json.NewDecoder(strings.NewReader(raw)).Decode(f))
	require.NoError(t, json.NewEncoder(buf).Encode(f))
	assert.Equal(t, raw, strings.TrimSpace(buf.String()))
}
