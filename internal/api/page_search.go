package api

import (
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/western/http-here/v2/internal/conf"
	"github.com/western/http-here/v2/internal/model2"
	"github.com/western/http-here/v2/internal/util"

	"github.com/charlievieth/fastwalk"
	"github.com/gofiber/fiber/v3"
)

func PageSearch(c fiber.Ctx) error {

	searchString := c.Query("s")

	// -------------------------------------------------------------------------------------------------------------------------

	result_list := FileFastSearchResult(searchString)

	model2.EventLogAdd(c, 200, "GetSearch", "Get search '"+searchString+"'")

	return c.Render("view/search", fiber.Map{

		"searchString": searchString,
		"result_list":  result_list,
	}, "view/layout/default")
}

// ---------------------------------------------------------------------------------------------------------------------------------------------

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
