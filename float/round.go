// Deprecated: use github.com/Luzifer/go_helpers/float instead.
package float

import "math"

// Round returns a float rounded according to "Round to nearest, ties away from zero" IEEE floaing point rounding rule
//
// Deprecated: Starting with Go1.10 this should be replaced with math.Round()
func Round(x float64) float64 {
	return math.Round(x)
}
