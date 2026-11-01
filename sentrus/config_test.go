package sentrus

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOptions(t *testing.T) {
	cfg := defaultConfig()
	require.NoError(t, WithDSN("dsn")(&cfg))
	require.NoError(t, WithEnvironment("production")(&cfg))
	require.NoError(t, WithLevels(logrus.InfoLevel, logrus.InfoLevel, logrus.TraceLevel)(&cfg))
	require.NoError(t, WithQueueLength(23)(&cfg))
	require.NoError(t, WithRelease("release")(&cfg))
	require.NoError(t, WithTimeout(time.Second)(&cfg))

	assert.Equal(t, "dsn", cfg.dsn)
	assert.Equal(t, "production", cfg.environment)
	assert.Equal(t, []logrus.Level{logrus.InfoLevel, logrus.TraceLevel}, cfg.levels)
	assert.Equal(t, 23, cfg.queueLength)
	assert.Equal(t, "release", cfg.release)
	assert.Equal(t, time.Second, cfg.timeout)
}

func TestOptionsRejectInvalidValues(t *testing.T) {
	for name, option := range map[string]Option{
		"empty levels":     WithLevels(),
		"invalid level":    WithLevels(logrus.Level(99)),
		"queue length":     WithQueueLength(0),
		"queue negative":   WithQueueLength(-1),
		"timeout":          WithTimeout(0),
		"timeout negative": WithTimeout(-time.Second),
	} {
		t.Run(name, func(t *testing.T) {
			cfg := defaultConfig()
			require.Error(t, option(&cfg))
		})
	}
}

func TestValidLevels(t *testing.T) {
	for _, level := range logrus.AllLevels {
		assert.True(t, isValidLevel(level))
	}
	assert.False(t, isValidLevel(logrus.Level(99)))
}
