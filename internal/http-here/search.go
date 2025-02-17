package app

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

	"github.com/western/http-here/internal/model"

	"github.com/gofiber/fiber/v2"
)

func GetSearch(c *fiber.Ctx) error {

	db, err := model.ConnectToSQLite()
	if err != nil {
		panic(err)
	}

	arg_fold := ""
	arg_fold = c.Locals("arg_fold").(string)

	s := c.Query("s")

	result_list := model.FileSearchResult(db, arg_fold, s)

	//fmt.Println("result_list=", result_list[0])
	LogPrefix(c, "200", "Get search '"+s+"'")

	return c.Render("view/search", fiber.Map{

		"s":           s,
		"result_list": result_list,
	}, "view/layout/default")
}
