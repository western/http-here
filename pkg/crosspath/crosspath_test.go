package crosspath

import (
	"path"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBase(t *testing.T) {
	tests := []struct {
		expect, input string
	}{
		{".", ""},
		{"justafile.txt", "justafile.txt"},
		{"with spaces.txt", "/Linux path/with spaces.txt"},
		{"awésom.tar.gz", "C:\\ünicode\\is\\awésom.tar.gz"},
		{"Resource with ext.ension", "\\\\?\\UNC\\ComputerName\\SharedFolder\\Resource with ext.ension"},
	}
	for _, test := range tests {
		assert.Equal(t, test.expect, Base(test.input))
	}
}

func TestDir(t *testing.T) {
	// Just to show how path.Dir() behaves:
	assert.Equal(t, ".", path.Dir(""))
	assert.Equal(t, ".", path.Dir("justafile.txt"))

	tests := []struct {
		expect, input string
	}{
		// Follow path.Dir() when it comes to empty directories:
		{".", ""},
		{".", "justafile.txt"},

		{"/", "/"},
		{"/", "/file-at-root"},
		{"C:", "C:\\file-at-root"},
		{"/Linux path", "/Linux path/with spaces.txt"},
		{"/Mixed path/with", "/Mixed path\\with/slash.txt"},
		{"C:/ünicode/is", "C:\\ünicode\\is\\awésom.tar.gz"},
		{"//SERVER/ünicode/is", "\\\\SERVER\\ünicode\\is\\awésom.tar.gz"},
		{"//?/UNC/ComputerName/SharedFolder", "\\\\?\\UNC\\ComputerName\\SharedFolder\\Resource"},
	}
	for _, test := range tests {
		assert.Equal(t,
			test.expect, Dir(test.input),
			"for input %q", test.input)
	}
}

func TestJoin(t *testing.T) {
	// Just to show how path.Join() behaves:
	assert.Equal(t, "", path.Join())
	assert.Equal(t, "", path.Join(""))
	assert.Equal(t, "", path.Join("", ""))
	assert.Equal(t, "a/b", path.Join("", "", "a", "", "b", ""))

	tests := []struct {
		expect string
		input  []string
	}{
		// Should behave the same as path.Join():
		{"", []string{}},
		{"", []string{""}},
		{"", []string{"", ""}},
		{"a/b", []string{"", "", "a", "", "b", ""}},

		{"/file-at-root", []string{"/", "file-at-root"}},
		{"C:/file-at-root", []string{"C:", "file-at-root"}},

		{"/Linux path/with spaces.txt", []string{"/Linux path", "with spaces.txt"}},
		{"C:/ünicode/is/awésom.tar.gz", []string{"C:\\ünicode", "is\\awésom.tar.gz"}},
		{"//SERVER/mount/dir/file.txt", []string{"\\\\SERVER", "mount", "dir", "file.txt"}},
		{"//?/UNC/ComputerName/SharedFolder/Resource", []string{"\\\\?\\UNC\\ComputerName", "SharedFolder\\Resource"}},
	}
	for _, test := range tests {
		assert.Equal(t,
			test.expect, Join(test.input...),
			"for input %q", test.input)
	}
}

func TestStem(t *testing.T) {
	tests := []struct {
		expect, input string
	}{
		{"", ""},
		{"stem", "stem.txt"},
		{"stem.tar", "stem.tar.gz"},
		{"file", "/path/to/file.txt"},
		{"file", "C:\\path\\to\\file.txt"},
		{"file", "C:\\path\\to/mixed/slashes/file.txt"},
		{"file", "C:\\path/to\\mixed/slashes\\file.txt"},
		{"Resource with ext", "\\\\?\\UNC\\ComputerName\\SharedFolder\\Resource with ext.ension"},
	}
	for _, test := range tests {
		assert.Equal(t,
			test.expect, Stem(test.input),
			"for input %q", test.input)
	}
}

func TestToNative_native_backslash(t *testing.T) {
	if filepath.Separator != '\\' {
		t.Skipf("skipping backslash-specific test on %q with path separator %q",
			runtime.GOOS, filepath.Separator)
	}

	tests := []struct {
		expect, input string
	}{
		{"", ""},
		{".", "."},
		{"\\some\\simple\\path", "/some/simple/path"},
		{"C:\\path\\to\\file.txt", "C:\\path\\to\\file.txt"},
		{"C:\\path\\to\\mixed\\slashes\\file.txt", "C:\\path\\to/mixed/slashes/file.txt"},
		{"\\\\?\\UNC\\ComputerName\\SharedFolder\\Resource with ext.ension",
			"\\\\?\\UNC\\ComputerName\\SharedFolder\\Resource with ext.ension"},
		{"\\\\?\\UNC\\ComputerName\\SharedFolder\\Resource with ext.ension",
			"//?/UNC/ComputerName/SharedFolder/Resource with ext.ension"},
	}
	for _, test := range tests {
		assert.Equal(t,
			test.expect, ToNative(test.input),
			"for input %q", test.input)
	}
}

func TestToNative_native_slash(t *testing.T) {
	if filepath.Separator != '/' {
		t.Skipf("skipping backslash-specific test on %q with path separator %q",
			runtime.GOOS, filepath.Separator)
	}

	tests := []struct {
		expect, input string
	}{
		{"", ""},
		{".", "."},
		{"/some/simple/path", "/some/simple/path"},
		{"C:/path/to/file.txt", "C:\\path\\to\\file.txt"},
		{"C:/path/to/mixed/slashes/file.txt", "C:\\path\\to/mixed/slashes/file.txt"},
		{"//?/UNC/ComputerName/SharedFolder/Resource with ext.ension",
			"\\\\?\\UNC\\ComputerName\\SharedFolder\\Resource with ext.ension"},
		{"//?/UNC/ComputerName/SharedFolder/Resource with ext.ension",
			"//?/UNC/ComputerName/SharedFolder/Resource with ext.ension"},
	}
	for _, test := range tests {
		assert.Equal(t,
			test.expect, ToNative(test.input),
			"for input %q", test.input)
	}
}

// This test should be skipped on every platform. It's there just to detect that
// the above two tests haven't run.
func TestToNative_unsupported(t *testing.T) {
	if filepath.Separator == '/' || filepath.Separator == '\\' {
		t.Skipf("skipping test on %q with path separator %q",
			runtime.GOOS, filepath.Separator)
	}

	t.Fatalf("ToNative not supported on this platform %q with path separator %q",
		runtime.GOOS, filepath.Separator)
}

func TestIsRoot(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{"empty", "", false},

		{"UNIX", "/", true},
		{"Drive only", "C:", true},
		{"Drive and slash", "C:/", true},
		{"Drive and backslash", `C:\`, true},

		{"backslash", `\`, false},
		{"just letters", "just letters", false},
		{"subdir of root", "/subdir", false},
		{"subdir of drive", "C:\\subdir", false},
		{"relative subdir of drive", "C:subdir", false},

		{"indirectly root", "/subdir/..", false},
		{"UNC notation", `\\NAS\Share\`, false},
		{"Slashed UNC notation", `//NAS/Share/`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRoot(tt.path); got != tt.want {
				t.Errorf("IsRoot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToPlatform(t *testing.T) {
	type args struct {
		path     string
		platform string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty-win", args{``, "windows"}, ``},
		{"empty-lnx", args{``, "linux"}, ``},
		{"single-win", args{`path with spaces`, "windows"}, `path with spaces`},
		{"single-lnx", args{`path with spaces`, "linux"}, `path with spaces`},
		{"native-win", args{`native\path`, "windows"}, `native\path`},
		{"native-lnx", args{`native/path`, "linux"}, `native/path`},
		{"opposite-win", args{`opposite/path`, "windows"}, `opposite\path`},
		{"opposite-lnx", args{`opposite\path`, "linux"}, `opposite/path`},
		{"mixed-win", args{`F:/mixed/path\to\file.blend`, "windows"}, `F:\mixed\path\to\file.blend`},
		{"mixed-lnx", args{`F:/mixed/path\to\file.blend`, "linux"}, `F:/mixed/path/to/file.blend`},
		{"absolute-win", args{`F:/absolute/path`, "windows"}, `F:\absolute\path`},
		{"absolute-lnx", args{`/absolute/path`, "linux"}, `/absolute/path`},
		{"relative-win", args{`/absolute/path`, "windows"}, `\absolute\path`},

		// Trailing path separators should not be removed if it's only a drive
		// letter, as concatenation rules are tricky there. `F:path` is not the same
		// as `F:\path`.
		{"drive-root-win", args{`F:\`, "windows"}, `F:\`},
		{"trailing-win", args{`F:\directory\`, "linux"}, `F:/directory`},
		{"trailing-lnx", args{`/dir/path/`, "windows"}, `\dir\path`},

		// UNC notation should survive, even when it no longer makes sense (like on Linux).
		{"unc-win", args{`\\NAS\share\path`, "windows"}, `\\NAS\share\path`},
		{"unc-lnx", args{`\\NAS\share\path`, "linux"}, `//NAS/share/path`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToPlatform(tt.args.path, tt.args.platform); got != tt.want {
				t.Errorf("ToPlatform() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrimTrailingSep(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty", args{""}, ""},
		{"single-slash", args{`/`}, `/`},
		{"single-backslash", args{`\`}, `\`},
		{"multiple-slashes", args{`///`}, `/`},
		{"multiple-backslashes", args{`\\\`}, `\`},
		{"multiple-mixed-slash-first", args{`/\\`}, `/`},
		{"multiple-mixed-backslash-first", args{`\//`}, `\`},
		{"nothing-trailing", args{`nothing/trailing\here`}, `nothing/trailing\here`},
		{"trailing-space", args{`trailing/space/ `}, `trailing/space/ `},
		{"trailing-slash", args{`trailing/slash/`}, `trailing/slash`},
		{"trailing-slashes", args{`trailing/slashes///`}, `trailing/slashes`},
		{"trailing-backslash", args{`trailing\backslash\`}, `trailing\backslash`},
		{"trailing-backslashes", args{`trailing\backslashes\\\`}, `trailing\backslashes`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TrimTrailingSep(tt.args.path); got != tt.want {
				t.Errorf("TrimTrailingSep() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnsureDriveAbsolute(t *testing.T) {
	tests := []struct {
		name      string
		inputPath string
		want      string
	}{
		// Windows paths expected to change:
		{"windows-drive-relative", `F:path\to\file`, `F:\path\to\file`},
		{"windows-drive-relative-mixed", "F:path/to/file", `F:\path/to/file`},
		{"windows-drive-only", "F:", `F:\`},

		// No-op paths:
		{"empty", "", ""},
		{"linux-root", "/", "/"},
		{"linux-path", "/some/path", "/some/path"},
		{"linux-unicode-path", "/söme/path", "/söme/path"},
		{"one-letter", "F", "F"},
		{"windows-drive-invalid", `©:path\to\thing`, `©:path\to\thing`},
		{"windows-unc", `\\NAS\Flamenco\path`, `\\NAS\Flamenco\path`},
		{"unicode-one-letter", "€", "€"},
		{"windows-drive-absolute", `F:\some\path`, `F:\some\path`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EnsureDriveAbsolute(tt.inputPath); got != tt.want {
				t.Errorf("EnsureDriveAbsolute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsUNCNotation(t *testing.T) {
	assert.True(t, IsUNCNotation(`\\SERVER\Share\path\file.txt`))
	assert.True(t, IsUNCNotation(`\\.\NUL`))
	assert.True(t, IsUNCNotation(`\\?\UNC\localhost\c$\temp\foo.txt`))
	assert.True(t, IsUNCNotation(`\\`)) // Dubious, not sure this is a valid path.
	assert.False(t, IsUNCNotation(`//blendfile/relative.path`))
	assert.False(t, IsUNCNotation(`C:\normal\drive\path`))
	assert.False(t, IsUNCNotation(`/posix/path.txt`))
	assert.False(t, IsUNCNotation(``))
	assert.False(t, IsUNCNotation(`\`))
}
