package app

import (
	"bufio"
	//"fmt"
	"html/template"
	"net/url"
	"os"
	"os/exec"
	"path"
	"regexp"
	"runtime"
	"strings"

	"github.com/western/http-here/v2/internal/api"
	"github.com/western/http-here/v2/internal/conf"
	"github.com/western/http-here/v2/internal/model2"
	"github.com/western/http-here/v2/internal/util"

	"github.com/gofiber/fiber/v2"
)

var isOfficeRegex = regexp.MustCompile("^(html|rtf|doc|docx|odt)$")

func GetEditDoc(c *fiber.Ctx) error {

	c_path, err := url.QueryUnescape(c.Path())
	if err != nil {

		model2.EventLogAdd(c, 500, "GetEditDoc", "Error "+path.Join(conf.ArgFold, c_path)+" "+err.Error())

		return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")
	}

	c_path = util.CleanDirtyPath(c_path)

	c_path = strings.ReplaceAll(c_path, "/__doc", "")

	if _, err := os.Stat(path.Join(conf.ArgFold, c_path)); err != nil {

		model2.EventLogAdd(c, 404, "GetEditDoc", path.Join(conf.ArgFold, c_path))

		return c.Status(fiber.StatusNotFound).Render("view/404", fiber.Map{}, "view/layout/error")
	}

	file_ext := util.GetExtNorm(c_path)
	orig_filename := util.GetFileName(c_path)

	// --------------------------------------------------------------------------------------------------------------------------------

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

	filepath_tmp := path.Join(conf.ConfigRoot, "temp")

	if runtime.GOOS == "windows" {

		// --------------------------------------------------------------------------------------------------------------------------------

		//filepath_tmp := util.RotateSlash(path.Join(conf.ConfigRoot, "temp"))
		//arg_fold_path := util.RotateSlash(path.Join(conf.ArgFold, c_path))

		arg_fold_path := path.Join(conf.ArgFold, c_path)

		util.RunAnyCommandUnderWin(`"C:/Program Files/LibreOffice/program/soffice.exe" --headless --norestore --nologo --convert-to html --outdir "` + filepath_tmp + `" "` + arg_fold_path + `"`)

		b, err := os.ReadFile(path.Join(filepath_tmp, orig_filename+".html"))
		if err != nil {

			model2.EventLogAdd(c, 500, "GetEditDoc", "Error libreoffice, open file "+err.Error())

			return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")
		}

		// --------------------------------------------------------------------------------------------------------------------------------

		model2.EventLogAdd(c, 200, "GetEditDoc", "Open for edit "+path.Join(conf.ArgFold, c_path))

		return c.Render("view/edit/edit_doc", fiber.Map{
			"file_name": orig_filename + "." + file_ext,
			"full_path": c_path,

			"file_data": template.HTML(string(b)),
		}, "view/layout/default")

	} else {

		// --------------------------------------------------------------------------------------------------------------------------------

		cmd := exec.Command("bash", "-c", "libreoffice --headless --norestore --nologo --convert-to html --outdir "+filepath_tmp+" \""+path.Join(conf.ArgFold, c_path)+"\"")
		cmd.Dir = conf.ArgFold

		stderr, _ := cmd.StderrPipe()
		if err := cmd.Start(); err != nil {
			panic(err)
		}

		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			//fmt.Println("libreoffice:", scanner.Text())
		}

		b, err := os.ReadFile(path.Join(filepath_tmp, orig_filename+".html"))
		if err != nil {

			model2.EventLogAdd(c, 500, "GetEditDoc", "Error libreoffice, open file "+err.Error())

			return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")
		}

		// --------------------------------------------------------------------------------------------------------------------------------

		model2.EventLogAdd(c, 200, "GetEditDoc", "Open for edit "+path.Join(conf.ArgFold, c_path))

		return c.Render("view/edit/edit_doc", fiber.Map{
			"file_name": orig_filename + "." + file_ext,
			"full_path": c_path,

			"file_data": template.HTML(string(b)),
		}, "view/layout/default")
	}

}

var isSimpleSourceRegex = regexp.MustCompile("^(html|txt|js|css|md|sh|json)$")

