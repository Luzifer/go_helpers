package fieldcollection

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
)

// Int64 tries to read key name as int64
func (f *FieldCollection) Int64(name string) (int64, error) {
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
		return int64(v), nil

	case int16:
		return int64(v), nil

	case int32:
		return int64(v), nil

	case int64:
		return v, nil

	case string:
		pv, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("parsing value: %w", err)
		}
		return pv, nil
	}

	return 0, ErrValueMismatch
}

// CanInt64 tries to read key name as int64 and checks whether error is nil
func (f *FieldCollection) CanInt64(name string) bool {
	_, err := f.Int64(name)
	return err == nil
}

// MustInt64 is a wrapper around Int64 and panics if an error was returned
func (f *FieldCollection) MustInt64(name string, defVal *int64) int64 {
	v, err := f.Int64(name)
	if err != nil {
		if defVal != nil {
			return *defVal
		}
		panic(err)
	}
	return v
}
