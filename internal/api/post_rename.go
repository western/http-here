package api

import (
	"net/url"
	"os"
	"path"
	_ "path/filepath"

	"github.com/western/http-here/internal/model"
	"github.com/western/http-here/internal/util"

	"github.com/gofiber/fiber/v2"

	"gorm.io/gorm"
)

func PostRename(c *fiber.Ctx) error {

	arg_fold := ""
	arg_fold = c.Locals("arg_fold").(string)

	prefix := c.Locals("prefix").(string)

	referer := c.Get("Referer")

	db := c.Locals("db").(*gorm.DB)

	// already decoded
	u, err := url.Parse(referer)
	if err != nil {

		model.EventLogAdd(db, c, "500", "PostRename", "Error url parse "+referer+" "+err.Error())

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

	to := c.FormValue("to")
	to = util.CleanDirtyPath(to)

	if len(to) == 0 {

		model.EventLogAdd(db, c, "500", "PostRename", "to is empty")

		return c.JSON(fiber.Map{
			"code": 500,
			"msg":  "to is empty",
		}, "application/json")
	}

	_, err = os.Stat(path.Join(arg_fold, to))
	if err == nil {

		model.EventLogAdd(db, c, "500", "PostRename", "'"+path.Join(arg_fold, to)+"' already exists ")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "'" + path.Join(arg_fold, to) + "' already exists",
		}, "application/json")
	}

	name := c.FormValue("name")
	name = util.CleanDirtyPath(name)

	if len(name) == 0 {

		model.EventLogAdd(db, c, "500", "PostRename", "name is empty")

		return c.JSON(fiber.Map{
			"code": 500,
			"msg":  "name is empty",
		}, "application/json")
	}

	_, err = os.Stat(path.Join(arg_fold, u_path, name))
	if err != nil {

		model.EventLogAdd(db, c, "500", "PostRename", "'"+path.Join(arg_fold, u_path, name)+"' not exists "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "'" + path.Join(arg_fold, u_path, name) + "' not exists",
		}, "application/json")
	}

	src_file_path := path.Join(arg_fold, u_path, name)
	target_file_path := path.Join(arg_fold, u_path, to)

	err = os.Rename(src_file_path, target_file_path)

	if err != nil {

		model.EventLogAdd(db, c, "500", "PostRename", "Rename error "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Rename error",
		}, "application/json")

	} else {

		model.EventLogAdd(db, c, "200", "PostRename", "Rename '"+src_file_path+"' => "+target_file_path)
	}

	go model.FileChkAsync(db, prefix)

	return c.JSON(fiber.Map{
		"code": 200,
	}, "application/json")

}
