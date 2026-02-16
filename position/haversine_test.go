package position_test

import (
	"testing"

	. "github.com/Luzifer/go_helpers/position"
	"github.com/stretchr/testify/assert"
)

func TestHaversine(t *testing.T) {
	testCases := []struct {
		SourceLat float64
		SourceLon float64
		DestLat   float64
		DestLon   float64
		Distance  float64
	}{
		{50.066389, -5.714722, 58.643889, -3.070000, 968.8535441168448},
		{50.063995, -5.609464, 53.553027, 9.993782, 1137.894906816002},
		{53.553027, 9.993782, 53.554528, 9.991357, 0.23133816528015647},
		{50, 9, 51, 9, 111.19492664455873},
		{0, 9, 0, 10, 111.19492664455873},
		{1, 0, -1, 0, 222.38985328911747},
	}

	for i := range testCases {
		tc := testCases[i]
		assert.Equal(t, tc.Distance, Haversine(tc.SourceLon, tc.SourceLat, tc.DestLon, tc.DestLat))
	}
}
