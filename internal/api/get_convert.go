package api

import (
	"bufio"
	"html/template"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/western/http-here/internal/model"
	"github.com/western/http-here/internal/util"

	"github.com/gofiber/fiber/v2"
)

func GetConvert(c *fiber.Ctx) error {

	arg_fold := ""
	arg_fold = c.Locals("arg_fold").(string)

	db, err := model.ConnectToSQLite()
	if err != nil {
		panic(err)
	}

	homepath, err := os.UserHomeDir()
	if err != nil {

		//util.LogPrefix(c, "500", "Error homepath detect "+err.Error())
		model.EventLogAdd(db, c, "500", "GetConvert", "Error homepath detect "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
		}, "application/json")
	}

	c_path, err := url.QueryUnescape(c.Path())
	if err != nil {

		//util.LogPrefix(c, "500", "Error "+filepath.Join(arg_fold, c_path)+" "+err.Error())
		model.EventLogAdd(db, c, "500", "GetConvert", "Error "+filepath.Join(arg_fold, c_path)+" "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
		}, "application/json")
	}

	c_path = util.CleanDirtyPath(c_path)

	c_path = strings.ReplaceAll(c_path, "/__convert", "")

	if _, err := os.Stat(filepath.Join(arg_fold, c_path)); err != nil {

		//util.LogPrefix(c, "404", filepath.Join(arg_fold, c_path))
		model.EventLogAdd(db, c, "404", "GetConvert", filepath.Join(arg_fold, c_path))

		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code": 404,
		}, "application/json")
	}

	file_ext := util.GetExtNorm(c_path)
	orig_filename := util.GetFileName(c_path)

	is_office_match, _ := regexp.MatchString("^(rtf|doc|docx|odt)$", file_ext)

	if is_office_match {

		// --------------------------------------------------------------------------------------------------------------------------------

		_, err := exec.Command("bash", "-c", "libreoffice --help").Output()
		if err != nil {

			//util.LogPrefix(c, "500", "Error libreoffice not found "+err.Error())
			model.EventLogAdd(db, c, "500", "GetConvert", "Error libreoffice not found "+err.Error())

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
			}, "application/json")
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

			//util.LogPrefix(c, "500", "Error libreoffice, open file "+err.Error())
			model.EventLogAdd(db, c, "500", "GetConvert", "Error libreoffice, open file "+err.Error())

			panic(err)
		}

		// --------------------------------------------------------------------------------------------------------------------------------

		//util.LogPrefix(c, "200", "Convert file to html "+filepath.Join(arg_fold, c_path))
		model.EventLogAdd(db, c, "200", "GetConvert", "Convert file to html "+filepath.Join(arg_fold, c_path))

		return c.JSON(fiber.Map{
			"code":      200,
			"file_data": template.HTML(string(b)),
		}, "application/json")

	}

	is_code_match, _ := regexp.MatchString("^(html|txt|js|css|md)$", file_ext)

	if is_code_match {

		filepath_tmp := filepath.Join(homepath, ".httphere", "temp")

		// --------------------------------------------------------------------------------------------------------------------------------

		util.CopyFile(
			filepath.Join(arg_fold, c_path),
			filepath.Join(filepath_tmp, orig_filename+"."+file_ext),
		)

		b, err := os.ReadFile(filepath.Join(filepath_tmp, orig_filename+"."+file_ext))
		if err != nil {

			//util.LogPrefix(c, "500", "Error open temp source file: "+err.Error())
			model.EventLogAdd(db, c, "500", "GetConvert", "Error open temp source file: "+err.Error())

			panic(err)
		}

		// --------------------------------------------------------------------------------------------------------------------------------

		//util.LogPrefix(c, "200", "Convert for edit "+filepath.Join(arg_fold, c_path))
		model.EventLogAdd(db, c, "200", "GetConvert", "Convert for edit "+filepath.Join(arg_fold, c_path))

		return c.JSON(fiber.Map{
			"code":      200,
			"file_data": template.HTML(string(b)),
		}, "application/json")

	}

	//util.LogPrefix(c, "500", "Error: format of file is not for edit")
	model.EventLogAdd(db, c, "500", "GetConvert", "Error: format of file is not for edit")

	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"code": 500,
		"msg":  "Error: format of file is not for edit",
	}, "application/json")

}
