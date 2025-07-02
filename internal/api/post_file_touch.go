package api

import (
	"net/url"
	"os"
	"path"
	_ "path/filepath"
	"regexp"
	"strings"

	"github.com/western/http-here/internal/model"
	"github.com/western/http-here/internal/util"

	"github.com/gofiber/fiber/v2"

	"gorm.io/gorm"
)

func PostFileTouch(c *fiber.Ctx) error {

	arg_fold := GetStringFromLocals(c, "arg_fold", "")

	referer := c.Get("Referer")

	db := c.Locals("db").(*gorm.DB)

	// already decoded
	u, err := url.Parse(referer)
	if err != nil {

		model.EventLogAdd(db, c, "500", "PostFileTouch", "Error url parse "+referer+" "+err.Error())

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

		model.EventLogAdd(db, c, "500", "PostFileTouch", "name is empty")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "name is empty",
		}, "application/json")
	}

	if fileInfo, err := os.Stat(path.Join(arg_fold, u_path, name)); err == nil {

		if fileInfo.IsDir() {
			model.EventLogAdd(db, c, "500", "PostFileTouch", "'"+path.Join(arg_fold, u_path, name)+"' already exists")

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  path.Join(u_path, name) + " already exists",
			}, "application/json")
		}
	}

	myfile, err := os.Create(path.Join(arg_fold, u_path, name))
	if err != nil {

		model.EventLogAdd(db, c, "500", "PostFileTouch", "Error file create "+path.Join(arg_fold, u_path, name)+" "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Error file create " + path.Join(arg_fold, u_path, name),
		}, "application/json")
	}
	myfile.WriteString("\n")
	myfile.Close()

	model.EventLogAdd(db, c, "200", "PostFileTouch", "Create file '"+path.Join(arg_fold, u_path, name)+"'")
	RemoveCacheDir(path.Join(arg_fold, u_path))

	return c.JSON(fiber.Map{
		"code": 200,
	}, "application/json")
}
