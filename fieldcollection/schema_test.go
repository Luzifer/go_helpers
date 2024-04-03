package fieldcollection

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

//nolint:funlen
func TestSchemaValidation(t *testing.T) {
	fc := FieldCollectionFromData(map[string]any{
		"anyZero":         nil,
		"bool":            true,
		"duration":        time.Second,
		"durationZero":    time.Duration(0),
		"int64":           int64(12),
		"int64Zero":       int64(0),
		"string":          "ohai",
		"stringZero":      "",
		"stringSlice":     []string{"ohai"},
		"stringSliceZero": []string{},
		"stringSliceNil":  nil,
	})

	// No validations
	assert.NoError(t, fc.ValidateSchema())

	// Non-existing field
	assert.ErrorContains(t, fc.ValidateSchema(
		MustHaveField(SchemaField{Name: "foo"}),
	), "field foo does not exist")

	// Non-existing field but can
	assert.NoError(t, fc.ValidateSchema(
		CanHaveField(SchemaField{Name: "foo"}),
	))

	// No unexpected fields (none given)
	assert.ErrorContains(t, fc.ValidateSchema(
		MustHaveNoUnknowFields,
	), "found unexpected fields: anyZero, bool, duration, durationZero, int64, int64Zero, string, stringSlice, stringSliceNil, stringSliceZero, stringZero")

	// No unexpected fields (all given)
	assert.NoError(t, fc.ValidateSchema(
		CanHaveField(SchemaField{Name: "anyZero"}),
		CanHaveField(SchemaField{Name: "bool"}),
		CanHaveField(SchemaField{Name: "duration"}),
		CanHaveField(SchemaField{Name: "durationZero"}),
		CanHaveField(SchemaField{Name: "int64"}),
		CanHaveField(SchemaField{Name: "int64Zero"}),
		CanHaveField(SchemaField{Name: "string"}),
		CanHaveField(SchemaField{Name: "stringSlice"}),
		CanHaveField(SchemaField{Name: "stringSliceNil"}),
		CanHaveField(SchemaField{Name: "stringSliceZero"}),
		CanHaveField(SchemaField{Name: "stringZero"}),
		MustHaveNoUnknowFields,
	))

	// Field must exist in any type and not be zero
	assert.NoError(t, fc.ValidateSchema(
		MustHaveField(SchemaField{Name: "string", NonEmpty: true}),
	))

	// Field must exist in any type and not be zero but is zero
	assert.ErrorContains(t, fc.ValidateSchema(
		MustHaveField(SchemaField{Name: "anyZero", NonEmpty: true}),
	), "field anyZero is empty")

	// Fields must exist and be of correct type
	assert.NoError(t, fc.ValidateSchema(
		MustHaveField(SchemaField{Name: "bool", Type: SchemaFieldTypeBool}),
		MustHaveField(SchemaField{Name: "duration", Type: SchemaFieldTypeDuration}),
		MustHaveField(SchemaField{Name: "int64", Type: SchemaFieldTypeInt64}),
		MustHaveField(SchemaField{Name: "string", Type: SchemaFieldTypeString}),
		MustHaveField(SchemaField{Name: "stringSlice", Type: SchemaFieldTypeStringSlice}),
	))
	assert.Error(t, fc.ValidateSchema(
		MustHaveField(SchemaField{Name: "bool", Type: SchemaFieldTypeDuration}),
	))
	assert.Error(t, fc.ValidateSchema(
		MustHaveField(SchemaField{Name: "duration", Type: SchemaFieldTypeBool}),
	))
	assert.Error(t, fc.ValidateSchema(
		MustHaveField(SchemaField{Name: "int64", Type: SchemaFieldTypeStringSlice}),
	))
	assert.Error(t, fc.ValidateSchema(
		MustHaveField(SchemaField{Name: "string", Type: SchemaFieldTypeInt64}),
	))
	assert.Error(t, fc.ValidateSchema(
		MustHaveField(SchemaField{Name: "stringSlice", Type: SchemaFieldTypeString}),
	))

	// Fields must not be zero
	assert.ErrorContains(t, fc.ValidateSchema(
		MustHaveField(SchemaField{Name: "duration", NonEmpty: true, Type: SchemaFieldTypeDuration}),
		MustHaveField(SchemaField{Name: "durationZero", NonEmpty: true, Type: SchemaFieldTypeDuration}),
	), "field durationZero is empty")
	assert.ErrorContains(t, fc.ValidateSchema(
		MustHaveField(SchemaField{Name: "int64", NonEmpty: true, Type: SchemaFieldTypeInt64}),
		MustHaveField(SchemaField{Name: "int64Zero", NonEmpty: true, Type: SchemaFieldTypeInt64}),
	), "field int64Zero is empty")
	assert.ErrorContains(t, fc.ValidateSchema(
		MustHaveField(SchemaField{Name: "string", NonEmpty: true, Type: SchemaFieldTypeString}),
		MustHaveField(SchemaField{Name: "stringZero", NonEmpty: true, Type: SchemaFieldTypeString}),
	), "field stringZero is empty")
	assert.ErrorContains(t, fc.ValidateSchema(
		MustHaveField(SchemaField{Name: "stringSlice", NonEmpty: true, Type: SchemaFieldTypeStringSlice}),
		MustHaveField(SchemaField{Name: "stringSliceZero", NonEmpty: true, Type: SchemaFieldTypeStringSlice}),
	), "field stringSliceZero is empty")

	// Invalid field type
	assert.ErrorContains(t, fc.ValidateSchema(
		MustHaveField(SchemaField{Name: "stringSlice", NonEmpty: true, Type: 99999}),
	), "unknown field type specified")
}
