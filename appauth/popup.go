package appauth

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
)

var (
	//go:embed closePage.html.gotmpl
	closePage string
	//go:embed postMessagePage.html.gotmpl
	postMessagePage string
)

func renderTemplate(w io.Writer, tpl string, fields any) error {
	t, err := template.New("").Funcs(template.FuncMap{
		"toJSON": func(data any) (template.JS, error) {
			render, err := json.Marshal(data)
			if err != nil {
				return "", fmt.Errorf("marshalling JSON: %w", err)
			}
			return template.JS(render), nil //#nosec:G203 // JSON is program made
		},
	}).Parse(tpl)
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	if err = t.Execute(w, fields); err != nil {
		return fmt.Errorf("rendering template: %w", err)
	}

	return nil
}

func writeClosePage(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Avoid caching a token/error page
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")

	_ = renderTemplate(w, closePage, map[string]string{"message": msg})
}

func writePostMessageAndClose(w http.ResponseWriter, targetOrigin string, payload map[string]any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")

	_ = renderTemplate(w, postMessagePage, map[string]any{
		"data":   payload,
		"origin": targetOrigin,
	})
}
