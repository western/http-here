package api

import (
	"io"
	"net/url"
	"os"
	"path"
	_ "path/filepath"
	"regexp"
	"strings"

	"github.com/western/http-here/internal/model"
	"github.com/western/http-here/internal/util"

	"github.com/gofiber/fiber/v2"
)

func PostUpload(c *fiber.Ctx) error {

	arg_fold := ""
	arg_fold = c.Locals("arg_fold").(string)

	arg_crypt := ""
	if c.Locals("arg_crypt") != nil {
		arg_crypt = c.Locals("arg_crypt").(string)
	}

	db, err := model.ConnectToSQLite()
	if err != nil {
		panic(err)
	}

	// -------------------------------------------------------------------------------------------------------------------------

	referer := c.Get("Referer")

	// already decoded
	u, err := url.Parse(referer)
	if err != nil {

		//util.LogPrefix(c, "500", "Error url parse "+referer+" "+err.Error())
		//model.EventLogMsg(db, c, "500", "PostUpload", "Error url parse "+referer+" "+err.Error())
		model.EventLogAdd(db, c, "500", "PostUpload", "Error url parse "+referer+" "+err.Error())

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
	files := form.File["fileBlob"]

	for _, file := range files {

		file_ext := util.GetExtNorm(file.Filename)
		originalFileName := util.GetFileName(file.Filename)

		originalFileName = strings.ReplaceAll(originalFileName, "/", "")
		re := regexp.MustCompile("\\s+")
		originalFileName = re.ReplaceAllLiteralString(originalFileName, "-")
		re = regexp.MustCompile("[\\-]{2,}")
		originalFileName = re.ReplaceAllLiteralString(originalFileName, "-")

		filename := originalFileName + "." + file_ext
		filename = util.CleanDirtyPath(filename)

		// -------------------------------------------------------------------------------------------------------------------------

		if fileInfo, err := os.Stat(path.Join(arg_fold, u_path, filename)); err == nil {

			if !fileInfo.IsDir() {
				//util.LogPrefix(c, "200", "'"+path.Join(arg_fold, u_path, filename)+"' already exists. It will be rewrite.")
				//model.EventLogMsg(db, c, "200", "PostUpload", "'"+path.Join(arg_fold, u_path, filename)+"' already exists. It will be rewrite.")
				model.EventLogAdd(db, c, "200", "PostUpload", "'"+path.Join(arg_fold, u_path, filename)+"' already exists. It will be rewrite.")
			}
		}

		// -------------------------------------------------------------------------------------------------------------------------

		code := c.Cookies("code")

		if arg_crypt == "" && len(code) > 0 {

			cookie := new(fiber.Cookie)
			cookie.Name = "code"
			c.Cookie(cookie)
		}

		if arg_crypt == "1" && len(code) > 0 {

			//panic(code)

			f, err := os.CreateTemp("", "httphere_crypt*")
			if err != nil {
				panic(err)
			}
			//fmt.Println("Crypt Temp file name:", f.Name())
			defer os.Remove(f.Name())

			readerFile, _ := file.Open()
			_, err = io.Copy(f, readerFile)
			if err != nil {
				panic(err)
			}
			f.Close()

			isOk, err := util.CryptFile(f.Name(), code)
			if !isOk {

				//util.LogPrefix(c, "500", "Error CryptFile "+err.Error())
				//model.EventLogMsg(db, c, "500", "PostUpload", "Error CryptFile "+err.Error())
				model.EventLogAdd(db, c, "500", "PostUpload", "Error CryptFile "+err.Error())

				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code": 500,
					"msg":  "Error CryptFile",
				}, "application/json")
			}

			out, err := os.Create(path.Join(arg_fold, u_path, filename+".crypt"))
			if err != nil {

				//util.LogPrefix(c, "500", "Error create "+path.Join(arg_fold, u_path, filename+".crypt")+" "+err.Error())
				//model.EventLogMsg(db, c, "500", "PostUpload", "Error create "+path.Join(arg_fold, u_path, filename+".crypt")+" "+err.Error())
				model.EventLogAdd(db, c, "500", "PostUpload", "Error create "+path.Join(arg_fold, u_path, filename+".crypt")+" "+err.Error())

				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code": 500,
					"msg":  "Error create " + path.Join(arg_fold, u_path, filename+".crypt"),
				}, "application/json")
			}
			defer out.Close()

			readerFile, _ = os.Open(f.Name())

			_, err = io.Copy(out, readerFile)
			if err != nil {
				panic(err)
			}
			f.Close()

			//util.LogPrefix(c, "200", "Save encrypted '"+path.Join(arg_fold, u_path, filename+".crypt")+"'")
			//model.EventLogMsg(db, c, "200", "PostUpload", "Save encrypted '"+path.Join(arg_fold, u_path, filename+".crypt")+"'")
			model.EventLogAdd(db, c, "200", "PostUpload", "Save encrypted '"+path.Join(arg_fold, u_path, filename+".crypt")+"'")

		} else {

			out, err := os.Create(path.Join(arg_fold, u_path, filename))
			if err != nil {

				//util.LogPrefix(c, "500", "Error create "+path.Join(arg_fold, u_path, filename)+" "+err.Error())
				//model.EventLogMsg(db, c, "500", "PostUpload", "Error create "+path.Join(arg_fold, u_path, filename)+" "+err.Error())
				model.EventLogAdd(db, c, "500", "PostUpload", "Error create "+path.Join(arg_fold, u_path, filename)+" "+err.Error())

				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code": 500,
					"msg":  "Error create " + path.Join(arg_fold, u_path, filename),
				}, "application/json")
			}
			defer out.Close()

			readerFile, _ := file.Open()
			_, err = io.Copy(out, readerFile)
			if err != nil {

				//util.LogPrefix(c, "500", "Error copy "+path.Join(arg_fold, u_path, filename)+" "+err.Error())
				//model.EventLogMsg(db, c, "500", "PostUpload", "Error copy "+path.Join(arg_fold, u_path, filename)+" "+err.Error())
				model.EventLogAdd(db, c, "500", "PostUpload", "Error copy "+path.Join(arg_fold, u_path, filename)+" "+err.Error())

				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code": 500,
					"msg":  "Error copy " + path.Join(arg_fold, u_path, filename),
				}, "application/json")
			}

			//util.LogPrefix(c, "200", "Save '"+path.Join(arg_fold, u_path, filename)+"'")
			//model.EventLogMsg(db, c, "200", "PostUpload", "Save '"+path.Join(arg_fold, u_path, filename)+"'")
			model.EventLogAdd(db, c, "200", "PostUpload", "Save '"+path.Join(arg_fold, u_path, filename)+"'")

		}

	}

	return c.JSON(fiber.Map{
		"code": 200,
	}, "application/json")
}
