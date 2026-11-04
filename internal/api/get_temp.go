package api

import (
	"net/url"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/western/http-here/v2/internal/conf"
	"github.com/western/http-here/v2/internal/model2"
	"github.com/western/http-here/v2/internal/util"

	"github.com/gofiber/fiber/v3"
)

func GetTemp(c fiber.Ctx) error {

	c_Path, err := url.QueryUnescape(c.Path())
	c_Path = strings.TrimLeft(c_Path, "/__temp")

	if err != nil {

		model2.EventLogAdd(c, 500, "__temp", "Error "+c_Path+" "+err.Error())
		return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")
	}

	c_Path = util.CleanDirtyPath(c_Path)
	targetTempPath := filepath.Join(conf.ConfigRoot, "temp", c_Path)

	if runtime.GOOS == "windows" {
		targetTempPath = util.RotateSlash(targetTempPath)
	}

	// ------------------------------------------------------------------------------------------------------------------------------------------------------

	Root := conf.TempRootFS

	fileInfo, err := Root.Stat(strings.TrimPrefix(c_Path, "/"))

	if err != nil {

		model2.EventLogAdd(c, 404, "__temp", "Root.Stat err "+err.Error())

		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code": 404,
			"msg":  c_Path + " not found",
		})
	}

	fileMode := fileInfo.Mode()

	if !fileMode.IsRegular() {

		model2.EventLogAdd(c, 500, "__temp", "Is not a regular file. "+targetTempPath)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Is not a regular file",
		})
	}

	// ------------------------------------------------------------------------------------------------------------------------------------------------------

	model2.EventLogAdd(c, 200, "__temp", "Temp get "+targetTempPath)

	return c.SendFile(targetTempPath)

}
