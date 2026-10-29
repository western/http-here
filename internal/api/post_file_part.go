package api

import (
	//"fmt"
	"io"
	//"net/url"
	"os"
	"path"
	//"regexp"
	//"strings"

	"github.com/western/http-here/v2/internal/conf"
	//"github.com/western/http-here/v2/internal/crypt"
	"github.com/western/http-here/v2/internal/model2"
	"github.com/western/http-here/v2/internal/util"

	"github.com/gofiber/fiber/v3"
)

func PostFilePart(c fiber.Ctx) error {

	form, err := c.MultipartForm()
	if err != nil {

		model2.EventLogAdd(c, 500, "PostFilePart", "Invalid form data")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Invalid form data",
		})
	}

	// Get a text field value (returns map[string][]string)
	/*
	   nameValues := form.Value["username"]
	   var username string
	   if len(nameValues) > 0 {
	       username = nameValues[0]
	   }
	*/

	// -------------------------------------------------------------------------------------------------------------------------

	// file_id
	// `${file.name}--${file.size}--${file.lastModified}`
	// file_id= FFF.zip--119028570--1772832382300

	file_id := c.FormValue("file_id")
	if len(file_id) == 0 {

		model2.EventLogAdd(c, 500, "PostFilePart", "Field file_id is empty")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Field file_id is empty",
		})
	}
	//fmt.Println("file_id=", file_id)

	// chunk_index
	// int

	chunk_index := c.FormValue("chunk_index")
	if !util.IsInt(chunk_index) {

		model2.EventLogAdd(c, 500, "PostFilePart", "Field chunk_index is not number")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Field chunk_index is not number",
		})
	}
	//fmt.Println("chunk_index=", chunk_index)

	// data
	// bytes

	/*
		    data := c.FormValue("data")
		    fmt.Println("data=", data)
		    if len(data) == 0 {

		        model2.EventLogAdd(c, 500, "PostFilePart", "Field data is empty")

				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code": 500,
					"msg":  "Field data is empty",
				})
		    }
	*/

	/*
	   dataArr := form.Value["data"]
	   fmt.Println("len(dataArr)=", len(dataArr))

	   //var data []byte
	   var data string
	   if len(dataArr) > 0 {
	       data = dataArr[0]
	   }
	   fmt.Println("data=", data)
	*/

	files := form.File["data"]
	if len(files) == 0 {

		model2.EventLogAdd(c, 500, "PostFilePart", "Field data is empty")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Field data is empty",
		})
	}

	data := files[0]
	//fmt.Println("data=", data)

	// -------------------------------------------------------------------------------------------------------------------------

	root, err := os.OpenRoot(path.Join(conf.ConfigRoot, "parts"))
	if err != nil {

		model2.EventLogAdd(c, 500, "PostFilePart", "open storage root err "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "open storage root err",
		})
	}

	// extract EXT from string "FFF.zip--119028570--1772832382300"

	/*
	   preResult := strings.Split(file_id, "--")
	   fileExt := "xxx"
	   if len(preResult) == 3 {

	       preResult2 := strings.Split(preResult[0], ".")
	       if len(preResult2) == 2 {
	           fileExt = preResult2[1]
	       }
	   }
	   fmt.Println("fileExt=", fileExt)
	*/

	//readTargetFilename := path.Join(readTarget, filename)
	fullFilename := file_id + "." + chunk_index
	//fmt.Println("fullFilename=", fullFilename)

	// ----------------------------------------------------------------------------------------------------------------------

	/*
		out, err := os.Create(readTargetFilename)
		if err != nil {

			model2.EventLogAdd(c, 500, "PostFilePart", "Error create "+readTargetFilename+" "+err.Error())

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "Error create " + readTargetFilename,
			}, "application/json")
		}
		defer out.Close()
	*/

	out, err := root.Create(fullFilename)
	if err != nil {

		model2.EventLogAdd(c, 500, "PostFilePart", "Error create "+fullFilename+" "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Error create " + fullFilename,
		})
	}

	readerFile, _ := data.Open()
	_, err = io.Copy(out, readerFile)
	if err != nil {

		model2.EventLogAdd(c, 500, "PostFilePart", "Error copy "+fullFilename+" "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Error copy " + fullFilename,
		})
	}

	model2.EventLogAdd(c, 200, "PostFilePart", "Save '"+fullFilename+"'")

	// -------------------------------------------------------------------------------------------------------------------------

	return c.JSON(fiber.Map{
		"code": 200,
	})
}
