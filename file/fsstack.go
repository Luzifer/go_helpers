package file

import (
	"io"
	"io/fs"

	"github.com/pkg/errors"
)

// FSStack represents layers of fs.FS to open a file from. The first
// layer (starting at index 0, going up) responding other than
// fs.ErrNotExist will determine the response of this stack.
type FSStack []fs.FS

var _ fs.FS = FSStack{}

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

			return nil, errors.Wrapf(err, "opening file from layer %d", i)
		}

		return file, nil
	}

	return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
}

// ReadFile is a convenice wrapper around Open and returns the content
// of the file if any is available.
func (f FSStack) ReadFile(name string) ([]byte, error) {
	file, err := f.Open(name)
	if err != nil {
		return nil, errors.Wrap(err, "opening file")
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	return content, errors.Wrap(err, "reading content")
}
