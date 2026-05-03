package fieldcollection

import "fmt"

// Implement YAML marshalling to plain underlying map[string]any

// MarshalYAML implements yaml.Marshaller interface
func (f *FieldCollection) MarshalYAML() (any, error) {
	return f.Data(), nil
}

// UnmarshalYAML implements yaml.Unmarshaller interface
func (f *FieldCollection) UnmarshalYAML(unmarshal func(any) error) error {
	data := make(map[string]any)
	if err := unmarshal(&data); err != nil {
		return fmt.Errorf("unmarshalling from YAML: %w", err)
	}

	f.SetFromData(data)
	return nil
}
