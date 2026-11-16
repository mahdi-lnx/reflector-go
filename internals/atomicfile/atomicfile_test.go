package atomicfile

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
)

func createTempFile(t *testing.T) (path string, fileMode os.FileMode){
	path = fmt.Sprintf("%s/hello.txt", t.TempDir())
	fileMode = 0o777

	return path, fileMode
}

func TestNewAtomicFile(t *testing.T){
	path, fileMode := createTempFile(t)

	af, err := NewAtomicFile(path, fileMode)

	if err != nil{
		t.Fatalf("failed to create an atomicfile: %v", err)
	}

	stat, err := af.file.Stat()

	if err != nil{
		t.Fatalf("could not complete the test: %v", err)
	}

	if stat.IsDir(){
		t.Fatal("expected the created temp file to be a file, not a dir")
	}

	if mode := stat.Mode(); mode != fileMode{
		t.Errorf("expected file mode to be %o, got %o", fileMode, mode)
	}

	if basename := stat.Name(); !strings.HasPrefix(basename, "mirrorlist.temp"){
		t.Errorf("expected the temp file basename to be 'mirrorlist.temp', got '%s'", basename)
	}
}

func TestAtomicFileWrite(t *testing.T){
	path, fileMode := createTempFile(t)

	af, err := NewAtomicFile(path, fileMode)

	if err != nil{
		t.Fatalf("failed to create an atomicfile: %v", err)
	}

	data := "some data"

	n, err := af.Write([]byte(data))

	if err != nil {
		t.Fatalf("failed to write '%s' to the atomicfile: %v", data, err)
	}

	if nb := len(data); n != nb {
		t.Fatalf("atomicfile wrote only %d bytes of %d to the atomicfile", n, nb)
	}
}

func TestAtomicFileClose(t *testing.T){
	path, fileMode := createTempFile(t)

	af, err := NewAtomicFile(path, fileMode)

	if err != nil{
		t.Fatalf("failed to create an atomicfile: %v", err)
	}

	err = af.Close()

	if err != nil{
		t.Fatalf("failed to close a non-closed atomicfile: %v", err)
	}

	err = af.Close()

	if err != nil{
		t.Fatalf("failed to close a closed atomicfile: %v", err)
	}

	_, err = os.Stat(af.tempPath)

	if !errors.Is(err, os.ErrNotExist){
		t.Fatalf("expected temp file %s to be removed but stat returned: %v", af.tempPath, err)
	}
}

func TestAtomicFileCommit(t *testing.T){

}

