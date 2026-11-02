package api

import (
	"bufio"
	"html/template"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/western/http-here/v2/internal/conf"
	"github.com/western/http-here/v2/internal/model2"
	//"github.com/western/http-here/v2/internal/safefs"
	"github.com/western/http-here/v2/internal/util"

	"github.com/gofiber/fiber/v3"
)

var (
	isOfficeRegex = regexp.MustCompile("^(rtf|doc|docx|odt)$")

	isSimpleSourceRegex = regexp.MustCompile("^(html|txt|js|css|md|sh|json)$")
)

func GetFileConvert(c fiber.Ctx) error {

	c_Path, err := url.QueryUnescape(c.Path())
	if err != nil {

		model2.EventLogAdd(c, 500, "GetFileConvert", "Error QueryUnescape "+c_Path+" "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
		})
	}

	c_Path = util.CleanDirtyPath(c_Path)

	c_Path = strings.ReplaceAll(c_Path, "/api/file/convert", "")

	targetPath := filepath.Join(conf.ArgFold, c_Path)

	file_ext := util.GetExtNorm(c_Path)

	// --------------------------------------------------------------------------------------------------------------------------------

	Root := conf.ArgFoldRootFS

	fileInfo, err := Root.Stat(strings.TrimPrefix(c_Path, "/"))
	if err != nil {

		model2.EventLogAdd(c, 500, "GetFileConvert", "Stat Root err: "+err.Error())

		// RETURN 404 for security
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code": 404,
		})
	}

	fileMode := fileInfo.Mode()

	if !fileMode.IsRegular() {

		model2.EventLogAdd(c, 500, "GetFileConvert", "Error "+targetPath+" is NOT regular file")

		// RETURN 404 for security
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code": 404,
		})
	}

	// --------------------------------------------------------------------------------------------------------------------------------

	is_office_match := isOfficeRegex.MatchString(file_ext)
	if is_office_match {

		return convertOfficeToHTML(c, c_Path)
	}

	is_simple_source_match := isSimpleSourceRegex.MatchString(file_ext)
	if is_simple_source_match {

		return slurpFile(c, c_Path)
	}

	// --------------------------------------------------------------------------------------------------------------------------------

	model2.EventLogAdd(c, 500, "GetFileConvert", "Error: format of file is not for edit")

	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"code": 500,
		"msg":  "Error: format of file is not for edit",
	})
}

func convertOfficeToHTML(c fiber.Ctx, c_path string) error {

	orig_filename := util.GetFileName(c_path)

	// --------------------------------------------------------------------------------------------------------------------------------

	if runtime.GOOS == "windows" {

		_, err := os.Stat("C:\\Program Files\\LibreOffice\\program\\soffice.exe")
		if err != nil {

			model2.EventLogAdd(c, 500, "convertOfficeToHTML", "Error libreoffice not found "+err.Error())

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "Error libreoffice not found",
			})
		}

	} else {

		_, err := exec.Command("bash", "-c", "libreoffice --help").Output()
		if err != nil {

			model2.EventLogAdd(c, 500, "convertOfficeToHTML", "Error libreoffice not found "+err.Error())

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "Error libreoffice not found",
			})
		}
	}

	// --------------------------------------------------------------------------------------------------------------------------------

	if runtime.GOOS == "windows" {

		filepath_tmp := util.RotateSlash(filepath.Join(conf.ConfigRoot, "temp"))
		arg_fold_path := util.RotateSlash(filepath.Join(conf.ArgFold, c_path))

		util.RunAnyCommandUnderWin(`"C:/Program Files/LibreOffice/program/soffice.exe" --headless --norestore --nologo --convert-to html --outdir "` + filepath_tmp + `" "` + arg_fold_path + `"`)

		// --------------------------------------------------------------------------------------------------------------------------------

		b, err := os.ReadFile(filepath.Join(filepath_tmp, orig_filename+".html"))
		if err != nil {

			model2.EventLogAdd(c, 500, "convertOfficeToHTML", "Error libreoffice, open file "+err.Error())
			panic(err)
		}

		model2.EventLogAdd(c, 200, "convertOfficeToHTML", "Convert file to html "+arg_fold_path)

		return c.JSON(fiber.Map{
			"code":      200,
			"file_data": template.HTML(string(b)),
		}, "application/json")

	} else {

		filepath_tmp := filepath.Join(conf.ConfigRoot, "temp")
		arg_fold_path := filepath.Join(conf.ArgFold, c_path)

		cmd := exec.Command("bash", "-c", "libreoffice --headless --norestore --nologo --convert-to html --outdir "+filepath_tmp+" \""+arg_fold_path+"\"")
		cmd.Dir = conf.ArgFold

		stderr, _ := cmd.StderrPipe()
		if err := cmd.Start(); err != nil {
			panic(err)
		}

		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			//fmt.Println("libreoffice:", scanner.Text())
		}

		// --------------------------------------------------------------------------------------------------------------------------------

		b, err := os.ReadFile(filepath.Join(filepath_tmp, orig_filename+".html"))
		if err != nil {

			model2.EventLogAdd(c, 500, "convertOfficeToHTML", "Error libreoffice, open file "+err.Error())
			panic(err)
		}

		model2.EventLogAdd(c, 200, "convertOfficeToHTML", "Convert file to html "+arg_fold_path)

		return c.JSON(fiber.Map{
			"code":      200,
			"file_data": template.HTML(string(b)),
		}, "application/json")

	}

}

func slurpFile(c fiber.Ctx, c_path string) error {

	filepath_tmp := filepath.Join(conf.ConfigRoot, "temp")

	file_ext := util.GetExtNorm(c_path)
	orig_filename := util.GetFileName(c_path)

	// --------------------------------------------------------------------------------------------------------------------------------

	util.CopyFile(
		path.Join(conf.ArgFold, c_path),
		path.Join(filepath_tmp, orig_filename+"."+file_ext),
	)

	b, err := os.ReadFile(filepath.Join(filepath_tmp, orig_filename+"."+file_ext))
	if err != nil {

		model2.EventLogAdd(c, 500, "slurpFile", "Error open temp source file: "+err.Error())

		panic(err)
	}

	// --------------------------------------------------------------------------------------------------------------------------------

	model2.EventLogAdd(c, 200, "slurpFile", "Convert for edit "+filepath.Join(conf.ArgFold, c_path))

	return c.JSON(fiber.Map{
		"code":      200,
		"file_data": template.HTML(string(b)),
	}, "application/json")
}
