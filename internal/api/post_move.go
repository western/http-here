package api

import (
	//"fmt"
	//"net/url"
	"os"
	"path"
	"strings"

	"github.com/western/http-here/internal/model"
	"github.com/western/http-here/internal/util"

	"github.com/gofiber/fiber/v2"

	"gorm.io/gorm"
)

func PostMove(c *fiber.Ctx) error {

	arg_fold := GetStringFromLocals(c, "arg_fold", "")
	prefix := GetStringFromLocals(c, "prefix", "")

	db := c.Locals("db").(*gorm.DB)

	// -------------------------------------------------------------------------------------------------------------------------

	from_path := c.FormValue("from_path")
	from_path = util.CleanDirtyPath(from_path)

	if len(from_path) == 0 {

		model.EventLogAdd(db, c, "500", "PostMove", "from_path is empty")

		return c.JSON(fiber.Map{
			"code": 500,
			"msg":  "from_path is empty",
		}, "application/json")
	}

	_, err := os.Stat(path.Join(arg_fold, from_path))
	if err != nil {

		model.EventLogAdd(db, c, "500", "PostMove", "'"+path.Join(arg_fold, from_path)+"' not exists "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "'" + path.Join(from_path) + "' not exists",
		}, "application/json")
	}

	// -------------------------------------------------------------------------------------------------------------------------

	to_path := c.FormValue("to_path")
	to_path = util.CleanDirtyPath(to_path)

	if len(to_path) == 0 {

		model.EventLogAdd(db, c, "500", "PostMove", "to_path is empty")

		return c.JSON(fiber.Map{
			"code": 500,
			"msg":  "to_path is empty",
		}, "application/json")
	}

	_, err = os.Stat(path.Join(arg_fold, to_path))
	if err != nil {

		model.EventLogAdd(db, c, "500", "PostMove", "'"+path.Join(arg_fold, to_path)+"' not exists "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "'" + path.Join(to_path) + "' not exists",
		}, "application/json")
	}

	// -------------------------------------------------------------------------------------------------------------------------

	if from_path == to_path {

		model.EventLogAdd(db, c, "500", "PostMove", "Source and target paths '"+from_path+"' are identical")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Source and target paths '" + from_path + "' are identical",
		}, "application/json")
	}

	// -------------------------------------------------------------------------------------------------------------------------

	form, _ := c.MultipartForm()
	names := form.Value["name"]

	if len(names) == 0 {

		model.EventLogAdd(db, c, "500", "PostMove", "form is empty")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "form is empty",
		}, "application/json")
	}

	var err_list []string

	for _, name := range names {

		name := strings.ReplaceAll(name, "/", "")
		name = util.CleanDirtyPath(name)

		if len(name) == 0 {
			model.EventLogAdd(db, c, "500", "PostMove", "name is empty")
			err_list = append(err_list, "name is empty")
			continue
		}

		src_file_path := path.Join(arg_fold, from_path, name)
		target_file_path := path.Join(arg_fold, to_path, name)

		src_stat, err := os.Stat(src_file_path)
		if err != nil {
			model.EventLogAdd(db, c, "500", "PostMove", "Source '"+src_file_path+"' not exists")
			err_list = append(err_list, "Source '"+src_file_path+"' not exists")
			continue
		}

		_, err = os.Stat(target_file_path)
		if err == nil {

			model.EventLogAdd(db, c, "200", "PostMove", "Target '"+target_file_path+"' is exists. It will be rewrite.")

			if err := os.RemoveAll(target_file_path); err != nil {
				model.EventLogAdd(db, c, "500", "PostMove", "'"+target_file_path+"' err "+err.Error())
				err_list = append(err_list, "'"+target_file_path+"' err "+err.Error())
				continue
			}
		}

		if src_stat.IsDir() {
			err := util.CopyDir(src_file_path, target_file_path)

			if err != nil {
				model.EventLogAdd(db, c, "500", "PostMove", "CopyDir '"+src_file_path+"' to '"+target_file_path+"' err "+err.Error())
				err_list = append(err_list, "CopyDir '"+src_file_path+"' to '"+target_file_path+"' err "+err.Error())
				continue
			}

			// remove source folder
			if err := os.RemoveAll(src_file_path); err != nil {
				model.EventLogAdd(db, c, "500", "PostMove", "rm folder '"+src_file_path+"' err "+err.Error())
				err_list = append(err_list, "'"+src_file_path+"' err "+err.Error())
				continue
			}

			model.EventLogAdd(db, c, "200", "PostMove", "Move dir '"+src_file_path+"' to "+target_file_path)

		} else {
			err := util.MoveFile(src_file_path, target_file_path)

			if err != nil {
				model.EventLogAdd(db, c, "500", "PostMove", "MoveFile '"+src_file_path+"' to '"+target_file_path+"' err "+err.Error())
				err_list = append(err_list, "MoveFile '"+src_file_path+"' to '"+target_file_path+"' err "+err.Error())
				continue
			}

			model.EventLogAdd(db, c, "200", "PostMove", "Move file '"+src_file_path+"' to "+target_file_path)
		}

	}

	go model.FileChkAsync(db, prefix)
	RemoveCacheDir(path.Join(arg_fold, from_path))
	RemoveCacheDir(path.Join(arg_fold, to_path))

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
