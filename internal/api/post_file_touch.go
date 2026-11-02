package api

import (
	//"os"
	"path"
	"regexp"
	"strings"

	"github.com/western/http-here/v2/internal/conf"
	"github.com/western/http-here/v2/internal/model2"
	"github.com/western/http-here/v2/internal/safefs"
	"github.com/western/http-here/v2/internal/util"

	"github.com/gofiber/fiber/v3"
)

func PostFileTouch(c fiber.Ctx) error {

	url_Path, err := util.RefererOrPath(c)
	if err != nil {

		model2.EventLogAdd(c, 500, "PostFileTouch", "RefererOrPath error: "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "RefererOrPath error",
		})
	}

	targetPath := path.Join(conf.ArgFold, url_Path)

	// -------------------------------------------------------------------------------------------------------------------------
	// check root with trust ArgFold

	Root := conf.ArgFoldRootFS

	if url_Path != "/" {

		fileInfo, err := Root.Stat(strings.TrimPrefix(url_Path, "/"))
		if err != nil {

			model2.EventLogAdd(c, 500, "PostFileTouch", "stat err "+err.Error())

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "stat err",
			})
		}

		if !fileInfo.IsDir() {

			model2.EventLogAdd(c, 500, "PostFileTouch", "It is not directory. "+targetPath)

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "It is not directory",
			})
		}
	}

	// -------------------------------------------------------------------------------------------------------------------------
	// reopen targetRoot with full targetPath

	//targetRoot, err := os.OpenRoot(targetPath)
	targetRoot, err := safefs.NewSafeFS(targetPath)
	if err != nil {

		model2.EventLogAdd(c, 500, "PostFileTouch", "open storage targetRoot err1. "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "open storage targetRoot err1",
		})
	}

	defer targetRoot.Close()

	// -------------------------------------------------------------------------------------------------------------------------

	var (
		regexpSpaceSymbols = regexp.MustCompile("\\s+")
		//regexpDoubleDash   = regexp.MustCompile("[\\-]{2,}")
	)

	formName := c.FormValue("name")

	formName = strings.ReplaceAll(formName, "/", "")
	formName = regexpSpaceSymbols.ReplaceAllLiteralString(formName, " ")

	formName = util.CleanDirtyPath(formName)

	if len(formName) == 0 {

		model2.EventLogAdd(c, 500, "PostFileTouch", "name is empty")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "name is empty",
		})
	}

	if fileInfo, err := targetRoot.Stat(formName); err == nil {

		if fileInfo.IsDir() {

			model2.EventLogAdd(c, 500, "PostFileTouch", "'"+path.Join(targetPath, formName)+"' already exists")

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  path.Join(url_Path, formName) + " already exists",
			})
		}
	}

	out, err := targetRoot.Create(formName)
	if err != nil {

		model2.EventLogAdd(c, 500, "PostFileTouch", "Error create "+formName+" "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Error create " + formName,
		})
	}
	defer out.Close()

	out.WriteString("\n")

	model2.EventLogAdd(c, 200, "PostFileTouch", "Create file '"+path.Join(targetPath, formName)+"'")
	RemoveCacheDir(targetPath)

	return c.JSON(fiber.Map{
		"code": 200,
	})
}
