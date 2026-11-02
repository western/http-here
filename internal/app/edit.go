package app

import (
	"bufio"
	//"fmt"
	"html/template"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/western/http-here/v2/internal/api"
	"github.com/western/http-here/v2/internal/conf"
	"github.com/western/http-here/v2/internal/model2"
	"github.com/western/http-here/v2/internal/util"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/csrf"
)

var isOfficeRegex = regexp.MustCompile("^(html|rtf|doc|docx|odt)$")

// trimRootPath - make path relative for use under RootFS
func trimRootPath(p string) string {
	return strings.TrimPrefix(p, "/")
}

func GetEditDoc(c fiber.Ctx) error {

	c_path, err := url.QueryUnescape(c.Path())
	if err != nil {

		model2.EventLogAdd(c, 500, "GetEditDoc", "Error "+filepath.Join(conf.ArgFold, c_path)+" "+err.Error())

		return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")
	}

	c_path = util.CleanDirtyPath(c_path)

	c_path = strings.ReplaceAll(c_path, "/__doc", "")

	// --------------------------------------------------------------------------------------------------------------------------------

	Root := conf.ArgFoldRootFS

	// stat under confined root of ArgFold
	if _, err := Root.Stat(trimRootPath(c_path)); err != nil {

		model2.EventLogAdd(c, 404, "GetEditDoc", filepath.Join(conf.ArgFold, c_path)+" "+err.Error())

		return c.Status(fiber.StatusNotFound).Render("view/404", fiber.Map{}, "view/layout/error")
	}

	// --------------------------------------------------------------------------------------------------------------------------------

	file_ext := util.GetExtNorm(c_path)
	orig_filename := util.GetFileName(c_path)

	// html|rtf|doc|docx|odt
	is_office_match := isOfficeRegex.MatchString(file_ext)
	if !is_office_match {

		model2.EventLogAdd(c, 500, "GetEditDoc", "Error: format of file is not for edit")
		return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")
	}

	// --------------------------------------------------------------------------------------------------------------------------------

	if runtime.GOOS == "windows" {

		_, err := os.Stat("C:/Program Files/LibreOffice/program/soffice.exe")
		if err != nil {

			model2.EventLogAdd(c, 500, "GetEditDoc", "Error libreoffice not found "+err.Error())

			return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")
		}

	} else {

		_, err := exec.Command("bash", "-c", "libreoffice --help").Output()
		if err != nil {

			model2.EventLogAdd(c, 500, "GetEditDoc", "Error libreoffice not found "+err.Error())

			return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")
		}
	}

	filepath_tmp := filepath.Join(conf.ConfigRoot, "temp")

	arg_csrf_disable := GetBoolFromLocals(c, "arg_csrf_disable")

	csrf_token := ""
	if !arg_csrf_disable {
		csrf_token = csrf.TokenFromContext(c)
	}

	tempRoot := conf.TempRootFS

	if runtime.GOOS == "windows" {

		// --------------------------------------------------------------------------------------------------------------------------------

		//filepath_tmp := util.RotateSlash(path.Join(conf.ConfigRoot, "temp"))
		//arg_fold_path := util.RotateSlash(path.Join(conf.ArgFold, c_path))

		arg_fold_path := filepath.Join(conf.ArgFold, c_path)

		util.RunAnyCommandUnderWin(`"C:/Program Files/LibreOffice/program/soffice.exe" --headless --norestore --nologo --convert-to html --outdir "` + filepath_tmp + `" "` + arg_fold_path + `"`)

		b, err := tempRoot.ReadFile(orig_filename + ".html")
		if err != nil {

			model2.EventLogAdd(c, 500, "GetEditDoc", "Error libreoffice, open file "+err.Error())

			return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")
		}

		// --------------------------------------------------------------------------------------------------------------------------------

		model2.EventLogAdd(c, 200, "GetEditDoc", "Open for edit "+filepath.Join(conf.ArgFold, c_path))

		return c.Render("view/edit/edit_doc", fiber.Map{
			"file_name": orig_filename + "." + file_ext,
			"full_path": c_path,

			"file_data": template.HTML(string(b)),

			"csrf_token": csrf_token,
		}, "view/layout/default")

	} else {

		// --------------------------------------------------------------------------------------------------------------------------------

		cmd := exec.Command("bash", "-c", "libreoffice --headless --norestore --nologo --convert-to html --outdir "+filepath_tmp+" \""+filepath.Join(conf.ArgFold, c_path)+"\"")
		cmd.Dir = conf.ArgFold

		stderr, _ := cmd.StderrPipe()
		if err := cmd.Start(); err != nil {
			panic(err)
		}

		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			//fmt.Println("libreoffice:", scanner.Text())
		}

		b, err := tempRoot.ReadFile(orig_filename + ".html")
		if err != nil {

			model2.EventLogAdd(c, 500, "GetEditDoc", "Error libreoffice, open file "+err.Error())

			return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")
		}

		// --------------------------------------------------------------------------------------------------------------------------------

		model2.EventLogAdd(c, 200, "GetEditDoc", "Open for edit "+filepath.Join(conf.ArgFold, c_path))

		return c.Render("view/edit/edit_doc", fiber.Map{
			"file_name": orig_filename + "." + file_ext,
			"full_path": c_path,

			"file_data": template.HTML(string(b)),

			"csrf_token": csrf_token,
		}, "view/layout/default")
	}

}

