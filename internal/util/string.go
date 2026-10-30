package util

import (
	"net/url"
	"regexp"
	"strings"
)

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
