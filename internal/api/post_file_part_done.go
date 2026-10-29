package api

import (
	//"fmt"
	//"io"
	//"net/url"
	"os"
	"path"
	//"regexp"
	"strconv"
	"strings"

	"github.com/western/http-here/v2/internal/conf"
	//"github.com/western/http-here/v2/internal/crypt"
	"github.com/western/http-here/v2/internal/model2"
	"github.com/western/http-here/v2/internal/util"

	"github.com/gofiber/fiber/v3"
)

func PostFilePartDone(c fiber.Ctx) error {

	// file_id
	// name
	// to_path
	// chunk_cnt

	// file_id
	// `${file.name}--${file.size}--${file.lastModified}`
	// file_id= FFF.zip--119028570--1772832382300

	file_id := c.FormValue("file_id")
	if len(file_id) == 0 {

		model2.EventLogAdd(c, 500, "PostFilePartDone", "Field file_id is empty")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Field file_id is empty",
		})
	}
	//fmt.Println("file_id=", file_id)

	// name
	// `${file.name}`

	name := c.FormValue("name")
	if len(name) == 0 {

		model2.EventLogAdd(c, 500, "PostFilePartDone", "Field name is empty")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Field name is empty",
		})
	}
	//fmt.Println("name=", name)

	// to_path
	// string, place for copy

	to_path := c.FormValue("to_path")
	to_path = util.CleanDirtyPath(to_path)
	if len(to_path) == 0 {

		model2.EventLogAdd(c, 500, "PostFilePartDone", "Field to_path is empty")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Field to_path is empty",
		})
	}
	//fmt.Println("to_path=", to_path)

	// chunk_cnt
	// int

	chunk_cnt := c.FormValue("chunk_cnt")
	if !util.IsInt(chunk_cnt) {

		model2.EventLogAdd(c, 500, "PostFilePartDone", "Field chunk_cnt is not number")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Field chunk_cnt is not number",
		})
	}
	//fmt.Println("chunk_cnt=", chunk_cnt)

	chunk_cnt_int, _ := strconv.Atoi(chunk_cnt)

	// -------------------------------------------------------------------------------------------------------------------------

	srcRoot, err := os.OpenRoot(path.Join(conf.ConfigRoot, "parts"))
	if err != nil {

		model2.EventLogAdd(c, 500, "PostFilePartDone", "open storage srcRoot err "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "open storage srcRoot err",
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
	*/
	//fmt.Println("fileExt=", fileExt)

	/*
	   fileExt := "xxx"
	   preResult2 := strings.Split(name, ".")
	   if len(preResult2) == 2 {
	       fileExt = preResult2[1]
	   }
	   //fmt.Println("fileExt=", fileExt)
	*/

	//readTargetFilename := path.Join(readTarget, filename)
	//fullFilename := file_id + "." + chunk_index
	//fmt.Println("fullFilename=", fullFilename)

	// ----------------------------------------------------------------------------------------------------------------------
	// assembleFile

	out, err := srcRoot.OpenFile(name, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {

		model2.EventLogAdd(c, 500, "PostFilePartDone", "Error create "+name+" "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Error create " + name,
		})
	}

	for indx := range chunk_cnt_int {

		//fmt.Println("indx=", indx)

		readTargetFile := file_id + "." + strconv.Itoa(indx)
		//fmt.Println("readTargetFile=", readTargetFile)

		//content, err := os.ReadFile("example.txt")
		content, err := srcRoot.ReadFile(readTargetFile)
		if err != nil {
			panic(err)
		}

		if _, err := out.Write(content); err != nil {
			panic(err)
		}

	}

	out.Close()

	// -------------------------------------------------------------------------------------------------------------------------
	// open root with trust ArgFold

	Root, err := os.OpenRoot(conf.ArgFold)
	if err != nil {

		model2.EventLogAdd(c, 500, "PostFilePartDone", "open storage Root err1. "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "open storage Root err1",
		})
	}

	//fmt.Println("to_path=", to_path)
	//fmt.Println("to_path=", strings.TrimPrefix(to_path, "/"))

	fileInfo, err := Root.Stat(strings.TrimPrefix(to_path, "/"))
	if err != nil {

		model2.EventLogAdd(c, 500, "PostFilePartDone", "stat err "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "stat err",
		})
	}

	if !fileInfo.IsDir() {

		model2.EventLogAdd(c, 500, "PostFilePartDone", "It is not directory. "+to_path)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "It is not directory",
		})
	}

	// -------------------------------------------------------------------------------------------------------------------------
	// reopen dstRoot with full targetPath

	dstRoot, err := os.OpenRoot(path.Join(conf.ArgFold, to_path))
	if err != nil {

		model2.EventLogAdd(c, 500, "PostFilePartDone", "open storage dstRoot err1. "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "open storage dstRoot err1",
		})
	}

	// ----------------------------------------------------------------------------------------------------------------------
	// copyAssembledTo

	/*
		dstRoot, err := os.OpenRoot(path.Join(conf.ArgFold, to_path))
		if err != nil {

			model2.EventLogAdd(c, 500, "PostFilePartDone", "open storage dstRoot err "+err.Error())

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "open storage dstRoot err",
			})
		}
	*/

	//src_file_path := path.Join(conf.ArgFold, from_path, name)
	//target_file_path := path.Join(conf.ArgFold, to_path, name)

	// ROOTCopyFile(srcRoot *os.Root, src string, dstRoot *os.Root, dst string)

	err = util.ROOTCopyFile(srcRoot, name, dstRoot, name)

	if err != nil {
		panic(err)
	}

	// ----------------------------------------------------------------------------------------------------------------------
	// remove assembled file

	err = srcRoot.Remove(name)
	if err != nil {
		panic(err)
	}

	// ----------------------------------------------------------------------------------------------------------------------
	// clearAssembledFileParts

	for indx := range chunk_cnt_int {

		readTargetFile := file_id + "." + strconv.Itoa(indx)

		err := srcRoot.Remove(readTargetFile)
		if err != nil {
			panic(err)
		}
	}

	// ----------------------------------------------------------------------------------------------------------------------

	model2.EventLogAdd(c, 200, "PostFilePartDone", "File assembled and copied to '"+path.Join(conf.ArgFold, to_path, name)+"'")

	return c.JSON(fiber.Map{
		"code": 200,
	})
}
