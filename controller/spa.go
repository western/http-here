package controller

import (
	_ "encoding/json"
	_ "errors"
	_ "fmt"
	_ "html/template"
	_ "io"
	_ "log"
	_ "net/url"
	_ "os"
	_ "path/filepath"
	_ "reflect"
	_ "regexp"
	_ "sort"
	_ "strings"
	_ "time"

	_ "github.com/western/http-here/model"

	"github.com/gofiber/fiber/v2"
)

func GetSpa(c *fiber.Ctx) error {

	/*
			arg_fold := ""
			arg_fold = c.Locals("arg_fold").(string)

			arg_upload_disable := ""
			if c.Locals("arg_upload_disable") != nil {
				arg_upload_disable = c.Locals("arg_upload_disable").(string)
			}

			arg_folder_make_disable := ""
			if c.Locals("arg_folder_make_disable") != nil {
				arg_folder_make_disable = c.Locals("arg_folder_make_disable").(string)
			}

			arg_extend_mode := ""
			if c.Locals("arg_extend_mode") != nil {
				arg_extend_mode = c.Locals("arg_extend_mode").(string)
			}

			arg_crypt := ""
			if c.Locals("arg_crypt") != nil {
				arg_crypt = c.Locals("arg_crypt").(string)
			}

			db, err := model.ConnectToSQLite()
		    if err != nil {
		        panic(err)
		    }
	*/

	return c.Render("view/spa", fiber.Map{

		"files_count_max":     100,
		"fieldSize_max":       14 * 1024 * 1024 * 1024,
		"fieldSize_max_human": "14 Gb",

		//"arg_upload_disable":      arg_upload_disable,
		//"arg_folder_make_disable": arg_folder_make_disable,
	})

}