func GetEditCode(c *fiber.Ctx) error {

	c_path, err := url.QueryUnescape(c.Path())
	if err != nil {

		model2.EventLogAdd(c, 500, "GetEditCode", "Error "+path.Join(conf.ArgFold, c_path)+" "+err.Error())

		return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")
	}

	c_path = util.CleanDirtyPath(c_path)

	c_path = strings.ReplaceAll(c_path, "/__code", "")

	if _, err := os.Stat(path.Join(conf.ArgFold, c_path)); err != nil {

		model2.EventLogAdd(c, 404, "GetEditCode", path.Join(conf.ArgFold, c_path))

		return c.Status(fiber.StatusNotFound).Render("view/404", fiber.Map{}, "view/layout/error")
	}

	file_ext := util.GetExtNorm(c_path)
	orig_filename := util.GetFileName(c_path)

	// --------------------------------------------------------------------------------------------------------------------------------

	// html|txt|js|css|md
	is_simple_source_match := isSimpleSourceRegex.MatchString(file_ext)
	if !is_simple_source_match {

		model2.EventLogAdd(c, 500, "GetEditCode", "Error: format of file is not for edit")
		return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")
	}

	// --------------------------------------------------------------------------------------------------------------------------------

	filepath_tmp := path.Join(conf.ConfigRoot, "temp")

	util.CopyFile(
		path.Join(conf.ArgFold, c_path),
		path.Join(filepath_tmp, orig_filename+"."+file_ext),
	)

	b, err := os.ReadFile(path.Join(filepath_tmp, orig_filename+"."+file_ext))
	if err != nil {

		model2.EventLogAdd(c, 500, "GetEditCode", "Error open temp source file: "+err.Error())

		return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")
	}

	// --------------------------------------------------------------------------------------------------------------------------------

	model2.EventLogAdd(c, 200, "GetEditCode", "Open for edit "+path.Join(conf.ArgFold, c_path))

	return c.Render("view/edit/edit_code", fiber.Map{
		"file_name": orig_filename + "." + file_ext,
		"full_path": c_path,

		"file_data": template.HTML(string(b)),
	}, "view/layout/default")

}

func GetEditMd(c *fiber.Ctx) error {

	c_path, err := url.QueryUnescape(c.Path())
	if err != nil {

		model2.EventLogAdd(c, 500, "GetEditMd", "Error "+path.Join(conf.ArgFold, c_path)+" "+err.Error())

		return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")
	}

	c_path = util.CleanDirtyPath(c_path)

	c_path = strings.ReplaceAll(c_path, "/__md", "")

	if _, err := os.Stat(path.Join(conf.ArgFold, c_path)); err != nil {

		model2.EventLogAdd(c, 404, "GetEditMd", path.Join(conf.ArgFold, c_path))

		return c.Status(fiber.StatusNotFound).Render("view/404", fiber.Map{}, "view/layout/error")
	}

	file_ext := util.GetExtNorm(c_path)
	orig_filename := util.GetFileName(c_path)

	// --------------------------------------------------------------------------------------------------------------------------------

	is_md_match, _ := regexp.MatchString("^(md)$", file_ext)

	if !is_md_match {

		model2.EventLogAdd(c, 500, "GetEditMd", "Error: format of file is not for edit")
		return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")
	}

	// --------------------------------------------------------------------------------------------------------------------------------

	filepath_tmp := path.Join(conf.ConfigRoot, "temp")

	util.CopyFile(
		path.Join(conf.ArgFold, c_path),
		path.Join(filepath_tmp, orig_filename+"."+file_ext),
	)

	b, err := os.ReadFile(path.Join(filepath_tmp, orig_filename+"."+file_ext))
	if err != nil {

		model2.EventLogAdd(c, 500, "GetEditMd", "Error open temp source file: "+err.Error())

		return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")
	}

	// --------------------------------------------------------------------------------------------------------------------------------

	model2.EventLogAdd(c, 200, "GetEditMd", "Open for edit "+path.Join(conf.ArgFold, c_path))

	return c.Render("view/edit/edit_md", fiber.Map{
		"file_name": orig_filename + "." + file_ext,
		"full_path": c_path,

		"file_data": template.HTML(string(b)),
	}, "view/layout/default")

}

