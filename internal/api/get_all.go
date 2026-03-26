package api

import (
	//"fmt"
	"errors"
	"html/template"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/western/http-here/v2/internal/conf"
	"github.com/western/http-here/v2/internal/crypt"
	"github.com/western/http-here/v2/internal/model2"
	"github.com/western/http-here/v2/internal/util"

	"github.com/gofiber/fiber/v2"
)

func GetAll(c *fiber.Ctx) error {

	c_path, err := url.QueryUnescape(c.Path())
	if err != nil {

		model2.EventLogAdd(c, 500, "CORE", "Error "+path.Join(conf.ArgFold, c_path)+" "+err.Error())
		return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")
	}

	c_path = util.CleanDirtyPath(c_path)

	readTarget := path.Join(conf.ArgFold, c_path)

	// -------------------------------------------------------------------------------------------------------------------------------------------

	fileInfo, err := os.Lstat(readTarget)

	if errors.Is(err, os.ErrNotExist) {

		model2.EventLogAdd(c, 404, "CORE", readTarget)

		return c.Status(fiber.StatusNotFound).Render("view/404", fiber.Map{
			"File": c_path,
		}, "view/layout/error")
	}

	if err != nil {

		model2.EventLogAdd(c, 500, "CORE", "Error "+readTarget+" "+err.Error())
		return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")
	}

	fileMode := fileInfo.Mode()

	if fileMode.IsDir() {

		return serveDirectory(c)
	}

	if fileMode.IsRegular() {

		return serveFile(c, fileMode)
	}

	// FILE is not Regular and not Directory
	// 500 Internal Server Error

	model2.EventLogAdd(c, 500, "CORE", "Error "+readTarget+" is NOT regular file and NOT directory")
	return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")
}

// -------------------------------------------------------------------------------------------------------------------------------------------

func serveDirectory(c *fiber.Ctx) error {

	arg_upload_disable := GetBoolFromLocals(c, "arg_upload_disable")
	arg_folder_make_disable := GetBoolFromLocals(c, "arg_folder_make_disable")
	arg_extend_mode := GetBoolFromLocals(c, "arg_extend_mode")
	arg_crypt := GetBoolFromLocals(c, "arg_crypt")
	arg_spa := GetBoolFromLocals(c, "arg_spa")
	arg_usedb := GetBoolFromLocals(c, "arg_usedb")

	c_path, err := url.QueryUnescape(c.Path())
	if err != nil {

		model2.EventLogAdd(c, 500, "CORE", "Error "+path.Join(conf.ArgFold, c_path)+" "+err.Error())
		return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")
	}

	c_path = util.CleanDirtyPath(c_path)

	readTarget := path.Join(conf.ArgFold, c_path)

	// -------------------------------------------------------------------------------------------------------------------------------------------

	if arg_spa {

		model2.EventLogAdd(c, 302, "CORE", "SPA application, redirect to /#!"+c_path)
		return c.Redirect("/#!"+c_path, 302)
	}

	// -------------------------------------------------------------------------------------------------------------------------------------------

	// check index.html inside
	if _, err := os.Stat(path.Join(readTarget, "index.html")); err == nil {

		model2.EventLogAdd(c, 200, "CORE", "Index file found for path '"+c_path+"', SendFile "+path.Join(readTarget, "index.html"))
		return c.SendFile(path.Join(readTarget, "index.html"), false)
	}

	// -------------------------------------------------------------------------------------------------------------------------------------------

	model2.EventLogAdd(c, 200, "CORE", "Dir "+readTarget)

	breadcrumb := ""
	separator := "/"

	res1 := strings.Split(c_path, separator)
	pt := ""
	for indx, el := range res1 {
		if indx == 0 {
			continue
		}
		pt += separator + el
		breadcrumb += `<li class="breadcrumb-item"><a class="nodecor" href="` + pt + `">` + el + `</a></li>`
	}

	template_file := "index"

	var mode string
	var s_sort string

	if arg_extend_mode {

		template_file = "index_extend"

		mode = c.Cookies("mode")
		if len(mode) == 0 {
			mode = "list"
		}

		q_mode := c.Query("mode")
		if len(q_mode) > 0 {
			mode = q_mode
		}

		setCookie(c, "mode", mode)

		s_sort = c.Cookies("sort")
		if len(s_sort) == 0 {
			s_sort = "name"
		}

		q_sort := c.Query("sort")
		if len(q_sort) > 0 {
			s_sort = q_sort
		}

		setCookie(c, "sort", s_sort)
	}

	rows, err := generateRows(c, s_sort)
	if err != nil {

		if os.IsPermission(err) {

			model2.EventLogAdd(c, 403, "CORE", "Forbidden for read "+readTarget)
			return c.Status(fiber.StatusForbidden).Render("view/error", fiber.Map{
				"Title":   "403",
				"Message": "Forbidden",
			}, "view/layout/error")
		}

		model2.EventLogAdd(c, 500, "CORE", "Read folder err: "+err.Error())
		return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")
	}

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

	return c.Render("view/"+template_file, fiber.Map{

		"Breadcrumb": template.HTML(breadcrumb),

		"rows":            rows,
		"arg_extend_mode": arg_extend_mode,
		"arg_crypt":       arg_crypt,
		"mode_thumb":      mode_thumb,
		"mode_list":       mode_list,

		"sort_name":     sort_name,
		"sort_modified": sort_modified,
		"sort_size":     sort_size,

		"files_count_max":     conf.FilesCountMax,
		"fieldSize_max":       conf.FieldSizeMax,
		"fieldSize_max_human": conf.FieldSizeMaxHuman,

		"arg_upload_disable":      arg_upload_disable,
		"arg_folder_make_disable": arg_folder_make_disable,
		"arg_usedb":               arg_usedb,
	}, "view/layout/default")

}

