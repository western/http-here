package util

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// -------------------------------------------------------------------------------------------------------------------------
// helpers

func writeFileForTest(t *testing.T, name string, data []byte) string {

	t.Helper()

	full := filepath.Join(t.TempDir(), name)

	if err := os.WriteFile(full, data, 0644); err != nil {
		t.Fatalf("write test file %s: %v", name, err)
	}

	return full
}

func writeFileForTest2(t *testing.T, full_name string, data []byte) {

	t.Helper()

	if err := os.WriteFile(full_name, data, 0644); err != nil {
		t.Fatalf("write test file %s: %v", full_name, err)
	}

}

func assertFileContent(t *testing.T, path string, want []byte) {

	t.Helper()

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	if !bytes.Equal(got, want) {
		t.Errorf("content of %s = %q, want %q", path, got, want)
	}
}

// -------------------------------------------------------------------------------------------------------------------------
// CopyFile

func TestCopyFile(t *testing.T) {

	dir := t.TempDir()
	src := writeFileForTest(t, "src.txt", []byte("copy me"))
	dst := filepath.Join(dir, "dst.txt")

	if err := CopyFile(src, dst); err != nil {
		t.Fatalf("CopyFile: %v", err)
	}

	assertFileContent(t, dst, []byte("copy me"))
}

func TestCopyFile_missingSource(t *testing.T) {

	dir := t.TempDir()

	err := CopyFile(filepath.Join(dir, "nope.txt"), filepath.Join(dir, "dst.txt"))

	if err == nil {
		t.Error("CopyFile with missing source expected error, got nil")
	}
}

// -------------------------------------------------------------------------------------------------------------------------
// ROOTCopyFile

func TestROOTCopyFile(t *testing.T) {

	dir := t.TempDir()

	srcRoot, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatalf("OpenRoot src: %v", err)
	}
	defer srcRoot.Close()

	dstDir := t.TempDir()

	dstRoot, err := os.OpenRoot(dstDir)
	if err != nil {
		t.Fatalf("OpenRoot dst: %v", err)
	}
	defer dstRoot.Close()

	if err := os.WriteFile(filepath.Join(dir, "src.txt"), []byte("root copy"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := ROOTCopyFile(srcRoot, "src.txt", dstRoot, "dst.txt"); err != nil {
		t.Fatalf("ROOTCopyFile: %v", err)
	}

	assertFileContent(t, filepath.Join(dstDir, "dst.txt"), []byte("root copy"))
}

// -------------------------------------------------------------------------------------------------------------------------
// MoveFile

func TestMoveFile(t *testing.T) {

	dir := t.TempDir()
	src := writeFileForTest(t, "src.txt", []byte("move me"))
	dst := filepath.Join(dir, "dst.txt")

	if err := MoveFile(src, dst); err != nil {
		t.Fatalf("MoveFile: %v", err)
	}

	assertFileContent(t, dst, []byte("move me"))

	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Errorf("source %s still exists after MoveFile", src)
	}
}

// -------------------------------------------------------------------------------------------------------------------------
// CopyDir

func TestCopyDir(t *testing.T) {

	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")

	if err := os.MkdirAll(filepath.Join(src, "sub"), 0755); err != nil {
		t.Fatal(err)
	}

	_, err := os.Stat(filepath.Join(src, "sub"))
	if err != nil {
		t.Fatal(err)
	}

	writeFileForTest2(t, filepath.Join(src, "a.txt"), []byte("a"))
	writeFileForTest2(t, filepath.Join(src, "sub", "b.txt"), []byte("b"))

	if err := CopyDir(src, dst); err != nil {
		t.Fatalf("CopyDir: %v", err)
	}

	assertFileContent(t, filepath.Join(dst, "a.txt"), []byte("a"))
	assertFileContent(t, filepath.Join(dst, "sub", "b.txt"), []byte("b"))
}

// -------------------------------------------------------------------------------------------------------------------------
// GetMd5File

func TestGetMd5File(t *testing.T) {

	// md5("abc") = 900150983cd24fb0d6963f7d28e17f72
	src := writeFileForTest(t, "abc.txt", []byte("abc"))

	if got := GetMd5File(src); got != fmt.Sprintf("%x", md5.Sum([]byte("abc"))) {
		t.Errorf("GetMd5File = %q, want md5 of \"abc\"", got)
	}
}

