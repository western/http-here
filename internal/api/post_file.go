package api

import (
	//"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/western/http-here/v2/internal/conf"
	"github.com/western/http-here/v2/internal/crypt"
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

	var (
		regexpSpaceSymbols = regexp.MustCompile("\\s+")
		//regexpDoubleDash   = regexp.MustCompile("[\\-]{2,}")
	)

	// file = type *multipart.FileHeader
	for _, file := range files {

		fileFilename, err := url.QueryUnescape(file.Filename)
		if err != nil {

			model2.EventLogAdd(c, 500, "PostFileUpload", "Error unescape file name "+err.Error())

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "Error unescape file name ",
			}, "application/json")
		}

		file_ext := util.GetExtNorm(fileFilename)
		originalFileName := util.GetFileName(fileFilename)

		originalFileName = strings.ReplaceAll(originalFileName, "/", "")
		originalFileName = regexpSpaceSymbols.ReplaceAllLiteralString(originalFileName, " ")
		//originalFileName = regexpDoubleDash.ReplaceAllLiteralString(originalFileName, "-")

		filename := originalFileName + "." + file_ext
		filename = util.CleanDirtyPath(filename)

		// -------------------------------------------------------------------------------------------------------------------------

		// -------------------------------------------------------------------------------------------------------------------------

		code := c.Cookies("code")

		// reset code for safety
		if !arg_crypt && len(code) > 0 {

			setCookie(c, "code", "")
		}

		if arg_crypt && len(code) > 0 {

			// ------------------------------------------------------------------------------------------------------------------

			fileOrig, err := os.CreateTemp("", "httphere_orig*")
			if err != nil {
				panic(err)
			}
			defer os.Remove(fileOrig.Name())

			readerFile, _ := file.Open()
			_, err = io.Copy(fileOrig, readerFile)
			if err != nil {
				panic(err)
			}
			readerFile.Close()

			// ------------------------------------------------------------------------------------------------------------------

			fileEncrypt, err := os.CreateTemp("", "httphere_encrypt*")
			if err != nil {
				panic(err)
			}
			defer os.Remove(fileEncrypt.Name())

			// ------------------------------------------------------------------------------------------------------------------

			if err := crypt.GCMEncryptFile([]byte(code), fileOrig.Name(), fileEncrypt.Name()); err != nil {

				model2.EventLogAdd(c, 500, "PostFileUpload", "Error CryptFile "+err.Error())

				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code": 500,
					"msg":  "Error CryptFile",
				}, "application/json")
			}

			// ------------------------------------------------------------------------------------------------------------------

			readTargetFilenameCrypt := path.Join(readTarget, filename+".crypt")

			out, err := os.Create(readTargetFilenameCrypt)
			if err != nil {

				model2.EventLogAdd(c, 500, "PostFileUpload", "Error create "+readTargetFilenameCrypt+" "+err.Error())

				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code": 500,
					"msg":  "Error create " + readTargetFilenameCrypt,
				}, "application/json")
			}
			defer out.Close()

			readerFile, _ = os.Open(fileEncrypt.Name())

			_, err = io.Copy(out, readerFile)
			if err != nil {
				panic(err)
			}
			readerFile.Close()

			model2.EventLogAdd(c, 200, "PostFileUpload", "Save encrypted '"+readTargetFilenameCrypt+"'")

		} else {

			readTargetFilename := path.Join(readTarget, filename)

			if fileInfo, err := os.Stat(readTargetFilename); err == nil {

				if !fileInfo.IsDir() {

					model2.EventLogAdd(c, 200, "PostFileUpload", "'"+readTargetFilename+"' already exists. It will be rewrite.")
				}
			}

			// ----------------------------------------------------------------------------------------------------------------------

			out, err := os.Create(readTargetFilename)
			if err != nil {

				model2.EventLogAdd(c, 500, "PostFileUpload", "Error create "+readTargetFilename+" "+err.Error())

				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code": 500,
					"msg":  "Error create " + readTargetFilename,
				}, "application/json")
			}
			defer out.Close()

			readerFile, _ := file.Open()
			_, err = io.Copy(out, readerFile)
			if err != nil {

				model2.EventLogAdd(c, 500, "PostFileUpload", "Error copy "+readTargetFilename+" "+err.Error())

				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code": 500,
					"msg":  "Error copy " + readTargetFilename,
				}, "application/json")
			}

			model2.EventLogAdd(c, 200, "PostFileUpload", "Save '"+readTargetFilename+"'")

		}

	}

	RemoveCacheDir(readTarget)

	return c.JSON(fiber.Map{
		"code": 200,
	}, "application/json")
}
