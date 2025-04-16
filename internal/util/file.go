package util

import (
	"crypto/md5"
	"fmt"
	"io"
	_ "io/fs"
	"io/ioutil"
	"mime/multipart"
	"os"
	"path"
	_ "path/filepath"
	"time"
)

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
