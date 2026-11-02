package api

import (
	"net/url"
	//"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/western/http-here/v2/internal/conf"
	"github.com/western/http-here/v2/internal/model2"
	"github.com/western/http-here/v2/internal/safefs"
	"github.com/western/http-here/v2/internal/util"

	"github.com/gofiber/fiber/v3"
)

func PostRename(c fiber.Ctx) error {

	referer := c.Get("Referer")

	// already decoded
	u, err := url.Parse(referer)
	if err != nil {

		model2.EventLogAdd(c, 500, "PostRename", "Error url parse "+referer+" "+err.Error())

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

		_, err := Root.Stat(strings.TrimPrefix(url_Path, "/"))
		if err != nil {

			model2.EventLogAdd(c, 500, "PostRename", "stat err "+err.Error())

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "stat err",
			})
		}

		/*
			if !fileInfo.IsDir() {

				model2.EventLogAdd(c, 500, "PostRename", "It is not directory. "+targetPath)

				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code": 500,
					"msg":  "It is not directory",
				})
			}
		*/
	}

	// -------------------------------------------------------------------------------------------------------------------------
	// REOPEN targetRoot for use below

	targetRoot, err := safefs.NewSafeFS(targetPath)
	if err != nil {

		model2.EventLogAdd(c, 500, "PostRename", "Open storage targetRoot err: "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "stat err",
		})
	}

	// -------------------------------------------------------------------------------------------------------------------------

	fromName := c.FormValue("name")
	fromName = util.CleanDirtyPath(fromName)

	if len(fromName) == 0 {

		model2.EventLogAdd(c, 500, "PostRename", "name is empty")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "name is empty",
		})
	}

	/*
		_, err = os.Stat(path.Join(targetPath, fromName))
		if err != nil {

			model2.EventLogAdd(c, 500, "PostRename", "'"+path.Join(targetPath, name)+"' not exists "+err.Error())

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "'" + path.Join(targetPath, name) + "' not exists",
			})
		}
	*/

	_, err = targetRoot.Stat(fromName)
	if err != nil {

		model2.EventLogAdd(c, 500, "PostRename", "'"+path.Join(targetPath, fromName)+"' not exists "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "'" + path.Join(url_Path, fromName) + "' not exists",
		})
	}

	// -------------------------------------------------------------------------------------------------------------------------

	toName := c.FormValue("to")
	toName = util.CleanDirtyPath(toName)

	if len(toName) == 0 {

		model2.EventLogAdd(c, 500, "PostRename", "to is empty")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "to is empty",
		})
	}

	/*
		_, err = os.Stat(path.Join(targetPath, to))
		if err == nil {

			model2.EventLogAdd(c, 500, "PostRename", "'"+path.Join(conf.ArgFold, to)+"' already exists ")

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "'" + path.Join(conf.ArgFold, to) + "' already exists",
			})
		}
	*/

	_, err = targetRoot.Stat(toName)
	if err == nil {

		model2.EventLogAdd(c, 500, "PostRename", "'"+path.Join(targetPath, toName)+"' already exists "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "'" + path.Join(url_Path, toName) + "' already exists",
		})
	}

	// -------------------------------------------------------------------------------------------------------------------------
	// open root with path.Join(targetPath, to)

	// -------------------------------------------------------------------------------------------------------------------------

	src_file_path := path.Join(targetPath, fromName)
	target_file_path := path.Join(targetPath, toName)

	//err = os.Rename(src_file_path, target_file_path)
	err = targetRoot.Rename(fromName, toName)

	if err != nil {

		model2.EventLogAdd(c, 500, "PostRename", "Rename error "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Rename error",
		})

	} else {

		model2.EventLogAdd(c, 200, "PostRename", "Rename '"+src_file_path+"' => "+target_file_path)
	}

	RemoveCacheDir(targetPath)

	return c.JSON(fiber.Map{
		"code": 200,
	})

}
