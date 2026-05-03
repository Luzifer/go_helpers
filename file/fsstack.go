// Package file contains helpers for filesystem access and file watching.
package file

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
)

// FSStack represents layers of fs.FS to open a file from. The first
// layer (starting at index 0, going up) responding other than
// fs.ErrNotExist will determine the response of this stack.
type FSStack []fs.FS

var _ fs.FS = (FSStack)(nil)

// Open iterates the FSStack starting at index 0, going up and returns
// the first non fs.ErrNotExist response. If all layers responds with
// fs.ErrNotExist a fs.PathError wrapping fs.ErrNotExist is returned.
func (f FSStack) Open(name string) (fs.File, error) {
	for i, fse := range f {
		file, err := fse.Open(name)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}

			return nil, fmt.Errorf("opening file from layer %d: %w", i, err)
		}

		return file, nil
	}

	return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
}

// ReadFile is a convenice wrapper around Open and returns the content
// of the file if any is available.
func (f FSStack) ReadFile(name string) (content []byte, err error) {
	file, err := f.Open(name)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("closing file: %w", closeErr))
		}
	}()

	content, err = io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("reading content: %w", err)
	}

	return content, nil
}
