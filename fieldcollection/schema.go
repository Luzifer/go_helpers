package fieldcollection

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Luzifer/go_helpers/v2/str"
)

const (
	knownFields = "knownFields"
)

type (
	// SchemaField defines how a field is expected to be
	SchemaField struct {
		// Name of the field to validate
		Name string
		// If set to true the field must i.e. not be "" for a string field
		NonEmpty bool
		// The expected type of the field
		Type SchemaFieldType
	}

	// SchemaFieldType is a collection of known field types for which
	// can be checked
	SchemaFieldType uint64

	// ValidateOpt is a validation function to be executed during the
	// validation call
	ValidateOpt func(f, validateStore *FieldCollection) error
)

// Collection of known field types for which can be checked
const (
	SchemaFieldTypeAny SchemaFieldType = iota
	SchemaFieldTypeBool
	SchemaFieldTypeDuration
	SchemaFieldTypeFloat64
	SchemaFieldTypeInt64
	SchemaFieldTypeString
	SchemaFieldTypeStringSlice
)

// CanHaveField validates the type of the field if it exists and puts
// the field to the allow-list for MustHaveNoUnknowFields
func CanHaveField(field SchemaField) ValidateOpt {
	return func(f, validateStore *FieldCollection) error {
		validateStore.Set(knownFields, append(validateStore.MustStringSlice(knownFields, nil), field.Name))

		if !f.HasAll(field.Name) {
			// It is allowed to not exist, and if it does not we don't need
			// to type-check it
			return nil
		}

		return validateFieldType(f, field)
	}
}

// MustHaveField validates the type of the field and puts the field to
// the allow-list for MustHaveNoUnknowFields
func MustHaveField(field SchemaField) ValidateOpt {
	return func(f, validateStore *FieldCollection) error {
		validateStore.Set(knownFields, append(validateStore.MustStringSlice(knownFields, nil), field.Name))

		if !f.HasAll(field.Name) {
			// It must exist and does not
			return fmt.Errorf("field %s does not exist", field.Name)
		}

		return validateFieldType(f, field)
	}
}

// MustHaveNoUnknowFields validates no fields are present which are
// not previously allow-listed through CanHaveField or MustHaveField
// and therefore should be put as the last ValidateOpt
func MustHaveNoUnknowFields(f *FieldCollection, validateStore *FieldCollection) error {
	var unexpected []string

	for _, k := range f.Keys() {
		if !str.StringInSlice(k, validateStore.MustStringSlice(knownFields, nil)) {
			unexpected = append(unexpected, k)
		}
	}

	sort.Strings(unexpected)

	if len(unexpected) > 0 {
		return fmt.Errorf("found unexpected fields: %s", strings.Join(unexpected, ", "))
	}

	return nil
}

// ValidateSchema can be used to validate the contents of the
// FieldCollection by passing in field definitions which may be there
// or must be there and to check whether there are no surplus fields
func (f *FieldCollection) ValidateSchema(opts ...ValidateOpt) error {
	validateStore := NewFieldCollection()
	validateStore.Set(knownFields, []string{})

	for _, opt := range opts {
		if err := opt(f, validateStore); err != nil {
			return err
		}
	}

	return nil
}

//nolint:gocognit,gocyclo // These are quite simple checks
func validateFieldType(f *FieldCollection, field SchemaField) (err error) {
	switch field.Type {
	case SchemaFieldTypeAny:
		v, err := f.Get(field.Name)
		if err != nil {
			return fmt.Errorf("getting field %s: %w", field.Name, err)
		}

		if field.NonEmpty && v == nil {
			return fmt.Errorf("field %s is empty", field.Name)
		}

	case SchemaFieldTypeBool:
		if !f.CanBool(field.Name) {
			return fmt.Errorf("field %s is not of type bool", field.Name)
		}

	case SchemaFieldTypeDuration:
		v, err := f.Duration(field.Name)
		if err != nil {
			return fmt.Errorf("field %s is not of type time.Duration: %w", field.Name, err)
		}

		if field.NonEmpty && v == 0 {
			return fmt.Errorf("field %s is empty", field.Name)
		}

	case SchemaFieldTypeFloat64:
		v, err := f.Float64(field.Name)
		if err != nil {
			return fmt.Errorf("field %s is not of type float64: %w", field.Name, err)
		}

		if field.NonEmpty && v == 0 {
			return fmt.Errorf("field %s is empty", field.Name)
		}

	case SchemaFieldTypeInt64:
		v, err := f.Int64(field.Name)
		if err != nil {
			return fmt.Errorf("field %s is not of type int64: %w", field.Name, err)
		}

		if field.NonEmpty && v == 0 {
			return fmt.Errorf("field %s is empty", field.Name)
		}

	case SchemaFieldTypeString:
		v, err := f.String(field.Name)
		if err != nil {
			return fmt.Errorf("field %s is not of type string: %w", field.Name, err)
		}

		if field.NonEmpty && v == "" {
			return fmt.Errorf("field %s is empty", field.Name)
		}

	case SchemaFieldTypeStringSlice:
		v, err := f.StringSlice(field.Name)
		if err != nil {
			return fmt.Errorf("field %s is not of type []string: %w", field.Name, err)
		}

		if field.NonEmpty && len(v) == 0 {
			return fmt.Errorf("field %s is empty", field.Name)
		}

	default:
		return fmt.Errorf("unknown field type specified")
	}

	return nil
}
