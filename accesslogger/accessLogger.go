// Package accesslogger provides helpers to record HTTP response status and size.
package accesslogger

import (
	"fmt"
	"net/http"
	"strconv"
)

// AccessLogResponseWriter wraps an http.ResponseWriter and records response metadata.
type AccessLogResponseWriter struct {
	StatusCode int
	Size       int

	http.ResponseWriter
}

// New wraps res in an AccessLogResponseWriter initialized with HTTP status 200.
func New(res http.ResponseWriter) *AccessLogResponseWriter {
	return &AccessLogResponseWriter{
		StatusCode:     http.StatusOK,
		Size:           0,
		ResponseWriter: res,
	}
}

// HTTPResponseType returns the status class in access-log form, such as "2xx".
func (a *AccessLogResponseWriter) HTTPResponseType() string {
	return fmt.Sprintf("%cxx", strconv.FormatInt(int64(a.StatusCode), 10)[0])
}

func (a *AccessLogResponseWriter) Write(out []byte) (int, error) {
	s, err := a.ResponseWriter.Write(out)
	a.Size += s
	return s, err //nolint:wrapcheck // we're just a thin wrapper, don't taint the inner error
}

// WriteHeader records code before forwarding it to the wrapped ResponseWriter.
func (a *AccessLogResponseWriter) WriteHeader(code int) {
	a.StatusCode = code
	a.ResponseWriter.WriteHeader(code)
}
