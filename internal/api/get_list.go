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

	// -------------------------------------------------------------------------------------------------------------------------

	if _, err := os.Stat(path.Join(arg_fold, u_path)); err != nil {

		model.EventLogAdd(db, c, "500", "GetList", "'"+path.Join(arg_fold, u_path)+"' not exists")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  path.Join(u_path) + " not exists",
		}, "application/json")
	}

	entries, err := os.ReadDir(path.Join(arg_fold, u_path))
	if err != nil {

		model.EventLogAdd(db, c, "500", "GetList", "Error "+path.Join(arg_fold, u_path)+" "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  path.Join(u_path) + " readdir error",
		}, "application/json")
	}

	var s_sort string = "name"

	s_sort = c.Cookies("sort")
	if len(s_sort) == 0 {
		s_sort = "name"
	}

	rows := listGenerateView(db, arg_fold, u_path, entries, s_sort)

	model.EventLogAdd(db, c, "200", "GetList", "Get list '"+path.Join(arg_fold, u_path)+"'")

	return c.JSON(fiber.Map{
		"code": 200,
		"rows": rows,
	}, "application/json")

}
