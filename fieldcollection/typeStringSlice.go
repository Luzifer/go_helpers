package fieldcollection

import "github.com/pkg/errors"

// CanStringSlice tries to read key name as []string and checks whether error is nil
func (f *FieldCollection) CanStringSlice(name string) bool {
	_, err := f.StringSlice(name)
	return err == nil
}

// MustStringSlice is a wrapper around StringSlice and panics if an error was returned
func (f *FieldCollection) MustStringSlice(name string, defVal *[]string) []string {
	v, err := f.StringSlice(name)
	if err != nil {
		if defVal != nil {
			return *defVal
		}
		panic(err)
	}
	return v
}

// StringSlice tries to read key name as []string
func (f *FieldCollection) StringSlice(name string) ([]string, error) {
	if f == nil || f.data == nil {
		return nil, errors.New("uninitialized field collection")
	}

	f.lock.RLock()
	defer f.lock.RUnlock()

	v, ok := f.data[name]
	if !ok {
		return nil, ErrValueNotSet
	}

	switch v := v.(type) {
	case []string:
		return v, nil

	case []any:
		var out []string

		for _, iv := range v {
			sv, ok := iv.(string)
			if !ok {
				return nil, errors.New("value in slice was not string")
			}
			out = append(out, sv)
		}

		return out, nil
	}

	return nil, ErrValueMismatch
}
