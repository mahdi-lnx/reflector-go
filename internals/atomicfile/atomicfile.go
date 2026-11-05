// atomicfile provides atomic file replacement operations
package atomicfile

import (
	"errors"
	"os"
	"path/filepath"
)

// AtomicFile defines a temp file that can atomically replace
// an original file when commited
type AtomicFile struct {
	// originalPath is the original file path
	originalPath string

	// tempPath is the temp file path
	tempPath string

	// file is the underlying temp file
	file *os.File
}

// Write writes the given data into the temp file
func (af *AtomicFile) Write(b []byte) (int, error) {
	return af.file.Write(b)
}

// Commit synchronizes and closes the temp file.
// It then atomically replaces the destination file with it
//
// If synchronization or closing fails, [AtomicFile.originalPath]
// is left unchanged
//
// [AtomicFile.Close] should always be called for cleanup
// regardless of wether Commit succeeded or not
func (af *AtomicFile) Commit() error {
	if af.file == nil {
		return errors.New("the atomic file is closed")
	}

	err := af.file.Sync()

	if err != nil {
		return err
	}

	err = af.file.Close()

	af.file = nil // go treats the file as closed regardless of err

	// do not rename the temp file if Close failed even when Sync succeeded

	if err != nil {
		return err
	}

	return os.Rename(af.tempPath, af.originalPath)
}

// Close closes and removes the temp file without commiting it
//
// Close returns nil if the the temp file is closed and it does not exist
func (af *AtomicFile) Close() error {

	if af.file == nil {
		// the file has already been closed

		// we try to remove the temp file if it still exists

		err := os.Remove(af.tempPath)

		if errors.Is(err, os.ErrNotExist) {
			// the temp does not exists
			return nil
		}
		return err
	}

	closeErr := af.file.Close()

	af.file = nil // go treats the file as closed regardless of err

	// Note that on POSIX complient systems removing a file is safe even if Close failed

	rmErr := os.Remove(af.tempPath)

	return errors.Join(closeErr, rmErr)
}

// NewAtomicFile creates a new [AtomicFile]
func NewAtomicFile(path string, mode os.FileMode) (*AtomicFile, error) {

	f, err := os.CreateTemp(filepath.Dir(path), "mirrorlist.temp")

	if err != nil {
		return nil, err
	}

	err = f.Chmod(mode)

	if err != nil {
		return nil, err
	}

	return &AtomicFile{
		originalPath: path,
		tempPath:     f.Name(),
		file:         f,
	}, nil
}
