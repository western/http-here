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

	root, err := NewSafeFS(dir)
	if err != nil {
		t.Fatalf("NewSafeFS(%s): %v", dir, err)
	}

	t.Cleanup(func() {
		root.Close()
	})

	return root, dir
}

// -------------------------------------------------------------------------------------------------------------------------
// NewSafeFS

func TestNewSafeFS_ok(t *testing.T) {

	root, dir := openTestRoot(t)

	if root == nil {
		t.Fatal("NewSafeFS returned nil root")
	}

	if root.path != dir {
		t.Errorf("root.path = %q, want %q (absolute)", root.path, dir)
	}

	if got := root.Name(); got != dir {
		t.Errorf("Name() = %q, want %q", got, dir)
	}
}

func TestNewSafeFS_emptyPath(t *testing.T) {

	for _, p := range []string{"", "   ", "\t"} {

		root, err := NewSafeFS(p)

		if !errors.Is(err, ErrInvalidPath) {
			t.Errorf("NewSafeFS(%q) error = %v, want ErrInvalidPath", p, err)
		}

		if root != nil {
			t.Errorf("NewSafeFS(%q) returned non-nil root", p)
		}
	}
}

func TestNewSafeFS_nonexistentPath(t *testing.T) {

	root, err := NewSafeFS(filepath.Join(t.TempDir(), "does-not-exist"))

	if err == nil {
		t.Error("NewSafeFS(nonexistent) expected error, got nil")
		root.Close()
	}
}

// -------------------------------------------------------------------------------------------------------------------------
// cleanPath: path validation

