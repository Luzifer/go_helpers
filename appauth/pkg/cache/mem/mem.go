package mem

import (
	"sync"

	"github.com/Luzifer/go_helpers/v2/appauth/pkg/cache"
)

type (
	// Cache implements a very simple in-memory cache not suitable for
	// surviving restarts or multi-instance applications
	Cache struct {
		sess map[string]*cache.Session
		lock sync.RWMutex
	}
)

var _ cache.Cache = &Cache{}

// New creates a new in-mem Cache
func New() *Cache {
	return &Cache{
		sess: make(map[string]*cache.Session),
	}
}

// GetSession returns the session by the given ID or an error
func (c *Cache) GetSession(id string) (cache.Session, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	s, ok := c.sess[id]
	if !ok {
		return cache.Session{}, cache.ErrSessionNotFound
	}

	return *s, nil
}

// RemoveSession removes the session by its ID from the cache
func (c *Cache) RemoveSession(id string) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	delete(c.sess, id)
	return nil
}

// SetSession stores the given session by its ID
func (c *Cache) SetSession(id string, sess cache.Session) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.sess[id] = &sess
	return nil
}
