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

	"gorm.io/gorm"
)

func PostFileUpload(c *fiber.Ctx) error {

	arg_fold := ""
	arg_fold = c.Locals("arg_fold").(string)

	arg_crypt := ""
	if c.Locals("arg_crypt") != nil {
		arg_crypt = c.Locals("arg_crypt").(string)
	}

	db := c.Locals("db").(*gorm.DB)

	// -------------------------------------------------------------------------------------------------------------------------

	referer := c.Get("Referer")

	// already decoded
	u, err := url.Parse(referer)
	if err != nil {

		model.EventLogAdd(db, c, "500", "PostFileUpload", "Error url parse "+referer+" "+err.Error())

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

	readTarget := path.Join(arg_fold, u_path)

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

		if fileInfo, err := os.Stat(path.Join(readTarget, filename)); err == nil {

			if !fileInfo.IsDir() {

				model.EventLogAdd(db, c, "200", "PostFileUpload", "'"+path.Join(readTarget, filename)+"' already exists. It will be rewrite.")
			}
		}

		// -------------------------------------------------------------------------------------------------------------------------

		code := c.Cookies("code")

		if arg_crypt == "" && len(code) > 0 {

			setCookie(c, "code", "")
		}

		if arg_crypt == "1" && len(code) > 0 {

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

			isOk, err := util.EncryptFile(f.Name(), code)
			if !isOk {

				model.EventLogAdd(db, c, "500", "PostFileUpload", "Error CryptFile "+err.Error())

				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code": 500,
					"msg":  "Error CryptFile",
				}, "application/json")
			}

			out, err := os.Create(path.Join(readTarget, filename+".crypt"))
			if err != nil {

				model.EventLogAdd(db, c, "500", "PostFileUpload", "Error create "+path.Join(readTarget, filename+".crypt")+" "+err.Error())

				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code": 500,
					"msg":  "Error create " + path.Join(readTarget, filename+".crypt"),
				}, "application/json")
			}
			defer out.Close()

			readerFile, _ = os.Open(f.Name())

			_, err = io.Copy(out, readerFile)
			if err != nil {
				panic(err)
			}
			f.Close()

			model.EventLogAdd(db, c, "200", "PostFileUpload", "Save encrypted '"+path.Join(readTarget, filename+".crypt")+"'")

		} else {

			out, err := os.Create(path.Join(readTarget, filename))
			if err != nil {

				model.EventLogAdd(db, c, "500", "PostFileUpload", "Error create "+path.Join(readTarget, filename)+" "+err.Error())

				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code": 500,
					"msg":  "Error create " + path.Join(readTarget, filename),
				}, "application/json")
			}
			defer out.Close()

			readerFile, _ := file.Open()
			_, err = io.Copy(out, readerFile)
			if err != nil {

				model.EventLogAdd(db, c, "500", "PostFileUpload", "Error copy "+path.Join(readTarget, filename)+" "+err.Error())

				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code": 500,
					"msg":  "Error copy " + path.Join(readTarget, filename),
				}, "application/json")
			}

			model.EventLogAdd(db, c, "200", "PostFileUpload", "Save '"+path.Join(readTarget, filename)+"'")

		}

	}

	return c.JSON(fiber.Map{
		"code": 200,
	}, "application/json")
}
