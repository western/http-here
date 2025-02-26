package api

import (
	"encoding/json"
	"errors"
	_ "fmt"
	"html/template"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	_ "runtime"
	"sort"
	"strings"
	"time"

	"github.com/western/http-here/internal/conf"
	"github.com/western/http-here/internal/model"
	"github.com/western/http-here/internal/util"

	"github.com/gofiber/fiber/v2"
)

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

	arg_crypt := ""
	if c.Locals("arg_crypt") != nil {
		arg_crypt = c.Locals("arg_crypt").(string)
	}

	arg_spa := ""
	if c.Locals("arg_spa") != nil {
		arg_spa = c.Locals("arg_spa").(string)
	}

	db, err := model.ConnectToSQLite()
	if err != nil {
		panic(err)
	}

	c_path, err := url.QueryUnescape(c.Path())
	if err != nil {

		model.EventLogAdd(db, c, "500", "CORE", "Error "+path.Join(arg_fold, c_path)+" "+err.Error())
		return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")
	}

	c_path = util.CleanDirtyPath(c_path)

	if fileInfo, err := os.Stat(path.Join(arg_fold, c_path)); err == nil {

		if fileInfo.IsDir() {

			//fmt.Println("c_path=", c_path)

			if arg_spa == "1" {

				model.EventLogAdd(db, c, "302", "CORE", "SPA application, redirect to /#!"+c_path)
				return c.Redirect("/#!"+c_path, 302)
			}

			// check index.html inside
			if _, err := os.Stat(path.Join(arg_fold, c_path, "index.html")); err == nil {

				model.EventLogAdd(db, c, "200", "CORE", "Index file found for path '"+c_path+"', SendFile "+path.Join(arg_fold, c_path, "index.html"))
				return c.SendFile(path.Join(arg_fold, c_path, "index.html"), false)
			}

			model.EventLogAdd(db, c, "200", "CORE", "Dir "+path.Join(arg_fold, c_path))

			breadcrumb := ""
			//separator := string(os.PathSeparator)
			separator := "/"
			//fmt.Println("os.PathSeparator=", separator)

			res1 := strings.Split(c_path, separator)
			pt := ""
			for indx, el := range res1 {
				if indx == 0 {
					continue
				}
				pt += separator + el
				breadcrumb += `<li class="breadcrumb-item"><a class="nodecor" href="` + pt + `">` + el + `</a></li>`
			}

			entries, err := os.ReadDir(path.Join(arg_fold, c_path))
			if err != nil {

				model.EventLogAdd(db, c, "500", "CORE", "Error "+path.Join(arg_fold, c_path)+" "+err.Error())
				return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")
			}

			template_file := "index"

			var rows []FileRow
			var mode string
			var s_sort string

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

				s_sort = c.Cookies("sort")
				if len(s_sort) == 0 {
					s_sort = "name"
				}

				q_sort := c.Query("sort")
				if len(q_sort) > 0 {
					s_sort = q_sort
				}

				cookie = new(fiber.Cookie)
				cookie.Name = "sort"
				cookie.Value = s_sort
				c.Cookie(cookie)
			}

			rows = listGenerateView(arg_fold, c_path, entries, s_sort)

			mode_thumb := false
			if mode == "thumb" {
				mode_thumb = true
			}

			mode_list := false
			if mode == "list" {
				mode_list = true
			}

			sort_name := false
			if s_sort == "name" {
				sort_name = true
			}

			sort_modified := false
			if s_sort == "modified" {
				sort_modified = true
			}

			sort_size := false
			if s_sort == "size" {
				sort_size = true
			}

			var folderTree_js []byte

			if arg_extend_mode == "1" {
				folderTree := util.WalkAndTreeBuild(arg_fold, "/", 1)

				folderTree_js, err = json.Marshal(folderTree)
				if err != nil {

					panic(err)
					return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")
				}
			}

			return c.Render("view/"+template_file, fiber.Map{

				"Breadcrumb":    template.HTML(breadcrumb),
				"folderTree_js": template.HTML(folderTree_js),

				"rows":            rows,
				"arg_extend_mode": arg_extend_mode,
				"arg_crypt":       arg_crypt,
				"mode_thumb":      mode_thumb,
				"mode_list":       mode_list,

				"sort_name":     sort_name,
				"sort_modified": sort_modified,
				"sort_size":     sort_size,

				"files_count_max":     conf.Files_count_max,
				"fieldSize_max":       conf.FieldSize_max,
				"fieldSize_max_human": conf.FieldSize_max_human,

				"arg_upload_disable":      arg_upload_disable,
				"arg_folder_make_disable": arg_folder_make_disable,
			}, "view/layout/default")

		} else {

			code := c.Cookies("code")

			q_code := c.Query("code")
			if len(q_code) > 0 {
				code = q_code
			}

			is_crypt_ext, _ := regexp.MatchString("\\.crypt$", c_path)

			if arg_crypt == "" && len(code) > 0 {

				cookie := new(fiber.Cookie)
				cookie.Name = "code"
				c.Cookie(cookie)
			}

			if (is_crypt_ext && arg_crypt == "1" && len(code) > 0) || (is_crypt_ext && len(code) > 0) {

				f, err := os.CreateTemp("", "httphere_decrypt*")
				if err != nil {
					panic(err)
				}
				defer os.Remove(f.Name())

				readerFile, _ := os.Open(path.Join(arg_fold, c_path))
				_, err = io.Copy(f, readerFile)
				if err != nil {
					panic(err)
				}
				f.Close()

				isOk, err := util.DecryptFile(f.Name(), code)
				if !isOk {

					model.EventLogAdd(db, c, "500", "CORE", "Error DecryptFile "+err.Error())
					return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")
				}

				model.EventLogAdd(db, c, "200", "CORE", "SendFile decrypt "+path.Join(arg_fold, c_path))

				fname := filepath.Base(c_path)
				fname = strings.Replace(fname, ".crypt", "", 1)
				c.Set(fiber.HeaderContentDisposition, `attachment; filename="`+fname+`"`)

				return c.SendFile(f.Name(), false)

			} else {

				model.EventLogAdd(db, c, "200", "CORE", "SendFile "+path.Join(arg_fold, c_path))

				return c.SendFile(path.Join(arg_fold, c_path), false)
			}

		}

	} else if errors.Is(err, os.ErrNotExist) {

		model.EventLogAdd(db, c, "404", "CORE", path.Join(arg_fold, c_path))

		return c.Status(fiber.StatusNotFound).Render("view/404", fiber.Map{
			"File": c_path,
		}, "view/layout/error")
	}

	return c.Status(fiber.StatusNotFound).Render("view/404", fiber.Map{}, "view/layout/error")

}

