package float

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRound(t *testing.T) {
	t.Run("should match the example table of IEEE 754 rules", func(t *testing.T) {
		assert.InDelta(t, 12.0, Round(11.5), 0)
		assert.InDelta(t, 13.0, Round(12.5), 0)
		assert.InDelta(t, -12.0, Round(-11.5), 0)
		assert.InDelta(t, -13.0, Round(-12.5), 0)
	})

	t.Run("should have correct rounding for numbers near 0.5", func(t *testing.T) {
		assert.InDelta(t, 0.0, Round(0.499999999997), 0)
		assert.InDelta(t, 0.0, Round(-0.499999999997), 0)
	})

	t.Run("should be able to handle +/-Inf", func(t *testing.T) {
		assert.InDelta(t, math.Inf(1), Round(math.Inf(1)), 0)
		assert.InDelta(t, math.Inf(-1), Round(math.Inf(-1)), 0)
	})

	t.Run("should be able to handle NaN", func(t *testing.T) {
		assert.True(t, math.IsNaN(Round(math.NaN())))
	})
}
