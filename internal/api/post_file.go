package api

import (
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	_ "github.com/western/http-here/internal/model"
	"github.com/western/http-here/internal/util"

	"github.com/gofiber/fiber/v2"
)

func PostFile(c *fiber.Ctx) error {

	arg_fold := ""
	arg_fold = c.Locals("arg_fold").(string)

	referer := c.Get("Referer")

	// already decoded
	u, err := url.Parse(referer)
	if err != nil {

		util.LogPrefix(c, "500", "Error url parse "+referer+" "+err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Error url parse " + referer,
		}, "application/json")
	}

	u_path := util.CleanDirtyPath(u.Path)

	// -------------------------------------------------------------------------------------------------------------------------

	form_path := c.FormValue("path")
	form_path = util.CleanDirtyPath(form_path)
	if len(form_path) > 0 {
		u_path = form_path
	}

	// -------------------------------------------------------------------------------------------------------------------------

	name := c.FormValue("name")

	name = strings.ReplaceAll(name, "/", "")
	re := regexp.MustCompile("\\s+")
	name = re.ReplaceAllLiteralString(name, " ")

	name = util.CleanDirtyPath(name)

	if len(name) == 0 {
		util.LogPrefix(c, "500", "name is empty")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "name is empty",
		}, "application/json")
	}

	if fileInfo, err := os.Stat(filepath.Join(arg_fold, u_path, name)); err == nil {

		if fileInfo.IsDir() {
			util.LogPrefix(c, "500", "'"+filepath.Join(arg_fold, u_path, name)+"' already exists")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  filepath.Join(u_path, name) + " already exists",
			}, "application/json")
		}
	}

	myfile, err := os.Create(filepath.Join(arg_fold, u_path, name))
	if err != nil {

		util.LogPrefix(c, "500", "Error file create "+filepath.Join(arg_fold, u_path, name)+" "+err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Error file create " + filepath.Join(arg_fold, u_path, name),
		}, "application/json")
	}
	myfile.WriteString("\n")
	myfile.Close()

	util.LogPrefix(c, "200", "Create file '"+filepath.Join(arg_fold, u_path, name)+"'")

	return c.JSON(fiber.Map{
		"code": 200,
	}, "application/json")
}
