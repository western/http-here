package api

import (
	_ "fmt"
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

func PostCopy(c *fiber.Ctx) error {

	arg_fold := ""
	arg_fold = c.Locals("arg_fold").(string)

	prefix := c.Locals("prefix").(string)

	db := c.Locals("db").(*gorm.DB)

	referer := c.Get("Referer")

	// already decoded
	u, err := url.Parse(referer)
	if err != nil {

		model.EventLogAdd(db, c, "500", "PostCopy", "Error url parse "+referer+" "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Error url parse " + referer,
		}, "application/json")
	}

	u_path := util.CleanDirtyPath(u.Path)

	// -------------------------------------------------------------------------------------------------------------------------

	from_path := c.FormValue("from_path")
	from_path = util.CleanDirtyPath(from_path)

	// -------------------------------------------------------------------------------------------------------------------------

	to := c.FormValue("to")
	to = util.CleanDirtyPath(to)

	if len(to) == 0 {

		model.EventLogAdd(db, c, "500", "PostCopy", "to is empty")

		return c.JSON(fiber.Map{
			"code": 500,
			"msg":  "to is empty",
		}, "application/json")
	}

	_, err = os.Stat(path.Join(arg_fold, to))
	if err != nil {

		model.EventLogAdd(db, c, "500", "PostCopy", "'"+path.Join(arg_fold, to)+"' not exists "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "'" + path.Join(arg_fold, to) + "' not exists",
		}, "application/json")
	}

	// -------------------------------------------------------------------------------------------------------------------------

	src_file_path := u_path
	if len(from_path) > 0 {
		src_file_path = from_path
	}

	if src_file_path == to {

		model.EventLogAdd(db, c, "500", "PostCopy", "Source and target paths '"+to+"' are identical")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Source and target paths '" + to + "' are identical",
		}, "application/json")
	}

	// -------------------------------------------------------------------------------------------------------------------------

	form, _ := c.MultipartForm()
	names := form.Value["name"]

	if len(names) == 0 {

		model.EventLogAdd(db, c, "500", "PostCopy", "form is empty")

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
				model.EventLogAdd(db, c, "500", "PostCopy", "name is empty")
				err_list = append(err_list, "name is empty")
				continue
			}

			src_file_path := path.Join(arg_fold, u_path, name)
			if len(from_path) > 0 {
				src_file_path = path.Join(arg_fold, from_path, name)
			}

			target_file_path := path.Join(arg_fold, to, name)

			src_stat, err := os.Stat(src_file_path)
			if err != nil {
				model.EventLogAdd(db, c, "500", "PostCopy", "Source '"+src_file_path+"' not exists")
				err_list = append(err_list, "Source '"+src_file_path+"' not exists")
				continue
			}

			_, err = os.Stat(target_file_path)
			if err == nil {

				model.EventLogAdd(db, c, "200", "PostCopy", "Target '"+target_file_path+"' is exists. It will be rewrite.")

				if err := os.RemoveAll(target_file_path); err != nil {
					model.EventLogAdd(db, c, "500", "PostCopy", "'"+target_file_path+"' err "+err.Error())
					err_list = append(err_list, "'"+target_file_path+"' err "+err.Error())
					continue
				}
			}

			if src_stat.IsDir() {
				err := util.CopyDir(src_file_path, target_file_path)

				if err != nil {
					model.EventLogAdd(db, c, "500", "PostCopy", "CopyDir '"+src_file_path+"' to '"+target_file_path+"' err "+err.Error())
					err_list = append(err_list, "CopyDir '"+src_file_path+"' to '"+target_file_path+"' err "+err.Error())
					continue
				}

				model.EventLogAdd(db, c, "200", "PostCopy", "Copy dir '"+src_file_path+"' to "+target_file_path)

			} else {
				err := util.CopyFile(src_file_path, target_file_path)

				if err != nil {
					model.EventLogAdd(db, c, "500", "PostCopy", "CopyFile '"+src_file_path+"' to '"+target_file_path+"' err "+err.Error())
					err_list = append(err_list, "CopyFile '"+src_file_path+"' to '"+target_file_path+"' err "+err.Error())
					continue
				}

				model.EventLogAdd(db, c, "200", "PostCopy", "Copy file '"+src_file_path+"' to "+target_file_path)
			}

		}
	}

	go model.FileChkAsync(db, prefix)
	
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
