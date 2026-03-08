package appauth

import (
	"context"
	"testing"
	"time"

	"github.com/Luzifer/go_helpers/appauth/pkg/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCache struct {
	sess      map[string]cache.Session
	removeIDs []string
}

func newTestCache() *testCache {
	return &testCache{sess: map[string]cache.Session{}}
}

func (c *testCache) GetSession(id string) (cache.Session, error) {
	s, ok := c.sess[id]
	if !ok {
		return cache.Session{}, cache.ErrSessionNotFound
	}
	return s, nil
}

func (c *testCache) RemoveSession(id string) error {
	c.removeIDs = append(c.removeIDs, id)
	delete(c.sess, id)
	return nil
}

func (c *testCache) SetSession(id string, sess cache.Session) error {
	c.sess[id] = sess
	return nil
}

func TestExchangeTokenThroughCacheIdleTimeout(t *testing.T) {
	tc := newTestCache()
	tc.sess["a"] = cache.Session{
		AccessToken: "token",
		Expires:     time.Now().Add(time.Hour),
		CreatedAt:   time.Now().Add(-2 * time.Hour),
		LastSeen:    time.Now().Add(-2 * time.Hour),
	}

	a := &Auth{
		cfg: Config{
			SessionIdleTimeout: time.Minute,
		},
		sessionCache: tc,
	}

	_, err := a.exchangeTokenThroughCache(context.Background(), "a")
	require.Error(t, err)
	require.Len(t, tc.removeIDs, 1)
	assert.Equal(t, "a", tc.removeIDs[0])
}

func TestExchangeTokenThroughCacheAbsoluteTimeout(t *testing.T) {
	tc := newTestCache()
	tc.sess["a"] = cache.Session{
		AccessToken: "token",
		Expires:     time.Now().Add(time.Hour),
		CreatedAt:   time.Now().Add(-2 * time.Hour),
		LastSeen:    time.Now().Add(-10 * time.Second),
	}

	a := &Auth{
		cfg: Config{
			SessionAbsoluteTimeout: time.Minute,
		},
		sessionCache: tc,
	}

	_, err := a.exchangeTokenThroughCache(context.Background(), "a")
	require.Error(t, err)
	require.Len(t, tc.removeIDs, 1)
	assert.Equal(t, "a", tc.removeIDs[0])
}

func TestExchangeTokenThroughCacheTimeoutsDisabled(t *testing.T) {
	tc := newTestCache()
	tc.sess["a"] = cache.Session{
		AccessToken: "token",
		Expires:     time.Now().Add(time.Hour),
		CreatedAt:   time.Now().Add(-24 * time.Hour),
		LastSeen:    time.Now().Add(-24 * time.Hour),
	}

	a := &Auth{
		cfg: Config{
			SessionIdleTimeout:     0,
			SessionAbsoluteTimeout: 0,
		},
		sessionCache: tc,
	}

	tok, err := a.exchangeTokenThroughCache(context.Background(), "a")
	require.NoError(t, err)
	assert.Equal(t, "token", tok)
	assert.Empty(t, tc.removeIDs)
}