var isSimpleSourceRegex = regexp.MustCompile("^(html|txt|js|css|md|sh|json)$")

func GetEditCode(c fiber.Ctx) error {

	c_path, err := url.QueryUnescape(c.Path())
	if err != nil {

		model2.EventLogAdd(c, 500, "GetEditCode", "Error "+filepath.Join(conf.ArgFold, c_path)+" "+err.Error())

		return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")
	}

	c_path = util.CleanDirtyPath(c_path)

	c_path = strings.ReplaceAll(c_path, "/__code", "")

	// --------------------------------------------------------------------------------------------------------------------------------

	Root := conf.ArgFoldRootFS

	// stat under confined root of ArgFold
	if _, err := Root.Stat(trimRootPath(c_path)); err != nil {

		model2.EventLogAdd(c, 404, "GetEditCode", filepath.Join(conf.ArgFold, c_path))

		return c.Status(fiber.StatusNotFound).Render("view/404", fiber.Map{}, "view/layout/error")
	}

	// --------------------------------------------------------------------------------------------------------------------------------

	file_ext := util.GetExtNorm(c_path)
	orig_filename := util.GetFileName(c_path)

	// html|txt|js|css|md
	is_simple_source_match := isSimpleSourceRegex.MatchString(file_ext)
	if !is_simple_source_match {

		model2.EventLogAdd(c, 500, "GetEditCode", "Error: format of file is not for edit")
		return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")
	}

	// --------------------------------------------------------------------------------------------------------------------------------

	// read source file under confined root of ArgFold
	b, err := Root.ReadFile(trimRootPath(c_path))
	if err != nil {

		model2.EventLogAdd(c, 500, "GetEditCode", "Error open source file: "+err.Error())

		return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")
	}

	// --------------------------------------------------------------------------------------------------------------------------------

	arg_csrf_disable := GetBoolFromLocals(c, "arg_csrf_disable")

	csrf_token := ""
	if !arg_csrf_disable {
		csrf_token = csrf.TokenFromContext(c)
	}

	model2.EventLogAdd(c, 200, "GetEditCode", "Open for edit "+filepath.Join(conf.ArgFold, c_path))

	return c.Render("view/edit/edit_code", fiber.Map{
		"file_name": orig_filename + "." + file_ext,
		"full_path": c_path,

		"file_data": template.HTML(string(b)),

		"csrf_token": csrf_token,
	}, "view/layout/default")

}

