package fieldcollection

import (
	"fmt"

	"github.com/pkg/errors"
)

// CanString tries to read key name as string and checks whether error is nil
func (f *FieldCollection) CanString(name string) bool {
	_, err := f.String(name)
	return err == nil
}

// MustString is a wrapper around String and panics if an error was returned
func (f *FieldCollection) MustString(name string, defVal *string) string {
	v, err := f.String(name)
	if err != nil {
		if defVal != nil {
			return *defVal
		}
		panic(err)
	}
	return v
}

// String tries to read key name as string
func (f *FieldCollection) String(name string) (string, error) {
	if f == nil || f.data == nil {
		return "", errors.New("uninitialized field collection")
	}

	f.lock.RLock()
	defer f.lock.RUnlock()

	v, ok := f.data[name]
	if !ok {
		return "", ErrValueNotSet
	}

	if sv, ok := v.(string); ok {
		return sv, nil
	}

	if iv, ok := v.(fmt.Stringer); ok {
		return iv.String(), nil
	}

	return "", ErrValueMismatch
}
