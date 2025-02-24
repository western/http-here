package api

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/western/http-here/internal/model"
	"github.com/western/http-here/internal/util"

	"github.com/gofiber/fiber/v2"
)

func PostDelete(c *fiber.Ctx) error {

	arg_fold := ""
	arg_fold = c.Locals("arg_fold").(string)

	referer := c.Get("Referer")

	db, err := model.ConnectToSQLite()
	if err != nil {
		panic(err)
	}

	// already decoded
	u, err := url.Parse(referer)
	if err != nil {

		//util.LogPrefix(c, "500", "Error url parse "+referer+" "+err.Error())
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

		//util.LogPrefix(c, "500", "form is empty")
		model.EventLogAdd(db, c, "500", "PostDelete", "form is empty")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "form is empty",
		}, "application/json")
	}

	for _, val := range names {

		if len(val) > 0 {
			//fmt.Println("val="+val)

			name := strings.ReplaceAll(val, "/", "")
			//re := regexp.MustCompile("\\s+")
			//name = re.ReplaceAllLiteralString(name, " ")

			name = util.CleanDirtyPath(name)

			if len(name) == 0 {
				//util.LogPrefix(c, "500", "name is empty")
				model.EventLogAdd(db, c, "500", "PostDelete", "name is empty")
				continue
			}

			fileInfo, err := os.Stat(filepath.Join(arg_fold, u_path, name))

			if err != nil {
				//util.LogPrefix(c, "500", "'"+filepath.Join(arg_fold, u_path, name)+"' not exists")
				model.EventLogAdd(db, c, "500", "PostDelete", "'"+filepath.Join(arg_fold, u_path, name)+"' not exists")
				continue
			}

			if fileInfo.IsDir() {

				// remove fold and all inside data

				if err := os.RemoveAll(filepath.Join(arg_fold, u_path, name)); err != nil {

					//util.LogPrefix(c, "500", "'"+filepath.Join(arg_fold, u_path, name)+"' err "+err.Error())
					model.EventLogAdd(db, c, "500", "PostDelete", "'"+filepath.Join(arg_fold, u_path, name)+"' err "+err.Error())
					continue
				}

				//util.LogPrefix(c, "200", "Remove fold '"+filepath.Join(arg_fold, u_path, name)+"'")
				model.EventLogAdd(db, c, "200", "PostDelete", "Remove fold '"+filepath.Join(arg_fold, u_path, name)+"'")

			} else {

				// remove one file

				if err := os.Remove(filepath.Join(arg_fold, u_path, name)); err != nil {

					//util.LogPrefix(c, "500", "'"+filepath.Join(arg_fold, u_path, name)+"' err "+err.Error())
					model.EventLogAdd(db, c, "500", "PostDelete", "'"+filepath.Join(arg_fold, u_path, name)+"' err "+err.Error())
					continue
				}

				//util.LogPrefix(c, "200", "Remove '"+filepath.Join(arg_fold, u_path, name)+"'")
				model.EventLogAdd(db, c, "200", "PostDelete", "Remove '"+filepath.Join(arg_fold, u_path, name)+"'")

			}

		}
	}

	go model.FileChkAsync(db)

	return c.JSON(fiber.Map{
		"code": 200,
	}, "application/json")

}
