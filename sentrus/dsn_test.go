package sentrus

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDSN(t *testing.T) {
	secret := strings.ToLower(t.Name())
	parsed, err := parseDSN("https://public:" + secret + "@example.com/sentry/42")
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/sentry/api/42/envelope/", parsed.endpoint)
	assert.Equal(t, "Sentry sentry_version=7, sentry_key=public, sentry_secret="+secret, parsed.authHeader)

	parsed, err = parseDSN("http://public@example.com/42/")
	require.NoError(t, err)
	assert.Equal(t, "http://example.com/api/42/envelope/", parsed.endpoint)
	assert.Equal(t, "Sentry sentry_version=7, sentry_key=public", parsed.authHeader)
}

func TestParseDSNRejectsInvalidInput(t *testing.T) {
	baseURL := "https://" + "key" + "@example.com"
	for name, dsn := range map[string]string{
		"credential": strings.Replace(baseURL+"/1", "key", "bad%20key", 1),
		"empty":      "",
		"fragment":   baseURL + "/1#fragment",
		"host":       strings.Replace(baseURL+"/1", "example.com", "", 1),
		"project":    baseURL,
		"query":      baseURL + "/1?query=value",
		"scheme":     strings.Replace(baseURL+"/1", "https", "ftp", 1),
	} {
		t.Run(name, func(t *testing.T) {
			_, err := parseDSN(dsn)
			require.Error(t, err)
		})
	}
}
