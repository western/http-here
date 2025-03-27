package app

import (
	"bufio"
	_ "fmt"
	"html/template"
	"net/url"
	"os"
	"os/exec"
	"path"
	_ "path/filepath"
	"regexp"
	"runtime"
	_ "strconv"
	"strings"

	"github.com/western/http-here/internal/model"
	"github.com/western/http-here/internal/util"

	"github.com/gofiber/fiber/v2"

	"gorm.io/gorm"
)

func GetEditDoc(c *fiber.Ctx) error {

	arg_fold := ""
	arg_fold = c.Locals("arg_fold").(string)

	prefix := c.Locals("prefix").(string)

	db := c.Locals("db").(*gorm.DB)

	/*
		homepath, err := os.UserHomeDir()
		if err != nil {

			model.EventLogAdd(db, c, "500", "GetEditDoc", "Error hmepath detect "+err.Error())

			return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")
		}
	*/

	c_path, err := url.QueryUnescape(c.Path())
	if err != nil {

		model.EventLogAdd(db, c, "500", "GetEditDoc", "Error "+path.Join(arg_fold, c_path)+" "+err.Error())

		return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")
	}

	c_path = util.CleanDirtyPath(c_path)

	c_path = strings.ReplaceAll(c_path, "/__doc", "")

	if _, err := os.Stat(path.Join(arg_fold, c_path)); err != nil {

		model.EventLogAdd(db, c, "404", "GetEditDoc", path.Join(arg_fold, c_path))

		return c.Status(fiber.StatusNotFound).Render("view/404", fiber.Map{}, "view/layout/error")
	}

	file_ext := util.GetExtNorm(c_path)
	orig_filename := util.GetFileName(c_path)

	is_office_match, _ := regexp.MatchString("^(html|rtf|doc|docx|odt)$", file_ext)

	if is_office_match {

		// --------------------------------------------------------------------------------------------------------------------------------

		if runtime.GOOS == "windows" {

			_, err := os.Stat("C:/Program Files/LibreOffice/program/soffice.exe")
			if err != nil {

				model.EventLogAdd(db, c, "500", "GetEditDoc", "Error libreoffice not found "+err.Error())

				return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")
			}

		} else {

			_, err := exec.Command("bash", "-c", "libreoffice --help").Output()
			if err != nil {

				model.EventLogAdd(db, c, "500", "GetEditDoc", "Error libreoffice not found "+err.Error())

				return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")
			}
		}

		if runtime.GOOS == "windows" {

			// --------------------------------------------------------------------------------------------------------------------------------

			//filepath_tmp := util.RotateSlash(path.Join(prefix, "temp"))
			//arg_fold_path := util.RotateSlash(path.Join(arg_fold, c_path))

			filepath_tmp := path.Join(prefix, "temp")
			arg_fold_path := path.Join(arg_fold, c_path)

			util.RunAnyCommandUnderWin(`"C:/Program Files/LibreOffice/program/soffice.exe" --headless --norestore --nologo --convert-to html --outdir "` + filepath_tmp + `" "` + arg_fold_path + `"`)

			b, err := os.ReadFile(path.Join(filepath_tmp, orig_filename+".html"))
			if err != nil {

				model.EventLogAdd(db, c, "500", "GetEditDoc", "Error libreoffice, open file "+err.Error())

				return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")
			}

			// --------------------------------------------------------------------------------------------------------------------------------

			model.EventLogAdd(db, c, "200", "GetEditDoc", "Open for edit "+path.Join(arg_fold, c_path))

			return c.Render("view/edit/edit_doc", fiber.Map{
				"file_name": orig_filename + "." + file_ext,
				"full_path": c_path,

				"file_data": template.HTML(string(b)),
			}, "view/layout/default")

		} else {

			// --------------------------------------------------------------------------------------------------------------------------------

			filepath_tmp := path.Join(prefix, "temp")

			cmd := exec.Command("bash", "-c", "libreoffice --headless --norestore --nologo --convert-to html --outdir "+filepath_tmp+" \""+path.Join(arg_fold, c_path)+"\"")
			cmd.Dir = arg_fold

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

				model.EventLogAdd(db, c, "500", "GetEditDoc", "Error libreoffice, open file "+err.Error())

				return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")
			}

			// --------------------------------------------------------------------------------------------------------------------------------

			model.EventLogAdd(db, c, "200", "GetEditDoc", "Open for edit "+path.Join(arg_fold, c_path))

			return c.Render("view/edit/edit_doc", fiber.Map{
				"file_name": orig_filename + "." + file_ext,
				"full_path": c_path,

				"file_data": template.HTML(string(b)),
			}, "view/layout/default")
		}
	}

	model.EventLogAdd(db, c, "500", "GetEditDoc", "Error: format of file is not for edit")

	return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")

}

