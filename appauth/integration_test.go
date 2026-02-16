package appauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gorilla/mux"
	"github.com/oauth2-proxy/mockoidc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testLogger struct{ t *testing.T }

func (t testLogger) Printf(format string, args ...any) { t.t.Logf(format, args...) }

func TestIntegration(t *testing.T) {
	// Create Mock-OIDC server with default user logged in
	oidcSrv, err := mockoidc.Run()
	require.NoError(t, err, "running mockoidc")
	t.Cleanup(func() {
		_ = oidcSrv.Shutdown()
	})

	// Create HTTP server to mount our popup and test endpoint
	r := mux.NewRouter()
	httpSrv := httptest.NewServer(r)
	t.Cleanup(httpSrv.Close)

	// Create Auth
	a, err := New(Config{
		IssuerURL:                 oidcSrv.Issuer(),
		ClientID:                  oidcSrv.ClientID,
		ClientSecret:              oidcSrv.ClientSecret,
		PopupRedirectURL:          httpSrv.URL,
		Scopes:                    []string{oidc.ScopeOpenID, "profile", "email", "groups"}, // roles is not supported
		AllowedPostMessageOrigins: []string{"http://localhost"},
		Logger:                    testLogger{t},
	})
	require.NoError(t, err, "creating Auth")

	// Add endpoints
	r.HandleFunc("/", a.ServePopup)
	r.Handle("/testAuth", a.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := UserFromContext(r.Context())
		if !ok {
			http.Error(w, "missing user", http.StatusInternalServerError)
			return
		}

		_ = json.NewEncoder(w).Encode(user)
	}), Opts{
		AnyGroup: []string{"engineering"},
	}))

	// Set up a cookie jar for the roundtrip and a non-redirect HTTP client
	jar, err := cookiejar.New(nil)
	require.NoError(t, err, "creating cookie jar")

	client := &http.Client{
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		Jar:           jar,
	}

	// Call the popup
	body, headers, status, err := testReq(client, http.MethodGet, httpSrv.URL+"?origin=http://localhost", nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusFound, status, body)

	authRedirect := headers.Get("location")
	assert.NotEmpty(t, authRedirect)

	// Call the OIDC redirect
	body, headers, status, err = testReq(client, http.MethodGet, authRedirect, nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusFound, status, body)

	authRedirect = headers.Get("location")
	require.NotEmpty(t, authRedirect)

	// Back to the Popup
	body, _, status, err = testReq(client, http.MethodGet, authRedirect, nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, status)

	assert.Contains(t, body, `token: "`, "there should be an access token")

	// Get the access token from the body
	match := regexp.MustCompile(`token: "([^"]+)"`).FindStringSubmatch(body)
	require.Len(t, match, 2)
	token := match[1]

	// Require auth with the token we got
	body, _, status, err = testReq(client, http.MethodGet, httpSrv.URL+"/testAuth", http.Header{
		"authorization": []string{"Session " + token},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, status, body)

	for _, claim := range []string{
		`"email":"jane.doe@example.com"`,
		`"groups":["engineering","design"]`,
		`"preferred_username":"jane.doe"`,
	} {
		assert.Contains(t, body, claim)
	}
}

func testReq(client *http.Client, method, url string, inHdr http.Header) (body string, headers http.Header, status int, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return "", nil, 0, fmt.Errorf("creating request: %w", err)
	}

	if inHdr != nil {
		req.Header = inHdr
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", nil, 0, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	bodyRaw, err := io.ReadAll(resp.Body)
	if err != nil {
		return body, nil, 0, fmt.Errorf("reading body: %w", err)
	}

	return string(bodyRaw), resp.Header, resp.StatusCode, nil
}
