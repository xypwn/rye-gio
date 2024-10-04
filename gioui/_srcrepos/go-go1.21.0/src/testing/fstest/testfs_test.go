// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fstest

import (
	"internal/testenv"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestSymlink(t *testing.T) {
	testenv.MustHaveSymlink(t)

	tmp := t.TempDir()
	tmpfs := os.DirFS(tmp)

	if err := os.WriteFile(filepath.Join(tmp, "hello"), []byte("hello, world\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.Symlink(filepath.Join(tmp, "hello"), filepath.Join(tmp, "hello.link")); err != nil {
		t.Fatal(err)
	}

	if err := TestFS(tmpfs, "hello", "hello.link"); err != nil {
		t.Fatal(err)
	}
}

func TestDash(t *testing.T) {
	m := MapFS{
		"a-b/a": {Data: []byte("a-b/a")},
	}
	if err := TestFS(m, "a-b/a"); err != nil {
		t.Error(err)
	}
}

type shuffledFS MapFS

func (fsys shuffledFS) Open(name string) (fs.File, error) {
	f, err := MapFS(fsys).Open(name)
	if err != nil {
		return nil, err
	}
	return &shuffledFile{File: f}, nil
}

type shuffledFile struct{ fs.File }

func (f *shuffledFile) ReadDir(n int) ([]fs.DirEntry, error) {
	dirents, err := f.File.(fs.ReadDirFile).ReadDir(n)
	// Shuffle in a deterministic way, all we care about is making sure that the
	// list of directory entries is not is the lexicographic order.
	//
	// We do this to make sure that the TestFS test suite is not affected by the
	// order of directory entries.
	sort.Slice(dirents, func(i, j int) bool {
		return dirents[i].Name() > dirents[j].Name()
	})
	return dirents, err
}

func TestShuffledFS(t *testing.T) {
	fsys := shuffledFS{
		"tmp/one":   {Data: []byte("1")},
		"tmp/two":   {Data: []byte("2")},
		"tmp/three": {Data: []byte("3")},
	}
	if err := TestFS(fsys, "tmp/one", "tmp/two", "tmp/three"); err != nil {
		t.Error(err)
	}
}
