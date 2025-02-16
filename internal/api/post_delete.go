

package api

import (
	_ "archive/zip"
	_ "bufio"
	_ "crypto/md5"
	_ "encoding/hex"
	_ "errors"
	_ "fmt"
	_ "io"
	_ "log"
	"net/url"
	"os"
	_ "os/exec"
	"path/filepath"
	_ "regexp"
	_ "strconv"
	"strings"
	_ "time"
	_ "html/template"

	
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



func PostDelete(c *fiber.Ctx) error {

	arg_fold := ""
	arg_fold = c.Locals("arg_fold").(string)

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

	u_path := CleanDirtyPath(u.Path)

	// -------------------------------------------------------------------------------------------------------------------------

	form_path := c.FormValue("path")
	form_path = CleanDirtyPath(form_path)
	if len(form_path) > 0 {
		u_path = form_path
	}

	// -------------------------------------------------------------------------------------------------------------------------

	form, _ := c.MultipartForm()
	names := form.Value["name"]

	if len(names) == 0 {

		LogPrefix(c, "500", "form is empty")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "form is empty",
		}, "application/json")
	}

	for _, val := range names {

		if len(val) > 0 {
			//fmt.Println("val="+val)

			name := strings.ReplaceAll(val, "/", "")
			//re := regexp.MustCompile("\\s+")
			//name = re.ReplaceAllLiteralString(name, " ")

			name = CleanDirtyPath(name)

			if len(name) == 0 {
				LogPrefix(c, "500", "name is empty")
				continue
			}

			fileInfo, err := os.Stat(filepath.Join(arg_fold, u_path, name))

			if err != nil {
				LogPrefix(c, "500", "'"+filepath.Join(arg_fold, u_path, name)+"' not exists")
				continue
			}

			if fileInfo.IsDir() {

				// remove fold and all inside data

				if err := os.RemoveAll(filepath.Join(arg_fold, u_path, name)); err != nil {

					LogPrefix(c, "500", "'"+filepath.Join(arg_fold, u_path, name)+"' err "+err.Error())
					continue
				}

				LogPrefix(c, "200", "Remove fold '"+filepath.Join(arg_fold, u_path, name)+"'")

			} else {

				// remove one file

				if err := os.Remove(filepath.Join(arg_fold, u_path, name)); err != nil {

					LogPrefix(c, "500", "'"+filepath.Join(arg_fold, u_path, name)+"' err "+err.Error())
					continue
				}

				LogPrefix(c, "200", "Remove '"+filepath.Join(arg_fold, u_path, name)+"'")

			}

		}
	}

	db, err := model.ConnectToSQLite()
	if err != nil {
		panic(err)
	}

	go model.FileChkAsync(db)

	return c.JSON(fiber.Map{
		"code": 200,
	}, "application/json")

}


