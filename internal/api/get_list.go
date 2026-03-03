package api

import (
	"os"
	"path"

	"github.com/western/http-here/v2/internal/conf"
	"github.com/western/http-here/v2/internal/model2"
	"github.com/western/http-here/v2/internal/util"

	"github.com/gofiber/fiber/v2"
)

func GetList(c *fiber.Ctx) error {

	// -------------------------------------------------------------------------------------------------------------------------

	u_path := c.FormValue("path")

	u_path = util.CleanDirtyPath(u_path)

	readTarget := path.Join(conf.ArgFold, u_path)

	// -------------------------------------------------------------------------------------------------------------------------

	if _, err := os.Stat(readTarget); err != nil {

		model2.EventLogAdd(c, 500, "GetList", "'"+readTarget+"' not exists")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  path.Join(u_path) + " not exists",
		}, "application/json")
	}

	var s_sort string = "name"

	s_sort = c.Cookies("sort")
	if len(s_sort) == 0 {
		s_sort = "name"
	}

	rows, err := generateRows(c, s_sort)
	if err != nil {

		if os.IsPermission(err) {

			model2.EventLogAdd(c, 403, "GetList", "Forbidden for read "+readTarget)

			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"code": 403,
				"msg":  "Forbidden",
			}, "application/json")
		}

		model2.EventLogAdd(c, 500, "GetList", "generateRows err: "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "error",
		}, "application/json")
	}

	model2.EventLogAdd(c, 200, "GetList", "Get list '"+readTarget+"'")

	return c.JSON(fiber.Map{
		"code": 200,
		"rows": rows,
	}, "application/json")

}
