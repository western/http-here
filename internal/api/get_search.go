package api

import (
	_ "archive/zip"
	_ "bufio"
	_ "crypto/md5"
	_ "encoding/hex"
	_ "errors"
	_ "fmt"
	_ "html/template"
	_ "io"
	_ "log"
	_ "net/url"
	_ "os"
	_ "os/exec"
	_ "path/filepath"
	_ "regexp"
	_ "strconv"
	_ "strings"
	_ "time"

	"github.com/western/http-here/internal/model"

	"github.com/gofiber/fiber/v2"

	_ "github.com/edwvee/exiffix"
	_ "golang.org/x/image/draw"
	_ "image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	_ "math"
)

func GetSearch(c *fiber.Ctx) error {

	arg_fold := ""
	arg_fold = c.Locals("arg_fold").(string)

	db, err := model.ConnectToSQLite()
	if err != nil {
		panic(err)
	}

	/*
		referer := c.Get("Referer")

		// already decoded
		u, err := url.Parse(referer)
		if err != nil {

			LogPrefix(c, "500", "Error url parse "+referer+" "+err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "Error url parse " + referer,
			}, "application/json")
		}
	*/

	//u_path := CleanDirtyPath(u.Path)

	// -------------------------------------------------------------------------------------------------------------------------

	s := c.Query("s")

	// -------------------------------------------------------------------------------------------------------------------------

	/*
		name := c.FormValue("name")

		name = strings.ReplaceAll(name, "/", "")
		re := regexp.MustCompile("\\s+")
		name = re.ReplaceAllLiteralString(name, " ")

		name = CleanDirtyPath(name)
	*/

	result_list := model.FileSearchResult(db, arg_fold, s)

	LogPrefix(c, "200", "Get api search '"+s+"'")

	return c.JSON(fiber.Map{
		"code":        200,
		"result_list": result_list,
	}, "application/json")

}
