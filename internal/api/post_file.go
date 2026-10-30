package api

import (
	//"fmt"
	"io"
	"mime/multipart"
	//"net/url"
	"os"
	"path"
	//"regexp"
	"strings"

	"github.com/western/http-here/v2/internal/conf"
	"github.com/western/http-here/v2/internal/crypt"
	"github.com/western/http-here/v2/internal/model2"
	"github.com/western/http-here/v2/internal/util"

	"github.com/gofiber/fiber/v3"
)

func PostFile(c fiber.Ctx) error {

	url_Path, err := util.RefererOrPath(c)
	if err != nil {

		model2.EventLogAdd(c, 500, "PostFile", "RefererOrPath error: "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "RefererOrPath error",
		})
	}

	targetPath := path.Join(conf.ArgFold, url_Path)

	// -------------------------------------------------------------------------------------------------------------------------
	// open root with trust ArgFold

	Root, err := os.OpenRoot(conf.ArgFold)
	if err != nil {

		model2.EventLogAdd(c, 500, "PostFile", "open storage Root err1. "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "open storage Root err1",
		})
	}

	if url_Path != "/" {

		fileInfo, err := Root.Stat(strings.TrimPrefix(url_Path, "/"))
		if err != nil {

			model2.EventLogAdd(c, 500, "PostFile", "stat err "+err.Error())

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "stat err",
			})
		}

		if !fileInfo.IsDir() {

			model2.EventLogAdd(c, 500, "PostFile", "It is not directory. "+targetPath)

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "It is not directory",
			})
		}
	}

	// -------------------------------------------------------------------------------------------------------------------------
	// reopen targetRoot with full targetPath

	targetRoot, err := os.OpenRoot(targetPath)
	if err != nil {

		model2.EventLogAdd(c, 500, "PostFile", "open storage targetRoot err1. "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "open storage targetRoot err1",
		})
	}

	// -------------------------------------------------------------------------------------------------------------------------

	form, _ := c.MultipartForm()
	files := form.File["fileBlob"]

	// file = type *multipart.FileHeader
	for _, file := range files {

		fileName, err := util.GetClearFilename(file.Filename)
		if err != nil {

			model2.EventLogAdd(c, 500, "PostFile", "Error clear file name "+err.Error())

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "Error clear file name ",
			})
		}

		// -------------------------------------------------------------------------------------------------------------------------

		arg_crypt := GetBoolFromLocals(c, "arg_crypt")
		code := c.Cookies("code")

		// reset code for safety
		if !arg_crypt && len(code) > 0 {

			setCookie(c, "code", "")
		}

		if arg_crypt && len(code) > 0 {

			err := uploadAndEncrypt(c, file, targetRoot, fileName, code)
			if err != nil {
				return err
			}

		} else {

			err := uploadOnly(c, file, targetRoot, fileName)
			if err != nil {
				return err
			}
		}

	}

	RemoveCacheDir(targetPath)

	return c.JSON(fiber.Map{
		"code": 200,
	})
}

func uploadOnly(c fiber.Ctx, file *multipart.FileHeader, targetRoot *os.Root, fileName string) error {

	//fmt.Println("targetRoot.Name()=", targetRoot.Name())

	targetPathFilename := path.Join(targetRoot.Name(), fileName)

	/*
		targetPathFilename := path.Join(targetPath, fileName)

		if fileInfo, err := os.Stat(targetPathFilename); err == nil {

			if !fileInfo.IsDir() {

				model2.EventLogAdd(c, 200, "PostFile", "'"+targetPathFilename+"' already exists. It will be rewrite.")
			}
		}
	*/

	// ----------------------------------------------------------------------------------------------------------------------

	/*
		out, err := os.Create(targetPathFilename)
		if err != nil {

			model2.EventLogAdd(c, 500, "PostFile", "Error create "+targetPathFilename+" "+err.Error())

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "Error create " + targetPathFilename,
			})
		}
		defer out.Close()
	*/

	out, err := targetRoot.Create(fileName)
	if err != nil {

		model2.EventLogAdd(c, 500, "PostFile", "Error create "+fileName+" "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Error create " + fileName,
		})
	}
	defer out.Close()

	readerFile, _ := file.Open()
	_, err = io.Copy(out, readerFile)
	if err != nil {

		model2.EventLogAdd(c, 500, "PostFile", "Error copy "+fileName+" "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Error copy " + fileName,
		})
	}

	model2.EventLogAdd(c, 200, "PostFile", "Saved '"+targetPathFilename+"'")

	return nil
}

func uploadAndEncrypt(c fiber.Ctx, file *multipart.FileHeader, targetRoot *os.Root, fileName, code string) error {

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

		model2.EventLogAdd(c, 500, "PostFile", "Error EncryptFile "+fileName+" "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Error EncryptFile " + fileName,
		})
	}

	// ------------------------------------------------------------------------------------------------------------------

	fileName_crypt := fileName + ".crypt"
	targetPathFilenameCrypt := path.Join(targetRoot.Name(), fileName_crypt)

	/*
		targetPathFilenameCrypt := path.Join(targetPath, fileName+".crypt")

		out, err := os.Create(targetPathFilenameCrypt)
		if err != nil {

			model2.EventLogAdd(c, 500, "PostFile", "Error create "+targetPathFilenameCrypt+" "+err.Error())

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "Error create " + targetPathFilenameCrypt,
			})
		}
		defer out.Close()
	*/

	out, err := targetRoot.Create(fileName_crypt)
	if err != nil {

		model2.EventLogAdd(c, 500, "PostFile", "Error create "+fileName_crypt+" "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Error create " + fileName_crypt,
		})
	}
	defer out.Close()

	readerFile, _ = os.Open(fileEncrypt.Name())

	_, err = io.Copy(out, readerFile)
	if err != nil {
		panic(err)
	}
	readerFile.Close()

	model2.EventLogAdd(c, 200, "PostFile", "Save encrypted '"+targetPathFilenameCrypt+"'")

	return nil

}
