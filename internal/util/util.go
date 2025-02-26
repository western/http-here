package util

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"mime/multipart"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	_ "strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/fatih/color"
	_ "github.com/gofiber/fiber/v2"

	"crypto/md5"
	_ "encoding/hex"
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

/*
func LogPrefix(c *fiber.Ctx, status, msg string) {

	green_clr := color.New(color.FgGreen).SprintFunc()
	red_clr := color.New(color.FgRed).SprintFunc()

	pref := ""

	if status != "200" {
		pref += red_clr("| ")
	} else {
		pref += green_clr("| ")
	}

	pid := strconv.Itoa(os.Getpid())

	pref += "[" + pid + "] "

	pref += "[" + time.Now().Format("2006-01-02 15:04:05.000") + "] "

	pref += "[" + c.IP() + "] "

	_user := ""
	if c.Locals("username") != nil {
		_user = green_clr(c.Locals("username").(string))
	}

	if len(_user) > 0 {
		pref += "[" + _user + "] "
	} else {
		pref += "[] "
	}

	if status != "200" {
		pref += "[" + red_clr(status) + "] "
	} else {
		pref += "[" + green_clr(status) + "] "
	}

	//pref += "[" + tag + "] "

	pref += msg

	fmt.Println(pref)
}
*/

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

func CopyFile(src, dst string) error {
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	if srcfd, err = os.Open(src); err != nil {
		return err
	}
	defer srcfd.Close()

	if dstfd, err = os.Create(dst); err != nil {
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}
	return os.Chmod(dst, srcinfo.Mode())
}

func MoveFile(src, dst string) error {
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	if srcfd, err = os.Open(src); err != nil {
		return err
	}
	defer srcfd.Close()

	if dstfd, err = os.Create(dst); err != nil {
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	srcfd.Close()
	dstfd.Close()
	if srcinfo, err = os.Stat(src); err != nil {
		return err
	} else {

		if err = os.Remove(src); err != nil {
			return err
		}
	}

	return os.Chmod(dst, srcinfo.Mode())
}

func CopyDir(src string, dst string) error {
	var err error
	var fds []os.FileInfo
	var srcinfo os.FileInfo

	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}

	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}

	if fds, err = ioutil.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = CopyDir(srcfp, dstfp); err != nil {
				fmt.Println(err)

			}
		} else {
			if err = CopyFile(srcfp, dstfp); err != nil {
				fmt.Println(err)

			}
		}
	}
	return nil
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

