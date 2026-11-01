// Package sentrus sends selected Logrus entries as Sentry-compatible issue events.
package sentrus

import (
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	defaultQueueLength = 100
	defaultTimeout     = 2 * time.Second
)

type (
	config struct {
		dsn         string
		environment string
		levels      []logrus.Level
		queueLength int
		release     string
		timeout     time.Duration
	}

	// Option configures a Hook.
	Option func(*config) error
)

var errDSNRequired = errors.New("DSN is required")

// WithDSN configures the Sentry-compatible DSN used to submit events.
func WithDSN(dsn string) Option {
	return func(cfg *config) error {
		cfg.dsn = dsn
		return nil
	}
}

// WithEnvironment adds environment to every submitted event.
func WithEnvironment(environment string) Option {
	return func(cfg *config) error {
		cfg.environment = environment
		return nil
	}
}

// WithLevels configures the exact Logrus levels submitted as issue events.
func WithLevels(levels ...logrus.Level) Option {
	return func(cfg *config) error {
		if len(levels) == 0 {
			return errors.New("levels must not be empty")
		}

		seen := make(map[logrus.Level]struct{}, len(levels))
		cfg.levels = nil
		for _, level := range levels {
			if !isValidLevel(level) {
				return fmt.Errorf("invalid Logrus level: %d", level)
			}
			if _, ok := seen[level]; ok {
				continue
			}

			seen[level] = struct{}{}
			cfg.levels = append(cfg.levels, level)
		}

		return nil
	}
}

// WithQueueLength configures the number of events waiting for delivery.
func WithQueueLength(length int) Option {
	return func(cfg *config) error {
		if length <= 0 {
			return errors.New("queue length must be positive")
		}

		cfg.queueLength = length
		return nil
	}
}

// WithRelease adds release to every submitted event.
func WithRelease(release string) Option {
	return func(cfg *config) error {
		cfg.release = release
		return nil
	}
}

// WithTimeout configures the timeout of each delivery attempt.
func WithTimeout(timeout time.Duration) Option {
	return func(cfg *config) error {
		if timeout <= 0 {
			return errors.New("timeout must be positive")
		}

		cfg.timeout = timeout
		return nil
	}
}

func defaultConfig() config {
	return config{
		levels: []logrus.Level{
			logrus.ErrorLevel,
			logrus.FatalLevel,
			logrus.PanicLevel,
		},
		queueLength: defaultQueueLength,
		timeout:     defaultTimeout,
	}
}

func isValidLevel(level logrus.Level) bool {
	return level >= logrus.PanicLevel && level <= logrus.TraceLevel
}
