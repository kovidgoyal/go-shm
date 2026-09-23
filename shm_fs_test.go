// License: GPLv3 Copyright: 2022, Kovid Goyal, <kovid at kovidgoyal.net>
//go:build linux || netbsd || openbsd || dragonfly || illumos

package shm

import (
	"path/filepath"
	"testing"
)

func TestOpenBSDShmPath(t *testing.T) {
	// The digests are of the names with their leading slash included, as
	// produced by makeshmpath() in OpenBSD libc. Independently verifiable
	// with: printf '/foo' | sha256sum
	for name, digest := range map[string]string{
		"/foo": "6f64c6e6261f492ac220b0a4cd9a14c6373181b92a4a8040c1fcde5db31ffc94",
		"foo":  "6f64c6e6261f492ac220b0a4cd9a14c6373181b92a4a8040c1fcde5db31ffc94",
		"/":    "8a5edab282632443219e051e4ade2d1d5bbc671c781051bf1437897cbdfea0f1",
	} {
		expected := filepath.Join(SHM_DIR, digest+".shm")
		if actual := openbsd_shm_path(name); actual != expected {
			t.Errorf("The path for %#v was %#v instead of %#v", name, actual, expected)
		}
	}
}
