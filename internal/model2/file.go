package model2

import (
	//"fmt"
	//"os"
	//"path"
	"html/template"
	//"github.com/western/http-here/v2/internal/conf"
)

func FileChkAsync() {

}

func FileAddAsync(FullPath string) error {

	return nil
}

func FileDelMd5Async(prefix, FullPath string) {

}

func FileDelAsync(FullPath string) {

}

type FileSearch struct {
	//IsDir    int
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

func FileSearchResult(s string) []FileSearch {

	var ret []FileSearch

	return ret
}
