package api

import (
	"net/url"
	"os"
	"path/filepath"

	"github.com/western/http-here/internal/model"
	"github.com/western/http-here/internal/util"

	"github.com/gofiber/fiber/v2"
)

func PostRename(c *fiber.Ctx) error {

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

	to := c.FormValue("to")
	to = util.CleanDirtyPath(to)

	if len(to) == 0 {
		util.LogPrefix(c, "500", "to is empty")
		return c.JSON(fiber.Map{
			"code": 500,
			"msg":  "to is empty",
		}, "application/json")
	}

	_, err = os.Stat(filepath.Join(arg_fold, to))
	if err == nil {
		util.LogPrefix(c, "500", "'"+filepath.Join(arg_fold, to)+"' already exists ")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "'" + filepath.Join(arg_fold, to) + "' already exists",
		}, "application/json")
	}

	name := c.FormValue("name")
	name = util.CleanDirtyPath(name)

	if len(name) == 0 {
		util.LogPrefix(c, "500", "name is empty")
		return c.JSON(fiber.Map{
			"code": 500,
			"msg":  "name is empty",
		}, "application/json")
	}

	_, err = os.Stat(filepath.Join(arg_fold, u_path, name))
	if err != nil {
		util.LogPrefix(c, "500", "'"+filepath.Join(arg_fold, u_path, name)+"' not exists "+err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "'" + filepath.Join(arg_fold, u_path, name) + "' not exists",
		}, "application/json")
	}

	src_file_path := filepath.Join(arg_fold, u_path, name)
	target_file_path := filepath.Join(arg_fold, u_path, to)

	err = os.Rename(src_file_path, target_file_path)

	if err != nil {

		util.LogPrefix(c, "500", "Rename error "+err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Rename error",
		}, "application/json")

	} else {

		util.LogPrefix(c, "200", "Rename '"+src_file_path+"' => "+target_file_path)
	}

	db, err := model.ConnectToSQLite()
	if err != nil {
		panic(err)
	}

	go model.FileChkAsync(db)

	return c.JSON(fiber.Map{
		"code": 200,
	}, "application/json")

}
