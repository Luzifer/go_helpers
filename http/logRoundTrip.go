package http

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"

	"github.com/pkg/errors"
)

type (
	// LogRoundTripper is a drop-in for the Transport of a http.Client
	// which then will log requests and responses to the given writer
	// (os.Stdout / os.Stderr / logrus writer / ...)
	LogRoundTripper struct {
		next   http.RoundTripper
		output io.Writer
	}
)

var _ http.RoundTripper = LogRoundTripper{}

// NewLogRoundTripper creates a new LogRoundTripper with the given
// next transport and output writer. If no next transport is given
// (next == nil) the http.DefaultTransport is used. If no writer
// is given the io.Discard writer is used.
func NewLogRoundTripper(next http.RoundTripper, out io.Writer) LogRoundTripper {
	if next == nil {
		next = http.DefaultTransport
	}

	if out == nil {
		// Makes no sense but ensures we don't fail writing to nil
		out = io.Discard
	}

	return LogRoundTripper{next, out}
}

// RoundTrip implements http.RoundTripper interface
func (l LogRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	dump, err := httputil.DumpRequest(req, true)
	if err != nil {
		return nil, errors.Wrap(err, "dumping request")
	}

	fmt.Fprintf(l.output, "---- 8< REQ >8 ----\n%s", dump)

	resp, err := l.next.RoundTrip(req)
	if err != nil {
		return resp, err
	}
	if dump, err = httputil.DumpResponse(resp, true); err != nil {
		return nil, errors.Wrap(err, "dumping response")
	}

	fmt.Fprintf(l.output, "---- 8< RES >8 ----\n%s---- 8< END >8 ----\n", dump)

	return resp, err
}
