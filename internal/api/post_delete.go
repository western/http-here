package api

import (
	"net/url"
	"os"
	"path"
	_ "path/filepath"
	"strings"

	"github.com/western/http-here/internal/model"
	"github.com/western/http-here/internal/util"

	"github.com/gofiber/fiber/v2"

	"gorm.io/gorm"
)

func PostDelete(c *fiber.Ctx) error {

	arg_fold := ""
	arg_fold = c.Locals("arg_fold").(string)

	prefix := c.Locals("prefix").(string)

	referer := c.Get("Referer")

	db := c.Locals("db").(*gorm.DB)

	// already decoded
	u, err := url.Parse(referer)
	if err != nil {

		model.EventLogAdd(db, c, "500", "PostDelete", "Error url parse "+referer+" "+err.Error())

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

	// -------------------------------------------------------------------------------------------------------------------------

	form, _ := c.MultipartForm()
	names := form.Value["name"]

	if len(names) == 0 {

		model.EventLogAdd(db, c, "500", "PostDelete", "form is empty")

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
				model.EventLogAdd(db, c, "500", "PostDelete", "name is empty")
				err_list = append(err_list, "name is empty")
				continue
			}

			fileInfo, err := os.Stat(path.Join(arg_fold, u_path, name))

			if err != nil {
				model.EventLogAdd(db, c, "500", "PostDelete", "'"+path.Join(arg_fold, u_path, name)+"' not exists")
				err_list = append(err_list, "'"+path.Join(arg_fold, u_path, name)+"' not exists")
				continue
			}

			if fileInfo.IsDir() {

				// remove fold and all inside data

				if err := os.RemoveAll(path.Join(arg_fold, u_path, name)); err != nil {
					model.EventLogAdd(db, c, "500", "PostDelete", "'"+path.Join(arg_fold, u_path, name)+"' err "+err.Error())
					err_list = append(err_list, "'"+path.Join(arg_fold, u_path, name)+"' err "+err.Error())
					continue
				}

				model.EventLogAdd(db, c, "200", "PostDelete", "Remove fold '"+path.Join(arg_fold, u_path, name)+"'")

			} else {

				// remove one file

				if err := os.Remove(path.Join(arg_fold, u_path, name)); err != nil {
					model.EventLogAdd(db, c, "500", "PostDelete", "'"+path.Join(arg_fold, u_path, name)+"' err "+err.Error())
					err_list = append(err_list, "'"+path.Join(arg_fold, u_path, name)+"' err "+err.Error())
					continue
				}

				model.EventLogAdd(db, c, "200", "PostDelete", "Remove '"+path.Join(arg_fold, u_path, name)+"'")

			}

		}
	}

	go model.FileChkAsync(db, prefix)

	if len(err_list) > 0 {

		//panic(err_list[0])

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  err_list[0],
		}, "application/json")
	}

	return c.JSON(fiber.Map{
		"code": 200,
	}, "application/json")

}
