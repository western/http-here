package api

import (
	"net/url"
	"os"
	"path"

	"github.com/western/http-here/v2/internal/conf"
	"github.com/western/http-here/v2/internal/model2"
	"github.com/western/http-here/v2/internal/util"

	"github.com/gofiber/fiber/v2"
)

func PostRename(c *fiber.Ctx) error {

	referer := c.Get("Referer")

	// already decoded
	u, err := url.Parse(referer)
	if err != nil {

		model2.EventLogAdd(c, 500, "PostRename", "Error url parse "+referer+" "+err.Error())

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

		model2.EventLogAdd(c, 500, "PostRename", "to is empty")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "to is empty",
		}, "application/json")
	}

	_, err = os.Stat(path.Join(conf.ArgFold, to))
	if err == nil {

		model2.EventLogAdd(c, 500, "PostRename", "'"+path.Join(conf.ArgFold, to)+"' already exists ")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "'" + path.Join(conf.ArgFold, to) + "' already exists",
		}, "application/json")
	}

	name := c.FormValue("name")
	name = util.CleanDirtyPath(name)

	if len(name) == 0 {

		model2.EventLogAdd(c, 500, "PostRename", "name is empty")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "name is empty",
		}, "application/json")
	}

	_, err = os.Stat(path.Join(conf.ArgFold, u_path, name))
	if err != nil {

		model2.EventLogAdd(c, 500, "PostRename", "'"+path.Join(conf.ArgFold, u_path, name)+"' not exists "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "'" + path.Join(conf.ArgFold, u_path, name) + "' not exists",
		}, "application/json")
	}

	src_file_path := path.Join(conf.ArgFold, u_path, name)
	target_file_path := path.Join(conf.ArgFold, u_path, to)

	err = os.Rename(src_file_path, target_file_path)

	if err != nil {

		model2.EventLogAdd(c, 500, "PostRename", "Rename error "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Rename error",
		}, "application/json")

	} else {

		model2.EventLogAdd(c, 200, "PostRename", "Rename '"+src_file_path+"' => "+target_file_path)
	}

	go model2.FileChkAsync()
	RemoveCacheDir(path.Join(conf.ArgFold, u_path))

	return c.JSON(fiber.Map{
		"code": 200,
	}, "application/json")

}
