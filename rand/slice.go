// Package rand provides helpers for cryptographic random choices.
package rand

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// EntryFromSlice returns one cryptographically random entry from s.
func EntryFromSlice[T any](s []T) (v T, err error) {
	if len(s) == 0 {
		return v, fmt.Errorf("cannot choose from zero-length slice")
	}

	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(s))))
	if err != nil {
		return v, fmt.Errorf("getting random entry: %w", err)
	}

	return s[n.Int64()], nil
}
