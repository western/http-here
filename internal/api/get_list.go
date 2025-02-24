package api

import (
	"os"
	"path/filepath"

	_ "github.com/western/http-here/internal/model"
	"github.com/western/http-here/internal/util"

	"github.com/gofiber/fiber/v2"
)

func GetList(c *fiber.Ctx) error {

	arg_fold := ""
	arg_fold = c.Locals("arg_fold").(string)

	/*
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
	*/

	//u_path := CleanDirtyPath(u.Path)

	// -------------------------------------------------------------------------------------------------------------------------

	u_path := c.FormValue("path")

	u_path = util.CleanDirtyPath(u_path)

	// -------------------------------------------------------------------------------------------------------------------------

	/*
		name := c.FormValue("name")

		name = strings.ReplaceAll(name, "/", "")
		re := regexp.MustCompile("\\s+")
		name = re.ReplaceAllLiteralString(name, " ")

		name = CleanDirtyPath(name)
	*/

	if _, err := os.Stat(filepath.Join(arg_fold, u_path)); err != nil {

		util.LogPrefix(c, "500", "'"+filepath.Join(arg_fold, u_path)+"' not exists")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  filepath.Join(u_path) + " not exists",
		}, "application/json")
	}

	entries, err := os.ReadDir(filepath.Join(arg_fold, u_path))
	if err != nil {

		util.LogPrefix(c, "500", "Error "+filepath.Join(arg_fold, u_path)+" "+err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  filepath.Join(u_path) + " readdir error",
		}, "application/json")
	}

	var s_sort string = "name"

	s_sort = c.Cookies("sort")
	if len(s_sort) == 0 {
		s_sort = "name"
	}

	rows := listGenerateView(arg_fold, u_path, entries, s_sort)

	util.LogPrefix(c, "200", "Get list '"+filepath.Join(arg_fold, u_path)+"'")

	return c.JSON(fiber.Map{
		"code": 200,
		"rows": rows,
	}, "application/json")

}
