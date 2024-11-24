package controller

import (
	"archive/zip"
	"fmt"
	"io"
	_ "io/ioutil"
	"math"
	"math/rand"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/gofiber/fiber/v2"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func RandStringRunes(n int) string {
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
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

func LogPrefix(c *fiber.Ctx, status string, addition string) {

	green := color.New(color.FgGreen).SprintFunc()
	//magenta := color.New(color.FgMagenta).SprintFunc()
	cian := color.New(color.FgCyan).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	pref := ""

	pid := strconv.Itoa(os.Getpid())

	pref += "[" + pid + "] "

	pref += "[" + green(time.Now().Format("2006-01-02 15:04:05")) + "] "

	pref += "[" + green(c.IP()) + "] "

	_user := ""
	if c.Locals("username") != nil {
		_user = green(c.Locals("username").(string))
	}

	if len(_user) > 0 {
		pref += "[" + _user + "] "
	} else {
		pref += "[] "
	}

	if status != "200" {
		pref += "[" + yellow(status) + "] "
	} else {
		pref += "[" + status + "] "
	}

	pref += cian(addition)

	fmt.Println(pref)

	// [2024-10-29 18:22:46] [192.168.0.101] [loginT] [200] Dir /tmp/fold1/fold3
	//fmt.Printf("[%s] [%s] [%s] [%s] %s\n", magenta("warning"), red("error"))

}

func GetExt(path string) string {

	ext := filepath.Ext(path)
	ext = strings.ToLower(ext)
	ext = strings.Replace(ext, ".", "", -1)

	return ext
}

func prettyByteSize(b int64) string {
	bf := float64(b)
	for _, unit := range []string{"", "Ki", "Mi", "Gi", "Ti", "Pi", "Ei", "Zi"} {
		if math.Abs(bf) < 1024.0 {
			return fmt.Sprintf("%3.1f %sB", bf, unit)
		}
		bf /= 1024.0
	}
	return fmt.Sprintf("%.1fYiB", bf)
}

func addFilesToZip(w *zip.Writer, basePath, baseInZip string) {
	// Open the Directory
	//files, err := ioutil.ReadDir(basePath)
	files, err := os.ReadDir(basePath)
	if err != nil {
		fmt.Println(err)
	}

	for _, file := range files {
		//fmt.Println(  filepath.Join(basePath, file.Name())   )
		if !file.IsDir() {
			//dat, err := ioutil.ReadFile(basePath + file.Name())
			dat, err := os.ReadFile(filepath.Join(basePath, file.Name()))
			if err != nil {
				fmt.Println(err)
			}

			// Add some files to the archive.
			f, err := w.Create(filepath.Join(baseInZip, file.Name()))
			if err != nil {
				fmt.Println(err)
			}
			_, err = f.Write(dat)
			if err != nil {
				fmt.Println(err)
			}
		} else if file.IsDir() {

			// Recurse
			//newBase := basePath + file.Name() + "/"
			newBase := filepath.Join(basePath, file.Name()) + "/"
			//fmt.Println("Recursing and Adding SubDir: " + file.Name())
			//fmt.Println("Recursing and Adding SubDir: " + newBase)

			//addFiles(w, newBase, baseInZip+file.Name()+"/")
			addFilesToZip(w, newBase, filepath.Join(baseInZip, file.Name())+"/")
		}
	}
}

func WalkAndClear(path string) {

	files, err := os.ReadDir(path)
	if err != nil {
		fmt.Println(err)
	}

	for _, file := range files {

		full_name := filepath.Join(path, file.Name())

		if !file.IsDir() {

			fileInfo, _ := os.Stat(full_name)

			if time.Now().Sub(fileInfo.ModTime()) > 30*24*time.Hour {
				//fmt.Println( "to del:", full_name )
				os.Remove(full_name)
			} else {
				//fmt.Println( "file:", full_name )
			}

		} else if file.IsDir() {
			WalkAndClear(full_name)
		}
	}
}

type TreeRow struct {
	Text string `json:"text"`
	Path string `json:"path"`

	Expanded bool `json:"expanded"`

	Nodes []TreeRow `json:"nodes"`
}

func WalkAndTreeBuild(path string, prev_path string, deep int) []TreeRow {

	//fmt.Println("WalkAndTreeBuild ", path)

	var node_list []TreeRow

	files, err := os.ReadDir(path)
	if err != nil {
		//fmt.Println(err)
		return node_list
	}

	if deep > 5 {
		return node_list
	}

	for _, file := range files {

		if !file.IsDir() {

			/*
				        node_list = append( node_list, TreeRow{

							Text:         file.Name(),

						})
			*/

		} else if file.IsDir() {

			nodes := WalkAndTreeBuild(filepath.Join(path, file.Name()), filepath.Join(prev_path, file.Name()), deep+1)

			node_list = append(node_list, TreeRow{

				Text: file.Name(),
				Path: filepath.Join(prev_path, file.Name()),

				Nodes: nodes,
			})

		}
	}

	return node_list
}

func MultipartToFile(file *multipart.FileHeader) *os.File {

	f, err := os.CreateTemp("", "becloud_convert*")
	fmt.Println("MultipartToFile Temp file name:", f.Name())
	defer os.Remove(f.Name())

	readerFile, _ := file.Open()
	_, err = io.Copy(f, readerFile)
	if err != nil {
		//return false, err
		panic(err)
	}
	f.Close()

	file2, _ := os.Open(f.Name())

	return file2
}
