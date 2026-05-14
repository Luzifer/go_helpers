package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Luzifer/go_helpers/appauth"
)

const (
	appauthClientID     = "appauth-test"
	appauthClientSecret = "appauth-secret" //#nosec:G101 // Only Test-Secret
	appauthRedirectURL  = "http://127.0.0.1:64557"
	dexIssuerURL        = "http://127.0.0.1:64556/dex"
	dexPassword         = "password"
	dexUserEmail        = "jane.doe@example.com"
	dexUsername         = "jane.doe"
)

type testLogger struct{ t *testing.T }

func (t testLogger) Printf(format string, args ...any) { t.t.Logf(format, args...) }

func TestIntegration(t *testing.T) {
	// Create HTTP server to mount our popup and test endpoint
	r := mux.NewRouter()
	listener, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:64557")
	require.NoError(t, err, "opening appauth test listener")

	httpSrv := httptest.NewUnstartedServer(r)
	httpSrv.Listener = listener
	httpSrv.Start()
	t.Cleanup(httpSrv.Close)

	// Create Auth
	a, err := appauth.New(appauth.Config{
		InsecureCookie:            true, // required on Go 1.25 tests
		IssuerURL:                 dexIssuerURL,
		ClientID:                  appauthClientID,
		ClientSecret:              appauthClientSecret,
		PopupRedirectURL:          appauthRedirectURL,
		Scopes:                    []string{oidc.ScopeOpenID, "profile", "email", "groups"}, // roles is not supported
		AllowedPostMessageOrigins: []string{"http://localhost"},
		Logger:                    testLogger{t},
	})
	require.NoError(t, err, "creating Auth")

	// Add endpoints
	r.HandleFunc("/", a.ServePopup)
	r.Handle("/testAuth", a.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := appauth.UserFromContext(r.Context())
		if !ok {
			http.Error(w, "missing user", http.StatusInternalServerError)
			return
		}

		_ = json.NewEncoder(w).Encode(user)
	}), appauth.Opts{
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

	authRedirect := headers.Get("Location")
	assert.NotEmpty(t, authRedirect)

	authRedirect, err = completeDexLogin(client, authRedirect)
	require.NoError(t, err)

	// Back to the Popup
	body, _, status, err = testReq(client, http.MethodGet, authRedirect, nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, status)

	assert.Contains(t, body, `token: "`, "there should be an access token")
	for _, claim := range []string{
		`user: {`,
		fmt.Sprintf(`"email":"%s"`, dexUserEmail),
		`"groups":["engineering"]`,
		fmt.Sprintf(`"preferred_username":"%s"`, dexUsername),
	} {
		assert.Contains(t, body, claim)
	}

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
		fmt.Sprintf(`"email":"%s"`, dexUserEmail),
		`"groups":["engineering"]`,
		fmt.Sprintf(`"preferred_username":"%s"`, dexUsername),
	} {
		assert.Contains(t, body, claim)
	}
}

func testReq(client *http.Client, method, reqURL string, inHdr http.Header) (body string, headers http.Header, status int, err error) {
	return testReqBody(client, method, reqURL, nil, inHdr)
}

func testReqForm(client *http.Client, reqURL string, values url.Values) (body string, headers http.Header, status int, err error) {
	return testReqBody(client, http.MethodPost, reqURL, strings.NewReader(values.Encode()), http.Header{
		"Content-Type": []string{"application/x-www-form-urlencoded"},
	})
}

func testReqBody(client *http.Client, method, reqURL string, inBody io.Reader, inHdr http.Header) (body string, headers http.Header, status int, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, reqURL, inBody)
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
	defer resp.Body.Close() //nolint:errcheck // only a test client

	bodyRaw, err := io.ReadAll(resp.Body)
	if err != nil {
		return body, nil, 0, fmt.Errorf("reading body: %w", err)
	}

	return string(bodyRaw), resp.Header, resp.StatusCode, nil
}
