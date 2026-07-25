// Package redis provides a Redis 8+ or Valkey 9+ backed appauth session cache.
package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/Luzifer/go_helpers/appauth/pkg/cache"
)

const defaultIdleTimeout = time.Hour

type (
	// Cache stores appauth sessions in a Redis hash.
	Cache struct {
		client      *redis.Client
		idleTimeout time.Duration
		hashKey     string
	}

	// Opt applies configuration to a Cache.
	Opt func(*Cache) error
)

var _ cache.Cache = (*Cache)(nil)

// New creates a Redis 8+ or Valkey 9+ backed appauth session cache.
func New(opts ...Opt) (c *Cache, err error) {
	c = &Cache{
		idleTimeout: defaultIdleTimeout,
	}

	for _, opt := range opts {
		if err = opt(c); err != nil {
			return nil, fmt.Errorf("applying option: %w", err)
		}
	}

	if c.client == nil {
		return nil, fmt.Errorf("cache initialized without redis client")
	}

	if c.hashKey == "" {
		return nil, fmt.Errorf("cache initialized without hash-key")
	}

	if c.idleTimeout <= 0 {
		return nil, fmt.Errorf("idle-timeout must be positive duration")
	}

	return c, nil
}

// WithHashKey configures the Redis hash key used to store sessions.
func WithHashKey(hashKey string) Opt {
	return func(c *Cache) error {
		c.hashKey = hashKey
		return nil
	}
}

// WithIdleTimeout configures how long an unused session is retained.
func WithIdleTimeout(d time.Duration) Opt {
	return func(c *Cache) error {
		c.idleTimeout = d
		return nil
	}
}

// WithRedisClient configures the Redis client used by the cache.
func WithRedisClient(client *redis.Client) Opt {
	return func(c *Cache) error {
		c.client = client
		return nil
	}
}

// GetSession returns the session for the given ID.
func (c Cache) GetSession(id string) (s cache.Session, err error) {
	rawSess, err := c.client.HGet(context.TODO(), c.hashKey, id).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return s, cache.ErrSessionNotFound
		}
		return s, fmt.Errorf("getting session: %w", err)
	}

	if err = json.Unmarshal(rawSess, &s); err != nil {
		return s, fmt.Errorf("decoding stored session: %w", err)
	}

	return s, nil
}

// RemoveSession removes the session for the given ID.
func (c Cache) RemoveSession(id string) (err error) {
	if err = c.client.HDel(context.TODO(), c.hashKey, id).Err(); err != nil {
		return fmt.Errorf("deleting hash-key field: %w", err)
	}

	return nil
}

// SetSession stores the session for the given ID.
func (c Cache) SetSession(id string, sess cache.Session) (err error) {
	//#nosec:G117 // OAuth tokens are the session payload and must be serialized into the protected Redis cache.
	rawSess, err := json.Marshal(sess)
	if err != nil {
		return fmt.Errorf("encoding session: %w", err)
	}

	if err = c.client.HSetEXWithArgs(context.TODO(), c.hashKey, &redis.HSetEXOptions{
		ExpirationType: redis.HSetEXExpirationEXAT,
		ExpirationVal:  sess.LastSeen.Add(c.idleTimeout).Unix(),
	}, id, string(rawSess)).Err(); err != nil {
		return fmt.Errorf("storing session: %w", err)
	}

	return nil
}