func GetEditMd(c fiber.Ctx) error {

	c_path, err := url.QueryUnescape(c.Path())
	if err != nil {

		model2.EventLogAdd(c, 500, "GetEditMd", "Error "+filepath.Join(conf.ArgFold, c_path)+" "+err.Error())

		return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")
	}

	c_path = util.CleanDirtyPath(c_path)

	c_path = strings.ReplaceAll(c_path, "/__md", "")

	// --------------------------------------------------------------------------------------------------------------------------------

	Root := conf.ArgFoldRootFS

	// stat under confined root of ArgFold
	if _, err := Root.Stat(trimRootPath(c_path)); err != nil {

		model2.EventLogAdd(c, 404, "GetEditMd", filepath.Join(conf.ArgFold, c_path))

		return c.Status(fiber.StatusNotFound).Render("view/404", fiber.Map{}, "view/layout/error")
	}

	// --------------------------------------------------------------------------------------------------------------------------------

	file_ext := util.GetExtNorm(c_path)
	orig_filename := util.GetFileName(c_path)

	is_md_match, _ := regexp.MatchString("^(md)$", file_ext)

	if !is_md_match {

		model2.EventLogAdd(c, 500, "GetEditMd", "Error: format of file is not for edit")
		return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")
	}

	// --------------------------------------------------------------------------------------------------------------------------------

	// read source file under confined root of ArgFold
	b, err := Root.ReadFile(trimRootPath(c_path))
	if err != nil {

		model2.EventLogAdd(c, 500, "GetEditMd", "Error open source file: "+err.Error())

		return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")
	}

	// --------------------------------------------------------------------------------------------------------------------------------

	arg_csrf_disable := GetBoolFromLocals(c, "arg_csrf_disable")

	csrf_token := ""
	if !arg_csrf_disable {
		csrf_token = csrf.TokenFromContext(c)
	}

	model2.EventLogAdd(c, 200, "GetEditMd", "Open for edit "+filepath.Join(conf.ArgFold, c_path))

	return c.Render("view/edit/edit_md", fiber.Map{
		"file_name": orig_filename + "." + file_ext,
		"full_path": c_path,

		"file_data": template.HTML(string(b)),

		"csrf_token": csrf_token,
	}, "view/layout/default")

}

