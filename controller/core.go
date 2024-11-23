package controller

import (
	"encoding/json"
	"errors"
	_ "fmt"
	"html/template"
	"log"
	"net/url"
	"os"
	"path/filepath"
	_ "reflect"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func OptionsAll(c *fiber.Ctx) error {

	return c.JSON(fiber.Map{
		"code":   200,
		"method": "OPTIONS",
	}, "application/json")
}

func GetAll(c *fiber.Ctx) error {

	arg_fold := ""
	arg_fold = c.Locals("arg_fold").(string)

	arg_upload_disable := ""
	if c.Locals("arg_upload_disable") != nil {
		arg_upload_disable = c.Locals("arg_upload_disable").(string)
	}

	arg_folder_make_disable := ""
	if c.Locals("arg_folder_make_disable") != nil {
		arg_folder_make_disable = c.Locals("arg_folder_make_disable").(string)
	}

	arg_extend_mode := ""
	if c.Locals("arg_extend_mode") != nil {
		arg_extend_mode = c.Locals("arg_extend_mode").(string)
	}

	//arg_fold := "/tmp"
	//arg_upload_disable := ""
	//arg_folder_make_disable := ""

	c_path, err := url.QueryUnescape(c.Path())
	if err != nil {

		log.Println(err)
		LogPrefix(c, "500", "Error "+filepath.Join(arg_fold, c_path))
		return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout")
	}

	c_path = CleanDirtyPath(c_path)

	if fileInfo, err := os.Stat(filepath.Join(arg_fold, c_path)); err == nil {

		if fileInfo.IsDir() {

			// check index.html inside
			if _, err := os.Stat(filepath.Join(arg_fold, c_path, "index.html")); err == nil {

				LogPrefix(c, "200", "Index file found for path '"+c_path+"', SendFile "+filepath.Join(arg_fold, c_path, "index.html"))
				return c.SendFile(filepath.Join(arg_fold, c_path, "index.html"), false)
			}

			LogPrefix(c, "200", "Dir "+filepath.Join(arg_fold, c_path))

			breadcrumb := ""

			res1 := strings.Split(c_path, "/")
			pt := ""
			for indx, el := range res1 {
				if indx == 0 {
					continue
				}
				pt += "/" + el
				breadcrumb += `<li class="breadcrumb-item"><a class="nodecor" href="` + pt + `">` + el + `</a></li>`
			}

			entries, err := os.ReadDir(filepath.Join(arg_fold, c_path))
			if err != nil {

				log.Println(err)
				LogPrefix(c, "500", "Error "+filepath.Join(arg_fold, c_path))
				return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout")
			}

			//folderlist := ""
			//filelist := ""

			//fl := template.HTML("")
			//var fl template.HTML

			//fmt.Println(reflect.TypeOf(fl))
			//fmt.Println(reflect.TypeOf(entries))
			template_file := "index"

			var rows []FileRow
			var mode string

			if arg_extend_mode == "1" {

				template_file = "index_extend"

				mode = c.Cookies("mode")
				if len(mode) == 0 {
					mode = "list"
				}

				q_mode := c.Query("mode")
				if len(q_mode) > 0 {
					mode = q_mode
				}

				cookie := new(fiber.Cookie)
				cookie.Name = "mode"
				cookie.Value = mode
				c.Cookie(cookie)

			}

			rows = listGenerateView(arg_fold, c_path, entries)

			mode_thumb := false
			if mode == "thumb" {
				mode_thumb = true
			}

			mode_list := false
			if mode == "list" {
				mode_list = true
			}

			//folderTree := WalkAndTreeBuild( filepath.Join(arg_fold, c_path), "/", 1 )
			folderTree := WalkAndTreeBuild(arg_fold, "/", 1)

			folderTree_js, err := json.Marshal(folderTree)
			if err != nil {
				//fmt.Println(err)
				panic(err)
				return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout")
			}

			return c.Render("view/"+template_file, fiber.Map{

				"Breadcrumb": template.HTML(breadcrumb),
				//"Filelist":   fl,
				//"folderTree_js": string(folderTree_js),
				"folderTree_js": template.HTML(folderTree_js),

				"rows":            rows,
				"arg_extend_mode": arg_extend_mode,
				"mode_thumb":      mode_thumb,
				"mode_list":       mode_list,

				"files_count_max":     20,
				"fieldSize_max":       7 * 1024 * 1024 * 1024,
				"fieldSize_max_human": "7 Gb",

				"arg_upload_disable":      arg_upload_disable,
				"arg_folder_make_disable": arg_folder_make_disable,
			}, "view/layout")

		} else {

			LogPrefix(c, "200", "SendFile "+filepath.Join(arg_fold, c_path))

			return c.SendFile(filepath.Join(arg_fold, c_path), false)
		}

	} else if errors.Is(err, os.ErrNotExist) {

		LogPrefix(c, "404", filepath.Join(arg_fold, c_path))

		return c.Status(fiber.StatusNotFound).Render("view/404", fiber.Map{
			"File": c_path,
		}, "view/layout")
	}

	return c.Status(fiber.StatusNotFound).Render("view/404", fiber.Map{}, "view/layout")

}

type FileRow struct {
	IsDir        bool
	FullPath     string
	Name         string
	SizeHuman    string
	ModTimeHuman string
	IsPreview    bool
}

func listGenerateView(arg_fold string, c_path string, entries []os.DirEntry) []FileRow {

	var rows_dir []FileRow
	var rows_file []FileRow

	for _, e := range entries {

		ext := filepath.Ext(e.Name())
		ext = strings.ToLower(ext)
		ext = strings.Replace(ext, ".", "", -1)

		is_preview_match, _ := regexp.MatchString("^(jpg|jpeg|png|gif)$", ext)

		if fileInfo2, err := os.Stat(filepath.Join(arg_fold, c_path, e.Name())); err == nil {

			modtime := fileInfo2.ModTime()
			modtime_human := modtime.Format("2006-01-02 15:04:05")

			size := fileInfo2.Size()
			size_human := prettyByteSize(size)

			if fileInfo2.IsDir() {
				rows_dir = append(rows_dir, FileRow{
					IsDir:        fileInfo2.IsDir(),
					FullPath:     filepath.Join(c_path, e.Name()),
					Name:         e.Name(),
					SizeHuman:    size_human,
					ModTimeHuman: modtime_human,
					IsPreview:    false,
				})
			} else {
				rows_file = append(rows_file, FileRow{
					IsDir:        fileInfo2.IsDir(),
					FullPath:     filepath.Join(c_path, e.Name()),
					Name:         e.Name(),
					SizeHuman:    size_human,
					ModTimeHuman: modtime_human,
					IsPreview:    is_preview_match,
				})
			}

		}

	}

	return append(rows_dir, rows_file...)
}
