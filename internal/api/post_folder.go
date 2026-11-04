package api

import (
	"net/url"
	"os"
	//"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/western/http-here/v2/internal/conf"
	"github.com/western/http-here/v2/internal/model2"
	"github.com/western/http-here/v2/internal/util"
	"github.com/western/http-here/v2/pkg/safefs"

	"github.com/gofiber/fiber/v3"
)

func PostFolder(c fiber.Ctx) error {

	referer := c.Get("Referer")

	// already decoded
	u, err := url.Parse(referer)
	if err != nil {

		model2.EventLogAdd(c, 500, "PostFolder", "Error url parse "+referer+" "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Error url parse " + referer,
		})
	}

	url_Path := util.CleanDirtyPath(u.Path)

	// -------------------------------------------------------------------------------------------------------------------------

	form_path := c.FormValue("path")
	form_path = util.CleanDirtyPath(form_path)
	if len(form_path) > 0 {
		url_Path = form_path
	}

	targetPath := filepath.Join(conf.ArgFold, url_Path)

	// -------------------------------------------------------------------------------------------------------------------------
	// open root with trust ArgFold

	Root := conf.ArgFoldRootFS

	if url_Path != "/" {

		fileInfo, err := Root.Stat(strings.TrimPrefix(url_Path, "/"))
		if err != nil {

			model2.EventLogAdd(c, 500, "PostFolder", "stat err "+err.Error())

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "stat err",
			})
		}

		if !fileInfo.IsDir() {

			model2.EventLogAdd(c, 500, "PostFolder", "It is not directory. "+targetPath)

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "It is not directory",
			})
		}
	}

	// -------------------------------------------------------------------------------------------------------------------------
	// REOPEN targetRoot for use below

	targetRoot, err := safefs.NewSafeFS(targetPath)
	if err != nil {

		model2.EventLogAdd(c, 500, "PostFolder", "Open storage targetRoot err: "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "stat err",
		})
	}

	// -------------------------------------------------------------------------------------------------------------------------

	var (
		regexpSpaceSymbols = regexp.MustCompile("\\s+")
		//regexpDoubleDash   = regexp.MustCompile("[\\-]{2,}")
	)

	name := c.FormValue("name")

	name = strings.ReplaceAll(name, "/", "")
	name = regexpSpaceSymbols.ReplaceAllLiteralString(name, " ")

	name = util.CleanDirtyPath(name)

	if len(name) == 0 {

		model2.EventLogAdd(c, 500, "PostFolder", "name is empty")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "name is empty",
		})
	}

	targetPathName := filepath.Join(targetPath, name)

	if fileInfo, err := targetRoot.Stat(name); err == nil {

		if fileInfo.IsDir() {

			model2.EventLogAdd(c, 500, "PostFolder", "'"+targetPathName+"' already exists")

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  filepath.Join(url_Path, name) + " already exists",
			})
		}
	}

	if err := targetRoot.Mkdir(name, os.ModePerm); err != nil {

		model2.EventLogAdd(c, 500, "PostFolder", "Error mkdir "+targetPathName+" "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Error mkdir " + filepath.Join(url_Path, name),
		})
	}

	model2.EventLogAdd(c, 200, "PostFolder", "Mkdir '"+targetPathName+"'")

	RemoveCacheDir(targetPath)

	return c.JSON(fiber.Map{
		"code": 200,
	})
}
