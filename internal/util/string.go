package util

import (
	"errors"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	//"fmt"

	"github.com/gofiber/fiber/v3"
)

func CleanPath(name string) (string, error) {

	//fmt.Println("CleanPath [", name, "]")

	if name == "" ||
		//filepath.IsAbs(name) ||
		strings.ContainsRune(name, '\x00') {
		return "", errors.New("Invalid path (1)")
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
		return "", errors.New("Invalid path (2)")
	}

	return name, nil
}

func RefererOrPath(c fiber.Ctx) (string, error) {

	referer := c.Get("Referer")

	// already decoded
	u, err := url.Parse(referer)
	if err != nil {
		return "", err
	}

	//url_Path := util.CleanDirtyPath(u.Path)
	url_Path, err := CleanPath(u.Path)
	if err != nil {
		return "", err
	}

	// -------------------------------------------------------------------------------------------------------------------------

	//form_Path := c.FormValue("path")
	//form_Path = CleanDirtyPath(form_Path)

	form_Path, _ := CleanPath(c.FormValue("path"))

	if len(form_Path) > 0 {
		url_Path = form_Path
	}

	return url_Path, nil
}

/*
fmt.Println("../../etc/passwd", util.CleanDirtyPath("../../etc/passwd"));
fmt.Println("..%2f..%2fsecret.txt", util.CleanDirtyPath("..%2f..%2fsecret.txt"));
fmt.Println("..\\..\\secret.txt", util.CleanDirtyPath("..\\..\\secret.txt"));
*/
func CleanDirtyPath(dirty_Str string) string {

	// double slash
	re := regexp.MustCompile("/+")
	dirty_Str = re.ReplaceAllLiteralString(dirty_Str, "/")

	// windows separators
	re2 := regexp.MustCompile("\\+")
	dirty_Str = re2.ReplaceAllLiteralString(dirty_Str, "")

	// points, from two
	re3 := regexp.MustCompile("\\.{2,}")
	dirty_Str = re3.ReplaceAllLiteralString(dirty_Str, ".")

	//dirty_Str = filepath.Clean(dirty_Str)

	return dirty_Str
}

/*
fileName, err := util.GetClearFilename("image1.jpg")
fmt.Println("fileName=", fileName, " err=",err)
fileName, err = util.GetClearFilename("/path/to/image2.jpg")
fmt.Println("fileName=", fileName, " err=",err)
fileName, err = util.GetClearFilename(`\path\to\ima  ge3.jpg`)
fmt.Println("fileName=", fileName, " err=",err)
*/
func GetClearFilename(dirty_Filename string) (string, error) {

	var (
		regexpSpaceSymbols = regexp.MustCompile("\\s+")
		//regexpDoubleDash   = regexp.MustCompile("[\\-]{2,}")
	)

	unescape_Filename, err := url.QueryUnescape(dirty_Filename)
	if err != nil {

		return "", err
	}

	file_ext := GetExtNorm(unescape_Filename)
	originalFileName := GetFileName(unescape_Filename)

	originalFileName = strings.ReplaceAll(originalFileName, `/`, "")
	originalFileName = strings.ReplaceAll(originalFileName, `\`, "")
	originalFileName = regexpSpaceSymbols.ReplaceAllLiteralString(originalFileName, " ")
	//originalFileName = regexpDoubleDash.ReplaceAllLiteralString(originalFileName, "-")

	filename := originalFileName + "." + file_ext
	filename = CleanDirtyPath(filename)

	return filename, nil
}
