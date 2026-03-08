package appauth

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
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

func setCookie(w http.ResponseWriter, r *http.Request, name, val string, ttl time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    url.QueryEscape(val),
		Path:     "/",
		HttpOnly: true,
		Secure:   !isLocalHost(r.Host),
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(ttl),
	})
}

func isLocalHost(hostport string) bool {
	host := hostport

	if h, _, err := net.SplitHostPort(hostport); err == nil {
		host = h
	}

	host = strings.Trim(host, "[]")

	switch host {
	case "localhost", "127.0.0.1", "::1":
		return true
	default:
		return false
	}
}