func TestGetMd5File_missingFile(t *testing.T) {

	if got := GetMd5File(filepath.Join(t.TempDir(), "nope.txt")); got != "" {
		t.Errorf("GetMd5File(missing) = %q, want empty", got)
	}
}

// -------------------------------------------------------------------------------------------------------------------------
// IsDirectory

func TestIsDirectory(t *testing.T) {

	dir := t.TempDir()

	if ok, err := IsDirectory(dir); err != nil || !ok {
		t.Errorf("IsDirectory(dir) = %v, %v; want true, nil", ok, err)
	}

	file := writeFileForTest(t, "file.txt", []byte("x"))

	if ok, err := IsDirectory(file); ok || err == nil {
		t.Errorf("IsDirectory(file) = %v, %v; want false, error", ok, err)
	}

	if _, err := IsDirectory(filepath.Join(dir, "nope")); err == nil {
		t.Error("IsDirectory(missing) expected error, got nil")
	}
}

// -------------------------------------------------------------------------------------------------------------------------
// WalkAndClearOld

func TestWalkAndClearOld(t *testing.T) {

	dir := t.TempDir()

	oldFile := filepath.Join(dir, "old.txt")
	freshFile := filepath.Join(dir, "fresh.txt")

	writeFileForTest2(t, oldFile, []byte("old"))
	writeFileForTest2(t, freshFile, []byte("fresh"))

	// make oldFile look 2 hours old
	past := time.Now().Add(-2 * time.Hour)

	if err := os.Chtimes(oldFile, past, past); err != nil {
		t.Fatalf("Chtimes: %v", err)
	}

	WalkAndClearOld(dir, 1*time.Hour)

	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Errorf("old file %s was not removed", oldFile)
	}

	if _, err := os.Stat(freshFile); err != nil {
		t.Errorf("fresh file %s was removed but should stay", freshFile)
	}
}

// -------------------------------------------------------------------------------------------------------------------------
// WalkAndClearZeroFile

func TestWalkAndClearZeroFile(t *testing.T) {

	dir := t.TempDir()

	zeroFile := filepath.Join(dir, "zero.txt")
	dataFile := filepath.Join(dir, "data.txt")

	writeFileForTest2(t, zeroFile, []byte(""))
	writeFileForTest2(t, dataFile, []byte("data"))

	if err := os.MkdirAll(filepath.Join(dir, "sub"), 0755); err != nil {
		t.Fatal(err)
	}

	nestedZero := filepath.Join(dir, "sub", "nested_zero.txt")

	if err := os.WriteFile(nestedZero, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	WalkAndClearZeroFile(dir, 0)

	if _, err := os.Stat(zeroFile); !os.IsNotExist(err) {
		t.Errorf("zero file %s was not removed", zeroFile)
	}

	if _, err := os.Stat(nestedZero); !os.IsNotExist(err) {
		t.Errorf("nested zero file %s was not removed", nestedZero)
	}

	if _, err := os.Stat(dataFile); err != nil {
		t.Errorf("data file %s was removed but should stay", dataFile)
	}
}

// -------------------------------------------------------------------------------------------------------------------------
// MultipartToFile

func TestMultipartToFile(t *testing.T) {

	var buf bytes.Buffer

	w := multipart.NewWriter(&buf)

	part, err := w.CreateFormFile("file", "test.txt")
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}

	if _, err := part.Write([]byte("hello multipart")); err != nil {
		t.Fatalf("part write: %v", err)
	}

	if err := w.Close(); err != nil {
		t.Fatalf("writer close: %v", err)
	}

	form, err := multipart.NewReader(&buf, w.Boundary()).ReadForm(1 << 20)
	if err != nil {
		t.Fatalf("ReadForm: %v", err)
	}

	fhs, ok := form.File["file"]
	if !ok || len(fhs) == 0 {
		t.Fatal("form has no 'file' entries")
	}

	f := MultipartToFile(fhs[0])
	defer f.Close()

	got, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("read result: %v", err)
	}

	if string(got) != "hello multipart" {
		t.Errorf("MultipartToFile content = %q, want %q", got, "hello multipart")
	}
}
