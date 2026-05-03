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

func setCookie(w http.ResponseWriter, name, val string, ttl time.Duration, useInsecureCookie bool) {
	http.SetCookie(w, &http.Cookie{ //#nosec:G124 // Secure defaults to true; callers must explicitly opt into insecure cookies for local HTTP/test deployments.
		Name:     name,
		Value:    url.QueryEscape(val),
		Path:     "/",
		HttpOnly: true,
		Secure:   !useInsecureCookie,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(ttl),
	})
}
