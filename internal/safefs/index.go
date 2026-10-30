package safefs

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrInvalidPath = errors.New("Invalid filesystem path")
	ErrRootClosed  = errors.New("Filesystem root is closed")
)

type RootFS struct {
	root *os.Root
	path string
}

func OpenRoot(path string) (*RootFS, error) {
	if strings.TrimSpace(path) == "" {
		return nil, ErrInvalidPath
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("Resolve root: %w", err)
	}

	/*
		if err := os.MkdirAll(abs, 0750); err != nil {
			return nil, fmt.Errorf("Create root: %w", err)
		}
	*/

	root, err := os.OpenRoot(abs)
	if err != nil {
		return nil, fmt.Errorf("Open root: %w", err)
	}

	return &RootFS{
		root: root,
		path: abs,
	}, nil
}

func (r *RootFS) valid() error {
	if r == nil || r.root == nil {
		return ErrRootClosed
	}
	return nil
}

func cleanPath(name string) (string, error) {
	if name == "" ||
		filepath.IsAbs(name) ||
		strings.ContainsRune(name, '\x00') {
		return "", ErrInvalidPath
	}

	/*
	   filepath.Clean
	   ------------------------------------------------------------------------------------------------
	   1. Reduces multiple consecutive slashes to a single one (e.g., `a//b` becomes `a/b`).
	   2. Removes references to the current directory (`.`) (e.g., `a/./b` becomes `a/b`).
	   3. Processes parent directory references (`..`) by collapsing them along with the preceding path segment (e.g., `a/b/../c` becomes `a/c`).
	   4. Returns `.` if an empty string (`""`) is provided as input.
	   5. Clears `..` references at the root level (e.g., `/../` becomes `/` on Unix-like systems).
	   6. Removes trailing slashes (unless the path is a root directory, such as `/` or `C:\`) — for example, `a/b/` becomes `a/b`.
	   7. Accounts for the operating system: uses forward slashes (`/`) on Unix or backslashes (`\`) on Windows.
	*/
	name = filepath.Clean(name)

	if name == "." || name == ".." ||
		strings.HasPrefix(name, ".."+string(filepath.Separator)) {
		return "", ErrInvalidPath
	}

	return name, nil
}

// Close releases the root handle.
func (r *RootFS) Close() error {
	if r == nil || r.root == nil {
		return nil
	}

	root := r.root
	r.root = nil
	return root.Close()
}

// MkdirAll creates directories inside the root.
func (r *RootFS) MkdirAll(name string, perm fs.FileMode) error {
	if err := r.valid(); err != nil {
		return err
	}

	name, err := cleanPath(name)
	if err != nil {
		return err
	}

	return r.root.MkdirAll(name, perm)
}

// Open opens a file for reading.
func (r *RootFS) Open(name string) (*os.File, error) {
	if err := r.valid(); err != nil {
		return nil, err
	}

	name, err := cleanPath(name)
	if err != nil {
		return nil, err
	}

	return r.root.Open(name)
}

// ReadFile reads a file.
func (r *RootFS) ReadFile(name string) ([]byte, error) {
	if err := r.valid(); err != nil {
		return nil, err
	}

	name, err = cleanPath(name)
	if err != nil {
		return nil, err
	}

	return r.root.ReadFile(name)
}

// WriteFile atomically replaces a file.
func (r *RootFS) WriteFile(
	name string,
	data []byte,
	perm fs.FileMode,
) error {
	if err := r.valid(); err != nil {
		return err
	}

	name, err := cleanPath(name)
	if err != nil {
		return err
	}

	dir := filepath.Dir(name)
	tmp, err := os.CreateTemp(
		filepath.Join(r.path, dir),
		".tmp-*",
	)
	if err != nil {
		return err
	}

	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}

	if err := tmp.Chmod(perm); err != nil {
		tmp.Close()
		return err
	}

	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}

	if err := tmp.Close(); err != nil {
		return err
	}

	// Re-open the temporary file through the confined root.
	relative, err := filepath.Rel(r.path, tmpName)
	if err != nil {
		return err
	}

	relative, err = cleanPath(relative)
	if err != nil {
		return err
	}

	f, err := r.root.OpenFile(
		relative,
		os.O_WRONLY|os.O_CREATE|os.O_EXCL,
		perm,
	)
	if err == nil {
		f.Close()
		return errors.New("Unexpected temporary file collision")
	}

	// Rename using paths relative to the root.
	return r.root.Rename(relative, name)
}

// Remove deletes a file or empty directory.
func (r *RootFS) Remove(name string) error {
	if err := r.valid(); err != nil {
		return err
	}

	name, err := cleanPath(name)
	if err != nil {
		return err
	}

	return r.root.Remove(name)
}

// remove folder and everything inside
func (r *RootFS) RemoveAll(name string) error {
	if err := r.valid(); err != nil {
		return err
	}

	name, err := cleanPath(name)
	if err != nil {
		return err
	}

	return r.root.RemoveAll(name)
}

// Stat retrieves file information.
func (r *RootFS) Stat(name string) (fs.FileInfo, error) {
	if err := r.valid(); err != nil {
		return nil, err
	}

	name, err := cleanPath(name)
	if err != nil {
		return nil, err
	}

	return r.root.Stat(name)
}

// Copy copies a file within the root.
func (r *RootFS) Copy(src, dst string) error {
	in, err := r.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	dst, err = cleanPath(dst)
	if err != nil {
		return err
	}

	out, err := r.root.OpenFile(
		dst,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		0640,
	)
	if err != nil {
		return err
	}

	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()

	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

// Walk walks the root directory.
func (r *RootFS) Walk(
	fn fs.WalkDirFunc,
) error {
	if err := r.valid(); err != nil {
		return err
	}

	return fs.WalkDir(r.root.FS(), ".", fn)
}
