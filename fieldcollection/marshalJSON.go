package fieldcollection

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

// Implement JSON marshalling to plain underlying map[string]any

// MarshalJSON implements json.Marshaller interface
func (f *FieldCollection) MarshalJSON() ([]byte, error) {
	if f == nil || f.data == nil {
		return []byte("{}"), nil
	}

	f.lock.RLock()
	defer f.lock.RUnlock()

	data, err := json.Marshal(f.data)
	if err != nil {
		return nil, fmt.Errorf("marshalling to JSON: %w", err)
	}

	return data, nil
}

// UnmarshalJSON implements json.Unmarshaller interface
func (f *FieldCollection) UnmarshalJSON(raw []byte) error {
	data := make(map[string]any)
	if err := json.Unmarshal(raw, &data); err != nil {
		return errors.Wrap(err, "unmarshalling from JSON")
	}

	f.SetFromData(data)
	return nil
}
