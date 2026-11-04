// Package crosspath deals with file/directory paths in a cross-platform way.
//
// This package tries to understand Windows paths on UNIX and vice versa.
// Returned paths may be using forward slashes as separators.
package crosspath

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"fmt"
	path_module "path" // import under other name so that parameters can be called 'path'
	"path/filepath"
	"strings"
)

// Base returns the last element of path. Trailing slashes are removed before
// extracting the last element. If the path is empty, Base returns ".". If the
// path consists entirely of slashes, Base returns "/".
func Base(path string) string {
	slashed := ToSlash(path)
	return path_module.Base(slashed)
}

// Dir returns all but the last element of path, typically the path's directory.
// If the path is empty, Dir returns ".".
func Dir(path string) string {
	if path == "" {
		return "."
	}

	slashed := ToSlash(path)

	// Don't use path.Dir(), as that cleans up the path and removes double
	// slashes. However, Windows UNC paths start with double backslashes, which
	// will translate to double slashes and should not be removed.
	dir, _ := path_module.Split(slashed)
	switch {
	case dir == "":
		return "."
	case len(dir) > 1:
		// Remove trailing slash.
		return dir[:len(dir)-1]
	default:
		return dir
	}
}

func Join(elem ...string) string {
	return ToSlash(path_module.Join(elem...))
}

// Stem returns the filename without extension.
func Stem(path string) string {
	base := Base(path)
	ext := path_module.Ext(base)
	return base[:len(base)-len(ext)]
}

// ToSlash replaces all backslashes with forward slashes.
// Contrary to filepath.ToSlash(), this also happens on Linux; it does not
// expect `path` to be in platform-native notation.
func ToSlash(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}

// ToNative replaces all path separators (forward and backward slashes) with the
// platform-native separator.
func ToNative(path string) string {
	switch filepath.Separator {
	case '/':
		return ToSlash(path)
	case '\\':
		return strings.ReplaceAll(path, "/", "\\")
	default:
		panic(fmt.Sprintf("this platform has an unknown path separator: %q", filepath.Separator))
	}
}

func validDriveLetter(r rune) bool {
	return ('A' <= r && r <= 'Z') || ('a' <= r && r <= 'z')
}

// IsRoot returns whether the given path is a root path or not.
// Paths "/", "C:", "C:\", and "C:/" are considered root, for all latin drive
// letters A-Z.
//
// NOTE: this does NOT resolve symlinks or `..` entries.
func IsRoot(path string) bool {
	switch {
	case path == "":
		return false
	case path == "/":
		return true
	// From here on, it can only be a DOS root, so must have a drive letter and a colon.
	case len(path) < 2, len(path) > 3:
		return false
	case path[1] != ':':
		return false
	// C:\ and C:/ are both considered roots.
	case len(path) == 3 && path[2] != '/' && path[2] != '\\':
		return false
	}

	runes := []rune(path)
	return validDriveLetter(runes[0])
}

// ToPlatform returns the path, with path separators adjusted for the given platform.
// It is assumed that all forward and backward slashes in the path are path
// separators, and that apart from the style of separators the path makes sense
// on the target platform.
func ToPlatform(path, platform string) string {
	if path == "" {
		return ""
	}

	components := strings.FieldsFunc(path, isPathSep)

	// FieldsFunc() removes leading path separators, turning an absolute path on
	// Linux to a relative path, and turning `\\NAS\share` on Windows into
	// `NAS\share`.
	extraComponents := []string{}
	for _, r := range path {
		if !isPathSep(r) {
			break
		}
		extraComponents = append(extraComponents, "")
	}
	components = append(extraComponents, components...)

	pathsep := pathSepForPlatform(platform)
	translated := strings.Join(components, pathsep)

	if platform == "windows" {
		return EnsureDriveAbsolute(translated)
	}

	return translated
}

// pathSepForPlatform returns the path separator for the given platform.
// This is rather simple, and just returns `\` on Windows and `/` on all other
// platforms.
func pathSepForPlatform(platform string) string {
	switch platform {
	case "windows":
		return `\`
	default:
		return "/"
	}
}

func isPathSep(r rune) bool {
	return r == '/' || r == '\\'
}

// TrimTrailingSep removes any trailing path separator.
func TrimTrailingSep(path string) string {
	if path == "" {
		return ""
	}

	trimmed := strings.TrimRightFunc(path, isPathSep)
	if trimmed == "" {
		return string([]rune(path)[0])
	}
	return trimmed
}

// EnsureDriveAbsolute ensures that a Windows path that starts with a drive
// letter also has an absolute path on that drive. For example, turns
// `F:path\to\file` into `F:\path\to\file`.
func EnsureDriveAbsolute(windowsPath string) string {
	runes := []rune(windowsPath)
	numRunes := len(runes)
	if numRunes < 2 {
		return windowsPath
	}

	if !validDriveLetter(runes[0]) || runes[1] != ':' {
		return windowsPath
	}

	pathSep := pathSepForPlatform("windows")
	if numRunes == 2 {
		return windowsPath + pathSep
	}
	if string(runes[2]) == pathSep {
		return windowsPath // Already F:\blabla
	}

	return string(runes[:2]) + pathSep + string(runes[2:])
}

// IsUNCNotation returns whether the path is in Windows UNC notation or not.
func IsUNCNotation(windowsPath string) bool {
	return strings.HasPrefix(windowsPath, `\\`)
}
