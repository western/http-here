package util

import (
	"strings"
	"testing"
)

// -------------------------------------------------------------------------------------------------------------------------
// RandStringRunes

func TestRandStringRunes(t *testing.T) {

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

	s := RandStringRunes(32)

	if len(s) != 32 {
		t.Errorf("RandStringRunes(32) length = %d, want 32", len(s))
	}

	for _, r := range s {

		if !strings.ContainsRune(charset, r) {
			t.Errorf("RandStringRunes produced unexpected rune %q", r)
		}
	}

	if got := RandStringRunes(0); got != "" {
		t.Errorf("RandStringRunes(0) = %q, want empty", got)
	}

	if RandStringRunes(32) == RandStringRunes(32) {
		t.Error("RandStringRunes(32) returned equal results twice (not random)")
	}
}

// -------------------------------------------------------------------------------------------------------------------------
// RotateSlash

func TestRotateSlash(t *testing.T) {

	tests := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"a/b/c", "a/b/c"},
		{`a\b\c`, "a/b/c"},
		{`C:\Users\test`, "C:/Users/test"},
		{`mixed/and\slash`, "mixed/and/slash"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {

			if got := RotateSlash(tt.input); got != tt.want {
				t.Errorf("RotateSlash(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// -------------------------------------------------------------------------------------------------------------------------
// GetExtNorm

func TestGetExtNorm(t *testing.T) {

	tests := []struct {
		input string
		want  string
	}{
		{"photo.jpg", "jpg"},
		{"photo.JPG", "jpg"},
		{"photo.JPEG", "jpg"},
		{"a.jpeg", "jpg"},
		{"archive.tar.gz", "gz"},
		{"noext", ""},
		{"", ""},
		{"/path/to/image.png", "png"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {

			if got := GetExtNorm(tt.input); got != tt.want {
				t.Errorf("GetExtNorm(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// -------------------------------------------------------------------------------------------------------------------------
// GetFileName

func TestGetFileName(t *testing.T) {

	tests := []struct {
		input string
		want  string
	}{
		{"image1.jpg", "image1"},
		{"/path/to/image2.jpg", "image2"},
		{"noext", "noext"},
		//{".hidden", ".hidden"},
		{"a.tar.gz", "a.tar"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {

			if got := GetFileName(tt.input); got != tt.want {
				t.Errorf("GetFileName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// -------------------------------------------------------------------------------------------------------------------------
// PrettyByteSize

func TestPrettyByteSize(t *testing.T) {

	tests := []struct {
		input int64
		want  string
	}{
		{0, "0.0 B"},
		{1023, "1023.0 B"},
		{1024, "1.0 KiB"},
		{1536, "1.5 KiB"},
		{1048576, "1.0 MiB"},
		{5 * 1024 * 1024 * 1024, "5.0 GiB"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {

			if got := PrettyByteSize(tt.input); got != tt.want {
				t.Errorf("PrettyByteSize(%d) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// -------------------------------------------------------------------------------------------------------------------------
// StringClearColor

func TestStringClearColor(t *testing.T) {

	tests := []struct {
		input string
		want  string
	}{
		{"plain", "plain"},
		{"\x1b[31mred\x1b[0m", "red"},
		{"\x1b[1;32mgreen\x1b[0m plain", "green plain"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {

			if got := StringClearColor(tt.input); got != tt.want {
				t.Errorf("StringClearColor(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// -------------------------------------------------------------------------------------------------------------------------
// IsInt

func TestIsInt(t *testing.T) {

	tests := []struct {
		input string
		want  bool
	}{
		{"42", true},
		{"-5", true},
		{"0", true},
		{"3.14", false},
		{"abc", false},
		{"", false},
		{"12a", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {

			if got := IsInt(tt.input); got != tt.want {
				t.Errorf("IsInt(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