func TestCleanPath_valid(t *testing.T) {

	cases := []struct {
		in   string
		want string
	}{
		{in: "a.txt", want: "a.txt"},
		{in: filepath.Join("a", "b"), want: filepath.Join("a", "b")},
		{in: "./a.txt", want: "a.txt"},
		{in: "a//b", want: filepath.Join("a", "b")},
		{in: "a/./b", want: filepath.Join("a", "b")},
		{in: "a/b/", want: filepath.Join("a", "b")},
		{in: "a/b/../c", want: filepath.Join("a", "c")},
	}

	for _, tc := range cases {

		got, err := cleanPath(tc.in)

		if err != nil {
			t.Errorf("cleanPath(%q) unexpected error: %v", tc.in, err)
			continue
		}

		if got != tc.want {
			t.Errorf("cleanPath(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestCleanPath_invalid(t *testing.T) {

	bad := []string{
		"",
		".",
		"..",
		"../outside.txt",
		"a/../../outside.txt",
		"a\x00b",
		filepath.Join(t.TempDir(), "abs.txt"), // absolute path
	}

	// "/abs.txt" is absolute only on Unix-like systems
	if filepath.IsAbs("/abs.txt") {
		bad = append(bad, "/abs.txt")
	}

	for _, name := range bad {

		got, err := cleanPath(name)

		if err == nil {
			t.Errorf("cleanPath(%q) expected error, got %q", name, got)
		}

		if got != "" {
			t.Errorf("cleanPath(%q) must return empty string on error, got %q", name, got)
		}
	}
}

// -------------------------------------------------------------------------------------------------------------------------
// cleanPath via API: traversal protection

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

func TestRootFS_mkdirNonRecursive(t *testing.T) {

	root, dir := openTestRoot(t)

	// simple Mkdir works
	if err := root.Mkdir("dir1", 0755); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}

	if fi, err := os.Stat(filepath.Join(dir, "dir1")); err != nil || !fi.IsDir() {
		t.Fatalf("Mkdir did not create dir: %v", err)
	}

	// Mkdir does not create parent dirs
	if err := root.Mkdir("a/b/c", 0755); err == nil {
		t.Error("Mkdir('a/b/c') expected error when parent is absent, got nil")
	}

	// traversal input is rejected
	if err := root.Mkdir("..", 0755); err == nil {
		t.Error("Mkdir('..') expected error, got nil")
	}
}

func TestRootFS_rename(t *testing.T) {

	root, _ := openTestRoot(t)

	if err := root.WriteFile("old.txt", []byte("rename me"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if err := root.Mkdir("sub", 0755); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}

	// rename into a subdir
	if err := root.Rename("old.txt", "sub/new.txt"); err != nil {
		t.Fatalf("Rename: %v", err)
	}

	if _, err := root.Stat("old.txt"); err == nil {
		t.Error("old.txt still exists after Rename")
	}

	got, err := root.ReadFile("sub/new.txt")
	if err != nil || string(got) != "rename me" {
		t.Errorf("Rename result = %q, %v", got, err)
	}

	// invalid input is rejected
	if err := root.Rename("../out.txt", "x.txt"); err == nil {
		t.Error("Rename('../out.txt', ...) expected error, got nil")
	}

	if err := root.Rename("sub/new.txt", ""); err == nil {
		t.Error("Rename(..., '') expected error, got nil")
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

	// source must survive Copy
	if _, err := root.Stat("src.txt"); err != nil {
		t.Errorf("src.txt missing after Copy: %v", err)
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

	// Move of absent source must fail
	if err := root.Move("absent.txt", "out.txt"); err == nil {
		t.Error("Move('absent.txt', ...) expected error, got nil")
	}
}

func TestRootFS_writeFileOverwrite(t *testing.T) {

	root, dir := openTestRoot(t)

	if err := root.WriteFile("file.txt", []byte("version-1"), 0644); err != nil {
		t.Fatalf("WriteFile v1: %v", err)
	}

	if err := root.WriteFile("file.txt", []byte("version-2-longer"), 0644); err != nil {
		t.Fatalf("WriteFile v2: %v", err)
	}

	got, err := root.ReadFile("file.txt")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if string(got) != "version-2-longer" {
		t.Errorf("ReadFile = %q, want %q", got, "version-2-longer")
	}

	// no temporary files must be left behind
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir(root dir): %v", err)
	}

	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".tmp-") {
			t.Errorf("leftover temporary file: %s", e.Name())
		}
	}
}

// -------------------------------------------------------------------------------------------------------------------------
// listing and walking

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

	entries, err = root.ReadRootDir()
	if err != nil {
		t.Fatalf("ReadRootDir: %v", err)
	}

	if len(entries) != 2 { // one.txt + sub
		t.Errorf("ReadRootDir count = %d, want 2", len(entries))
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
// RootStat

func TestRootFS_rootStat(t *testing.T) {

	root, _ := openTestRoot(t)

	fi, err := root.RootStat()
	if err != nil {
		t.Fatalf("RootStat: %v", err)
	}

	if !fi.IsDir() {
		t.Errorf("RootStat = %v, want a directory", fi)
	}

	// os.Root.Stat(".") may report the name as "." or as the base dir name
	// depending on the OS
	if name := fi.Name(); name != "." && name != filepath.Base(root.Name()) {
		t.Errorf("RootStat name = %q, want %q or %q", name, ".", filepath.Base(root.Name()))
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
		{"Mkdir", func() error { return root.Mkdir("x", 0755) }},
		{"MkdirAll", func() error { return root.MkdirAll("x", 0755) }},
		{"WriteFile", func() error { return root.WriteFile("x", nil, 0644) }},
		{"Rename", func() error { return root.Rename("x", "y") }},
		{"Move", func() error { return root.Move("x", "y") }},
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

	if _, err := root.RootStat(); !errors.Is(err, ErrRootClosed) {
		t.Errorf("RootStat after Close: err = %v, want ErrRootClosed", err)
	}

	if _, err := root.ReadDir("x"); !errors.Is(err, ErrRootClosed) {
		t.Errorf("ReadDir after Close: err = %v, want ErrRootClosed", err)
	}

	if _, err := root.ReadRootDir(); !errors.Is(err, ErrRootClosed) {
		t.Errorf("ReadRootDir after Close: err = %v, want ErrRootClosed", err)
	}

	if got := root.Name(); got != "" {
		t.Errorf("Name after Close = %q, want empty string", got)
	}
}

func TestRootFS_nilReceiver(t *testing.T) {

	var r *RootFS

	if err := r.Close(); err != nil {
		t.Errorf("nil Close: err = %v, want nil", err)
	}

	if got := r.Name(); got != "" {
		t.Errorf("nil Name = %q, want empty string", got)
	}

	if _, err := r.Open("x"); !errors.Is(err, ErrRootClosed) {
		t.Errorf("nil Open: err = %v, want ErrRootClosed", err)
	}

	if _, err := r.Stat("x"); !errors.Is(err, ErrRootClosed) {
		t.Errorf("nil Stat: err = %v, want ErrRootClosed", err)
	}

	if err := r.Remove("x"); !errors.Is(err, ErrRootClosed) {
		t.Errorf("nil Remove: err = %v, want ErrRootClosed", err)
	}
}

// -------------------------------------------------------------------------------------------------------------------------
// security

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
