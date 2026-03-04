package model2

import (
	"fmt"
	"os"
	//"path"
	"strconv"
	//"time"
	"encoding/json"
	"html/template"

	//"github.com/western/http-here/v2/internal/conf"
	"github.com/western/http-here/v2/internal/util"

	badger "github.com/dgraph-io/badger/v4"
)

type File struct {
	ID uint64 `json:"id"`

	FullPath string `json:"full_path"`
	Name     string `json:"name"`
	EXT      string `json:"ext"`

	Size      int64  `json:"size"`
	SizeHuman string `json:"size_human"`

	ModTime string `json:"mod_time"`
}

func FileFind(FullPath string) (File, bool) {

	if len(FullPath) == 0 {
		panic("Yous should set full_path")
	}

	var file File

	if Dbse.BadgerEnable {

		Dbse.Badger.View(func(txn *badger.Txn) error {

			it := txn.NewIterator(badger.DefaultIteratorOptions)
			defer it.Close()
			prefix := []byte("file_")

			for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
				item := it.Item()
				//k := item.Key()
				item.Value(func(v []byte) error {

					var el File
					err2 := json.Unmarshal(v, &el)
					if err2 != nil {
						fmt.Println("error:", err2)
					}

					if el.FullPath == FullPath {

						//fmt.Printf("key=%s, value=%+v\n", k, el)
						file = el
					}

					return nil
				})
			}
			return nil
		})

		if file.ID == 0 {
			return file, false
		} else {
			return file, true
		}

	}

	return file, false
}

func FileGetPrimaryId() uint64 {

	seq, err := Dbse.Badger.GetSequence([]byte("seq_file"), 1000)
	defer seq.Release()

	// uint64, err
	num, err := seq.Next()
	if err != nil {
		panic(err)
	}

	if num == 0 {
		num, err = seq.Next()
		if err != nil {
			panic(err)
		}
	}

	return num
}

func FileChkAsync() {

}

func FileAdd(FullPath string) error {

	if Dbse.BadgerEnable {

		_, isFound := FileFind(FullPath)
		if isFound {
			return nil
		}

		file_id := FileGetPrimaryId()

		name := util.GetFileName(FullPath)
		ext := util.GetExtNorm(FullPath)

		fileInfo, err := os.Stat(FullPath)
		if err != nil {
			panic(err)
		}

		size := fileInfo.Size()
		size_human := util.PrettyByteSize(size)

		mod_time := fileInfo.ModTime()
		mod_time_human := mod_time.Format("2006-01-02 15:04:05")

		el := File{
			ID: file_id,

			FullPath: FullPath,
			Name:     name,
			EXT:      ext,

			Size:      size,
			SizeHuman: size_human,

			ModTime: mod_time_human,
		}
		payload, _ := json.Marshal(el)

		key := "file_"
		key += strconv.FormatUint(file_id, 10)

		Dbse.Badger.Update(func(txn *badger.Txn) error {

			err := txn.Set([]byte(key), []byte(payload))
			if err != nil {
				panic(err)
			}

			return nil
		})

		//FileList()

	}

	return nil
}

func FileDelMd5Async(prefix, FullPath string) {

}

func FileDelAsync(FullPath string) {

}

type FileSearch struct {
	ID uint64 `json:"id"`

	FullPath string `json:"full_path"`
	Name     string `json:"name"`
	EXT      string `json:"ext"`

	Size      int64  `json:"size"`
	SizeHuman string `json:"size_human"`

	ModTime string `json:"mod_time"`

	OnlyFold     string        `json:"only_fold"`
	OnlyFoldHtml template.HTML `json:"only_fold_html"`

	NameHtml template.HTML `json:"name_html"`
}

func FileSearchResult(s string) []FileSearch {

	var ret []FileSearch

	return ret
}

func FileList() {

	if Dbse.BadgerEnable {

		Dbse.Badger.View(func(txn *badger.Txn) error {
			it := txn.NewIterator(badger.DefaultIteratorOptions)
			defer it.Close()
			prefix := []byte("file_")
			for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
				item := it.Item()
				k := item.Key()
				err := item.Value(func(v []byte) error {

					//fmt.Printf("key=%s, value=%s\n", k, v)

					var el File
					err2 := json.Unmarshal(v, &el)
					if err2 != nil {
						fmt.Println("error:", err2)
					}

					fmt.Printf("key=%s, value=%+v\n", k, el)

					return nil
				})
				if err != nil {
					return err
				}
			}
			return nil
		})

	}
}

func FileClear() {

	if Dbse.BadgerEnable {

		prefix := []byte("file_")

		err := Dbse.Badger.DropPrefix(prefix)
		if err != nil {
			panic(err)
		}

	}
}
