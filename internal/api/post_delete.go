package api

import (
	"net/url"
	"os"
	"path"
	_ "path/filepath"
	"strings"

	"github.com/western/http-here/v2/internal/conf"
	"github.com/western/http-here/v2/internal/model2"
	"github.com/western/http-here/v2/internal/util"

	"github.com/gofiber/fiber/v2"
)

func PostDelete(c *fiber.Ctx) error {

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

	u_path := util.CleanDirtyPath(u.Path)

	// -------------------------------------------------------------------------------------------------------------------------

	form_path := c.FormValue("path")
	form_path = util.CleanDirtyPath(form_path)
	if len(form_path) > 0 {
		u_path = form_path
	}

	readTarget := path.Join(conf.ArgFold, u_path)

	// -------------------------------------------------------------------------------------------------------------------------

	form, _ := c.MultipartForm()
	names := form.Value["name"]

	if len(names) == 0 {

		model2.EventLogAdd(c, 500, "PostDelete", "form is empty")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "form is empty",
		}, "application/json")
	}

	var err_list []string

	for _, val := range names {

		if len(val) > 0 {

			name := strings.ReplaceAll(val, "/", "")
			//re := regexp.MustCompile("\\s+")
			//name = re.ReplaceAllLiteralString(name, " ")

			name = util.CleanDirtyPath(name)

			if len(name) == 0 {
				model2.EventLogAdd(c, 500, "PostDelete", "name is empty")
				err_list = append(err_list, "name is empty")
				continue
			}

			fileInfo, err := os.Stat(path.Join(readTarget, name))

			if err != nil {
				model2.EventLogAdd(c, 500, "PostDelete", "'"+path.Join(readTarget, name)+"' not exists")
				err_list = append(err_list, "'"+path.Join(readTarget, name)+"' not exists")
				continue
			}

			if fileInfo.IsDir() {

				// remove fold and all inside data

				if err := os.RemoveAll(path.Join(readTarget, name)); err != nil {
					model2.EventLogAdd(c, 500, "PostDelete", "'"+path.Join(readTarget, name)+"' err "+err.Error())
					err_list = append(err_list, "'"+path.Join(readTarget, name)+"' err "+err.Error())
					continue
				}

				model2.EventLogAdd(c, 200, "PostDelete", "Remove fold '"+path.Join(readTarget, name)+"'")

			} else {

				// remove one file

				if err := os.Remove(path.Join(readTarget, name)); err != nil {
					model2.EventLogAdd(c, 500, "PostDelete", "'"+path.Join(readTarget, name)+"' err "+err.Error())
					err_list = append(err_list, "'"+path.Join(readTarget, name)+"' err "+err.Error())
					continue
				}

				model2.EventLogAdd(c, 200, "PostDelete", "Remove '"+path.Join(readTarget, name)+"'")

			}

		}
	}

	go model2.FileChkAsync()
	RemoveCacheDir(readTarget)

	if len(err_list) > 0 {

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  err_list[0],
		}, "application/json")
	}

	return c.JSON(fiber.Map{
		"code": 200,
	}, "application/json")

}
