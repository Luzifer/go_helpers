package cache

import (
	"fmt"
	"time"
)

type (
	// Cache describes what to implement when building a cache provider
	Cache interface {
		GetSession(id string) (Session, error)
		RemoveSession(id string) error
		SetSession(id string, sess Session) error
	}

	// Session holds the data for the stored session
	Session struct {
		AccessToken  string
		IDToken      string
		RefreshToken string

		Expires time.Time // AT expiry
	}
)

// ErrSessionNotFound is an error returned when the cache cannot find
// the given session ID
var ErrSessionNotFound = fmt.Errorf("session not found")
