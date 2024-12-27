package controller

import (
	"bufio"
	_ "errors"
	_ "fmt"
	"html/template"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	_ "strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func GetEdit(c *fiber.Ctx) error {

	arg_fold := ""
	arg_fold = c.Locals("arg_fold").(string)

	homepath, err := os.UserHomeDir()
	if err != nil {

		LogPrefix(c, "500", "Error hmepath detect "+err.Error())
		return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout")
	}

	c_path, err := url.QueryUnescape(c.Path())
	if err != nil {

		LogPrefix(c, "500", "Error "+filepath.Join(arg_fold, c_path)+" "+err.Error())
		return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout")
	}

	c_path = CleanDirtyPath(c_path)

	c_path = strings.ReplaceAll(c_path, "/__edit", "")

	if _, err := os.Stat(filepath.Join(arg_fold, c_path)); err != nil {

		LogPrefix(c, "404", filepath.Join(arg_fold, c_path))
		return c.Status(fiber.StatusNotFound).Render("view/404", fiber.Map{}, "view/layout")
	}

	file_ext := GetExtNorm(c_path)
	orig_filename := GetFileName(c_path)

	is_office_match, _ := regexp.MatchString("^(html|rtf|doc|docx|odt)$", file_ext)

	if is_office_match {

		// --------------------------------------------------------------------------------------------------------------------------------

		_, err := exec.Command("bash", "-c", "libreoffice --help").Output()
		if err != nil {

			LogPrefix(c, "500", "Error libreoffice not found "+err.Error())
			return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout")
		}

		// --------------------------------------------------------------------------------------------------------------------------------

		filepath_tmp := filepath.Join(homepath, ".httphere", "temp")

		cmd := exec.Command("bash", "-c", "libreoffice --headless --norestore --nologo --convert-to html --outdir "+filepath_tmp+" \""+filepath.Join(arg_fold, c_path)+"\"")
		cmd.Dir = arg_fold

		stderr, _ := cmd.StderrPipe()
		if err := cmd.Start(); err != nil {
			panic(err)
		}

		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			//fmt.Println("libreoffice:", scanner.Text())
		}

		b, err := os.ReadFile(filepath.Join(filepath_tmp, orig_filename+".html"))
		if err != nil {

			LogPrefix(c, "500", "Error libreoffice, open file "+err.Error())
			panic(err)
		}

		// --------------------------------------------------------------------------------------------------------------------------------

		LogPrefix(c, "200", "Open for edit "+filepath.Join(arg_fold, c_path))

		return c.Render("view/edit", fiber.Map{

			"full_path": c_path,

			//"file_data": string(b),
			"file_data": template.HTML(string(b)),
		}, "view/layout")

	}

	LogPrefix(c, "500", "Error: format of file are not for edit")

	return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout")

}

func PostEdit(c *fiber.Ctx) error {

	arg_fold := ""
	arg_fold = c.Locals("arg_fold").(string)

	homepath, err := os.UserHomeDir()
	if err != nil {

		LogPrefix(c, "500", "Error hmepath detect "+err.Error())
		return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout")
	}

	full_path := c.FormValue("full_path")

	if _, err := os.Stat(filepath.Join(arg_fold, full_path)); err != nil {

		LogPrefix(c, "500", filepath.Join(arg_fold, full_path))

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  filepath.Join(arg_fold, full_path) + " not found",
		}, "application/json")
	}

	file_ext := GetExtNorm(full_path)
	orig_filename := GetFileName(full_path)

	body := c.FormValue("body")

	filepath_tmp := filepath.Join(homepath, ".httphere", "temp")
	html_temp_file := filepath.Join(filepath_tmp, orig_filename+".html")

	f, err := os.Create(html_temp_file)
	if err != nil {
		panic(err)
	}

	_, err = f.WriteString(body)
	if err != nil {
		panic(err)
	}
	f.Close()

	LogPrefix(c, "200", "Update "+html_temp_file)

	is_office_match, _ := regexp.MatchString("^(html|rtf|doc|docx|odt)$", file_ext)

	if is_office_match {

		// --------------------------------------------------------------------------------------------------------------------------------

		_, err := exec.Command("bash", "-c", "libreoffice --help").Output()
		if err != nil {

			LogPrefix(c, "500", "Error libreoffice not found "+err.Error())

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "Libreoffice not found ",
			}, "application/json")
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

		//filepath_tmp := filepath.Join(homepath, ".httphere", "temp")
		//html_temp_file := filepath.Join(filepath_tmp, orig_filename+".html")

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

		// --------------------------------------------------------------------------------------------------------------------------------

		target_temp_file := filepath.Join(filepath_tmp, orig_filename+"."+file_ext)
		target_file := filepath.Join(arg_fold, full_path)

		err = os.Rename(target_temp_file, target_file)

		if err != nil {

			LogPrefix(c, "500", "Rename error "+err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "Rename error ",
			}, "application/json")

		} else {

			LogPrefix(c, "200", "Move "+target_temp_file+" to "+target_file)
			return c.JSON(fiber.Map{
				"code": 200,
			}, "application/json")
		}

		// --------------------------------------------------------------------------------------------------------------------------------

	}

	LogPrefix(c, "500", "Error: format of file are not for edit")

	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"code": 500,
		"msg":  "Error: format of file are not for edit",
	}, "application/json")

}
