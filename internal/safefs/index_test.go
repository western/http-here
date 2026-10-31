package safefs

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// openTestRoot creates a RootFS over a fresh temp dir.
func openTestRoot(t *testing.T) (*RootFS, string) {

	t.Helper()

	dir := t.TempDir()

	root, err := OpenRoot(dir)
	if err != nil {
		t.Fatalf("OpenRoot(%s): %v", dir, err)
	}

	t.Cleanup(func() {
		root.Close()
	})

	return root, dir
}

// -------------------------------------------------------------------------------------------------------------------------
// OpenRoot

func TestOpenRoot_ok(t *testing.T) {

	root, dir := openTestRoot(t)

	if root == nil {
		t.Fatal("OpenRoot returned nil root")
	}

	if root.path != dir {
		t.Errorf("root.path = %q, want %q (absolute)", root.path, dir)
	}
}

func TestOpenRoot_emptyPath(t *testing.T) {

	for _, p := range []string{"", "   ", "\t"} {

		root, err := OpenRoot(p)

		if !errors.Is(err, ErrInvalidPath) {
			t.Errorf("OpenRoot(%q) error = %v, want ErrInvalidPath", p, err)
		}

		if root != nil {
			t.Errorf("OpenRoot(%q) returned non-nil root", p)
		}
	}
}

func TestOpenRoot_nonexistentPath(t *testing.T) {

	root, err := OpenRoot(filepath.Join(t.TempDir(), "does-not-exist"))

	if err == nil {
		t.Error("OpenRoot(nonexistent) expected error, got nil")
		root.Close()
	}
}

// -------------------------------------------------------------------------------------------------------------------------
// cleanPath: traversal protection

func TestRootFS_traversalRejected(t *testing.T) {

	root, _ := openTestRoot(t)

	bad := []string{
		"",
		"..",
		"../outside.txt",
		"a/../../outside.txt",
		"a\x00b",
	}

	for _, name := range bad {

		if _, err := root.Open(name); err == nil {
			t.Errorf("Open(%q) expected error, got nil", name)
		}

		if err := root.MkdirAll(name, 0755); err == nil {
			t.Errorf("MkdirAll(%q) expected error, got nil", name)
		}

		if err := root.Remove(name); err == nil {
			t.Errorf("Remove(%q) expected error, got nil", name)
		}

		if err := root.RemoveAll(name); err == nil {
			t.Errorf("RemoveAll(%q) expected error, got nil", name)
		}

		if _, err := root.Stat(name); err == nil {
			t.Errorf("Stat(%q) expected error, got nil", name)
		}

		if _, err := root.ReadFile(name); err == nil {
			t.Errorf("ReadFile(%q) expected error, got nil", name)
		}

		if err := root.WriteFile(name, []byte("x"), 0644); err == nil {
			t.Errorf("WriteFile(%q) expected error, got nil", name)
		}

		if err := root.Copy(name, "dst.txt"); err == nil {
			t.Errorf("Copy(%q, ...) expected error, got nil", name)
		}
	}
}

func TestRootFS_absolutePathRejected(t *testing.T) {

	root, _ := openTestRoot(t)

	for _, name := range []string{"/abs.txt", filepath.Join(root.path, "x.txt")} {

		if _, err := root.Open(name); err == nil {
			t.Errorf("Open(%q) expected error for absolute path, got nil", name)
		}

		if err := root.WriteFile(name, []byte("x"), 0644); err == nil {
			t.Errorf("WriteFile(%q) expected error for absolute path, got nil", name)
		}
	}
}

// -------------------------------------------------------------------------------------------------------------------------
// basic operations

