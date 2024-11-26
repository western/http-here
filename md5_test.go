

package main

import (
	"testing"
    
    _ "fmt"
    "os"
    "regexp"
    "path/filepath"
    
	"github.com/western/http-here/controller"

	"crypto/md5"
	"encoding/hex"
)



func TestSimpleConcatMd5(t *testing.T) {
    
    path := "/tmp/rain_blue/"
    
    files, err := os.ReadDir(path)
	if err != nil {
		return
	}

	

	for _, file := range files {

		fileInfo, _ := os.Stat(filepath.Join(path, file.Name()))

		//fmt.Println("file.Name()=", file.Name())

		modtime := fileInfo.ModTime()
		modtime_human := modtime.Format("2006-01-02 15:04:05")

		size := fileInfo.Size()
		size_human := controller.PrettyByteSize(size)

		//fmt.Println("runtime.NumGoroutine=", runtime.NumGoroutine())

		if !file.IsDir() {

			c_width := "600"
			

			file_ext := controller.GetExtNorm(file.Name())
			orig_filename := controller.GetFileName(file.Name())

			is_preview_match, _ := regexp.MatchString("^(jpg|jpeg|png|gif)$", file_ext)
			if !is_preview_match {
				continue
			}

			hash_name := md5.Sum([]byte(orig_filename + modtime_human + size_human + c_width))
			hex_name := hex.EncodeToString(hash_name[:])
			
			_ = hex_name
		}
	}
}



func TestMd5BodyFile(t *testing.T) {
    
    path := "/tmp/rain_blue/"
    
    files, err := os.ReadDir(path)
	if err != nil {
		return
	}

	

	for _, file := range files {

		//fileInfo, _ := os.Stat(filepath.Join(path, file.Name()))

		//fmt.Println("file.Name()=", file.Name())

		//modtime := fileInfo.ModTime()
		//modtime_human := modtime.Format("2006-01-02 15:04:05")

		//size := fileInfo.Size()
		//size_human := controller.PrettyByteSize(size)

		//fmt.Println("runtime.NumGoroutine=", runtime.NumGoroutine())

		if !file.IsDir() {

			//c_width := "800"
			//i_width := 800

			file_ext := controller.GetExtNorm(file.Name())
			//orig_filename := controller.GetFileName(file.Name())

			is_preview_match, _ := regexp.MatchString("^(jpg|jpeg|png|gif)$", file_ext)
			if !is_preview_match {
				continue
			}

			//hash_name := md5.Sum([]byte(orig_filename + modtime_human + size_human + c_width))
			//hex_name := hex.EncodeToString(hash_name[:])
			
			
			
			hex_name := controller.GetMd5File( filepath.Join(path, file.Name()) )
			
			_ = hex_name
			
		}
	}
}