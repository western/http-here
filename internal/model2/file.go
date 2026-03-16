package model2

import (
	"fmt"
	"os"
	//"path"
	"path/filepath"
	//"strconv"
	//"time"
	"encoding/json"
	"golang.org/x/exp/slices"
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
	allowMD5ExtsRegex = regexp.MustCompile("^(jpg|jpeg|png|gif|pdf|rtf|doc|docx|xls|xlsx|odt|ods)$")
)

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

func FileFindByPath(searchFullPath string) (File, bool) {

	if len(searchFullPath) == 0 {
		panic("Yous should set searchFullPath")
	}

	var returnFile File

	if Dbse.BadgerEnable {

		foundBytes, isFound := BadgerGetOne("idx_file_fullpath_" + searchFullPath)
		if !isFound {

			//EventLogAdd(nil, 500, "FileFindByPath", "File by path not found (1)")
			return returnFile, false
		}

		file_key := string(foundBytes)

		foundBytes2, isFound2 := BadgerGetOne(file_key)
		if !isFound2 {

			//EventLogAdd(nil, 500, "FileFindByPath", "File by path not found (2)")
			return returnFile, false
		}

		err := json.Unmarshal(foundBytes2, &returnFile)
		if err != nil {

			//EventLogAdd(nil, 500, "FileFindByPath", "json.Unmarshal "+err.Error())
			return returnFile, false
		}

		return returnFile, true

	}

	return returnFile, false
}

func FileFindByMD5(findMD5 string) (File, bool) {

	if len(findMD5) == 0 {
		panic("Yous should set findMD5")
	}

	var file_ret File

	if Dbse.BadgerEnable {

		bytesFound, isFound := BadgerGetOne("idx_file_md5_" + findMD5)
		if isFound {

			bytesFound2, isFound2 := BadgerGetOne(string(bytesFound))
			if isFound2 {

				err := json.Unmarshal(bytesFound2, &file_ret)
				if err != nil {

					EventLogAdd(nil, 500, "FileFindByMD5", "json.Unmarshal "+err.Error())
				}

				return file_ret, true
			}
		}

	}

	EventLogAdd(nil, 500, "FileFindByMD5", "File by MD5 not found")
	return file_ret, false
}

func FileChkAsync() {

}

func FileAdd(FullPath string) error {

	if Dbse.BadgerEnable {

		_, isFound := FileFindByPath(FullPath)
		if isFound {
			return nil
		}

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

		is_match := allowMD5ExtsRegex.MatchString(ext)
		if is_match {

			md5_hash = util.GetMd5File(FullPath)

		} else if size <= 20*1024*1024 {

			md5_hash = util.GetMd5File(FullPath)
		}

		file_id := FileGetPrimaryId()
		if file_id == 0 {
			panic("FileAdd: file_id is zero")
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

		key := "file_" + fmt.Sprintf("%020d", file_id)

		// ------------------------------------------------------------------------------------------------------------------------

		Dbse.Badger.Update(func(txn *badger.Txn) error {

			err := txn.Set([]byte(key), []byte(payload))
			if err != nil {
				panic(err)
			}

			if len(el.FullPath) > 0 {
				err := txn.Set([]byte("idx_file_fullpath_"+el.FullPath), []byte(key))
				if err != nil {
					panic(err)
				}
			}

			if len(md5_hash) > 0 {
				err := txn.Set([]byte("idx_file_md5_"+md5_hash), []byte(key))
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

					var el File
					err2 := json.Unmarshal(v, &el)
					if err2 != nil {

						EventLogAdd(nil, 500, "FileList", "json.Unmarshal "+err2.Error())
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

func FileSelect(Filters []map[string]interface{}) []File {

	var ret []File

	if Dbse.BadgerEnable {

		Dbse.Badger.View(func(txn *badger.Txn) error {

			it := txn.NewIterator(badger.DefaultIteratorOptions)
			defer it.Close()
			prefix := []byte("file_")

			for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {

				item := it.Item()
				//k := item.Key()

				err := item.Value(func(v []byte) error {

					var el File
					err2 := json.Unmarshal(v, &el)
					if err2 != nil {

						EventLogAdd(nil, 500, "FileSelect", "json.Unmarshal "+err2.Error())
					}

					//fmt.Printf("key=%s, value=%+v\n", k, el)
					var isExpected []bool

					for _, f_val := range Filters {

						//fmt.Println("")
						//fmt.Println("item=", f_val)

						for k2, v2 := range f_val {
							//fmt.Println("k2=", k2, " v2=", v2)

							switch k2 {
							case "ID":
								if el.ID == v2 {
									isExpected = append(isExpected, true)
								} else {
									isExpected = append(isExpected, false)
								}
							case "MD5":
								if el.MD5 == v2 {
									isExpected = append(isExpected, true)
								} else {
									isExpected = append(isExpected, false)
								}
							case "FullPath":
								if el.FullPath == v2 {
									isExpected = append(isExpected, true)
								} else {
									isExpected = append(isExpected, false)
								}

							}

						}
					}

					//fmt.Println("isExpected=", isExpected)
					//fmt.Println();

					isPresentFalse := slices.Contains(isExpected, false)
					if !isPresentFalse {
						// false is absent
						ret = append(ret, el)
					}

					return nil
				})
				if err != nil {
					return err
				}
			}
			return nil
		})

	}

	return ret
}

func FileTruncate() {

	if Dbse.BadgerEnable {

		prefix := []byte("file_")

		err := Dbse.Badger.DropPrefix(prefix)
		if err != nil {
			panic(err)
		}

		// -------------------------------------------------------------------------------------------------------------------------------------------

		prefix = []byte("idx_file_fullpath_")

		err = Dbse.Badger.DropPrefix(prefix)
		if err != nil {
			panic(err)
		}

		// -------------------------------------------------------------------------------------------------------------------------------------------

		prefix = []byte("idx_file_md5_")

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

		EventLogAdd(nil, 200, "FileTruncate", "File storage recreated")

	}
}
