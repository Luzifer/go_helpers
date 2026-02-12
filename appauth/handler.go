package appauth

import (
	"context"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/Luzifer/go_helpers/appauth/v2/pkg/cache"
	"golang.org/x/oauth2"
)

const (
	flowCookieTimeout = 5 * time.Minute
	stateLength       = 32
	verifierLength    = 64
	sessionIDLength   = 64
)

// RequireAuth shields the given next Handler with the given auth
// requirements. The identified user is available through UserFromContext
// from the request context in the next Handler
func (a *Auth) RequireAuth(next http.Handler, opts Opts) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		tokenType, token, ok := strings.Cut(r.Header.Get("Authorization"), " ")
		if !ok {
			a.logf("auth: missing authorization path=%s", r.URL.Path)
			http.NotFound(w, r)
			return
		}

		switch tokenType {
		case "Bearer":
			// That's expected from API-clients with direct OIDC-Provider
			// access such as server-to-server or desktop applications, we
			// use the token directly in this case.

		case "Session":
			// We got a session identifier and need to fetch a token from
			// the cache and possibly renew it

			var err error
			if token, err = a.exchangeTokenThroughCache(r.Context(), token); err != nil {
				a.logf("auth: exchanging session for token path=%s type=%s", r.URL.Path, tokenType)
				http.NotFound(w, r)
				return
			}

		default:
			a.logf("auth: invalid token type path=%s type=%s", r.URL.Path, tokenType)
			http.NotFound(w, r)
			return
		}

		u, err := a.verifyAccessToken(r.Context(), token)
		if err != nil {
			a.logf("auth: invalid token path=%s err=%v", r.URL.Path, err)
			http.NotFound(w, r)
			return
		}

		if !a.authorize(u, opts) {
			a.logf("auth: forbidden path=%s sub=%s need_roles=%v need_groups=%v have_roles=%v have_groups=%v",
				r.URL.Path, u.Sub, opts.AnyRole, opts.AnyGroup, u.Roles, u.Groups,
			)
			http.NotFound(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), userKey, u)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ServePopup is a mountable HTTP HandleFunc which initiates the
// redirect to the OIDC server and on return to the same URL exchanges
// the code for the token, then passes the token back to the requesting
// Javascript through the window.opener.PostMessage function.
func (a *Auth) ServePopup(w http.ResponseWriter, r *http.Request) {
	// Are we currently in the callback-state of the flow?
	if r.URL.Query().Get("code") != "" || r.URL.Query().Get("error") != "" {
		a.popupCallback(w, r)
		return
	}

	// We were not, lets start the roundtrip!
	a.popupStart(w, r)
}

func (a *Auth) allowedOrigin(origin string) (string, bool) {
	if origin == "" {
		return "", false
	}

	if slices.Contains(a.cfg.AllowedPostMessageOrigins, origin) {
		return origin, true
	}

	return "", false
}

func (a *Auth) popupStart(w http.ResponseWriter, r *http.Request) {
	state, err := randB64(stateLength)
	if err != nil {
		http.Error(w, "state", http.StatusInternalServerError)
		return
	}

	verifier, err := randB64(verifierLength)
	if err != nil {
		http.Error(w, "pkce", http.StatusInternalServerError)
		return
	}

	origin := r.URL.Query().Get("origin") // the opener's origin

	// Store values for callback validation
	setCookie(w, "oidc_state", state, flowCookieTimeout)
	setCookie(w, "oidc_verifier", verifier, flowCookieTimeout)
	if origin != "" {
		setCookie(w, "oidc_origin", origin, flowCookieTimeout)
	}

	challenge := pkceChallengeS256(verifier)

	authURL := a.oauth2.AuthCodeURL(
		state,
		oauth2.SetAuthURLParam("code_challenge", challenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)

	http.Redirect(w, r, authURL, http.StatusFound)
}

func (a *Auth) popupCallback(w http.ResponseWriter, r *http.Request) {
	if e := r.URL.Query().Get("error"); e != "" {
		a.logf("popup: oidc error=%s desc=%s", e, r.URL.Query().Get("error_description"))
		writeClosePage(w, "Login failed.")
		return
	}

	code := r.URL.Query().Get("code")
	stateQ := r.URL.Query().Get("state")
	if code == "" || stateQ == "" {
		a.logf("popup: missing code/state")
		writeClosePage(w, "Bad callback.")
		return
	}

	stateC, err := readCookie(r, "oidc_state")
	if err != nil || stateC != stateQ {
		a.logf("popup: bad state err=%v", err)
		writeClosePage(w, "Bad state.")
		return
	}

	verifier, err := readCookie(r, "oidc_verifier")
	if err != nil || verifier == "" {
		a.logf("popup: missing verifier err=%v", err)
		writeClosePage(w, "Bad verifier.")
		return
	}

	tok, err := a.oauth2.Exchange(r.Context(), code,
		oauth2.SetAuthURLParam("code_verifier", verifier),
	)
	if err != nil {
		a.logf("popup: exchange failed err=%v", err)
		writeClosePage(w, "Exchange failed.")
		return
	}

	access, _ := tok.Extra("access_token").(string)
	idt, _ := tok.Extra("id_token").(string)
	refresh, _ := tok.Extra("refresh_token").(string)

	origin, _ := readCookie(r, "oidc_origin")
	targetOrigin, ok := a.allowedOrigin(origin)
	if !ok {
		a.logf("popup: origin not allowed origin=%q", origin)
		// Refuse to deliver token
		writeClosePage(w, "Origin not allowed.")
		return
	}

	sessID, err := randB64(sessionIDLength)
	if err != nil {
		a.logf("popup: creating session err=%v", err)
		writeClosePage(w, "Creating session failed.")
		return
	}

	if err = a.sessionCache.SetSession(sessID, cache.Session{
		AccessToken:  access,
		IDToken:      idt,
		RefreshToken: refresh,
		Expires:      tok.Expiry,
	}); err != nil {
		a.logf("popup: writing session err=%v", err)
		writeClosePage(w, "Writing session failed.")
		return
	}

	payload := map[string]any{
		"access_token": sessID,
	}

	writePostMessageAndClose(w, targetOrigin, payload)
}
