package fieldcollection

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
)

// CanDuration tries to read key name as time.Duration and checks whether error is nil
func (f *FieldCollection) CanDuration(name string) bool {
	_, err := f.Duration(name)
	return err == nil
}

// Duration tries to read key name as time.Duration
func (f *FieldCollection) Duration(name string) (time.Duration, error) {
	if f == nil || f.data == nil {
		return 0, errors.New("uninitialized field collection")
	}

	switch {
	case !f.HasAll(name):
		return 0, ErrValueNotSet

	case f.CanInt64(name):
		return time.Duration(f.MustInt64(name, nil)), nil

	case f.CanString(name):
		v, err := time.ParseDuration(f.MustString(name, nil))
		if err != nil {
			return 0, fmt.Errorf("parsing value: %w", err)
		}
		return v, nil

	default:
		return 0, ErrValueMismatch
	}
}

// MustDuration is a wrapper around Duration and panics if an error was returned
func (f *FieldCollection) MustDuration(name string, defVal *time.Duration) time.Duration {
	v, err := f.Duration(name)
	if err != nil {
		if defVal != nil {
			return *defVal
		}
		panic(err)
	}
	return v
}
