package api

import (
	"net/url"
	"os"
	"path"
	_ "path/filepath"
	"regexp"
	"strings"

	"github.com/western/http-here/v2/internal/conf"
	"github.com/western/http-here/v2/internal/model2"
	"github.com/western/http-here/v2/internal/util"

	"github.com/gofiber/fiber/v2"
)

func PostFileTouch(c *fiber.Ctx) error {

	referer := c.Get("Referer")

	// already decoded
	u, err := url.Parse(referer)
	if err != nil {

		model2.EventLogAdd(c, 500, "PostFileTouch", "Error url parse "+referer+" "+err.Error())

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

		model2.EventLogAdd(c, 500, "PostFileTouch", "name is empty")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "name is empty",
		}, "application/json")
	}

	if fileInfo, err := os.Stat(path.Join(conf.ArgFold, u_path, name)); err == nil {

		if fileInfo.IsDir() {
			model2.EventLogAdd(c, 500, "PostFileTouch", "'"+path.Join(conf.ArgFold, u_path, name)+"' already exists")

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  path.Join(u_path, name) + " already exists",
			}, "application/json")
		}
	}

	myfile, err := os.Create(path.Join(conf.ArgFold, u_path, name))
	if err != nil {

		model2.EventLogAdd(c, 500, "PostFileTouch", "Error file create "+path.Join(conf.ArgFold, u_path, name)+" "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Error file create " + path.Join(conf.ArgFold, u_path, name),
		}, "application/json")
	}
	myfile.WriteString("\n")
	myfile.Close()

	model2.EventLogAdd(c, 200, "PostFileTouch", "Create file '"+path.Join(conf.ArgFold, u_path, name)+"'")
	RemoveCacheDir(path.Join(conf.ArgFold, u_path))

	return c.JSON(fiber.Map{
		"code": 200,
	}, "application/json")
}
