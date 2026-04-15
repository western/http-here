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
	"time"

	"github.com/western/http-here/v2/internal/conf"
)

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

func CleanDirtyPath(p string) string {

	re := regexp.MustCompile("/+")
	p = re.ReplaceAllLiteralString(p, "/")

	re2 := regexp.MustCompile("\\.{2,}")
	p = re2.ReplaceAllLiteralString(p, ".")

	//p = filepath.Clean(p)

	return p
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
