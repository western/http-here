package model2

import (
	"fmt"
	"os"
	//"path"
	"path/filepath"
	"strconv"
	//"time"
	"encoding/json"
	"html/template"
	"regexp"
	"strings"

	"github.com/western/http-here/v2/internal/conf"
	"github.com/western/http-here/v2/internal/util"

	"github.com/charlievieth/fastwalk"
	badger "github.com/dgraph-io/badger/v4"
	"github.com/jedib0t/go-pretty/v6/table"
)

type File struct {
	ID  uint64 `json:"id"`
	MD5 string `json:"md5"`

	FullPath string `json:"full_path"`
	Name     string `json:"name"`
	EXT      string `json:"ext"`

	Size      int64  `json:"size"`
	SizeHuman string `json:"size_human"`

	ModTime string `json:"mod_time"`
}

var (
	generateMd5AllRegex = regexp.MustCompile("^(jpg|jpeg|png|gif|pdf|rtf|doc|docx|xls|xlsx|odt|ods)$")
)

func FileFindByPath(FullPath string) (File, bool) {

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

func FileFindByMD5(findMD5 string) (File, bool) {

	if len(findMD5) == 0 {
		panic("Yous should set findMD5")
	}

	var file_ret File

	if Dbse.BadgerEnable {

		bytesFound, isFound := BadgerGetOne("idx_md5_file_" + findMD5)
		if isFound {

			bytesFound2, isFound2 := BadgerGetOne(string(bytesFound))
			if isFound2 {

				err := json.Unmarshal(bytesFound2, &file_ret)
				if err != nil {
					fmt.Println("error:", err)
				}

				return file_ret, true
			}
		}

	}

	return file_ret, false
}

func FileGetPrimaryId() uint64 {

	if !Dbse.BadgerEnable {
		return 0
	}

	seq, err := Dbse.Badger.GetSequence([]byte("seq_file"), 1000)
	if err != nil {
		//fmt.Println("FileGetPrimaryId GetSequence err:", err.Error())
		return 0
	}
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

		_, isFound := FileFindByPath(FullPath)
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

		md5_hash := ""

		//generateMd5AllRegex = regexp.MustCompile("^(jpg|jpeg|png|gif|pdf|rtf|doc|docx|xls|xlsx|odt|ods)$")
		is_match := generateMd5AllRegex.MatchString(ext)
		if is_match {

			md5_hash = util.GetMd5File(FullPath)

		} else if size <= 20*1024*1024 {

			md5_hash = util.GetMd5File(FullPath)
		}

		el := File{
			ID:  file_id,
			MD5: md5_hash,

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

		// ------------------------------------------------------------------------------------------------------------------------

		Dbse.Badger.Update(func(txn *badger.Txn) error {

			err := txn.Set([]byte(key), []byte(payload))
			if err != nil {
				panic(err)
			}

			if len(el.FullPath) > 0 {
				err := txn.Set([]byte("idx_fullpath_file_"+el.FullPath), []byte(key))
				if err != nil {
					panic(err)
				}
			}

			if len(md5_hash) > 0 {
				err := txn.Set([]byte("idx_md5_file_"+md5_hash), []byte(key))
				if err != nil {
					panic(err)
				}
			}

			return nil
		})

		// ------------------------------------------------------------------------------------------------------------------------

		//FileList()

	}

	return nil
}

func FileDelMd5Async(FullPath string) {

}

func FileDelAsync(FullPath string) {

}

type FileSearchType struct {
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

func FileSearchResult(s string) []FileSearchType {

	var ret []FileSearchType
	repl := regexp.MustCompile(`(?i)` + s)

	resultPathFoundRegex := regexp.MustCompile(s)

	FileSearchWalk := func(resultPath string, de os.DirEntry, err error) error {

		resultPathCut := strings.Replace(resultPath, conf.ArgFold, "", 1)
		name := util.GetFileName(resultPathCut)
		ext := util.GetExtNorm(resultPathCut)

		is_found := resultPathFoundRegex.MatchString(resultPathCut)
		if !is_found {

			return nil
		}

		// ----------------------------------------------------------------------------------------------

		if de.IsDir() {
			return nil
		}

		fileInfo, err := de.Info()
		if err != nil {
			panic(err)
		}

		//size := fileInfo.Size()
		//size_human := util.PrettyByteSize(size)

		mod_time := fileInfo.ModTime()
		mod_time_human := mod_time.Format("2006-01-02 15:04:05")

		// ----------------------------------------------------------------------------------------------

		OnlyFold := filepath.Dir(resultPathCut)

		OnlyFoldHtml := repl.ReplaceAllString(OnlyFold, `<span style="background-color:yellow">`+s+`</span>`)

		NameHtml := repl.ReplaceAllString(name, `<span style="background-color:yellow">`+s+`</span>`)

		// ----------------------------------------------------------------------------------------------

		ret = append(ret, FileSearchType{

			FullPath: resultPathCut,
			Name:     name,
			EXT:      ext,

			Size:      fileInfo.Size(),
			SizeHuman: util.PrettyByteSize(fileInfo.Size()),

			ModTime: mod_time_human,

			OnlyFold:     OnlyFold,
			OnlyFoldHtml: template.HTML(OnlyFoldHtml),

			NameHtml: template.HTML(NameHtml),
		})

		return nil
	}

	filepath.WalkDir(conf.ArgFold, FileSearchWalk)

	return ret
}

func FileFastSearchResult(s string) []FileSearchType {

	var ret []FileSearchType
	repl := regexp.MustCompile(`(?i)` + s)

	resultPathFoundRegex := regexp.MustCompile(`(?i)` + s)

	walkFn := func(resultPath string, de os.DirEntry, err error) error {

		resultPathCut := strings.Replace(resultPath, conf.ArgFold, "", 1)
		name := util.GetFileName(resultPathCut)
		ext := util.GetExtNorm(resultPathCut)

		is_found := resultPathFoundRegex.MatchString(resultPathCut)
		if !is_found {

			return nil
		}

		// ----------------------------------------------------------------------------------------------

		if de.IsDir() {
			return nil
		}

		fileInfo, err := de.Info()
		if err != nil {
			panic(err)
		}

		//size := fileInfo.Size()
		//size_human := util.PrettyByteSize(size)

		mod_time := fileInfo.ModTime()
		mod_time_human := mod_time.Format("2006-01-02 15:04:05")

		// ----------------------------------------------------------------------------------------------

		OnlyFold := filepath.Dir(resultPathCut)

		OnlyFoldHtml := repl.ReplaceAllString(OnlyFold, `<span style="background-color:yellow">`+s+`</span>`)

		NameHtml := repl.ReplaceAllString(name, `<span style="background-color:yellow">`+s+`</span>`)

		// ----------------------------------------------------------------------------------------------

		ret = append(ret, FileSearchType{

			FullPath: resultPathCut,
			Name:     name,
			EXT:      ext,

			Size:      fileInfo.Size(),
			SizeHuman: util.PrettyByteSize(fileInfo.Size()),

			ModTime: mod_time_human,

			OnlyFold:     OnlyFold,
			OnlyFoldHtml: template.HTML(OnlyFoldHtml),

			NameHtml: template.HTML(NameHtml),
		})

		return nil

	}

	fastConf := fastwalk.Config{
		Follow: true,
	}
	fastwalk.Walk(&fastConf, conf.ArgFold, walkFn)

	return ret
}

func FileList() {

	t := table.NewWriter()
	//t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleLight)

	// key=file_23, value={ID:23 MD5:b423c6288f1b9d97076ab406870eae01 FullPath:/tmp/folder1/spring/alexandru-tudorache-JdjdIjzJl94-unsplash.jpg Name:alexandru-tudorache-JdjdIjzJl94-unsplash EXT:jpg Size:1185441 SizeHuman:1.1 MiB ModTime:2025-08-03 20:25:35}
	t.AppendHeader(table.Row{"#", "MD5", "FullPath", "Name", "EXT", "SizeHuman", "ModTime"})

	if Dbse.BadgerEnable {

		Dbse.Badger.View(func(txn *badger.Txn) error {

			it := txn.NewIterator(badger.DefaultIteratorOptions)
			defer it.Close()
			prefix := []byte("file_")

			for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
				item := it.Item()
				//k := item.Key()
				err := item.Value(func(v []byte) error {

					//fmt.Printf("key=%s, value=%s\n", k, v)

					var el File
					err2 := json.Unmarshal(v, &el)
					if err2 != nil {
						fmt.Println("error:", err2)
					}

					//fmt.Printf("key=%s, value=%+v\n", k, el)
					t.AppendRows([]table.Row{
						{el.ID, el.MD5, el.FullPath, el.Name, el.EXT, el.SizeHuman, el.ModTime},
					})

					return nil
				})
				if err != nil {
					return err
				}
			}
			return nil
		})

		fmt.Println(t.Render())

	}
}

func FileClear() {

	if Dbse.BadgerEnable {

		prefix := []byte("file_")

		err := Dbse.Badger.DropPrefix(prefix)
		if err != nil {
			panic(err)
		}

		// -------------------------------------------------------------------------------------------------------------------------------------------

		prefix = []byte("idx_fullpath_file_")

		err = Dbse.Badger.DropPrefix(prefix)
		if err != nil {
			panic(err)
		}

		// -------------------------------------------------------------------------------------------------------------------------------------------

		prefix = []byte("idx_md5_file_")

		err = Dbse.Badger.DropPrefix(prefix)
		if err != nil {
			panic(err)
		}

		// -------------------------------------------------------------------------------------------------------------------------------------------

		prefix = []byte("seq_file")

		err = Dbse.Badger.DropPrefix(prefix)
		if err != nil {
			panic(err)
		}

	}
}
