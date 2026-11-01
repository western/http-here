package api

import (
	"os"
	"path"
	"strings"

	"github.com/western/http-here/v2/internal/conf"
	"github.com/western/http-here/v2/internal/model2"
	"github.com/western/http-here/v2/internal/safefs"
	"github.com/western/http-here/v2/internal/util"

	"github.com/gofiber/fiber/v3"
)

func GetList(c fiber.Ctx) error {

	// -------------------------------------------------------------------------------------------------------------------------

	form_Path := c.FormValue("path")

	form_Path = util.CleanDirtyPath(form_Path)

	targetPath := path.Join(conf.ArgFold, form_Path)

	// -------------------------------------------------------------------------------------------------------------------------

	Root := conf.ArgFoldRootFS

	if form_Path != "/" {

		fileInfo, err := Root.Stat(strings.TrimPrefix(form_Path, "/"))
		if err != nil {

			model2.EventLogAdd(c, 500, "GetList", "stat err "+err.Error())

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  form_Path + " not exists",
			})
		}

		fileMode := fileInfo.Mode()

		if !fileMode.IsDir() {

			model2.EventLogAdd(c, 500, "GetList", "It is not directory: "+form_Path)

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "It is not directory",
			})
		}
	}

	// -------------------------------------------------------------------------------------------------------------------------
	// reopen targetRoot with full targetPath

	targetRoot, err := safefs.OpenRoot(targetPath)
	if err != nil {

		model2.EventLogAdd(c, 500, "GetList", "Open storage targetRoot err: "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  form_Path + " not exists",
		})
	}

	// -------------------------------------------------------------------------------------------------------------------------

	var s_sort string = "name"

	s_sort = c.Cookies("sort")
	if len(s_sort) == 0 {
		s_sort = "name"
	}

	rows, err := generateRows(c, targetRoot, s_sort)
	if err != nil {

		if os.IsPermission(err) {

			model2.EventLogAdd(c, 403, "GetList", "Forbidden for read "+targetPath)

			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"code": 403,
				"msg":  "Forbidden",
			})
		}

		model2.EventLogAdd(c, 500, "GetList", "generateRows err: "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "error",
		})
	}

	model2.EventLogAdd(c, 200, "GetList", "Get list '"+targetPath+"'")

	return c.JSON(fiber.Map{
		"code": 200,
		"rows": rows,
	})

}
