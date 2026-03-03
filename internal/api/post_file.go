package api

import (
	"io"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/western/http-here/v2/internal/conf"
	"github.com/western/http-here/v2/internal/model2"
	"github.com/western/http-here/v2/internal/util"

	"github.com/gofiber/fiber/v2"
)

func PostFile(c *fiber.Ctx) error {

	arg_crypt := GetBoolFromLocals(c, "arg_crypt")

	// -------------------------------------------------------------------------------------------------------------------------

	referer := c.Get("Referer")

	// already decoded
	u, err := url.Parse(referer)
	if err != nil {

		model2.EventLogAdd(c, 500, "PostFileUpload", "Error url parse "+referer+" "+err.Error())

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

				model2.EventLogAdd(c, 200, "PostFileUpload", "'"+path.Join(readTarget, filename)+"' already exists. It will be rewrite.")
			}
		}

		// -------------------------------------------------------------------------------------------------------------------------

		code := c.Cookies("code")

		if arg_crypt && len(code) > 0 {

			setCookie(c, "code", "")
		}

		if arg_crypt && len(code) > 0 {

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

				model2.EventLogAdd(c, 500, "PostFileUpload", "Error CryptFile "+err.Error())

				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code": 500,
					"msg":  "Error CryptFile",
				}, "application/json")
			}

			out, err := os.Create(path.Join(readTarget, filename+".crypt"))
			if err != nil {

				model2.EventLogAdd(c, 500, "PostFileUpload", "Error create "+path.Join(readTarget, filename+".crypt")+" "+err.Error())

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

			model2.EventLogAdd(c, 200, "PostFileUpload", "Save encrypted '"+path.Join(readTarget, filename+".crypt")+"'")

		} else {

			out, err := os.Create(path.Join(readTarget, filename))
			if err != nil {

				model2.EventLogAdd(c, 500, "PostFileUpload", "Error create "+path.Join(readTarget, filename)+" "+err.Error())

				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code": 500,
					"msg":  "Error create " + path.Join(readTarget, filename),
				}, "application/json")
			}
			defer out.Close()

			readerFile, _ := file.Open()
			_, err = io.Copy(out, readerFile)
			if err != nil {

				model2.EventLogAdd(c, 500, "PostFileUpload", "Error copy "+path.Join(readTarget, filename)+" "+err.Error())

				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code": 500,
					"msg":  "Error copy " + path.Join(readTarget, filename),
				}, "application/json")
			}

			model2.EventLogAdd(c, 200, "PostFileUpload", "Save '"+path.Join(readTarget, filename)+"'")

		}

	}

	RemoveCacheDir(readTarget)

	return c.JSON(fiber.Map{
		"code": 200,
	}, "application/json")
}