func GetEditCode(c *fiber.Ctx) error {

	arg_fold := ""
	arg_fold = c.Locals("arg_fold").(string)

	prefix := c.Locals("prefix").(string)

	db := c.Locals("db").(*gorm.DB)

	/*
		homepath, err := os.UserHomeDir()
		if err != nil {

			model.EventLogAdd(db, c, "500", "GetEditCode", "Error homepath detect "+err.Error())

			return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")
		}*/

	c_path, err := url.QueryUnescape(c.Path())
	if err != nil {

		model.EventLogAdd(db, c, "500", "GetEditCode", "Error "+path.Join(arg_fold, c_path)+" "+err.Error())

		return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")
	}

	c_path = util.CleanDirtyPath(c_path)

	c_path = strings.ReplaceAll(c_path, "/__code", "")

	if _, err := os.Stat(path.Join(arg_fold, c_path)); err != nil {

		model.EventLogAdd(db, c, "404", "GetEditCode", path.Join(arg_fold, c_path))

		return c.Status(fiber.StatusNotFound).Render("view/404", fiber.Map{}, "view/layout/error")
	}

	file_ext := util.GetExtNorm(c_path)
	orig_filename := util.GetFileName(c_path)

	is_code_match, _ := regexp.MatchString("^(html|txt|js|css|md)$", file_ext)

	if is_code_match {

		filepath_tmp := path.Join(prefix, "temp")

		// --------------------------------------------------------------------------------------------------------------------------------

		util.CopyFile(
			path.Join(arg_fold, c_path),
			path.Join(filepath_tmp, orig_filename+"."+file_ext),
		)

		b, err := os.ReadFile(path.Join(filepath_tmp, orig_filename+"."+file_ext))
		if err != nil {

			model.EventLogAdd(db, c, "500", "GetEditCode", "Error open temp source file: "+err.Error())

			return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")
		}

		// --------------------------------------------------------------------------------------------------------------------------------

		model.EventLogAdd(db, c, "200", "GetEditCode", "Open for edit "+path.Join(arg_fold, c_path))

		return c.Render("view/edit/edit_code", fiber.Map{
			"file_name": orig_filename + "." + file_ext,
			"full_path": c_path,

			"file_data": template.HTML(string(b)),
		}, "view/layout/default")

	}

	model.EventLogAdd(db, c, "500", "GetEditCode", "Error: format of file is not for edit")

	return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")

}