func TestRootFS_mkdirOpenWriteReadRemove(t *testing.T) {

	root, dir := openTestRoot(t)

	// MkdirAll
	if err := root.MkdirAll("a/b", 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	if fi, err := os.Stat(filepath.Join(dir, "a", "b")); err != nil || !fi.IsDir() {
		t.Fatalf("MkdirAll did not create dir: %v", err)
	}

	// WriteFile into subdir
	data := []byte("hello safefs")

	if err := root.WriteFile("a/b/file.txt", data, 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// ReadFile
	got, err := root.ReadFile("a/b/file.txt")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if string(got) != string(data) {
		t.Errorf("ReadFile = %q, want %q", got, data)
	}

	// Open
	f, err := root.Open("a/b/file.txt")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	f.Close()

	// Stat
	fi, err := root.Stat("a/b/file.txt")
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}

	if fi.Size() != int64(len(data)) {
		t.Errorf("Stat size = %d, want %d", fi.Size(), len(data))
	}

	// Remove
	if err := root.Remove("a/b/file.txt"); err != nil {
		t.Fatalf("Remove: %v", err)
	}

	if _, err := root.Stat("a/b/file.txt"); err == nil {
		t.Error("file still exists after Remove")
	}

	// RemoveAll
	if err := root.RemoveAll("a"); err != nil {
		t.Fatalf("RemoveAll: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "a")); !os.IsNotExist(err) {
		t.Error("dir 'a' still exists after RemoveAll")
	}
}

func TestRootFS_copyMove(t *testing.T) {

	root, _ := openTestRoot(t)

	if err := root.WriteFile("src.txt", []byte("copy me"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Copy
	if err := root.Copy("src.txt", "dst.txt"); err != nil {
		t.Fatalf("Copy: %v", err)
	}

	got, err := root.ReadFile("dst.txt")
	if err != nil || string(got) != "copy me" {
		t.Errorf("Copy result = %q, %v", got, err)
	}

	// Move
	if err := root.Move("dst.txt", "moved.txt"); err != nil {
		t.Fatalf("Move: %v", err)
	}

	if _, err := root.Stat("dst.txt"); err == nil {
		t.Error("dst.txt still exists after Move")
	}

	got, err = root.ReadFile("moved.txt")
	if err != nil || string(got) != "copy me" {
		t.Errorf("Move result = %q, %v", got, err)
	}
}

func TestRootFS_readDirWalk(t *testing.T) {

	root, _ := openTestRoot(t)

	if err := root.MkdirAll("sub", 0755); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"one.txt", "sub/two.txt"} {
		if err := root.WriteFile(name, []byte(name), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// ReadDir of root
	entries, err := root.ReadDir(".")
	if err == nil {
		t.Fatalf("ReadDir: must not read POINT")
	}

	entries, err = root.ReadCurrentDir()
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}

	if len(entries) != 2 { // one.txt + sub
		t.Errorf("ReadDir('.') count = %d, want 2", len(entries))
	}

	// ReadDir of subdir
	entries, err = root.ReadDir("sub")
	if err != nil {
		t.Fatalf("ReadDir(sub): %v", err)
	}

	if len(entries) != 1 || entries[0].Name() != "two.txt" {
		t.Errorf("ReadDir('sub') = %v, want [two.txt]", entries)
	}

	// Walk
	visited := 0

	err = root.Walk(func(path string, d os.DirEntry, err error) error {

		if err != nil {
			return err
		}

		if !d.IsDir() {
			visited++
		}

		return nil
	})

	if err != nil {
		t.Fatalf("Walk: %v", err)
	}

	if visited != 2 {
		t.Errorf("Walk visited %d files, want 2", visited)
	}
}

// -------------------------------------------------------------------------------------------------------------------------
// Close / ErrRootClosed

func TestRootFS_closeAndUse(t *testing.T) {

	root, _ := openTestRoot(t)

	if err := root.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// second Close is a no-op
	if err := root.Close(); err != nil {
		t.Errorf("second Close returned error: %v", err)
	}

	closedOps := []struct {
		name string
		fn   func() error
	}{
		{"MkdirAll", func() error { return root.MkdirAll("x", 0755) }},
		{"WriteFile", func() error { return root.WriteFile("x", nil, 0644) }},
		{"Remove", func() error { return root.Remove("x") }},
		{"RemoveAll", func() error { return root.RemoveAll("x") }},
		{"Copy", func() error { return root.Copy("x", "y") }},
		{"Walk", func() error { return root.Walk(func(string, os.DirEntry, error) error { return nil }) }},
	}

	for _, op := range closedOps {

		if err := op.fn(); !errors.Is(err, ErrRootClosed) {
			t.Errorf("%s after Close: err = %v, want ErrRootClosed", op.name, err)
		}
	}

	if _, err := root.Open("x"); !errors.Is(err, ErrRootClosed) {
		t.Errorf("Open after Close: err = %v, want ErrRootClosed", err)
	}

	if _, err := root.ReadFile("x"); !errors.Is(err, ErrRootClosed) {
		t.Errorf("ReadFile after Close: err = %v, want ErrRootClosed", err)
	}

	if _, err := root.Stat("x"); !errors.Is(err, ErrRootClosed) {
		t.Errorf("Stat after Close: err = %v, want ErrRootClosed", err)
	}

	if _, err := root.ReadDir("x"); !errors.Is(err, ErrRootClosed) {
		t.Errorf("ReadDir after Close: err = %v, want ErrRootClosed", err)
	}
}

func TestRootFS_escapeIsImpossible(t *testing.T) {

	// security check: file created outside the root must NOT be accessible
	outside := t.TempDir()

	root, _ := openTestRoot(t)
	defer root.Close()

	secret := filepath.Join(outside, "secret.txt")

	if err := os.WriteFile(secret, []byte("top secret"), 0644); err != nil {
		t.Fatal(err)
	}

	// try to reach the secret via relative escape
	rel, err := filepath.Rel(root.path, secret)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(filepath.Clean(rel), "..") {
		t.Fatalf("test setup error: expected relative path outside root, got %q", rel)
	}

	if _, err := root.Open(rel); err == nil {
		t.Error("root.Open allowed escape outside of the root directory")
	}
}
