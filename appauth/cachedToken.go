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

	now := time.Now()
	if sess.CreatedAt.IsZero() {
		sess.CreatedAt = now
	}
	if sess.LastSeen.IsZero() {
		sess.LastSeen = now
	}

	if a.cfg.SessionAbsoluteTimeout > 0 && sess.CreatedAt.Add(a.cfg.SessionAbsoluteTimeout).Before(now) {
		_ = a.sessionCache.RemoveSession(sessID)
		return "", fmt.Errorf("session expired by absolute timeout")
	}

	if a.cfg.SessionIdleTimeout > 0 && sess.LastSeen.Add(a.cfg.SessionIdleTimeout).Before(now) {
		_ = a.sessionCache.RemoveSession(sessID)
		return "", fmt.Errorf("session expired by idle timeout")
	}

	sess.LastSeen = now

	if sess.Expires.After(now) {
		// Access token is still valid
		if err = a.sessionCache.SetSession(sessID, sess); err != nil {
			return "", fmt.Errorf("updating session usage: %w", err)
		}
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
