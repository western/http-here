
package api

import (
	"archive/zip"
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
	"path/filepath"
	"reflect"
	"regexp"
	_ "runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/gofiber/fiber/v2"

	"crypto/md5"
	"encoding/hex"

	"github.com/edwvee/exiffix"
	"golang.org/x/image/draw"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
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

func LogPrefix(c *fiber.Ctx, status string, addition string) {

	green_clr := color.New(color.FgGreen).SprintFunc()
	//magenta_clr := color.New(color.FgMagenta).SprintFunc()
	//cian_clr := color.New(color.FgCyan).SprintFunc()
	//yellow_clr := color.New(color.FgYellow).SprintFunc()
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

	//pref += cian(addition)
	pref += addition

	fmt.Println(pref)

	// [2024-10-29 18:22:46] [192.168.0.101] [loginT] [200] Dir /tmp/fold1/fold3
	//fmt.Printf("[%s] [%s] [%s] [%s] %s\n", magenta("warning"), red("error"))

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
		srcfp := filepath.Join(src, fd.Name())
		dstfp := filepath.Join(dst, fd.Name())

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

			fileInfo, err := os.Stat(filepath.Join(basePath, file.Name()))
			if err != nil {
				fmt.Println(err)
				//LogPrefix(c, "500", "'"+filepath.Join(basePath, file.Name())+"' not exists")
				//continue
			}

			header, err := zip.FileInfoHeader(fileInfo)
			if err != nil {
				fmt.Println(err)
				//LogPrefix(c, "500", "'"+filepath.Join(arg_fold, u_path, name)+"' err: "+err.Error())
				//continue
			}
			header.Method = zip.Store
			header.Name = filepath.Join(baseInZip, file.Name())

			// Add some files to the archive.
			//f, err := w.Create(filepath.Join(baseInZip, file.Name()))
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
			newBase := filepath.Join(basePath, file.Name()) + "/"
			//fmt.Println("Recursing and Adding SubDir: " + file.Name())
			//fmt.Println("Recursing and Adding SubDir: " + newBase)

			//addFiles(w, newBase, baseInZip+file.Name()+"/")
			addFilesToZip(w, newBase, filepath.Join(baseInZip, file.Name())+"/")
		}
	}
}

func WalkAndClearOld(path string) {

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

func WalkAndClearZeroFile(path string, deep int) {

	files, err := os.ReadDir(path)
	if err != nil {
		return
	}

	if deep > 5 {
		return
	}

	for _, file := range files {

		if !file.IsDir() {

			if NewFileInfo, err := os.Stat(filepath.Join(path, file.Name())); err == nil {
				if NewFileInfo.Size() == 0 {

					//fmt.Println("WalkAndClearZeroFile", "remove=", filepath.Join(path, file.Name())    )

					if err2 := os.Remove(filepath.Join(path, file.Name())); err2 != nil {
						panic("Problem of remove zero file " + filepath.Join(path, file.Name()) + " " + err2.Error())
					}
				}
			}

		} else if file.IsDir() {

			WalkAndClearZeroFile(filepath.Join(path, file.Name()), deep+1)

		}
	}

}

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

		fileInfo, err := os.Stat(filepath.Join(path, file.Name()))

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

			if _, err := os.Stat(filepath.Join(homepath, ".httphere", "thumb", hex_name)); err == nil {
				//fmt.Println("Is exists ", filepath.Join(homepath, ".httphere", "thumb", hex_name))
				continue
			}

			go func() {

				input, _ := os.Open(filepath.Join(path, file.Name()))
				defer input.Close()

				output, _ := os.Create(filepath.Join(homepath, ".httphere", "thumb", hex_name))
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

					err := os.Remove(filepath.Join(homepath, ".httphere", "thumb", hex_name))
					if err != nil {
						//log.Fatal(err)
						panic(err)
					}

					//fmt.Println("Too small for thumb make ", filepath.Join(path, file.Name()))
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

				//fmt.Println("Make thumb for ", filepath.Join(path, file.Name()))

				//fmt.Println("runtime.NumGoroutine=", runtime.NumGoroutine())

			}()

			//fmt.Println("Not wait for ", filepath.Join(path, file.Name()))

		} else if file.IsDir() {

			WalkAndMakeThumbnail(filepath.Join(path, file.Name()), deep+1)

		}
	}

	return
}

// openssl aes-256-cbc -a -salt -in file.txt -out file.txt.cr -pass pass:123
func CryptFile(path, pass string) (bool, error) {

	_, err := exec.Command("bash", "-c", "openssl --help").Output()
	if err != nil {
		return false, err
	}

	if len(path) == 0 {
		return false, errors.New("Path of file is empty")
	}

	if len(pass) == 0 {
		return false, errors.New("Pass is empty")
	}

	//panic("yyy")

	//fmt.Println( filepath.Dir(path) )
	//fmt.Println( filepath.Base(path) )
	from_file := filepath.Base(path)
	to_file := filepath.Base(path) + ".cr"

	cmd := exec.Command("bash", "-c", "openssl aes-256-cbc -a -salt -in "+from_file+" -out "+to_file+" -pass pass:"+pass)
	cmd.Dir = filepath.Dir(path)
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

	if err = os.Remove(path); err != nil {
		return false, err
	}

	if err = os.Rename(filepath.Join(filepath.Dir(path), to_file), path); err != nil {
		return false, err
	}

	return true, nil
}

// openssl aes-256-cbc -d -a -in file.txt.cr -out file.txt.new -pass pass:123
func DecryptFile(path, pass string) (bool, error) {

	_, err := exec.Command("bash", "-c", "openssl --help").Output()
	if err != nil {
		return false, err
	}

	if len(path) == 0 {
		return false, errors.New("Path of file is empty")
	}

	if len(pass) == 0 {
		return false, errors.New("Pass is empty")
	}

	//fmt.Println( filepath.Dir(path) )
	//fmt.Println( filepath.Base(path) )
	from_file := filepath.Base(path)
	to_file := filepath.Base(path)
	to_file = strings.Replace(to_file, ".cr", "", 1)
	to_file += ".decrypt"

	cmd := exec.Command("bash", "-c", "openssl aes-256-cbc -d -a  -in "+from_file+" -out "+to_file+" -pass pass:"+pass)
	cmd.Dir = filepath.Dir(path)
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

	if err = os.Remove(path); err != nil {
		return false, err
	}

	if err = os.Rename(filepath.Join(filepath.Dir(path), to_file), path); err != nil {
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