func PostFileEdit(c *fiber.Ctx) error {

	arg_fold := ""
	arg_fold = c.Locals("arg_fold").(string)

	prefix := c.Locals("prefix").(string)

	db := c.Locals("db").(*gorm.DB)

	/*
		homepath, err := os.UserHomeDir()
		if err != nil {

			model.EventLogAdd(db, c, "500", "PostFileEdit", "Error homepath detect "+err.Error())

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
			}, "application/json")
		}*/

	full_path := c.FormValue("full_path")

	if _, err := os.Stat(path.Join(arg_fold, full_path)); err != nil {

		model.EventLogAdd(db, c, "500", "PostFileEdit", path.Join(arg_fold, full_path))

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  path.Join(arg_fold, full_path) + " not found",
		}, "application/json")
	}

	file_ext := util.GetExtNorm(full_path)
	orig_filename := util.GetFileName(full_path)

	body := c.FormValue("body")

	save_as_source := c.FormValue("save_as_source")

	// --------------------------------------------------------------------------------------------------------------------------------

	if save_as_source == "1" {

		// --------------------------------------------------------------------------------------------------------------------------------

		filepath_tmp := path.Join(prefix, "temp")
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

		model.EventLogAdd(db, c, "200", "PostFileEdit", "Update "+source_temp_file)

		// --------------------------------------------------------------------------------------------------------------------------------

		target_file := path.Join(arg_fold, full_path)

		if runtime.GOOS == "windows" {
			target_file = util.RotateSlash(target_file)
		}

		err = util.MoveFile(source_temp_file, target_file)

		if err != nil {

			model.EventLogAdd(db, c, "500", "PostFileEdit", "Rename error "+err.Error())

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "Rename error ",
			}, "application/json")

		} else {

			go model.FileDelAsync(db, path.Join(arg_fold, full_path))

			model.EventLogAdd(db, c, "200", "PostFileEdit", "Move "+source_temp_file+" => "+target_file)

			return c.JSON(fiber.Map{
				"code": 200,
			}, "application/json")
		}

		// --------------------------------------------------------------------------------------------------------------------------------

	}

	// --------------------------------------------------------------------------------------------------------------------------------

	is_office_match, _ := regexp.MatchString("^(html|rtf|doc|docx|odt)$", file_ext)

	if is_office_match {

		// --------------------------------------------------------------------------------------------------------------------------------

		filepath_tmp := path.Join(prefix, "temp")
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

		model.EventLogAdd(db, c, "200", "PostFileEdit", "Update "+html_temp_file)

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

				model.EventLogAdd(db, c, "500", "PostFileEdit", "Error libreoffice not found "+err.Error())

				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code": 500,
					"msg":  "Libreoffice not found ",
				}, "application/json")
			}

		} else {

			_, err := exec.Command("bash", "-c", "libreoffice --help").Output()
			if err != nil {

				model.EventLogAdd(db, c, "500", "PostFileEdit", "Error libreoffice not found "+err.Error())

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

		//filepath_tmp := path.Join(prefix, "temp")
		//html_temp_file := path.Join(filepath_tmp, orig_filename+".html")

		if runtime.GOOS == "windows" {

			// --------------------------------------------------------------------------------------------------------------------------------

			//filepath_tmp := util.RotateSlash(path.Join(prefix, "temp"))
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
		target_file := path.Join(arg_fold, full_path)

		/*
			if runtime.GOOS == "windows" {
			    source_temp_file = util.RotateSlash(source_temp_file)
			    target_file = util.RotateSlash(target_file)
			}
		*/

		err = util.MoveFile(source_temp_file, target_file)

		if err != nil {

			model.EventLogAdd(db, c, "500", "PostFileEdit", "Rename error "+err.Error())

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "Rename error ",
			}, "application/json")

		} else {

			// remove temporary html
			if _, err := os.Stat(path.Join(filepath_tmp, orig_filename+".html")); err == nil {

				os.Remove(path.Join(filepath_tmp, orig_filename+".html"))
			}

			model.FileDelMd5Async(db, prefix, path.Join(arg_fold, full_path))

			model.FileDelAsync(db, path.Join(arg_fold, full_path))

			model.EventLogAdd(db, c, "200", "PostFileEdit", "Move "+source_temp_file+" => "+target_file)

			return c.JSON(fiber.Map{
				"code": 200,
			}, "application/json")
		}

		// --------------------------------------------------------------------------------------------------------------------------------

	}

	model.EventLogAdd(db, c, "500", "PostFileEdit", "Error: format of file is not for edit")

	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"code": 500,
		"msg":  "Error: format of file is not for edit",
	}, "application/json")

}
