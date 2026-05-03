// Package position provides helpers for geographic coordinates.
package position

import "math"

const (
	degreesInHalfCircle = 180
	earthRadius         = float64(6371)
)

// Haversine returns the great-circle distance in kilometers between two coordinates.
func Haversine(lonFrom float64, latFrom float64, lonTo float64, latTo float64) (distance float64) {
	deltaLat := (latTo - latFrom) * (math.Pi / degreesInHalfCircle)
	deltaLon := (lonTo - lonFrom) * (math.Pi / degreesInHalfCircle)

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(latFrom*(math.Pi/degreesInHalfCircle))*math.Cos(latTo*(math.Pi/degreesInHalfCircle))*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}
