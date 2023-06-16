package http

import (
	"net/http"
	"os"

	"github.com/pkg/errors"
)

type (
	// NoListFS wraps an http.FileSystem and ensures no directory
	// listings are printed out in order not to expose the contents
	// of the directory in a listing.
	NoListFS struct{ Next http.FileSystem }
)

var _ http.FileSystem = NoListFS{}

// Open wraps the Open of the inner http.FileSystem and returns an
// os.ErrNotExist in case a directory should be opened. This will
// result in a HTTP 404 when trying to open a directory causing the
// contents to be conceiled in an effective manner (zero exposure
// of known paths)
func (n NoListFS) Open(name string) (http.File, error) {
	f, err := n.Next.Open(name)
	if err != nil {
		return f, errors.Wrap(err, "opening file from inner fs")
	}

	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, errors.Wrap(err, "getting stat from opened file")
	}

	if info.IsDir() {
		f.Close()
		return nil, errors.Wrap(os.ErrNotExist, "refusing to open a directory")
	}

	return f, nil
}
