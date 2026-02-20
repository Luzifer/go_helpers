package str

import "slices"

// AppendIfMissing adds a string to a slice when it's not present yet
func AppendIfMissing(slice []string, s string) []string {
	if slices.Contains(slice, s) {
		return slice
	}
	return append(slice, s)
}

// StringInSlice checks for the existence of a string in the slice
//
// Deprecated: This is a wrapper around slices.Contains()
func StringInSlice(a string, list []string) bool {
	return slices.Contains(list, a)
}
