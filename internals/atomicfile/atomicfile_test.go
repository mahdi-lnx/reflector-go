package atomicfile

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestNewAtomicFile(t *testing.T){
	path := fmt.Sprintf("%s/hello.txt", t.TempDir())
	var fileMode os.FileMode = 0o777

	af, err := NewAtomicFile(path, fileMode)

	if err != nil{
		t.Fatalf("failed to create an atomicfile: %s", err)
	}

	stat, err := af.file.Stat()

	if err != nil{
		t.Fatalf("could not complete the test: %s", err)
	}

	if stat.IsDir(){
		t.Fatal("expected the created temp file to be a file, not a dir")
	}

	if mode := stat.Mode(); mode != fileMode{
		t.Errorf("expected file mode to be %o, got %o", fileMode, mode)
	}

	if basename := stat.Name(); !strings.Contains(basename, "mirrorlist.temp"){
		t.Errorf("expected the temp file basename to be 'mirrorlist.temp', got '%s'", basename)
	}
}
