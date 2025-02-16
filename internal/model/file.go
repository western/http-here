package model

import (
	"crypto/md5"
	_ "encoding/hex"
	"errors"
	"fmt"
	"html/template"
	"io"
	_ "io/ioutil"
	"math"
	"os"
	"path/filepath"
	_ "strconv"
	"strings"
	_ "time"

	_ "github.com/fatih/color"
	_ "github.com/gofiber/fiber/v2"

	_ "gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type File struct {
	ID uint `json:"id" gorm:"unique;primaryKey;autoIncrement"`

	MD5 string `gorm:"index:indx_file_md5"`

	FullPath string `gorm:"index:indx_file_full_path"`
	Name     string
	EXT      string

	IsDir   int
	Size    int64
	ModTime string
}

func (File) TableName() string {
	return "file"
}

func FileAddAsync(db *gorm.DB, FullPath string) error {

	IsDir := 0

	ext := filepath.Ext(FullPath)
	ext = strings.ToLower(ext)
	ext = strings.Replace(ext, ".", "", -1)

	name := filepath.Base(FullPath)

	modtime_human := ""
	var size int64 = 0

	if fileInfo, err := os.Stat(FullPath); err == nil {

		modtime := fileInfo.ModTime()
		modtime_human = modtime.Format("2006-01-02 15:04:05")

		size = fileInfo.Size()

		if fileInfo.IsDir() {
			IsDir = 1
		}

	} else if errors.Is(err, os.ErrNotExist) {

		return err
	}

	if result := db.Where("full_path = ?", FullPath).First(&File{}); result.Error == nil {

		return nil
	}

	md5_hash := GetMd5File(FullPath)

	el := &File{
		MD5: md5_hash,

		FullPath: FullPath,
		Name:     name,
		EXT:      ext,

		IsDir:   IsDir,
		Size:    size,
		ModTime: modtime_human,
	}

	result := db.Create(el)
	if result.Error != nil {
		return result.Error
	}
	return nil

}

func FileChkAsync(db *gorm.DB) {

	var rows []File

	if result := db.Find(&rows); result.Error == nil {

		for _, file := range rows {

			if _, err := os.Stat(file.FullPath); err == nil {

			} else if errors.Is(err, os.ErrNotExist) {

				db.Delete(&File{}, file.ID)
			}

		}

	}

}

func FileDelAsync(db *gorm.DB, FullPath string) {

	if result := db.Where("full_path = ?", FullPath).Delete(&File{}); result.Error != nil {
		fmt.Println("FileDelAsync", result.Error)
	}
}

type FileSearch struct {
	IsDir    int
	FullPath string
	Name     string
	EXT      string

	Size      int64
	SizeHuman string

	ModTime string

	OnlyFold     string
	OnlyFoldHtml template.HTML

	NameHtml template.HTML
}

func FileSearchResult(db *gorm.DB, arg_fold, s string) []FileSearch {

	var rows []File
	var ret []FileSearch

	if len(s) == 0 {
		return ret
	}

	if result := db.Order("full_path").Where("full_path like ? and full_path like ?", arg_fold+"%", "%"+s+"%").Find(&rows); result.Error != nil {
		fmt.Println(result.Error)
		return ret
	}

	for _, file := range rows {

		file.FullPath = strings.Replace(file.FullPath, arg_fold, "", 1)

		OnlyFold := filepath.Dir(file.FullPath)
		OnlyFoldHtml := strings.ReplaceAll(OnlyFold, s, `<span style="background-color:yellow">`+s+`</span>`)

		NameHtml := strings.ReplaceAll(file.Name, s, `<span style="background-color:yellow">`+s+`</span>`)
		NameHtml = string(NameHtml)

		ret = append(ret, FileSearch{

			IsDir:    0,
			FullPath: file.FullPath,
			Name:     file.Name,
			EXT:      file.EXT,

			Size:      file.Size,
			SizeHuman: PrettyByteSize(file.Size),

			ModTime: file.ModTime,

			OnlyFold:     OnlyFold,
			OnlyFoldHtml: template.HTML(OnlyFoldHtml),

			NameHtml: template.HTML(NameHtml),
		})
	}

	return ret

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
