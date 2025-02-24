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
)

func PostCopy(c *fiber.Ctx) error {

	arg_fold := ""
	arg_fold = c.Locals("arg_fold").(string)

	db, err := model.ConnectToSQLite()
	if err != nil {
		panic(err)
	}

	referer := c.Get("Referer")

	// already decoded
	u, err := url.Parse(referer)
	if err != nil {

		//util.LogPrefix(c, "500", "Error url parse "+referer+" "+err.Error())
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
		//util.LogPrefix(c, "500", "to is empty")
		model.EventLogAdd(db, c, "500", "PostCopy", "to is empty")

		return c.JSON(fiber.Map{
			"code": 500,
			"msg":  "to is empty",
		}, "application/json")
	}

	_, err = os.Stat(path.Join(arg_fold, to))
	if err != nil {
		//util.LogPrefix(c, "500", "'"+path.Join(arg_fold, to)+"' not exists "+err.Error())
		model.EventLogAdd(db, c, "500", "PostCopy", "'"+path.Join(arg_fold, to)+"' not exists "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "'" + path.Join(arg_fold, to) + "' not exists",
		}, "application/json")
	}

	form, _ := c.MultipartForm()
	names := form.Value["name"]

	if len(names) == 0 {

		//util.LogPrefix(c, "500", "form is empty")
		model.EventLogAdd(db, c, "500", "PostCopy", "form is empty")

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
				model.EventLogAdd(db, c, "500", "PostCopy", "name is empty")
				continue
			}

			src_file_path := path.Join(arg_fold, u_path, name)
			if len(from_path) > 0 {
				src_file_path = path.Join(arg_fold, from_path, name)
			}

			target_file_path := path.Join(arg_fold, to, name)

			src_stat, err := os.Stat(src_file_path)
			if err != nil {
				//util.LogPrefix(c, "500", "Source '"+src_file_path+"' not exists")
				model.EventLogAdd(db, c, "500", "PostCopy", "Source '"+src_file_path+"' not exists")
				continue
			}

			_, err = os.Stat(target_file_path)
			if err == nil {

				//util.LogPrefix(c, "200", "Target '"+target_file_path+"' is exists. It will be rewrite.")
				model.EventLogAdd(db, c, "200", "PostCopy", "Target '"+target_file_path+"' is exists. It will be rewrite.")

				if err := os.RemoveAll(target_file_path); err != nil {

					//util.LogPrefix(c, "500", "'"+target_file_path+"' err "+err.Error())
					model.EventLogAdd(db, c, "500", "PostCopy", "'"+target_file_path+"' err "+err.Error())
					continue
				}
			}

			if src_stat.IsDir() {
				err := util.CopyDir(src_file_path, target_file_path)

				if err != nil {
					//util.LogPrefix(c, "500", "CopyDir '"+src_file_path+"' to '"+target_file_path+"' err "+err.Error())
					model.EventLogAdd(db, c, "500", "PostCopy", "CopyDir '"+src_file_path+"' to '"+target_file_path+"' err "+err.Error())
					continue
				}

				//util.LogPrefix(c, "200", "Copy dir '"+src_file_path+"' to "+target_file_path)
				model.EventLogAdd(db, c, "200", "PostCopy", "Copy dir '"+src_file_path+"' to "+target_file_path)

			} else {
				err := util.CopyFile(src_file_path, target_file_path)

				if err != nil {
					//util.LogPrefix(c, "500", "CopyFile '"+src_file_path+"' to '"+target_file_path+"' err "+err.Error())
					model.EventLogAdd(db, c, "500", "PostCopy", "CopyFile '"+src_file_path+"' to '"+target_file_path+"' err "+err.Error())
					continue
				}

				//util.LogPrefix(c, "200", "Copy file '"+src_file_path+"' to "+target_file_path)
				model.EventLogAdd(db, c, "200", "PostCopy", "Copy file '"+src_file_path+"' to "+target_file_path)
			}

		}
	}

	go model.FileChkAsync(db)

	return c.JSON(fiber.Map{
		"code": 200,
	}, "application/json")

}
