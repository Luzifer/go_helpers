// Package fieldcollection contains a map[string]any with accessor
// methods to derive them into different formats
//
// Deprecated: use github.com/Luzifer/go_helpers/fieldcollection instead.
package fieldcollection

import (
	"strings"
	"sync"

	"github.com/pkg/errors"
)

var (
	// ErrValueNotSet signalizes the value does not exist in the map
	ErrValueNotSet = errors.New("specified value not found")
	// ErrValueMismatch signalizes the value has a different data type
	ErrValueMismatch = errors.New("specified value has different format")
)

type (
	// FieldCollection holds a map with integrated locking and can
	// therefore used in multiple Go-routines concurrently
	FieldCollection struct {
		data map[string]any
		lock sync.RWMutex
	}
)

// NewFieldCollection creates a new FieldCollection with empty data store
func NewFieldCollection() *FieldCollection {
	return &FieldCollection{data: make(map[string]any)}
}

// FieldCollectionFromData is a wrapper around NewFieldCollection and SetFromData
//
//revive:disable-next-line:exported
func FieldCollectionFromData(data map[string]any) *FieldCollection {
	o := NewFieldCollection()
	o.SetFromData(data)
	return o
}

// Clone is a wrapper around n.SetFromData(o.Data())
func (f *FieldCollection) Clone() *FieldCollection {
	out := new(FieldCollection)
	out.SetFromData(f.Data())
	return out
}

// Data creates a map-copy of the data stored inside the FieldCollection
func (f *FieldCollection) Data() map[string]any {
	if f == nil {
		return nil
	}

	f.lock.RLock()
	defer f.lock.RUnlock()

	out := make(map[string]any)
	for k := range f.data {
		out[k] = f.data[k]
	}

	return out
}

// Expect takes a list of keys and returns an error with all non-found names
func (f *FieldCollection) Expect(keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	if f == nil || f.data == nil {
		return errors.New("uninitialized field collection")
	}

	f.lock.RLock()
	defer f.lock.RUnlock()

	var missing []string

	for _, k := range keys {
		if _, ok := f.data[k]; !ok {
			missing = append(missing, k)
		}
	}

	if len(missing) > 0 {
		return errors.Errorf("missing key(s) %s", strings.Join(missing, ", "))
	}

	return nil
}

// HasAll takes a list of keys and returns whether all of them exist inside the FieldCollection
func (f *FieldCollection) HasAll(keys ...string) bool {
	return f.Expect(keys...) == nil
}

// Get retrieves the value of a key as "any" type or returns an error
// in case the field is not set
func (f *FieldCollection) Get(name string) (any, error) {
	if f == nil || f.data == nil {
		return nil, errors.New("uninitialized field collection")
	}

	f.lock.RLock()
	defer f.lock.RUnlock()

	v, ok := f.data[name]
	if !ok {
		return nil, ErrValueNotSet
	}

	return v, nil
}

// Keys returns a list of all known keys
func (f *FieldCollection) Keys() (keys []string) {
	f.lock.RLock()
	defer f.lock.RUnlock()

	for k := range f.data {
		keys = append(keys, k)
	}

	return keys
}

// Set sets a single key to specified value
func (f *FieldCollection) Set(key string, value any) {
	f.lock.Lock()
	defer f.lock.Unlock()

	if f.data == nil {
		f.data = make(map[string]any)
	}

	f.data[key] = value
}

// SetFromData takes a map of data and copies all data into the FieldCollection
func (f *FieldCollection) SetFromData(data map[string]any) {
	f.lock.Lock()
	defer f.lock.Unlock()

	if f.data == nil {
		f.data = make(map[string]any)
	}

	for key, value := range data {
		f.data[key] = value
	}
}