func GetMd5File(path string) string {

	f, err := os.Open(path)
	if err != nil {
		//panic(err)
		fmt.Println(err)
		return ""
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		//panic(err)
		fmt.Println(err)
		return ""
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

/*
func addFilesToZip(w *zip.Writer, basePath, baseInZip string) {
	// Open the Directory
	//files, err := ioutil.ReadDir(basePath)
	files, err := os.ReadDir(basePath)
	if err != nil {
		fmt.Println(err)
	}

	for _, file := range files {
		//fmt.Println(  path.Join(basePath, file.Name())   )
		if !file.IsDir() {
			//dat, err := ioutil.ReadFile(basePath + file.Name())
			dat, err := os.ReadFile(path.Join(basePath, file.Name()))
			if err != nil {
				fmt.Println(err)
			}

			fileInfo, err := os.Stat(path.Join(basePath, file.Name()))
			if err != nil {
				fmt.Println(err)
				//LogPrefix(c, "500", "'"+path.Join(basePath, file.Name())+"' not exists")
				//continue
			}

			header, err := zip.FileInfoHeader(fileInfo)
			if err != nil {
				fmt.Println(err)
				//LogPrefix(c, "500", "'"+path.Join(arg_fold, u_path, name)+"' err: "+err.Error())
				//continue
			}
			header.Method = zip.Store
			header.Name = path.Join(baseInZip, file.Name())

			// Add some files to the archive.
			//f, err := w.Create(path.Join(baseInZip, file.Name()))
			f, err := w.CreateHeader(header)
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
			newBase := path.Join(basePath, file.Name()) + "/"
			//fmt.Println("Recursing and Adding SubDir: " + file.Name())
			//fmt.Println("Recursing and Adding SubDir: " + newBase)

			//addFiles(w, newBase, baseInZip+file.Name()+"/")
			addFilesToZip(w, newBase, path.Join(baseInZip, file.Name())+"/")
		}
	}
}
*/

func RunAnyCommandUnderWin(c string) error {

	homepath, err := os.UserHomeDir()
	if err != nil {
		return errors.New("homepath detect error")
	}

	filepath_tmp := path.Join(homepath, ".httphere", "temp")

	if _, err := os.Stat(path.Join(homepath, ".httphere", "temp")); err != nil {
		if err := os.MkdirAll(path.Join(homepath, ".httphere", "temp"), os.ModePerm); err != nil {
			return err
		}
	}

	pid := strconv.Itoa(os.Getpid())
	cmd_filename := path.Join(filepath_tmp, "run"+pid+".cmd")

	myf, err := os.Create(cmd_filename)
	if err != nil {
		return err
	}

	myf.WriteString("@echo off\r\n")
	myf.WriteString("chcp 65001\r\n")

	myf.WriteString(c)
	myf.WriteString("\r\n")
	myf.Close()

	cmd := exec.Command(cmd_filename)

	stderr, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {

		fmt.Println("RunAnyCommandUnderWin 1:", err)
		//panic(err)
	}
	//fmt.Println("stderr:", stderr)

	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		fmt.Println("RunAnyCommandUnderWin 2:", scanner.Text())
		//panic("xxx")
	}

	return nil
}

func WalkAndClearOld(path_ string) {

	files, err := os.ReadDir(path_)
	if err != nil {
		fmt.Println(err)
	}

	for _, file := range files {

		full_name := path.Join(path_, file.Name())

		if !file.IsDir() {

			fileInfo, _ := os.Stat(full_name)

			if time.Now().Sub(fileInfo.ModTime()) > 30*24*time.Hour {
				//fmt.Println( "to del:", full_name )
				os.Remove(full_name)
			} else {
				//fmt.Println( "file:", full_name )
			}

		} else if file.IsDir() {
			WalkAndClearOld(full_name)
		}
	}
}

type TreeRow struct {
	Text string `json:"text"`
	Path string `json:"path"`

	Expanded bool `json:"expanded"`

	Nodes []TreeRow `json:"nodes"`
}

func WalkAndTreeBuild(path_ string, prev_path string, deep int) []TreeRow {

	//fmt.Println("WalkAndTreeBuild ", path)

	var node_list []TreeRow

	files, err := os.ReadDir(path_)
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

			nodes := WalkAndTreeBuild(path.Join(path_, file.Name()), path.Join(prev_path, file.Name()), deep+1)

			node_list = append(node_list, TreeRow{

				Text: file.Name(),
				Path: path.Join(prev_path, file.Name()),

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

func WalkAndClearZeroFile(path_ string, deep int) {

	files, err := os.ReadDir(path_)
	if err != nil {
		return
	}

	if deep > 5 {
		return
	}

	for _, file := range files {

		if !file.IsDir() {

			if NewFileInfo, err := os.Stat(path.Join(path_, file.Name())); err == nil {
				if NewFileInfo.Size() == 0 {

					//fmt.Println("WalkAndClearZeroFile", "remove=", path.Join(path_, file.Name())    )

					if err2 := os.Remove(path.Join(path_, file.Name())); err2 != nil {
						panic("Problem of remove zero file " + path.Join(path_, file.Name()) + " " + err2.Error())
					}
				}
			}

		} else if file.IsDir() {

			WalkAndClearZeroFile(path.Join(path_, file.Name()), deep+1)

		}
	}

}

/*
func WalkAndMakeThumbnail(path string, deep int) {

	//fmt.Println("WalkAndMakeThumbnail ", path)

	files, err := os.ReadDir(path)
	if err != nil {
		return
	}

	if deep > 5 {
		return
	}

	homepath, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	for _, file := range files {

		fileInfo, err := os.Stat(path.Join(path, file.Name()))

		modtime := fileInfo.ModTime()
		modtime_human := modtime.Format("2006-01-02 15:04:05")

		size := fileInfo.Size()
		size_human := PrettyByteSize(size)

		//fmt.Println("runtime.NumGoroutine=", runtime.NumGoroutine())

		if !file.IsDir() {

			c_width := "600"
			i_width := 600

			file_ext := GetExtNorm(file.Name())
			orig_filename := GetFileName(file.Name())

			is_preview_match, _ := regexp.MatchString("^(jpg|jpeg|png|gif)$", file_ext)
			if !is_preview_match {
				continue
			}

			hash_name := md5.Sum([]byte(orig_filename + modtime_human + size_human + c_width))
			hex_name := hex.EncodeToString(hash_name[:])

			if _, err := os.Stat(path.Join(homepath, ".httphere", "thumb", hex_name)); err == nil {
				//fmt.Println("Is exists ", path.Join(homepath, ".httphere", "thumb", hex_name))
				continue
			}

			go func() {

				input, _ := os.Open(path.Join(path, file.Name()))
				defer input.Close()

				output, _ := os.Create(path.Join(homepath, ".httphere", "thumb", hex_name))
				defer output.Close()

				var src image.Image

				// Decode the image (from PNG to image.Image):
				if file_ext == "png" {
					src, err = png.Decode(input)
					if err != nil {
						//log.Fatal(err)
						panic(err)
					}
				}

				if file_ext == "jpg" {

					// src, err = jpeg.Decode(input)
					src, _, err = exiffix.Decode(input)
					if err != nil {
						//log.Fatal(err)
						panic(err)
					}
				}

				if file_ext == "gif" {
					src, err = gif.Decode(input)
					if err != nil {
						//log.Fatal(err)
						panic(err)
					}
				}

				ratio := (float64)(src.Bounds().Max.Y) / (float64)(src.Bounds().Max.X)
				i_height := int(math.Round(float64(i_width) * ratio))

				var dst *image.RGBA

				if src.Bounds().Max.X > i_width || src.Bounds().Max.Y > i_height {
					dst = image.NewRGBA(image.Rect(0, 0, i_width, i_height))
				} else {

					err := os.Remove(path.Join(homepath, ".httphere", "thumb", hex_name))
					if err != nil {
						//log.Fatal(err)
						panic(err)
					}

					//fmt.Println("Too small for thumb make ", path.Join(path, file.Name()))
					return
				}

				// Resize:
				draw.NearestNeighbor.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)

				if file_ext == "png" {
					err = png.Encode(output, dst)
					if err != nil {
						//log.Fatal(err)
						panic(err)
					}
				}

				if file_ext == "jpg" {
					err = jpeg.Encode(output, dst, nil)
					if err != nil {
						//log.Fatal(err)
						panic(err)
					}
				}

				if file_ext == "gif" {
					err = gif.Encode(output, dst, nil)
					if err != nil {
						//log.Fatal(err)
						panic(err)
					}
				}

				output.Close()
				input.Close()

				src = nil
				dst = nil

				//fmt.Println("Make thumb for ", path.Join(path, file.Name()))

				//fmt.Println("runtime.NumGoroutine=", runtime.NumGoroutine())

			}()

			//fmt.Println("Not wait for ", path.Join(path, file.Name()))

		} else if file.IsDir() {

			WalkAndMakeThumbnail(path.Join(path, file.Name()), deep+1)

		}
	}

	return
}
*/

// openssl aes-256-cbc -a -salt -in file.txt -out file.txt.cr -pass pass:123
func CryptFile(path_, pass string) (bool, error) {

	_, err := exec.Command("bash", "-c", "openssl --help").Output()
	if err != nil {
		return false, err
	}

	if len(path_) == 0 {
		return false, errors.New("Path of file is empty")
	}

	if len(pass) == 0 {
		return false, errors.New("Pass is empty")
	}

	from_file := filepath.Base(path_)
	to_file := filepath.Base(path_) + ".cr"

	cmd := exec.Command("bash", "-c", "openssl aes-256-cbc -a -salt -in "+from_file+" -out "+to_file+" -pass pass:"+pass)
	cmd.Dir = filepath.Dir(path_)
	//out, _ := cmd.Output()
	//_ = out

	stderr, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
		//panic(err)
		return false, err
	}

	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		//fmt.Println("openssl:", scanner.Text())
	}

	//fmt.Println("--------------------------------------------------------------------------------------------------")
	//fmt.Println("out=", out)

	if err = os.Remove(path_); err != nil {
		return false, err
	}

	if err = os.Rename(path.Join(filepath.Dir(path_), to_file), path_); err != nil {
		return false, err
	}

	return true, nil
}

// openssl aes-256-cbc -d -a -in file.txt.cr -out file.txt.new -pass pass:123
func DecryptFile(path_, pass string) (bool, error) {

	_, err := exec.Command("bash", "-c", "openssl --help").Output()
	if err != nil {
		return false, err
	}

	if len(path_) == 0 {
		return false, errors.New("Path of file is empty")
	}

	if len(pass) == 0 {
		return false, errors.New("Pass is empty")
	}

	from_file := filepath.Base(path_)
	to_file := filepath.Base(path_)
	to_file = strings.Replace(to_file, ".cr", "", 1)
	to_file += ".decrypt"

	cmd := exec.Command("bash", "-c", "openssl aes-256-cbc -d -a  -in "+from_file+" -out "+to_file+" -pass pass:"+pass)
	cmd.Dir = filepath.Dir(path_)
	//out, _ := cmd.Output()
	//_ = out

	stderr, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
		//panic(err)
		return false, err
	}

	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		//fmt.Println("openssl:", scanner.Text())
	}

	//fmt.Println("--------------------------------------------------------------------------------------------------")
	//fmt.Println("out=", out)

	if err = os.Remove(path_); err != nil {
		return false, err
	}

	if err = os.Rename(path.Join(filepath.Dir(path_), to_file), path_); err != nil {
		return false, err
	}

	return true, nil
}

func MutexUnLocked(m *sync.Mutex) bool {
	state := reflect.ValueOf(m).Elem().FieldByName("state")
	//fmt.Println("state=", state)
	//return state.Int()&mutexLocked == mutexLocked
	//return atomic.CompareAndSwapInt32(&m.state, 0, mutexLocked)
	//return false
	return state.Int() == 0
}