func PostFileEdit(c *fiber.Ctx) error {

	full_path := c.FormValue("full_path")

	if _, err := os.Stat(path.Join(conf.ArgFold, full_path)); err != nil {

		model2.EventLogAdd(c, 500, "PostFileEdit", path.Join(conf.ArgFold, full_path))

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  path.Join(conf.ArgFold, full_path) + " not found",
		}, "application/json")
	}

	file_ext := util.GetExtNorm(full_path)
	orig_filename := util.GetFileName(full_path)

	body := c.FormValue("body")

	save_as_source := c.FormValue("save_as_source")

	// --------------------------------------------------------------------------------------------------------------------------------

	if save_as_source == "1" {

		// --------------------------------------------------------------------------------------------------------------------------------

		filepath_tmp := path.Join(conf.ConfigRoot, "temp")
		source_temp_file := path.Join(filepath_tmp, orig_filename+"."+file_ext)

		if runtime.GOOS == "windows" {
			filepath_tmp = util.RotateSlash(filepath_tmp)
			source_temp_file = util.RotateSlash(source_temp_file)
		}

		f, err := os.Create(source_temp_file)
		if err != nil {
			panic(err)
		}

		_, err = f.WriteString(body)
		if err != nil {
			panic(err)
		}
		f.Close()

		model2.EventLogAdd(c, 200, "PostFileEdit", "Update "+source_temp_file)

		// --------------------------------------------------------------------------------------------------------------------------------

		target_file := path.Join(conf.ArgFold, full_path)

		if runtime.GOOS == "windows" {
			target_file = util.RotateSlash(target_file)
		}

		err = util.MoveFile(source_temp_file, target_file)

		if err != nil {

			model2.EventLogAdd(c, 500, "PostFileEdit", "Rename error "+err.Error())

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "Rename error ",
			}, "application/json")

		} else {

			go model2.FileDelAsync(path.Join(conf.ArgFold, full_path))

			model2.EventLogAdd(c, 200, "PostFileEdit", "Move "+source_temp_file+" => "+target_file)

			api.RemoveCacheDir(path.Dir(target_file))

			return c.JSON(fiber.Map{
				"code": 200,
			}, "application/json")
		}

		// --------------------------------------------------------------------------------------------------------------------------------

	}

	// --------------------------------------------------------------------------------------------------------------------------------

	// html|rtf|doc|docx|odt
	is_office_match := isOfficeRegex.MatchString(file_ext)

	if is_office_match {

		// --------------------------------------------------------------------------------------------------------------------------------

		filepath_tmp := path.Join(conf.ConfigRoot, "temp")
		html_temp_file := path.Join(filepath_tmp, orig_filename+".html")

		if runtime.GOOS == "windows" {
			filepath_tmp = util.RotateSlash(filepath_tmp)
			html_temp_file = util.RotateSlash(html_temp_file)
		}

		f, err := os.Create(html_temp_file)
		if err != nil {
			panic(err)
		}

		_, err = f.WriteString(body)
		if err != nil {
			panic(err)
		}
		f.Close()

		model2.EventLogAdd(c, 200, "PostFileEdit", "Update "+html_temp_file)

		// --------------------------------------------------------------------------------------------------------------------------------

		/*
			_, err = exec.Command("bash", "-c", "libreoffice --help").Output()
			if err != nil {


				model.EventLogAdd(db, c, "500", "PostFileEdit", "Error libreoffice not found "+err.Error())

				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code": 500,
					"msg":  "Libreoffice not found ",
				}, "application/json")
			}
		*/

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

		//filepath_tmp := path.Join(conf.ConfigRoot, "temp")
		//html_temp_file := path.Join(filepath_tmp, orig_filename+".html")

		if runtime.GOOS == "windows" {

			// --------------------------------------------------------------------------------------------------------------------------------

			//filepath_tmp := util.RotateSlash(path.Join(conf.ConfigRoot, "temp"))
			//html_temp_file := util.RotateSlash(path.Join(filepath_tmp, orig_filename+".html"))

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

		source_temp_file := path.Join(filepath_tmp, orig_filename+"."+file_ext)
		target_file := path.Join(conf.ArgFold, full_path)

		/*
			if runtime.GOOS == "windows" {
			    source_temp_file = util.RotateSlash(source_temp_file)
			    target_file = util.RotateSlash(target_file)
			}
		*/

		err = util.MoveFile(source_temp_file, target_file)

		if err != nil {

			model2.EventLogAdd(c, 500, "PostFileEdit", "Rename error "+err.Error())

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "Rename error ",
			}, "application/json")

		} else {

			// remove temporary html
			if _, err := os.Stat(path.Join(filepath_tmp, orig_filename+".html")); err == nil {

				os.Remove(path.Join(filepath_tmp, orig_filename+".html"))
			}

			model2.FileDelMd5Async(path.Join(conf.ArgFold, full_path))

			model2.FileDelAsync(path.Join(conf.ArgFold, full_path))

			model2.EventLogAdd(c, 200, "PostFileEdit", "Move "+source_temp_file+" => "+target_file)

			api.RemoveCacheDir(path.Dir(target_file))

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
