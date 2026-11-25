package app

import (
	//"embed"
	//"errors"
	//"flag"
	//"fmt"
	//"log"
	//"net"
	//"net/http"
	//"net/url"
	//"os"
	//"path"
	//"path/filepath"
	//"runtime"
	//"strconv"
	//"strings"
	//"time"

	"github.com/western/http-here/v2/internal/api"
	//api_v1 "github.com/western/http-here/v2/internal/api/v1"
	//api_v2 "github.com/western/http-here/v2/internal/api/v2"
	"github.com/western/http-here/v2/internal/conf"
	//"github.com/western/http-here/v2/internal/crypt"
	"github.com/western/http-here/v2/internal/model2"
	//"github.com/western/http-here/v2/internal/util"

	//"github.com/fatih/color"
	"github.com/gofiber/fiber/v3"
	//"github.com/gofiber/template/html/v3"
	//"github.com/jedib0t/go-pretty/v6/table"
)

func CoreFiberRoute(app *fiber.App) {

	// -------------------------------------------------------------------------------------------------------------------------------------------

	arg_extend_mode := conf.ArgsMap["extend-mode"].(bool)

	arg_index_disable := conf.ArgsMap["index-disable"].(bool)

	arg_upload_disable := conf.ArgsMap["upload-disable"].(bool)

	arg_folder_make_disable := conf.ArgsMap["folder-make-disable"].(bool)

	// -------------------------------------------------------------------------------------------------------------------------------------------

	if arg_extend_mode {

		app.Get("/__thumb/*", api.GetThumb)
	}

	app.Get("/__temp/*", api.GetTemp)

	// -------------------------------------------------------------------------------------------------------------------------------------------
	// TESTS

	/*
		app.Get("/api/test/get", api.GetTestGet)
		app.Get("/api/test/set", api.GetTestSet)
		app.Get("/api/test/del", api.GetTestDel)
	*/

	// -------------------------------------------------------------------------------------------------------------------------------------------
	// ROUTE

	app.Post("/api/jserr", api.PostJserr)

	if !arg_index_disable {

		if arg_extend_mode {
			app.Get("/__doc/*", GetEditDoc)
			app.Get("/__code/*", GetEditCode)
			app.Get("/__md/*", GetEditMd)
			app.Get("/__search/", api.PageSearch)
			//app.Get("/api/file/convert/*", api.GetFileConvert)
			app.Post("/api/file/edit", PostFileEdit)
		}

		//app.Get("/api/list", api.GetList)
	}

	if !arg_upload_disable {
		app.Post("/api/file", api.PostFile)

		//app.Post("/api/file/part/init", api.PostFilePartInit)
		//app.Post("/api/file/part", api.PostFilePart)
		//app.Post("/api/file/part/done", api.PostFilePartDone)
	}

	if !arg_folder_make_disable {
		app.Post("/api/folder", api.PostFolder)
		app.Post("/api/file/touch", api.PostFileTouch)
	}

	if arg_extend_mode {
		app.Post("/api/delete", api.PostDelete)
		app.Post("/api/move", api.PostMove)
		app.Post("/api/copy", api.PostCopy)
		app.Post("/api/rename", api.PostRename)
		app.Post("/api/zip", api.PostZip)

		//app.Get("/api/search", api.GetSearch)
	}

	if !arg_index_disable {

		/*
			if *arg_spa {
				app.Get("/", GetSpa)
				app.Get("/*", api.GetAll)
			} else {
				app.Get("/*", api.GetAll)
			}
		*/

		app.Get("/*", api.GetAll)
	}

	// -------------------------------------------------------------------------------------------------------------------------------------------
	// BAD REQUEST

	app.Use(func(c fiber.Ctx) error {

		model2.EventLogAdd(c, 404, "last_handle", c.Path())

		return c.Status(fiber.StatusNotFound).Render("view/404", fiber.Map{}, "view/layout/error")
	})

}
