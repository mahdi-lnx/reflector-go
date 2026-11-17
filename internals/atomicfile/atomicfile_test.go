package atomicfile

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func createTempFile(t *testing.T) (path string, fileMode os.FileMode) {
	t.Helper()

	path = filepath.Join(t.TempDir(), "hello.txt")
	fileMode = 0o777

	return path, fileMode
}

func writeTestData(f io.Writer, t *testing.T) string {
	t.Helper()

	data := "some data"

	n, err := f.Write([]byte(data))

	if err != nil {
		t.Fatalf("failed to write '%s' to the atomicfile: %v", data, err)
	}

	if nb := len(data); n != nb {
		t.Fatalf("atomicfile wrote only %d bytes of %d to the atomicfile", n, nb)
	}

	return data
}

func TestNewAtomicFile(t *testing.T) {
	path, fileMode := createTempFile(t)

	af, err := NewAtomicFile(path, fileMode)

	if err != nil {
		t.Fatalf("failed to create an atomicfile: %v", err)
	}

	defer af.Close()

	stat, err := af.file.Stat()

	if err != nil {
		t.Fatalf("could not complete the test: %v", err)
	}

	if stat.IsDir() {
		t.Fatal("expected the created temp file to be a file, not a dir")
	}

	if mode := stat.Mode(); mode != fileMode {
		t.Errorf("expected file mode to be %o, got %o", fileMode, mode)
	}

	if basename := stat.Name(); !strings.HasPrefix(basename, "mirrorlist.temp") {
		t.Errorf("expected the temp file basename to be 'mirrorlist.temp', got '%s'", basename)
	}
}

func TestAtomicFileWrite(t *testing.T) {
	path, fileMode := createTempFile(t)

	af, err := NewAtomicFile(path, fileMode)

	if err != nil {
		t.Fatalf("failed to create an atomicfile: %v", err)
	}

	defer af.Close()

	writeTestData(af, t)
}

func TestAtomicFileClose(t *testing.T) {
	path, fileMode := createTempFile(t)

	af, err := NewAtomicFile(path, fileMode)

	if err != nil {
		t.Fatalf("failed to create an atomicfile: %v", err)
	}

	err = af.Close()

	if err != nil {
		t.Fatalf("failed to close a non-closed atomicfile: %v", err)
	}

	err = af.Close()

	if err != nil {
		t.Fatalf("failed to close a closed atomicfile: %v", err)
	}

	_, err = os.Stat(af.tempPath)

	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected temp file %s to be removed, but stat returned: %v", af.tempPath, err)
	}
}

func TestAtomicFileCommit(t *testing.T) {

	commitAtomicFile := func(path string, fileMode os.FileMode, t *testing.T) {
		t.Helper()

		af, err := NewAtomicFile(path, fileMode)

		if err != nil {
			t.Fatalf("failed to create an atomicfile: %v", err)
		}

		defer af.Close()

		data := writeTestData(af, t)

		err = af.Commit()

		if err != nil {
			t.Fatalf("expected Commit to commit the changes, but got: %v", err)
		}

		err = af.Commit()

		if !errors.Is(err, ErrCommit) {
			t.Fatalf("expected Commit to return the ErrCommit sentinel, but got: %v", err)
		}

		_, err = os.Stat(af.tempPath)

		if !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("expected temp file %s to be removed, but stat returned: %v", af.tempPath, err)
		}

		content, err := os.ReadFile(path)

		if err != nil {
			t.Fatalf("failed to read the contents of %s: %v", path, err)
		}

		if content := string(content); content != data {
			t.Fatalf("expected %s content '%s' to be '%s'", path, content, data)
		}
	}

	t.Run("commit non-existing file", func(t *testing.T) {
		path, fileMode := createTempFile(t)
		commitAtomicFile(path, fileMode, t)
	})

	t.Run("commit existing file", func(t *testing.T) {
		path, fileMode := createTempFile(t)

		if err := os.WriteFile(path, []byte("some original data"), fileMode); err != nil {
			t.Fatalf("failed to create and write some original content to '%s': %v", path, err)
		}

		commitAtomicFile(path, fileMode, t)
	})
}