// -------------------------------------------------------------------------------------------------------------------------------------------

func serveFile(c *fiber.Ctx, fileMode os.FileMode) error {

	arg_crypt := GetBoolFromLocals(c, "arg_crypt")

	c_path, err := url.QueryUnescape(c.Path())
	if err != nil {

		model2.EventLogAdd(c, 500, "CORE", "Error "+path.Join(conf.ArgFold, c_path)+" "+err.Error())
		return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")
	}

	c_path = util.CleanDirtyPath(c_path)

	readTarget := path.Join(conf.ArgFold, c_path)

	// -------------------------------------------------------------------------------------------------------------------------------------------

	code := c.Cookies("code")

	q_code := c.Query("code")
	if len(q_code) > 0 {
		code = q_code
	}

	// reset code for safety
	if !arg_crypt && len(code) > 0 {

		setCookie(c, "code", "")
	}

	is_crypt_ext, _ := regexp.MatchString("\\.crypt$", c_path)

	if (is_crypt_ext && arg_crypt && len(code) > 0) || (is_crypt_ext && len(code) > 0) {

		fileDecrupt, err := os.CreateTemp("", "httphere_decrypt*")
		if err != nil {
			panic(err)
		}
		defer os.Remove(fileDecrupt.Name())

		// ------------------------------------------------------------------------------------------------------------------

		if err := crypt.GCMDecryptFile([]byte(code), readTarget, fileDecrupt.Name()); err != nil {

			model2.EventLogAdd(c, 500, "CORE", "Error DecryptFile "+readTarget+" "+err.Error())
			return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")
		}

		// ------------------------------------------------------------------------------------------------------------------

		fname := filepath.Base(c_path)
		fname = strings.Replace(fname, ".crypt", "", 1)

		c.Set(fiber.HeaderContentDisposition, `attachment; filename="`+fname+`"`)

		model2.EventLogAdd(c, 200, "CORE", "SendFile decrypt "+readTarget+" as "+fname)

		return c.SendFile(fileDecrupt.Name(), false)

	} else {

		if fileMode.Perm()&0444 == 0444 {

		} else {

			model2.EventLogAdd(c, 403, "CORE", "Forbidden for read "+readTarget)
			return c.Status(fiber.StatusForbidden).Render("view/error", fiber.Map{
				"Title":   "403",
				"Message": "Forbidden",
			}, "view/layout/error")
		}

		model2.EventLogAdd(c, 200, "CORE", "SendFile "+readTarget)

		return c.SendFile(readTarget, false)
	}
}

// -------------------------------------------------------------------------------------------------------------------------------------------

var (
	previewImgRegex = regexp.MustCompile("^(jpg|jpeg|png|gif)$")
	previewDocRegex = regexp.MustCompile("^(pdf|rtf|doc|docx|xls|xlsx|odt|ods)$")

	editDocRegex  = regexp.MustCompile("^(html|rtf|doc|docx|odt)$")
	editCodeRegex = regexp.MustCompile("^(html|txt|js|css|md|sh|json)$")
	editMdRegex   = regexp.MustCompile("^(md)$")
)

