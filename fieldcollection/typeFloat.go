package fieldcollection

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
)

// Float64 tries to read key name as float64
func (f *FieldCollection) Float64(name string) (float64, error) {
	if f == nil || f.data == nil {
		return 0, errors.New("uninitialized field collection")
	}

	f.lock.RLock()
	defer f.lock.RUnlock()

	v, ok := f.data[name]
	if !ok {
		return 0, ErrValueNotSet
	}

	switch v := v.(type) {
	case int:
		return float64(v), nil

	case int16:
		return float64(v), nil

	case int32:
		return float64(v), nil

	case int64:
		return float64(v), nil

	case float64:
		return v, nil

	case string:
		pv, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, fmt.Errorf("parsing value: %w", err)
		}
		return pv, nil
	}

	return 0, ErrValueMismatch
}

// CanFloat64 tries to read key name as float64 and checks whether error is nil
func (f *FieldCollection) CanFloat64(name string) bool {
	_, err := f.Float64(name)
	return err == nil
}

// MustFloat64 is a wrapper around Float64 and panics if an error was returned
func (f *FieldCollection) MustFloat64(name string, defVal *float64) float64 {
	v, err := f.Float64(name)
	if err != nil {
		if defVal != nil {
			return *defVal
		}
		panic(err)
	}
	return v
}
