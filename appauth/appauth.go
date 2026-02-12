// Package appauth contains a helper to add basic OIDC authentication
// to a single-page application with API
package appauth

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/Luzifer/go_helpers/v2/appauth/pkg/cache/mem"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// New creats a new Auth adapter
func New(cfg Config) (*Auth, error) {
	if cfg.IssuerURL == "" || cfg.ClientID == "" || cfg.ClientSecret == "" || cfg.PopupRedirectURL == "" {
		return nil, errors.New("IssuerURL, ClientID, ClientSecret, PopupRedirectURL are required")
	}

	if len(cfg.Scopes) == 0 {
		cfg.Scopes = []string{oidc.ScopeOpenID, "profile", "email"}
	}

	provider, err := oidc.NewProvider(context.Background(), cfg.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("creating OIDC provider: %w", err)
	}

	a := &Auth{
		cfg:      cfg,
		provider: provider,
		verifier: provider.Verifier(&oidc.Config{ClientID: cfg.ClientID}),
		oauth2: oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.PopupRedirectURL,
			Endpoint:     provider.Endpoint(),
			Scopes:       cfg.Scopes,
		},

		sessionCache: mem.New(),
	}

	if cfg.Cache != nil {
		a.sessionCache = cfg.Cache
	}

	return a, nil
}

func (*Auth) authorize(u *User, opts Opts) bool {
	// If no requirements, authenticated is enough
	if len(opts.AnyRole) == 0 && len(opts.AnyGroup) == 0 {
		return true
	}

	if len(opts.AnyRole) > 0 {
		for _, requiredRole := range opts.AnyRole {
			if slices.Contains(u.Roles, requiredRole) {
				return true
			}
		}
	}

	if len(opts.AnyGroup) > 0 {
		for _, requiredGroup := range opts.AnyGroup {
			if slices.Contains(u.Groups, requiredGroup) {
				return true
			}
		}
	}

	return false
}

func (a *Auth) verifyAccessToken(ctx context.Context, raw string) (*User, error) {
	// Verify signature + issuer etc. by parsing as an IDToken-ish structure.
	// This works for JWT access tokens because OIDC provider keys verify JWTs.
	_, err := a.verifier.Verify(ctx, raw)
	if err != nil {
		return nil, fmt.Errorf("verifying access token: %w", err)
	}

	ui, err := a.provider.UserInfo(ctx, oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: raw,
		TokenType:   "Bearer",
	}))
	if err != nil {
		return nil, fmt.Errorf("getting userinfo: %w", err)
	}

	// Claims into map so we can inspect aud/groups/roles flexibly
	var claims map[string]any
	if err := ui.Claims(&claims); err != nil {
		return nil, fmt.Errorf("parsing claims into map: %w", err)
	}

	u := &User{
		Sub:   str(claims["sub"]),
		Email: str(claims["email"]),
		Name:  str(claims["name"]),
		Raw:   claims,
	}

	u.Groups = extractStringSlice(claims["groups"])
	u.Roles = extractRoles(claims, a.cfg.ClientID)

	return u, nil
}
