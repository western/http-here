package controller

import (

	_ "fmt"

	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"


	"os"



	"io"
	"log"
	"net/url"
)

func PostUpload(c *fiber.Ctx) error {


	arg_fold := os.Getenv("arg_fold")

	referer := c.Get("Referer")
	//fmt.Println("referer= " + referer)

	// already decoded
	u, err := url.Parse(referer)
	if err != nil {
		panic(err)
	}
	//fmt.Println("u.path= " + u.Path)

	u_path := filepath.Clean(u.Path)
	//fmt.Println("u_path clean=" + u_path)



	form, _ := c.MultipartForm()
	files := form.File["fileBlob"]

	for _, file := range files {
		fileExt := filepath.Ext(file.Filename)
		fileExt = strings.ToLower(fileExt)

		originalFileName := strings.TrimSuffix(filepath.Base(file.Filename), filepath.Ext(file.Filename))
		//now := time.Now()
		//filename := strings.ReplaceAll(strings.ToLower(originalFileName), " ", "-") + "-" + fmt.Sprintf("%v", now.Unix()) + fileExt
		//filename := strings.ReplaceAll(strings.ToLower(originalFileName), " ", "-") + fileExt
		filename := strings.ReplaceAll(originalFileName, " ", "-") + fileExt



		//fmt.Println("filepath.Join=" + filepath.Join(arg_fold, u_path, filename))

		out, err := os.Create(filepath.Join(arg_fold, u_path, filename))
		if err != nil {
			log.Fatal(err)
		}
		defer out.Close()

		readerFile, _ := file.Open()
		_, err = io.Copy(out, readerFile)
		if err != nil {
			log.Fatal(err)
		}
	}

	return c.JSON(fiber.Map{
		"code": 200,
	}, "application/json")
}

func PostFolder(c *fiber.Ctx) error {


	arg_fold := os.Getenv("arg_fold")

	referer := c.Get("Referer")
	//fmt.Println("referer= " + referer)

	// already decoded
	u, err := url.Parse(referer)
	if err != nil {
		panic(err)
	}



	u_path := filepath.Clean(u.Path)
	//fmt.Println("u_path clean=" + u_path)


	name := c.FormValue("name")

	//fmt.Println("name= " + name)
	//fmt.Println("filepath.Join=" + filepath.Join(arg_fold, u_path, name))

	if err := os.Mkdir(filepath.Join(arg_fold, u_path, name), os.ModePerm); err != nil {
		log.Fatal(err)
	}

	return c.JSON(fiber.Map{
		"code": 200,
	}, "application/json")
}
