package util

import (
	"bufio"
	"encoding/json"
	//"errors"
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	//"reflect"
	"regexp"
	"strconv"
	"strings"
	//"sync"
	"context"
	"net/url"
	"time"

	"github.com/western/http-here/v2/internal/conf"
)

// transparent-png24-1px.png
// PNG image data, 1 x 1, 8-bit/color RGBA, non-interlaced
var PNGPixel = []byte{
	137, 80, 78, 71, 13, 10, 26, 10, 0, 0, 0, 13, 73, 72, 68, 82, 0, 0, 0, 1, 0, 0, 0, 1, 8, 6, 0, 0, 0, 31, 21, 196, 137, 0, 0, 0, 25, 116, 69, 88, 116, 83, 111, 102, 116, 119, 97, 114, 101, 0, 65, 100, 111, 98, 101,
	32, 73, 109, 97, 103, 101, 82, 101, 97, 100, 121, 113, 201, 101, 60, 0, 0, 3, 34, 105, 84, 88, 116, 88, 77, 76, 58, 99, 111, 109, 46, 97, 100, 111, 98, 101, 46, 120, 109, 112, 0, 0, 0, 0, 0, 60, 63, 120, 112,
	97, 99, 107, 101, 116, 32, 98, 101, 103, 105, 110, 61, 34, 239, 187, 191, 34, 32, 105, 100, 61, 34, 87, 53, 77, 48, 77, 112, 67, 101, 104, 105, 72, 122, 114, 101, 83, 122, 78, 84, 99, 122, 107, 99, 57, 100,
	34, 63, 62, 32, 60, 120, 58, 120, 109, 112, 109, 101, 116, 97, 32, 120, 109, 108, 110, 115, 58, 120, 61, 34, 97, 100, 111, 98, 101, 58, 110, 115, 58, 109, 101, 116, 97, 47, 34, 32, 120, 58, 120, 109, 112,
	116, 107, 61, 34, 65, 100, 111, 98, 101, 32, 88, 77, 80, 32, 67, 111, 114, 101, 32, 53, 46, 51, 45, 99, 48, 49, 49, 32, 54, 54, 46, 49, 52, 53, 54, 54, 49, 44, 32, 50, 48, 49, 50, 47, 48, 50, 47, 48, 54, 45, 49,
	52, 58, 53, 54, 58, 50, 55, 32, 32, 32, 32, 32, 32, 32, 32, 34, 62, 32, 60, 114, 100, 102, 58, 82, 68, 70, 32, 120, 109, 108, 110, 115, 58, 114, 100, 102, 61, 34, 104, 116, 116, 112, 58, 47, 47, 119, 119, 119,
	46, 119, 51, 46, 111, 114, 103, 47, 49, 57, 57, 57, 47, 48, 50, 47, 50, 50, 45, 114, 100, 102, 45, 115, 121, 110, 116, 97, 120, 45, 110, 115, 35, 34, 62, 32, 60, 114, 100, 102, 58, 68, 101, 115, 99, 114, 105,
	112, 116, 105, 111, 110, 32, 114, 100, 102, 58, 97, 98, 111, 117, 116, 61, 34, 34, 32, 120, 109, 108, 110, 115, 58, 120, 109, 112, 61, 34, 104, 116, 116, 112, 58, 47, 47, 110, 115, 46, 97, 100, 111, 98, 101,
	46, 99, 111, 109, 47, 120, 97, 112, 47, 49, 46, 48, 47, 34, 32, 120, 109, 108, 110, 115, 58, 120, 109, 112, 77, 77, 61, 34, 104, 116, 116, 112, 58, 47, 47, 110, 115, 46, 97, 100, 111, 98, 101, 46, 99, 111,
	109, 47, 120, 97, 112, 47, 49, 46, 48, 47, 109, 109, 47, 34, 32, 120, 109, 108, 110, 115, 58, 115, 116, 82, 101, 102, 61, 34, 104, 116, 116, 112, 58, 47, 47, 110, 115, 46, 97, 100, 111, 98, 101, 46, 99, 111,
	109, 47, 120, 97, 112, 47, 49, 46, 48, 47, 115, 84, 121, 112, 101, 47, 82, 101, 115, 111, 117, 114, 99, 101, 82, 101, 102, 35, 34, 32, 120, 109, 112, 58, 67, 114, 101, 97, 116, 111, 114, 84, 111, 111, 108,
	61, 34, 65, 100, 111, 98, 101, 32, 80, 104, 111, 116, 111, 115, 104, 111, 112, 32, 67, 83, 54, 32, 40, 87, 105, 110, 100, 111, 119, 115, 41, 34, 32, 120, 109, 112, 77, 77, 58, 73, 110, 115, 116, 97, 110,
	99, 101, 73, 68, 61, 34, 120, 109, 112, 46, 105, 105, 100, 58, 51, 49, 56, 70, 48, 54, 66, 48, 68, 49, 68, 52, 49, 49, 70, 49, 66, 55, 50, 52, 69, 69, 53, 57, 69, 68, 67, 54, 65, 57, 52, 53, 34, 32, 120, 109,
	112, 77, 77, 58, 68, 111, 99, 117, 109, 101, 110, 116, 73, 68, 61, 34, 120, 109, 112, 46, 100, 105, 100, 58, 51, 49, 56, 70, 48, 54, 66, 49, 68, 49, 68, 52, 49, 49, 70, 49, 66, 55, 50, 52, 69, 69, 53, 57, 69,
	68, 67, 54, 65, 57, 52, 53, 34, 62, 32, 60, 120, 109, 112, 77, 77, 58, 68, 101, 114, 105, 118, 101, 100, 70, 114, 111, 109, 32, 115, 116, 82, 101, 102, 58, 105, 110, 115, 116, 97, 110, 99, 101, 73, 68, 61,
	34, 120, 109, 112, 46, 105, 105, 100, 58, 51, 49, 56, 70, 48, 54, 65, 69, 68, 49, 68, 52, 49, 49, 70, 49, 66, 55, 50, 52, 69, 69, 53, 57, 69, 68, 67, 54, 65, 57, 52, 53, 34, 32, 115, 116, 82, 101, 102, 58, 100,
	111, 99, 117, 109, 101, 110, 116, 73, 68, 61, 34, 120, 109, 112, 46, 100, 105, 100, 58, 51, 49, 56, 70, 48, 54, 65, 70, 68, 49, 68, 52, 49, 49, 70, 49, 66, 55, 50, 52, 69, 69, 53, 57, 69, 68, 67, 54, 65, 57,
	52, 53, 34, 47, 62, 32, 60, 47, 114, 100, 102, 58, 68, 101, 115, 99, 114, 105, 112, 116, 105, 111, 110, 62, 32, 60, 47, 114, 100, 102, 58, 82, 68, 70, 62, 32, 60, 47, 120, 58, 120, 109, 112, 109, 101, 116, 97,
	62, 32, 60, 63, 120, 112, 97, 99, 107, 101, 116, 32, 101, 110, 100, 61, 34, 114, 34, 63, 62, 10, 212, 43, 91, 0, 0, 0, 16, 73, 68, 65, 84, 120, 218, 98, 248, 255, 255, 63, 3, 64, 128, 1, 0, 8, 252, 2, 254, 219,
	162, 77, 22, 0, 0, 0, 0, 73, 69, 78, 68, 174, 66, 96, 130,
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func RandStringRunes(n int) string {
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
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

// change all slashes \ to /
func RotateSlash(p string) string {
	return strings.Replace(p, "\\", "/", -1)
}

// get ext and normalize
func GetExtNorm(path string) string {

	ext := filepath.Ext(path)
	ext = strings.ToLower(ext)
	ext = strings.Replace(ext, ".", "", -1)
	if ext == "jpeg" {
		ext = "jpg"
	}

	return ext
}

/*
fileName := util.GetFileName("image1.jpg")
fmt.Println("fileName=", fileName)
fileName = util.GetFileName("/path/to/image2.jpg")
fmt.Println("fileName=", fileName)
fileName = util.GetFileName(`\path\to\ima  ge3.jpg`)
fmt.Println("fileName=", fileName)
*/
func GetFileName(path string) string {

	filename := strings.TrimSuffix(
		filepath.Base(path),
		filepath.Ext(path),
	)

	return filename
}

func PrettyByteSize(b int64) string {
	bf := float64(b)
	for _, unit := range []string{"", "Ki", "Mi", "Gi", "Ti", "Pi", "Ei", "Zi"} {
		if math.Abs(bf) < 1024.0 {
			return fmt.Sprintf("%3.1f %sB", bf, unit)
		}
		bf /= 1024.0
	}
	return fmt.Sprintf("%.1fYiB", bf)
}

func RunAnyCommandUnderWin(longCmdString string) error {

	filepath_tmp := path.Join(conf.ConfigRoot, "temp")

	pid := strconv.Itoa(os.Getpid())
	cmd_filename := path.Join(filepath_tmp, "run"+pid+".cmd")
	cmd_filename = RotateSlash(cmd_filename)

	myf, err := os.Create(cmd_filename)
	if err != nil {
		return err
	}

	//myf.WriteString("@echo off\r\n")
	myf.WriteString("chcp 65001\r\n")

	myf.WriteString(longCmdString)
	myf.WriteString("\r\n")
	myf.Close()

	// ----------------------------------------------------------------------------------------------------------------------------------

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "cmd", "/C", cmd_filename)

	//stdout, _ := cmd.StdoutPipe()

	stdout, _ := cmd.StderrPipe()
	//cmd.Stdout = cmd.Stderr   // combine ERR + OUT

	if err := cmd.Start(); err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			//fmt.Println("RunAnyCommandUnderWin output:", scanner.Text())
		}
	}()

	if err := cmd.Wait(); err != nil {
		//fmt.Println("RunAnyCommandUnderWin finished with error:", err)
	}

	return nil
}

func PrintPrettify(prefix string, payload interface{}) {

	s, _ := json.MarshalIndent(payload, "", "\t")
	fmt.Println(prefix, "= ", string(s))
}

// /[\u001b\u009b][[()#;?]*(?:[0-9]{1,4}(?:;[0-9]{0,4})*)?[0-9A-ORZcf-nqry=><]/g, ”
var regexpAnsiColor = regexp.MustCompile("[\u001b\u009b][[()#;?]*(?:[0-9]{1,4}(?:;[0-9]{0,4})*)?[0-9A-ORZcf-nqry=><]")

func StringClearColor(msg string) string {
	return regexpAnsiColor.ReplaceAllString(msg, "")
}

func IsInt(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}
