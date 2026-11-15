package atomicfile

import (
	"fmt"
	"os"
	"testing"
)

func TestNewAtomicFile(t *testing.T){

	t.Run("create an atomicfile for a existing file", func (t *testing.T){

	})

	t.Run("create an atomicfile for a non-existing file", func(t *testing.T) {
		path := fmt.Sprintf("%s/hello.txt", t.TempDir())
		var fileMode os.FileMode = 0x777

		f, err := NewAtomicFile(path, fileMode)

		if err != nil{
			t.Fatalf("could not complete the test: %s", err)
		}

		stat, err := f.file.Stat()

		if err != nil{
			t.Fatalf("could not complete the test: %s", err)
		}

		if stat.IsDir(){
			t.Fatal("expected the created temp file to be a file, not a dir")
		}

		if mode := stat.Mode(); mode != fileMode{
			t.Errorf("expected file mode to be %d, got %d", fileMode, mode)
		}

		if basename := stat.Name(); basename != "mirrorlist.temp"{
			t.Errorf("expected the temp file basename to be 'mirrorlist.temp', got '%s'", basename)
		}
	})
}