func generateRows(c *fiber.Ctx, s_sort string) ([]FileRow, error) {

	arg_extend_mode := GetBoolFromLocals(c, "arg_extend_mode")
	arg_cache_dir := GetIntFromLocals(c, "arg_cache_dir", 30)
	arg_crypt := GetBoolFromLocals(c, "arg_crypt")

	c_path, err := url.QueryUnescape(c.Path())
	if err != nil {
		panic(err.Error())
	}

	c_path = util.CleanDirtyPath(c_path)

	readTarget := path.Join(conf.ArgFold, c_path)

	// -------------------------------------------------------------------------------------------------------------------------------------------
	// FOR request with PARAM path=

	u_path := c.FormValue("path")

	u_path = util.CleanDirtyPath(u_path)

	if len(u_path) > 0 {
		c_path = u_path
		readTarget = path.Join(conf.ArgFold, u_path)
	}

	// -------------------------------------------------------------------------------------------------------------------------------------------
	// CACHE

	dirCacheMux.RLock()
	cached, exists := dirCache[readTarget]
	dirCacheMux.RUnlock()

	if !arg_crypt && exists && !cached.isExpired(arg_cache_dir) {

		rows_dir := cached.rows_dir
		rows_file := cached.rows_file

		// -------------------------------------------------------------------------------------------------------------------------------------------
		// SORT AND RETURN

		return sortAndCombine(rows_dir, rows_file, s_sort), nil
	}

	// -------------------------------------------------------------------------------------------------------------------------------------------
	// entries []os.DirEntry

	entries, err := os.ReadDir(readTarget)
	if err != nil {

		model2.EventLogAdd(c, 500, "generateRows", "Error "+path.Join(conf.ArgFold, c_path)+" "+err.Error())

		return []FileRow{}, err
	}

	// -------------------------------------------------------------------------------------------------------------------------------------------

	var rows_dir []FileRow
	var rows_file []FileRow

	// Pre-allocate slices with capacity for better performance
	rows_dir = make([]FileRow, 0, len(entries)/2)
	rows_file = make([]FileRow, 0, len(entries)/2)

	for _, e := range entries {

		ext := filepath.Ext(e.Name())
		ext = strings.ToLower(ext)
		ext = strings.Replace(ext, ".", "", -1)

		is_preview_img := previewImgRegex.MatchString(ext)
		is_preview_doc := previewDocRegex.MatchString(ext)

		is_edit_doc := editDocRegex.MatchString(ext)
		is_edit_code := editCodeRegex.MatchString(ext)
		is_edit_md := editMdRegex.MatchString(ext)

		if fileInfo, err := os.Stat(path.Join(readTarget, e.Name())); err == nil {

			modtime := fileInfo.ModTime()
			modtime_human := modtime.Format("2006-01-02 15:04:05")

			size := fileInfo.Size()
			size_human := util.PrettyByteSize(size)

			if fileInfo.IsDir() {
				rows_dir = append(rows_dir, FileRow{
					IsDir:    fileInfo.IsDir(),
					FullPath: path.Join(c_path, e.Name()),
					Name:     e.Name(),

					Size:      size,
					SizeHuman: size_human,

					ModTime:      modtime,
					ModTimeHuman: modtime_human,

					IsPreviewImg: false,
					IsPreviewDoc: false,
					IsEditDoc:    false,
					IsEditCode:   false,
					IsEditMd:     false,
					Rndm:         util.RandStringRunes(2),
				})
			} else {
				rows_file = append(rows_file, FileRow{
					IsDir:    fileInfo.IsDir(),
					FullPath: path.Join(c_path, e.Name()),
					Name:     e.Name(),

					Size:      size,
					SizeHuman: size_human,

					ModTime:      modtime,
					ModTimeHuman: modtime_human,

					IsPreviewImg: is_preview_img,
					IsPreviewDoc: is_preview_doc,
					IsEditDoc:    is_edit_doc,
					IsEditCode:   is_edit_code,
					IsEditMd:     is_edit_md,
					Rndm:         util.RandStringRunes(2),
				})

				if arg_extend_mode {
					//model2.FileAdd(path.Join(readTarget, e.Name()))
				}
			}

		}

	}

	// -------------------------------------------------------------------------------------------------------------------------------------------

	dirCacheMux.Lock()
	dirCache[readTarget] = &cachedDir{

		rows_dir:  rows_dir,
		rows_file: rows_file,

		timestamp: time.Now(),
	}
	dirCacheMux.Unlock()

	// -------------------------------------------------------------------------------------------------------------------------------------------

	return sortAndCombine(rows_dir, rows_file, s_sort), nil
}

func sortAndCombine(rows_dir, rows_file []FileRow, s_sort string) []FileRow {

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
