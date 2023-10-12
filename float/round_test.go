package float_test

import (
	"math"
	"testing"

	. "github.com/Luzifer/go_helpers/v2/float"
	"github.com/stretchr/testify/assert"
)

func TestRound(t *testing.T) {
	t.Run("should match the example table of IEEE 754 rules", func(t *testing.T) {
		assert.Equal(t, 12.0, Round(11.5))
		assert.Equal(t, 13.0, Round(12.5))
		assert.Equal(t, -12.0, Round(-11.5))
		assert.Equal(t, -13.0, Round(-12.5))
	})

	t.Run("should have correct rounding for numbers near 0.5", func(t *testing.T) {
		assert.Equal(t, 0.0, Round(0.499999999997))
		assert.Equal(t, 0.0, Round(-0.499999999997))
	})

	t.Run("should be able to handle +/-Inf", func(t *testing.T) {
		assert.Equal(t, math.Inf(1), math.Inf(1))
		assert.Equal(t, math.Inf(-1), math.Inf(-1))
	})

	t.Run("should be able to handle NaN", func(t *testing.T) {
		assert.True(t, math.IsNaN(Round(math.NaN())))
	})
}
