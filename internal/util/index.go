package util

import (
	"bufio"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
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
	cmd_filename = RotateSlash(cmd_filename)

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

func MutexUnLocked(m *sync.Mutex) bool {
	state := reflect.ValueOf(m).Elem().FieldByName("state")
	//fmt.Println("state=", state)
	//return state.Int()&mutexLocked == mutexLocked
	//return atomic.CompareAndSwapInt32(&m.state, 0, mutexLocked)
	//return false
	return state.Int() == 0
}
