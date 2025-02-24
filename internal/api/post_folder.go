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
)

func PostFolder(c *fiber.Ctx) error {

	arg_fold := ""
	arg_fold = c.Locals("arg_fold").(string)

	referer := c.Get("Referer")

	db, err := model.ConnectToSQLite()
	if err != nil {
		panic(err)
	}

	// already decoded
	u, err := url.Parse(referer)
	if err != nil {

		//util.LogPrefix(c, "500", "Error url parse "+referer+" "+err.Error())
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

	// -------------------------------------------------------------------------------------------------------------------------

	name := c.FormValue("name")

	name = strings.ReplaceAll(name, "/", "")
	re := regexp.MustCompile("\\s+")
	name = re.ReplaceAllLiteralString(name, " ")

	name = util.CleanDirtyPath(name)

	if len(name) == 0 {
		//util.LogPrefix(c, "500", "name is empty")
		model.EventLogAdd(db, c, "500", "PostFolder", "name is empty")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "name is empty",
		}, "application/json")
	}

	if fileInfo, err := os.Stat(path.Join(arg_fold, u_path, name)); err == nil {

		if fileInfo.IsDir() {
			//util.LogPrefix(c, "500", "'"+path.Join(arg_fold, u_path, name)+"' already exists")
			model.EventLogAdd(db, c, "500", "PostFolder", "'"+path.Join(arg_fold, u_path, name)+"' already exists")

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  path.Join(u_path, name) + " already exists",
			}, "application/json")
		}
	}

	if err := os.Mkdir(path.Join(arg_fold, u_path, name), os.ModePerm); err != nil {

		//util.LogPrefix(c, "500", "Error mkdir "+path.Join(arg_fold, u_path, name)+" "+err.Error())
		model.EventLogAdd(db, c, "500", "PostFolder", "Error mkdir "+path.Join(arg_fold, u_path, name)+" "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Error mkdir " + path.Join(arg_fold, u_path, name),
		}, "application/json")
	}

	//util.LogPrefix(c, "200", "Mkdir '"+path.Join(arg_fold, u_path, name)+"'")
	model.EventLogAdd(db, c, "200", "PostFolder", "Mkdir '"+path.Join(arg_fold, u_path, name)+"'")

	return c.JSON(fiber.Map{
		"code": 200,
	}, "application/json")
}
