package api

import (
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/western/http-here/v2/internal/conf"
	"github.com/western/http-here/v2/internal/model2"
	"github.com/western/http-here/v2/internal/util"

	"github.com/gofiber/fiber/v3"
)

func PostDelete(c fiber.Ctx) error {

	referer := c.Get("Referer")

	// already decoded
	u, err := url.Parse(referer)
	if err != nil {

		model2.EventLogAdd(c, 500, "PostDelete", "Error url parse "+referer+" "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Error url parse " + referer,
		}, "application/json")
	}

	url_Path := util.CleanDirtyPath(u.Path)

	// -------------------------------------------------------------------------------------------------------------------------

	form_path := c.FormValue("path")
	form_path = util.CleanDirtyPath(form_path)
	if len(form_path) > 0 {
		url_Path = form_path
	}

	targetPath := path.Join(conf.ArgFold, url_Path)

	// -------------------------------------------------------------------------------------------------------------------------
	// open root with trust ArgFold

	Root, err := os.OpenRoot(conf.ArgFold)
	if err != nil {

		model2.EventLogAdd(c, 500, "PostDelete", "open storage Root err1. "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "open storage Root err1",
		})
	}

	//fmt.Println("url_Path=", url_Path)
	//fmt.Println("url_Path=", strings.TrimPrefix(url_Path, "/"))

	fileInfo, err := Root.Stat(strings.TrimPrefix(url_Path, "/"))
	if err != nil {

		model2.EventLogAdd(c, 500, "PostDelete", "stat err "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "stat err",
		})
	}

	if !fileInfo.IsDir() {

		model2.EventLogAdd(c, 500, "PostDelete", "It is not directory. "+targetPath)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "It is not directory",
		})
	}

	// -------------------------------------------------------------------------------------------------------------------------
	// reopen targetRoot with full targetPath

	targetRoot, err := os.OpenRoot(targetPath)
	//_, err = os.OpenRoot(targetPath)
	if err != nil {

		model2.EventLogAdd(c, 500, "PostDelete", "open storage targetRoot err1. "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "open storage targetRoot err1",
		})
	}

	// -------------------------------------------------------------------------------------------------------------------------

	form, _ := c.MultipartForm()
	names := form.Value["name"]

	if len(names) == 0 {

		model2.EventLogAdd(c, 500, "PostDelete", "form is empty")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "form is empty",
		})
	}

	var err_list []string

	for _, name := range names {

		//name, _ = util.GetClearFilename(name)

		if len(name) == 0 {
			model2.EventLogAdd(c, 500, "PostDelete", "name is empty")
			err_list = append(err_list, "name is empty")
			continue
		}

		targetPathName := path.Join(targetRoot.Name(), name)

		fileInfo, err := targetRoot.Stat(name)

		if err != nil {
			model2.EventLogAdd(c, 500, "PostDelete", "'"+targetPathName+"' not exists")
			err_list = append(err_list, "'"+targetPathName+"' not exists")
			continue
		}

		if fileInfo.IsDir() {

			// remove fold and all inside data

			if err := targetRoot.RemoveAll(name); err != nil {
				model2.EventLogAdd(c, 500, "PostDelete", "'"+targetPathName+"' err "+err.Error())
				err_list = append(err_list, "'"+targetPathName+"' err "+err.Error())
				continue
			}

			RemoveCacheDir(path.Join(targetPath, name))

			model2.EventLogAdd(c, 200, "PostDelete", "Remove fold '"+targetPathName+"'")

		} else {

			// remove one file

			if err := targetRoot.Remove(name); err != nil {
				model2.EventLogAdd(c, 500, "PostDelete", "'"+targetPathName+"' err "+err.Error())
				err_list = append(err_list, "'"+targetPathName+"' err "+err.Error())
				continue
			}

			model2.EventLogAdd(c, 200, "PostDelete", "Remove '"+targetPathName+"'")

		}

	}

	RemoveCacheDir(targetPath)

	if len(err_list) > 0 {

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  err_list[0],
		})
	}

	return c.JSON(fiber.Map{
		"code": 200,
	})

}
