package util

import (
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
)

// -------------------------------------------------------------------------------------------------------------------------
// CleanPath

func TestCleanPath(t *testing.T) {

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "empty", input: "", wantErr: true},
		{name: "dot", input: ".", wantErr: true},
		{name: "dotdot", input: "..", wantErr: true},
		{name: "dotdot prefix", input: "../secret.txt", wantErr: true},
		{name: "deep dotdot", input: "a/../../etc/passwd", wantErr: true},
		{name: "nul byte", input: "a\x00b", wantErr: true},

		{name: "simple", input: "ok/path.txt", want: filepath.Clean("ok/path.txt")},
		{name: "double slash", input: "a//b", want: filepath.Clean("a/b")},
		{name: "dot segments", input: "a/./b", want: filepath.Clean("a/b")},
		{name: "inner dotdot", input: "a/b/../c", want: filepath.Clean("a/c")},
		{name: "trailing slash", input: "a/b/", want: filepath.Clean("a/b")},
		{name: "single file", input: "file.txt", want: "file.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := CleanPath(tt.input)

			if tt.wantErr {

				if err == nil {
					t.Errorf("CleanPath(%q) expected error, got %q", tt.input, got)
				}
				return
			}

			if err != nil {
				t.Errorf("CleanPath(%q) unexpected error: %v", tt.input, err)
				return
			}

			if got != tt.want {
				t.Errorf("CleanPath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// -------------------------------------------------------------------------------------------------------------------------
// CleanDirtyPath

func TestCleanDirtyPath(t *testing.T) {

	tests := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"normal/path", "normal/path"},
		{"a//b///c", "a/b/c"},
		{"a\\b\\c", "abc"},
		{"a...b", "a.b"},
		{"....", "."},
		{"../../etc/passwd", "././etc/passwd"},
		{"..%2f..%2fsecret.txt", ".%2f.%2fsecret.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {

			if got := CleanDirtyPath(tt.input); got != tt.want {
				t.Errorf("CleanDirtyPath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// -------------------------------------------------------------------------------------------------------------------------
// GetClearFilename

func TestGetClearFilename(t *testing.T) {

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "simple", input: "image1.jpg", want: "image1.jpg"},
		{name: "with path", input: "/path/to/image2.jpg", want: "image2.jpg"},
		{name: "url escaped", input: "my%20file.txt", want: "my file.txt"},
		{name: "double spaces", input: "ima  ge3.jpg", want: "ima ge3.jpg"},
		{name: "no ext", input: "noext", want: "noext."},
		{name: "upper ext normalized", input: "photo.JPEG", want: "photo.jpg"},
		{name: "bad escape", input: "%zz", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := GetClearFilename(tt.input)

			if tt.wantErr {

				if err == nil {
					t.Errorf("GetClearFilename(%q) expected error, got %q", tt.input, got)
				}
				return
			}

			if err != nil {
				t.Errorf("GetClearFilename(%q) unexpected error: %v", tt.input, err)
				return
			}

			if got != tt.want {
				t.Errorf("GetClearFilename(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// -------------------------------------------------------------------------------------------------------------------------
// RefererOrPath (via real fiber Ctx from app.Test)

func TestRefererOrPath_fromReferer(t *testing.T) {

	want := filepath.Clean("/docs/folder")

	app := fiber.New()

	var got string
	var gotErr error

	app.Get("/", func(c fiber.Ctx) error {

		got, gotErr = RefererOrPath(c)

		return c.SendString("ok")
	})

	req := httptest.NewRequest(fiber.MethodGet, "/", nil)
	req.Header.Set("Referer", "http://localhost/docs/folder")

	if _, err := app.Test(req); err != nil {
		t.Fatalf("app.Test error: %v", err)
	}

	if gotErr != nil {
		t.Fatalf("RefererOrPath unexpected error: %v", gotErr)
	}

	if got != want {
		t.Errorf("RefererOrPath = %q, want %q", got, want)
	}
}

func TestRefererOrPath_formOverrides(t *testing.T) {

	want := filepath.Clean("/custom/path")

	app := fiber.New()

	var got string
	var gotErr error

	app.Post("/", func(c fiber.Ctx) error {

		got, gotErr = RefererOrPath(c)

		return c.SendString("ok")
	})

	req := httptest.NewRequest(fiber.MethodPost, "/", strings.NewReader("path=/custom/path"))
	req.Header.Set("Referer", "http://localhost/ignored")
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationForm)

	if _, err := app.Test(req); err != nil {
		t.Fatalf("app.Test error: %v", err)
	}

	if gotErr != nil {
		t.Fatalf("RefererOrPath unexpected error: %v", gotErr)
	}

	if got != want {
		t.Errorf("RefererOrPath = %q, want %q", got, want)
	}
}

/*
func TestRefererOrPath_badFormPath(t *testing.T) {

	app := fiber.New()

	var gotErr error

	app.Post("/", func(c fiber.Ctx) error {

		_, gotErr = RefererOrPath(c)

		return c.SendString("ok")
	})

	req := httptest.NewRequest(fiber.MethodPost, "/", strings.NewReader("path=../evil"))
	req.Header.Set("Referer", "http://localhost/docs")
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationForm)

	if _, err := app.Test(req); err != nil {
		t.Fatalf("app.Test error: %v", err)
	}

	if gotErr == nil {
		t.Error("RefererOrPath expected error for traversal form path, got nil")
	}
}
*/