type FileRow struct {
	IsDir    bool
	FullPath string
	Name     string

	Size      int64
	SizeHuman string

	ModTime      time.Time
	ModTimeHuman string
	Md5          string

	IsPreviewImg bool
	IsPreviewDoc bool
	IsEditDoc    bool
	IsEditCode   bool
	Rndm         string
}

func listGenerateView(arg_fold string, c_path string, entries []os.DirEntry, s_sort string) []FileRow {

	db, err := model.ConnectToSQLite()
	if err != nil {
		panic(err)
	}

	var rows_dir []FileRow
	var rows_file []FileRow

	for _, e := range entries {

		ext := filepath.Ext(e.Name())
		ext = strings.ToLower(ext)
		ext = strings.Replace(ext, ".", "", -1)

		is_preview_img, _ := regexp.MatchString("^(jpg|jpeg|png|gif)$", ext)
		is_preview_doc, _ := regexp.MatchString("^(pdf|rtf|doc|docx|xls|xlsx|odt|ods)$", ext)

		is_edit_doc, _ := regexp.MatchString("^(html|rtf|doc|docx|odt)$", ext)
		is_edit_code, _ := regexp.MatchString("^(html|txt|js|css|md)$", ext)

		if fileInfo2, err := os.Stat(path.Join(arg_fold, c_path, e.Name())); err == nil {

			modtime := fileInfo2.ModTime()
			modtime_human := modtime.Format("2006-01-02 15:04:05")

			size := fileInfo2.Size()
			size_human := util.PrettyByteSize(size)

			//fmt.Println("path.Join=", path.Join(c_path, e.Name()))
			//fmt.Println("path.Join=", path.Join(c_path, e.Name()))

			if fileInfo2.IsDir() {
				rows_dir = append(rows_dir, FileRow{
					IsDir: fileInfo2.IsDir(),
					//FullPath: path.Join(c_path, e.Name()),
					FullPath: path.Join(c_path, e.Name()),
					Name:     e.Name(),

					Size:      size,
					SizeHuman: size_human,

					ModTime:      modtime,
					ModTimeHuman: modtime_human,

					//Md5:   "",
					IsPreviewImg: false,
					IsPreviewDoc: false,
					IsEditDoc:    false,
					IsEditCode:   false,
					Rndm:         util.RandStringRunes(2),
				})
			} else {
				rows_file = append(rows_file, FileRow{
					IsDir: fileInfo2.IsDir(),
					//FullPath: path.Join(c_path, e.Name()),
					FullPath: path.Join(c_path, e.Name()),
					Name:     e.Name(),

					Size:      size,
					SizeHuman: size_human,

					ModTime:      modtime,
					ModTimeHuman: modtime_human,

					//Md5:       GetMd5File( path.Join(arg_fold, c_path, e.Name()) ),
					IsPreviewImg: is_preview_img,
					IsPreviewDoc: is_preview_doc,
					IsEditDoc:    is_edit_doc,
					IsEditCode:   is_edit_code,
					Rndm:         util.RandStringRunes(2),
				})

				go model.FileAddAsync(db, path.Join(arg_fold, c_path, e.Name()))
			}

		}

	}

	if s_sort == "name" {

		sort.Slice(rows_dir, func(i, j int) bool {
			return rows_dir[i].Name < rows_dir[j].Name
		})

		sort.Slice(rows_file, func(i, j int) bool {
			return rows_file[i].Name < rows_file[j].Name
		})
	}

	if s_sort == "modified" {

		sort.Slice(rows_dir, func(i, j int) bool {
			return rows_dir[i].ModTime.Unix() < rows_dir[j].ModTime.Unix()
		})

		sort.Slice(rows_file, func(i, j int) bool {
			return rows_file[i].ModTime.Unix() < rows_file[j].ModTime.Unix()
		})
	}

	if s_sort == "size" {

		sort.Slice(rows_dir, func(i, j int) bool {
			return rows_dir[i].Size < rows_dir[j].Size
		})

		sort.Slice(rows_file, func(i, j int) bool {
			return rows_file[i].Size < rows_file[j].Size
		})
	}

	return append(rows_dir, rows_file...)
}