func PostFileEdit(c fiber.Ctx) error {

	full_path := util.CleanDirtyPath(c.FormValue("full_path"))

	// --------------------------------------------------------------------------------------------------------------------------------

	Root := conf.ArgFoldRootFS

	// stat under confined root of ArgFold
	if _, err := Root.Stat(trimRootPath(full_path)); err != nil {

		model2.EventLogAdd(c, 500, "PostFileEdit", filepath.Join(conf.ArgFold, full_path))

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  filepath.Join(conf.ArgFold, full_path) + " not found",
		}, "application/json")
	}

	// --------------------------------------------------------------------------------------------------------------------------------

	file_ext := util.GetExtNorm(full_path)
	orig_filename := util.GetFileName(full_path)

	body := c.FormValue("body")

	save_as_source := c.FormValue("save_as_source")

	// --------------------------------------------------------------------------------------------------------------------------------

	if save_as_source == "1" {

		// write body atomically under confined root of ArgFold
		err := Root.WriteFile(trimRootPath(full_path), []byte(body), 0644)

		if err != nil {

			model2.EventLogAdd(c, 500, "PostFileEdit", "Update error "+err.Error())

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "Update error ",
			}, "application/json")

		} else {

			//go model2.FileDelAsync(path.Join(conf.ArgFold, full_path))

			model2.EventLogAdd(c, 200, "PostFileEdit", "Update "+filepath.Join(conf.ArgFold, full_path))

			api.RemoveCacheDir(path.Dir(filepath.Join(conf.ArgFold, full_path)))

			return c.JSON(fiber.Map{
				"code": 200,
			}, "application/json")
		}

		// --------------------------------------------------------------------------------------------------------------------------------

	}

	tempRoot := conf.TempRootFS

	// --------------------------------------------------------------------------------------------------------------------------------

	// html|rtf|doc|docx|odt
	is_office_match := isOfficeRegex.MatchString(file_ext)

	if is_office_match {

		// --------------------------------------------------------------------------------------------------------------------------------

		filepath_tmp := filepath.Join(conf.ConfigRoot, "temp")

		// write html body atomically under confined root of temp
		if err := tempRoot.WriteFile(orig_filename+".html", []byte(body), 0644); err != nil {

			model2.EventLogAdd(c, 500, "PostFileEdit", "Update error "+err.Error())

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "Update error ",
			}, "application/json")
		}

		model2.EventLogAdd(c, 200, "PostFileEdit", "Update "+filepath.Join(filepath_tmp, orig_filename+".html"))

		// --------------------------------------------------------------------------------------------------------------------------------

		if runtime.GOOS == "windows" {

			_, err := os.Stat("C:/Program Files/LibreOffice/program/soffice.exe")
			if err != nil {

				model2.EventLogAdd(c, 500, "PostFileEdit", "Error libreoffice not found "+err.Error())

				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code": 500,
					"msg":  "Libreoffice not found ",
				}, "application/json")
			}

		} else {

			_, err := exec.Command("bash", "-c", "libreoffice --help").Output()
			if err != nil {

				model2.EventLogAdd(c, 500, "PostFileEdit", "Error libreoffice not found "+err.Error())

				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code": 500,
					"msg":  "Libreoffice not found ",
				}, "application/json")
			}
		}

		// --------------------------------------------------------------------------------------------------------------------------------

		convert_format := ""

		switch file_ext {
		case "doc":
			convert_format = "doc:MS Word 97"
		case "docx":
			convert_format = "docx:MS Word 2007 XML"
		case "odt":
			convert_format = "odt:writer8"
		case "rtf":
			convert_format = "rtf:Rich Text Format"
		}

		html_temp_file := filepath.Join(filepath_tmp, orig_filename+".html")

		if runtime.GOOS == "windows" {

			// --------------------------------------------------------------------------------------------------------------------------------

			util.RunAnyCommandUnderWin(`"C:/Program Files/LibreOffice/program/soffice.exe" --headless --norestore --nologo --convert-to "` + convert_format + `" --outdir "` + filepath_tmp + `" "` + html_temp_file + `"`)

		} else {

			// --------------------------------------------------------------------------------------------------------------------------------

			cmd := exec.Command("bash", "-c", "libreoffice --headless --norestore --nologo --convert-to \""+convert_format+"\" --outdir "+filepath_tmp+" \""+html_temp_file+"\"")
			cmd.Dir = filepath_tmp

			stderr, _ := cmd.StderrPipe()
			if err := cmd.Start(); err != nil {
				panic(err)
			}

			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				//fmt.Println("libreoffice:", scanner.Text())
			}

		}

		// --------------------------------------------------------------------------------------------------------------------------------

		// read converted file under confined root of temp
		b, err := tempRoot.ReadFile(orig_filename + "." + file_ext)

		if err != nil {

			model2.EventLogAdd(c, 500, "PostFileEdit", "Convert error "+err.Error())

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "Convert error ",
			}, "application/json")
		}

		// write converted result atomically under confined root of ArgFold
		err = Root.WriteFile(trimRootPath(full_path), b, 0644)

		if err != nil {

			model2.EventLogAdd(c, 500, "PostFileEdit", "Rename error "+err.Error())

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "Rename error ",
			}, "application/json")

		} else {

			// remove temporary html
			tempRoot.Remove(orig_filename + ".html")

			//model2.FileDelMd5Async(path.Join(conf.ArgFold, full_path))

			//model2.FileDelAsync(path.Join(conf.ArgFold, full_path))

			model2.EventLogAdd(c, 200, "PostFileEdit", "Update "+filepath.Join(conf.ArgFold, full_path))

			api.RemoveCacheDir(path.Dir(filepath.Join(conf.ArgFold, full_path)))

			return c.JSON(fiber.Map{
				"code": 200,
			}, "application/json")
		}

		// --------------------------------------------------------------------------------------------------------------------------------

	}

	model2.EventLogAdd(c, 500, "PostFileEdit", "Error: format of file is not for edit")

	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"code": 500,
		"msg":  "Error: format of file is not for edit",
	}, "application/json")

}
