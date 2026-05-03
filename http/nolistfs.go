package http //revive:disable-line:package-naming // kept for historical reasons

import (
	"errors"
	"fmt"
	"net/http"
	"os"
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
		return f, fmt.Errorf("opening file from inner fs: %w", err)
	}

	info, err := f.Stat()
	if err != nil {
		if closeErr := f.Close(); closeErr != nil {
			return nil, errors.Join(
				fmt.Errorf("getting stat from opened file: %w", err),
				fmt.Errorf("closing file: %w", closeErr),
			)
		}

		return nil, fmt.Errorf("getting stat from opened file: %w", err)
	}

	if info.IsDir() {
		if closeErr := f.Close(); closeErr != nil {
			return nil, errors.Join(
				fmt.Errorf("refusing to open a directory: %w", os.ErrNotExist),
				fmt.Errorf("closing file: %w", closeErr),
			)
		}

		return nil, fmt.Errorf("refusing to open a directory: %w", os.ErrNotExist)
	}

	return f, nil
}
