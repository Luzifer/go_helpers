package http //revive:disable-line:package-naming // kept for historical reasons

import (
	"crypto/md5" //#nosec:G501 // required for RFC-compatible Digest MD5
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
)

// GetDigestAuth builds a Digest authorization header for the given response challenge.
func GetDigestAuth(resp *http.Response, method, requestPath, user, password string) string {
	auth, err := GetDigestAuthWithError(resp, method, requestPath, user, password)
	if err != nil {
		return ""
	}

	return auth
}

// GetDigestAuthWithError builds a Digest authorization header for the given response challenge.
func GetDigestAuthWithError(resp *http.Response, method, requestPath, user, password string) (string, error) {
	params := make(map[string]string)
	for part := range strings.SplitSeq(resp.Header.Get("Www-Authenticate"), " ") {
		if !strings.Contains(part, `="`) {
			continue
		}
		spl := strings.Split(strings.Trim(part, " ,"), "=")
		if !slices.Contains([]string{"nonce", "realm", "qop"}, spl[0]) {
			continue
		}
		params[spl[0]] = strings.Trim(spl[1], `"`)
	}

	b := make([]byte, 8)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", fmt.Errorf("generating digest cnonce: %w", err)
	}

	params["cnonce"] = fmt.Sprintf("%x", b)
	params["nc"] = "1"
	params["uri"] = requestPath
	params["username"] = user
	params["response"] = getMD5([]string{
		getMD5([]string{params["username"], params["realm"], password}),
		params["nonce"],
		params["nc"],
		params["cnonce"],
		params["qop"],
		getMD5([]string{method, requestPath}),
	})

	var authParts []string
	for k, v := range params {
		authParts = append(authParts, fmt.Sprintf("%s=%q", k, v))
	}
	return "Digest " + strings.Join(authParts, ", "), nil
}

func getMD5(in []string) string {
	h := md5.New() //#nosec:G401 // required for RFC-compatible Digest MD5
	h.Write([]byte(strings.Join(in, ":")))
	return hex.EncodeToString(h.Sum(nil))
}
