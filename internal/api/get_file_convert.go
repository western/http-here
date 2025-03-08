package api

import (
	"bufio"
	"html/template"
	"net/url"
	"os"
	"os/exec"
	"path"
	_ "path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/western/http-here/internal/model"
	"github.com/western/http-here/internal/util"

	"github.com/gofiber/fiber/v2"

	"gorm.io/gorm"
)

func GetFileConvert(c *fiber.Ctx) error {

	arg_fold := ""
	arg_fold = c.Locals("arg_fold").(string)

	db := c.Locals("db").(*gorm.DB)

	homepath, err := os.UserHomeDir()
	if err != nil {

		model.EventLogAdd(db, c, "500", "GetFileConvert", "Error homepath detect "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
		}, "application/json")
	}

	c_path, err := url.QueryUnescape(c.Path())
	if err != nil {

		model.EventLogAdd(db, c, "500", "GetFileConvert", "Error "+path.Join(arg_fold, c_path)+" "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
		}, "application/json")
	}

	c_path = util.CleanDirtyPath(c_path)

	c_path = strings.ReplaceAll(c_path, "/api/file/convert", "")

	if _, err := os.Stat(path.Join(arg_fold, c_path)); err != nil {

		model.EventLogAdd(db, c, "404", "GetFileConvert", path.Join(arg_fold, c_path))

		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code": 404,
		}, "application/json")
	}

	file_ext := util.GetExtNorm(c_path)
	orig_filename := util.GetFileName(c_path)

	is_office_match, _ := regexp.MatchString("^(rtf|doc|docx|odt)$", file_ext)

	if is_office_match {

		// --------------------------------------------------------------------------------------------------------------------------------

		if runtime.GOOS == "windows" {

			_, err := os.Stat("C:\\Program Files\\LibreOffice\\program\\soffice.exe")
			if err != nil {

				model.EventLogAdd(db, c, "500", "GetFileConvert", "Error libreoffice not found "+err.Error())

				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code": 500,
					"msg":  "Error libreoffice not found",
				}, "application/json")
			}

		} else {

			_, err := exec.Command("bash", "-c", "libreoffice --help").Output()
			if err != nil {

				model.EventLogAdd(db, c, "500", "GetFileConvert", "Error libreoffice not found "+err.Error())

				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code": 500,
					"msg":  "Error libreoffice not found",
				}, "application/json")
			}
		}

		// --------------------------------------------------------------------------------------------------------------------------------

		if runtime.GOOS == "windows" {

			filepath_tmp := util.RotateSlash(path.Join(homepath, ".httphere", "temp"))
			arg_fold_path := util.RotateSlash(path.Join(arg_fold, c_path))

			util.RunAnyCommandUnderWin(`"C:/Program Files/LibreOffice/program/soffice.exe" --headless --norestore --nologo --convert-to html --outdir "` + filepath_tmp + `" "` + arg_fold_path + `"`)

			// --------------------------------------------------------------------------------------------------------------------------------

			b, err := os.ReadFile(path.Join(filepath_tmp, orig_filename+".html"))
			if err != nil {

				model.EventLogAdd(db, c, "500", "GetFileConvert", "Error libreoffice, open file "+err.Error())
				panic(err)
			}

			model.EventLogAdd(db, c, "200", "GetFileConvert", "Convert file to html "+arg_fold_path)

			return c.JSON(fiber.Map{
				"code":      200,
				"file_data": template.HTML(string(b)),
			}, "application/json")

		} else {

			filepath_tmp := path.Join(homepath, ".httphere", "temp")
			arg_fold_path := path.Join(arg_fold, c_path)

			cmd := exec.Command("bash", "-c", "libreoffice --headless --norestore --nologo --convert-to html --outdir "+filepath_tmp+" \""+arg_fold_path+"\"")
			cmd.Dir = arg_fold

			stderr, _ := cmd.StderrPipe()
			if err := cmd.Start(); err != nil {
				panic(err)
			}

			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				//fmt.Println("libreoffice:", scanner.Text())
			}

			// --------------------------------------------------------------------------------------------------------------------------------

			b, err := os.ReadFile(path.Join(filepath_tmp, orig_filename+".html"))
			if err != nil {

				model.EventLogAdd(db, c, "500", "GetFileConvert", "Error libreoffice, open file "+err.Error())
				panic(err)
			}

			model.EventLogAdd(db, c, "200", "GetFileConvert", "Convert file to html "+arg_fold_path)

			return c.JSON(fiber.Map{
				"code":      200,
				"file_data": template.HTML(string(b)),
			}, "application/json")

		}

	}

	is_code_match, _ := regexp.MatchString("^(html|txt|js|css|md)$", file_ext)

	if is_code_match {

		filepath_tmp := path.Join(homepath, ".httphere", "temp")

		// --------------------------------------------------------------------------------------------------------------------------------

		util.CopyFile(
			path.Join(arg_fold, c_path),
			path.Join(filepath_tmp, orig_filename+"."+file_ext),
		)

		b, err := os.ReadFile(path.Join(filepath_tmp, orig_filename+"."+file_ext))
		if err != nil {

			model.EventLogAdd(db, c, "500", "GetFileConvert", "Error open temp source file: "+err.Error())

			panic(err)
		}

		// --------------------------------------------------------------------------------------------------------------------------------

		model.EventLogAdd(db, c, "200", "GetFileConvert", "Convert for edit "+path.Join(arg_fold, c_path))

		return c.JSON(fiber.Map{
			"code":      200,
			"file_data": template.HTML(string(b)),
		}, "application/json")

	}

	model.EventLogAdd(db, c, "500", "GetFileConvert", "Error: format of file is not for edit")

	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"code": 500,
		"msg":  "Error: format of file is not for edit",
	}, "application/json")

}
