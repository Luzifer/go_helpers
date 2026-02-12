package appauth

import (
	"github.com/Luzifer/go_helpers/v2/appauth/pkg/cache"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type (
	// Auth contains the parts required for authentication and authorization
	// against an OIDC server
	Auth struct {
		cfg Config

		provider *oidc.Provider
		verifier *oidc.IDTokenVerifier // We will verify JWTs; access tokens are JWTs in KC by default.

		oauth2 oauth2.Config

		sessionCache cache.Cache
	}

	// Config holds the configuration for the Auth adapter
	Config struct {
		IssuerURL string

		// Popup client
		ClientID         string
		ClientSecret     string
		PopupRedirectURL string // MUST be the same route you mount the handler on

		Scopes []string // e.g. []string{oidc.ScopeOpenID, "profile", "email"}

		// Who may receive tokens via postMessage (strict allowlist)
		AllowedPostMessageOrigins []string

		Logger Logger      // optional
		Cache  cache.Cache // optional
	}

	// Logger defines what a log-provider must implement in order to be
	// usable for this library
	Logger interface {
		Printf(format string, v ...any)
	}

	// Opts controls the authorization within a route
	Opts struct {
		AnyRole  []string // realm or client roles
		AnyGroup []string
	}

	ctxKey int
)

const userKey ctxKey = 1
