package env_test

import (
	"sort"
	"testing"

	. "github.com/Luzifer/go_helpers/env"
	"github.com/stretchr/testify/assert"
)

func TestListToMap(t *testing.T) {
	var (
		list = []string{
			"FIRST_KEY=firstvalue",
			"SECOND_KEY=secondvalue",
			"WEIRD=",
			"NOVALUE",
			"",
		}
		emap = map[string]string{
			"FIRST_KEY":  "firstvalue",
			"SECOND_KEY": "secondvalue",
			"WEIRD":      "",
			"NOVALUE":    "",
		}
	)

	assert.Equal(t, emap, ListToMap(list))
}

func TestMapToList(t *testing.T) {
	var (
		list = []string{
			"FIRST_KEY=firstvalue",
			"SECOND_KEY=secondvalue",
			"WEIRD=",
		}
		emap = map[string]string{
			"FIRST_KEY":  "firstvalue",
			"SECOND_KEY": "secondvalue",
			"WEIRD":      "",
		}
	)

	l := MapToList(emap)
	sort.Strings(l) // Workaround: The test needs the elements to be in same order
	assert.Equal(t, list, l)
}
