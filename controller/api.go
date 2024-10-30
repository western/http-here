package controller

import (
	_ "errors"
	_ "fmt"

	"path/filepath"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"

	"os"

	"io"
	"log"
	"net/url"
)

func PostUpload(c *fiber.Ctx) error {

	//arg_fold := os.Getenv("arg_fold")
	arg_fold := ""
	arg_fold = c.Locals("arg_fold").(string)

	referer := c.Get("Referer")
	//fmt.Println("referer= " + referer)

	// already decoded
	u, err := url.Parse(referer)
	if err != nil {
		panic(err)
	}
	//fmt.Println("u.path= " + u.Path)

	//u_path := filepath.Clean(u.Path)
	u_path := CleanDirtyPath(u.Path)
	//fmt.Println("u_path clean=" + u_path)

	form, _ := c.MultipartForm()
	files := form.File["fileBlob"]

	for _, file := range files {
		fileExt := filepath.Ext(file.Filename)
		fileExt = strings.ToLower(fileExt)

		originalFileName := strings.TrimSuffix(
			filepath.Base(file.Filename),
			filepath.Ext(file.Filename),
		)

		originalFileName = strings.ReplaceAll(originalFileName, "/", "")
		re := regexp.MustCompile("\\s+")
		originalFileName = re.ReplaceAllLiteralString(originalFileName, "-")

		filename := originalFileName + fileExt
		filename = CleanDirtyPath(filename)

		//fmt.Println("filename=" + filename)
		//fmt.Println("filepath.Join=" + filepath.Join(arg_fold, u_path, filename))

		if fileInfo, err := os.Stat(filepath.Join(arg_fold, u_path, filename)); err == nil {

			if !fileInfo.IsDir() {
				LogPrefix(c, "200", "'"+filepath.Join(arg_fold, u_path, filename)+"' already exists. It will be rewrite.")
			}
		}

		out, err := os.Create(filepath.Join(arg_fold, u_path, filename))
		if err != nil {

			log.Println(err)
			LogPrefix(c, "500", "Error create "+filepath.Join(arg_fold, u_path, filename))
			return c.JSON(fiber.Map{
				"code": 500,
			}, "application/json")
		}
		defer out.Close()

		LogPrefix(c, "200", "Save '"+filepath.Join(arg_fold, u_path, filename)+"'")

		readerFile, _ := file.Open()
		_, err = io.Copy(out, readerFile)
		if err != nil {

			log.Println(err)
			LogPrefix(c, "500", "Error copy "+filepath.Join(arg_fold, u_path, filename))
			return c.JSON(fiber.Map{
				"code": 500,
			}, "application/json")
		}
	}

	return c.JSON(fiber.Map{
		"code": 200,
	}, "application/json")
}

func PostFolder(c *fiber.Ctx) error {

	//arg_fold := os.Getenv("arg_fold")
	arg_fold := ""
	arg_fold = c.Locals("arg_fold").(string)

	referer := c.Get("Referer")
	//fmt.Println("referer= " + referer)

	// already decoded
	u, err := url.Parse(referer)
	if err != nil {

		log.Println(err)
		LogPrefix(c, "500", "Error url parse "+referer)
		return c.JSON(fiber.Map{
			"code": 500,
		}, "application/json")
	}

	//u_path := filepath.Clean(u.Path)
	u_path := CleanDirtyPath(u.Path)
	//fmt.Println("u_path clean=" + u_path)

	name := c.FormValue("name")

	name = strings.ReplaceAll(name, "/", "")
	re := regexp.MustCompile("\\s+")
	name = re.ReplaceAllLiteralString(name, " ")

	name = CleanDirtyPath(name)

	//fmt.Println("name= " + name)
	//fmt.Println("filepath.Join=" + filepath.Join(arg_fold, u_path, name))

	if fileInfo, err := os.Stat(filepath.Join(arg_fold, u_path, name)); err == nil {

		if fileInfo.IsDir() {
			LogPrefix(c, "500", "'"+filepath.Join(arg_fold, u_path, name)+"' already exists")
			return c.JSON(fiber.Map{
				"code": 500,
				"msg":  filepath.Join(u_path, name) + " already exists",
			}, "application/json")
		}
	}

	if err := os.Mkdir(filepath.Join(arg_fold, u_path, name), os.ModePerm); err != nil {

		log.Println(err)
		LogPrefix(c, "500", "Error mkdir "+filepath.Join(arg_fold, u_path, name))
		return c.JSON(fiber.Map{
			"code": 500,
		}, "application/json")
	}

	LogPrefix(c, "200", "Mkdir '"+filepath.Join(arg_fold, u_path, name)+"'")

	return c.JSON(fiber.Map{
		"code": 200,
	}, "application/json")
}
