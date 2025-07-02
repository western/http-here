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

func PostFolder(c *fiber.Ctx) error {

	arg_fold := GetStringFromLocals(c, "arg_fold", "")

	referer := c.Get("Referer")

	db := c.Locals("db").(*gorm.DB)

	// already decoded
	u, err := url.Parse(referer)
	if err != nil {

		model.EventLogAdd(db, c, "500", "PostFolder", "Error url parse "+referer+" "+err.Error())

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

	readTarget := path.Join(arg_fold, u_path)

	// -------------------------------------------------------------------------------------------------------------------------

	name := c.FormValue("name")

	name = strings.ReplaceAll(name, "/", "")
	re := regexp.MustCompile("\\s+")
	name = re.ReplaceAllLiteralString(name, " ")

	name = util.CleanDirtyPath(name)

	if len(name) == 0 {

		model.EventLogAdd(db, c, "500", "PostFolder", "name is empty")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "name is empty",
		}, "application/json")
	}

	if fileInfo, err := os.Stat(path.Join(readTarget, name)); err == nil {

		if fileInfo.IsDir() {

			model.EventLogAdd(db, c, "500", "PostFolder", "'"+path.Join(readTarget, name)+"' already exists")

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  path.Join(u_path, name) + " already exists",
			}, "application/json")
		}
	}

	if err := os.Mkdir(path.Join(readTarget, name), os.ModePerm); err != nil {

		model.EventLogAdd(db, c, "500", "PostFolder", "Error mkdir "+path.Join(readTarget, name)+" "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Error mkdir " + path.Join(readTarget, name),
		}, "application/json")
	}

	model.EventLogAdd(db, c, "200", "PostFolder", "Mkdir '"+path.Join(readTarget, name)+"'")

	RemoveCacheDir(readTarget)

	return c.JSON(fiber.Map{
		"code": 200,
	}, "application/json")
}
