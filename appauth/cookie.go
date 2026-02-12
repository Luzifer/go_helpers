package appauth

import (
	"fmt"
	"net/http"
	"net/url"
	"time"
)

func readCookie(r *http.Request, name string) (string, error) {
	c, err := r.Cookie(name)
	if err != nil {
		return "", fmt.Errorf("getting cookie: %w", err)
	}

	v, err := url.QueryUnescape(c.Value)
	if err != nil {
		return "", fmt.Errorf("unescaping cookie value: %w", err)
	}

	return v, nil
}

func setCookie(w http.ResponseWriter, name, val string, ttl time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    url.QueryEscape(val),
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(ttl),
	})
}
