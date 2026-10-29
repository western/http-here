package api

import (
	//"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/western/http-here/v2/internal/conf"
	"github.com/western/http-here/v2/internal/crypt"
	"github.com/western/http-here/v2/internal/model2"
	"github.com/western/http-here/v2/internal/util"

	"github.com/gofiber/fiber/v3"
)

func PostFile(c fiber.Ctx) error {

	referer := c.Get("Referer")

	// already decoded
	u, err := url.Parse(referer)
	if err != nil {

		model2.EventLogAdd(c, 500, "PostFile", "Error url parse "+referer+" "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Error url parse " + referer,
		})
	}

	url_Path := util.CleanDirtyPath(u.Path)

	// -------------------------------------------------------------------------------------------------------------------------

	form_path := c.FormValue("path")
	form_path = util.CleanDirtyPath(form_path)
	if len(form_path) > 0 {
		url_Path = form_path
	}

	targetPath := path.Join(conf.ArgFold, url_Path)

	// -------------------------------------------------------------------------------------------------------------------------
	// check path if exists

	fileInfo, err := os.Stat(targetPath)
	if err != nil {

		model2.EventLogAdd(c, 500, "PostFile", "Path not found. "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Path not found",
		})
	}

	if !fileInfo.IsDir() {

		model2.EventLogAdd(c, 500, "PostFile", "It is not path. "+targetPath)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "It is not path",
		})
	}

	// -------------------------------------------------------------------------------------------------------------------------
	// reopen root with full targetPath

	/*
		targetRoot, err := os.OpenRoot(targetPath)
		if err != nil {

			model2.EventLogAdd(c, 500, "PostFile", "open storage targetRoot err2. "+err.Error())

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "open storage targetRoot err2",
			})
		}
	*/

	// -------------------------------------------------------------------------------------------------------------------------

	form, _ := c.MultipartForm()
	files := form.File["fileBlob"]

	// file = type *multipart.FileHeader
	for _, file := range files {

		fileName, err := GetClearFilename(file.Filename)
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

			err := uploadAndCrypt(c, file, targetPath, fileName, code)
			if err != nil {
				return err
			}

		} else {

			err := uploadOnly(c, file, targetPath, fileName)
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

func uploadOnly(c fiber.Ctx, file *multipart.FileHeader, targetPath, fileName string) error {

	targetPathFilename := path.Join(targetPath, fileName)

	if fileInfo, err := os.Stat(targetPathFilename); err == nil {

		if !fileInfo.IsDir() {

			model2.EventLogAdd(c, 200, "PostFile", "'"+targetPathFilename+"' already exists. It will be rewrite.")
		}
	}

	// ----------------------------------------------------------------------------------------------------------------------

	out, err := os.Create(targetPathFilename)
	if err != nil {

		model2.EventLogAdd(c, 500, "PostFile", "Error create "+targetPathFilename+" "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Error create " + targetPathFilename,
		})
	}
	defer out.Close()

	readerFile, _ := file.Open()
	_, err = io.Copy(out, readerFile)
	if err != nil {

		model2.EventLogAdd(c, 500, "PostFile", "Error copy "+targetPathFilename+" "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Error copy " + targetPathFilename,
		})
	}

	model2.EventLogAdd(c, 200, "PostFile", "Save '"+targetPathFilename+"'")

	return nil
}

func uploadAndCrypt(c fiber.Ctx, file *multipart.FileHeader, targetPath, fileName, code string) error {

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

	readerFile, _ = os.Open(fileEncrypt.Name())

	_, err = io.Copy(out, readerFile)
	if err != nil {
		panic(err)
	}
	readerFile.Close()

	model2.EventLogAdd(c, 200, "PostFile", "Save encrypted '"+targetPathFilenameCrypt+"'")

	return nil

}

/*
fileName, err := api.GetClearFilename("image1.jpg")
fmt.Println("fileName=", fileName, " err=",err)
fileName, err = api.GetClearFilename("/path/to/image2.jpg")
fmt.Println("fileName=", fileName, " err=",err)
fileName, err = api.GetClearFilename(`\path\to\ima  ge3.jpg`)
fmt.Println("fileName=", fileName, " err=",err)
*/
func GetClearFilename(dirty_Filename string) (string, error) {

	var (
		regexpSpaceSymbols = regexp.MustCompile("\\s+")
		//regexpDoubleDash   = regexp.MustCompile("[\\-]{2,}")
	)

	unescape_Filename, err := url.QueryUnescape(dirty_Filename)
	if err != nil {

		return "", err
	}

	file_ext := util.GetExtNorm(unescape_Filename)
	originalFileName := util.GetFileName(unescape_Filename)

	originalFileName = strings.ReplaceAll(originalFileName, `/`, "")
	originalFileName = strings.ReplaceAll(originalFileName, `\`, "")
	originalFileName = regexpSpaceSymbols.ReplaceAllLiteralString(originalFileName, " ")
	//originalFileName = regexpDoubleDash.ReplaceAllLiteralString(originalFileName, "-")

	filename := originalFileName + "." + file_ext
	filename = util.CleanDirtyPath(filename)

	return filename, nil
}
