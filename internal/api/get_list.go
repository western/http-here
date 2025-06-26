package api

import (
	"os"
	"path"
	_ "path/filepath"

	"github.com/western/http-here/internal/model"
	"github.com/western/http-here/internal/util"

	"github.com/gofiber/fiber/v2"

	"gorm.io/gorm"
)

func GetList(c *fiber.Ctx) error {

	arg_fold := ""
	arg_fold = c.Locals("arg_fold").(string)

	db := c.Locals("db").(*gorm.DB)

	// -------------------------------------------------------------------------------------------------------------------------

	u_path := c.FormValue("path")

	u_path = util.CleanDirtyPath(u_path)

	readFolder := path.Join(arg_fold, u_path)

	// -------------------------------------------------------------------------------------------------------------------------

	if _, err := os.Stat(readFolder); err != nil {

		model.EventLogAdd(db, c, "500", "GetList", "'"+readFolder+"' not exists")

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

	rows, err := generateRows(db, c, s_sort)
	if err != nil {

		if os.IsPermission(err) {

			model.EventLogAdd(db, c, "403", "GetList", "Forbidden for read "+readFolder)

			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"code": 403,
				"msg":  "Forbidden",
			}, "application/json")
		}

		model.EventLogAdd(db, c, "500", "GetList", "generateRows err: "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "error",
		}, "application/json")
	}

	model.EventLogAdd(db, c, "200", "GetList", "Get list '"+readFolder+"'")

	return c.JSON(fiber.Map{
		"code": 200,
		"rows": rows,
	}, "application/json")

}
