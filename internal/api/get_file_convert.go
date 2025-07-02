package api

import (
	"bufio"
	"html/template"
	"net/url"
	"os"
	"os/exec"
	"path"
	"regexp"
	"runtime"
	"strings"

	"github.com/western/http-here/internal/model"
	"github.com/western/http-here/internal/util"

	"github.com/gofiber/fiber/v2"

	"gorm.io/gorm"
)

var (
	isOfficeRegex = regexp.MustCompile("^(rtf|doc|docx|odt)$")

	isSimpleSourceRegex = regexp.MustCompile("^(html|txt|js|css|md)$")
)

func GetFileConvert(c *fiber.Ctx) error {

	arg_fold := GetStringFromLocals(c, "arg_fold", "")

	db := c.Locals("db").(*gorm.DB)

	c_path, err := url.QueryUnescape(c.Path())
	if err != nil {

		model.EventLogAdd(db, c, "500", "GetFileConvert", "Error "+path.Join(arg_fold, c_path)+" "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
		}, "application/json")
	}

	c_path = util.CleanDirtyPath(c_path)

	c_path = strings.ReplaceAll(c_path, "/api/file/convert", "")

	readTarget := path.Join(arg_fold, c_path)

	if _, err := os.Stat(readTarget); err != nil {

		model.EventLogAdd(db, c, "404", "GetFileConvert", readTarget)

		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code": 404,
		}, "application/json")
	}

	file_ext := util.GetExtNorm(c_path)

	// --------------------------------------------------------------------------------------------------------------------------------

	is_office_match := isOfficeRegex.MatchString(file_ext)

	if is_office_match {

		return convertOfficeToHTML(c, c_path)
	}

	is_simple_source_match := isSimpleSourceRegex.MatchString(file_ext)

	if is_simple_source_match {

		return slurpFile(c, c_path)
	}

	// --------------------------------------------------------------------------------------------------------------------------------

	model.EventLogAdd(db, c, "500", "GetFileConvert", "Error: format of file is not for edit")

	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"code": 500,
		"msg":  "Error: format of file is not for edit",
	}, "application/json")
}

func convertOfficeToHTML(c *fiber.Ctx, c_path string) error {

	arg_fold := GetStringFromLocals(c, "arg_fold", "")
	prefix := GetStringFromLocals(c, "prefix", "")

	db := c.Locals("db").(*gorm.DB)

	orig_filename := util.GetFileName(c_path)

	// --------------------------------------------------------------------------------------------------------------------------------

	if runtime.GOOS == "windows" {

		_, err := os.Stat("C:\\Program Files\\LibreOffice\\program\\soffice.exe")
		if err != nil {

			model.EventLogAdd(db, c, "500", "convertOfficeToHTML", "Error libreoffice not found "+err.Error())

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "Error libreoffice not found",
			}, "application/json")
		}

	} else {

		_, err := exec.Command("bash", "-c", "libreoffice --help").Output()
		if err != nil {

			model.EventLogAdd(db, c, "500", "convertOfficeToHTML", "Error libreoffice not found "+err.Error())

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "Error libreoffice not found",
			}, "application/json")
		}
	}

	// --------------------------------------------------------------------------------------------------------------------------------

	if runtime.GOOS == "windows" {

		filepath_tmp := util.RotateSlash(path.Join(prefix, "temp"))
		arg_fold_path := util.RotateSlash(path.Join(arg_fold, c_path))

		util.RunAnyCommandUnderWin(`"C:/Program Files/LibreOffice/program/soffice.exe" --headless --norestore --nologo --convert-to html --outdir "` + filepath_tmp + `" "` + arg_fold_path + `"`)

		// --------------------------------------------------------------------------------------------------------------------------------

		b, err := os.ReadFile(path.Join(filepath_tmp, orig_filename+".html"))
		if err != nil {

			model.EventLogAdd(db, c, "500", "convertOfficeToHTML", "Error libreoffice, open file "+err.Error())
			panic(err)
		}

		model.EventLogAdd(db, c, "200", "convertOfficeToHTML", "Convert file to html "+arg_fold_path)

		return c.JSON(fiber.Map{
			"code":      200,
			"file_data": template.HTML(string(b)),
		}, "application/json")

	} else {

		filepath_tmp := path.Join(prefix, "temp")
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

			model.EventLogAdd(db, c, "500", "convertOfficeToHTML", "Error libreoffice, open file "+err.Error())
			panic(err)
		}

		model.EventLogAdd(db, c, "200", "convertOfficeToHTML", "Convert file to html "+arg_fold_path)

		return c.JSON(fiber.Map{
			"code":      200,
			"file_data": template.HTML(string(b)),
		}, "application/json")

	}

}

func slurpFile(c *fiber.Ctx, c_path string) error {

	arg_fold := GetStringFromLocals(c, "arg_fold", "")
	prefix := GetStringFromLocals(c, "prefix", "")

	filepath_tmp := path.Join(prefix, "temp")

	db := c.Locals("db").(*gorm.DB)

	file_ext := util.GetExtNorm(c_path)
	orig_filename := util.GetFileName(c_path)

	// --------------------------------------------------------------------------------------------------------------------------------

	util.CopyFile(
		path.Join(arg_fold, c_path),
		path.Join(filepath_tmp, orig_filename+"."+file_ext),
	)

	b, err := os.ReadFile(path.Join(filepath_tmp, orig_filename+"."+file_ext))
	if err != nil {

		model.EventLogAdd(db, c, "500", "slurpFile", "Error open temp source file: "+err.Error())

		panic(err)
	}

	// --------------------------------------------------------------------------------------------------------------------------------

	model.EventLogAdd(db, c, "200", "slurpFile", "Convert for edit "+path.Join(arg_fold, c_path))

	return c.JSON(fiber.Map{
		"code":      200,
		"file_data": template.HTML(string(b)),
	}, "application/json")
}
