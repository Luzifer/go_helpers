// Package yaml contains a method to convert a YAML into a JSON object
//
// Deprecated: use github.com/Luzifer/go_helpers/yaml instead.
package yaml

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"go.yaml.in/yaml/v3"
)

// ToJSON takes an io.Reader containing YAML source and converts it into
// a JSON representation of the YAML object.
func ToJSON(in io.Reader) (io.Reader, error) {
	var body interface{}

	if err := yaml.NewDecoder(in).Decode(&body); err != nil {
		return nil, fmt.Errorf("unmarshaling YAML: %s", err)
	}

	body = convert(body)

	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(body); err != nil {
		return nil, fmt.Errorf("marshaling JSON: %s", err)
	}

	return buf, nil
}

// Source: https://stackoverflow.com/a/40737676/1741281
func convert(i interface{}) interface{} {
	switch x := i.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[k.(string)] = convert(v)
		}
		return m2
	case []interface{}:
		for i, v := range x {
			x[i] = convert(v)
		}
	}
	return i
}
