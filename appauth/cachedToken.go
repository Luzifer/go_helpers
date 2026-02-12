package appauth

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/oauth2"
)

func (a *Auth) exchangeTokenThroughCache(ctx context.Context, sessID string) (token string, err error) {
	sess, err := a.sessionCache.GetSession(sessID)
	if err != nil {
		return "", fmt.Errorf("getting session from cache: %w", err)
	}

	if sess.Expires.After(time.Now()) {
		// Access token is still valid
		return sess.AccessToken, nil
	}

	// Renew token and store session back
	seed := &oauth2.Token{
		RefreshToken: sess.RefreshToken,
	}

	tok, err := a.oauth2.TokenSource(ctx, seed).Token()
	if err != nil {
		return "", fmt.Errorf("refreshing token: %w", err)
	}

	sess.AccessToken = tok.AccessToken
	sess.Expires = tok.Expiry

	if tok.RefreshToken != "" {
		sess.RefreshToken = tok.RefreshToken
	}

	if idt, ok := tok.Extra("id_token").(string); ok {
		sess.IDToken = idt
	}

	if err = a.sessionCache.SetSession(sessID, sess); err != nil {
		return "", fmt.Errorf("updating session: %w", err)
	}

	return sess.AccessToken, nil
}
