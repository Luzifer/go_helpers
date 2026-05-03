// Package duration provides helpers to format time durations for humans.
package duration

import (
	"bytes"
	"fmt"
	"math"
	"strings"
	"text/template"
	"time"

	"github.com/leekchan/gtf"
)

const (
	defaultDurationFormat = `{{if gt .Years 0}}{{.Years}} year{{.Years|pluralize "s"}}, {{end}}` +
		`{{if gt .Days 0}}{{.Days}} day{{.Days|pluralize "s"}}, {{end}}` +
		`{{if gt .Hours 0}}{{.Hours}} hour{{.Hours|pluralize "s"}}, {{end}}` +
		`{{if gt .Minutes 0}}{{.Minutes}} minute{{.Minutes|pluralize "s"}}, {{end}}` +
		`{{if gt .Seconds 0}}{{.Seconds}} second{{.Seconds|pluralize "s"}}{{end}}`

	oneDay  = 24 * time.Hour
	oneYear = 365 * oneDay
)

// CustomHumanizeDuration formats in using tpl and duration fields.
func CustomHumanizeDuration(in time.Duration, tpl string) (string, error) {
	result := struct{ Years, Days, Hours, Minutes, Seconds int64 }{}

	in = time.Duration(math.Abs(float64(in)))

	for in > 0 {
		switch {
		case in > oneYear:
			result.Years = int64(in / oneYear)
			in -= time.Duration(result.Years) * oneYear
		case in > oneDay:
			result.Days = int64(in / oneDay)
			in -= time.Duration(result.Days) * oneDay
		case in > time.Hour:
			result.Hours = int64(in / time.Hour)
			in -= time.Duration(result.Hours) * time.Hour
		case in > time.Minute:
			result.Minutes = int64(in / time.Minute)
			in -= time.Duration(result.Minutes) * time.Minute
		default:
			result.Seconds = int64(in / time.Second)
			in = 0
		}
	}

	tmpl, err := template.New("timeformat").Funcs(gtf.GtfFuncMap).Parse(tpl)
	if err != nil {
		return "", fmt.Errorf("parsing template: %w", err)
	}

	buf := bytes.NewBuffer(nil)
	if err = tmpl.Execute(buf, result); err != nil {
		return "", fmt.Errorf("executing template: %w", err)
	}

	return buf.String(), nil
}

// HumanizeDuration formats in using the default human-readable duration template.
func HumanizeDuration(in time.Duration) string {
	f, err := CustomHumanizeDuration(in, defaultDurationFormat)
	if err != nil {
		panic(err)
	}

	return strings.Trim(f, " ,")
}
