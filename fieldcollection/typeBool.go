package fieldcollection

import (
	"errors"
	"fmt"
	"strconv"
)

// Bool tries to read key name as bool
func (f *FieldCollection) Bool(name string) (bool, error) {
	if f == nil || f.data == nil {
		return false, errors.New("uninitialized field collection")
	}

	f.lock.RLock()
	defer f.lock.RUnlock()

	v, ok := f.data[name]
	if !ok {
		return false, ErrValueNotSet
	}

	switch v := v.(type) {
	case bool:
		return v, nil
	case string:
		bv, err := strconv.ParseBool(v)
		if err != nil {
			return false, fmt.Errorf("parsing string to bool: %w", err)
		}

		return bv, nil
	}

	return false, ErrValueMismatch
}

// CanBool tries to read key name as bool and checks whether error is nil
func (f *FieldCollection) CanBool(name string) bool {
	_, err := f.Bool(name)
	return err == nil
}

// MustBool is a wrapper around Bool and panics if an error was returned
func (f *FieldCollection) MustBool(name string, defVal *bool) bool {
	v, err := f.Bool(name)
	if err != nil {
		if defVal != nil {
			return *defVal
		}
		panic(err)
	}
	return v
}
