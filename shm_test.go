// License: GPLv3 Copyright: 2022, Kovid Goyal, <kovid at kovidgoyal.net>

package shm

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"reflect"
	"strings"
	"testing"
)

var _ = fmt.Print

func TestSHM(t *testing.T) {
	data := make([]byte, 13347)
	_, _ = rand.Read(data)
	mm, err := CreateTemp("test-kitty-shm-", uint64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(mm.Name(), "/") || strings.Contains(mm.Name()[1:], "/") {
		t.Fatalf("The SHM name %#v is not of the POSIX mandated form: a single leading slash", mm.Name())
	}

	copy(mm.Slice(), data)
	err = mm.Flush()
	if err != nil {
		t.Fatalf("Failed to msync() with error: %v", err)
	}
	err = mm.Close()
	if err != nil {
		t.Fatalf("Failed to close with error: %v", err)
	}

	g, err := Open(mm.Name(), uint64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	data2 := g.Slice()
	if !reflect.DeepEqual(data, data2) {
		t.Fatalf("Could not read back written data: Written data length: %d Read data length: %d", len(data), len(data2))
	}
	err = g.Close()
	if err != nil {
		t.Fatalf("Failed to close with error: %v", err)
	}
	err = g.Unlink()
	if err != nil {
		t.Fatalf("Failed to unlink with error: %v", err)
	}
	// names without the leading slash must refer to the same object
	g, err = Open(strings.TrimPrefix(mm.Name(), "/"), uint64(len(data)))
	if err == nil {
		t.Fatalf("Unlinking failed could re-open the SHM data. Data equal: %v Data length: %d", reflect.DeepEqual(g.Slice(), data), len(g.Slice()))
	}
	if mm.IsFileSystemBacked() {
		_, err = os.Stat(mm.FileSystemName())
		if !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("Unlinking %s did not work", mm.Name())
		}
	}
}

func TestSHMNames(t *testing.T) {
	for _, pattern := range []string{"test-shm-", "/test-shm-", "//test-shm-", "test-*-shm", "/test-*-shm"} {
		mm, err := CreateTemp(pattern, 32)
		if err != nil {
			t.Fatalf("Failed to create SHM for pattern %#v with error: %v", pattern, err)
		}
		name := mm.Name()
		mm.Close()
		if !strings.HasPrefix(name, "/") || strings.Contains(name[1:], "/") {
			t.Errorf("The SHM name %#v generated from the pattern %#v does not have a single leading slash", name, pattern)
		}
		if !strings.HasPrefix(name[1:], "test-") {
			t.Errorf("The SHM name %#v generated from the pattern %#v does not start with the pattern prefix", name, pattern)
		}
		if err = ShmUnlink(name); err != nil {
			t.Errorf("Failed to unlink %#v with error: %v", name, err)
		}
	}
	if _, err := CreateTemp("a/b-", 32); !errors.Is(err, ErrPatternHasSeparator) {
		t.Errorf("Expected ErrPatternHasSeparator for a pattern containing a separator, got: %v", err)
	}
}
