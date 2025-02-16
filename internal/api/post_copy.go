

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



func PostCopy(c *fiber.Ctx) error {

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

	from_path := c.FormValue("from_path")
	from_path = CleanDirtyPath(from_path)

	// -------------------------------------------------------------------------------------------------------------------------

	to := c.FormValue("to")
	to = CleanDirtyPath(to)

	if len(to) == 0 {
		LogPrefix(c, "500", "to is empty")
		return c.JSON(fiber.Map{
			"code": 500,
			"msg":  "to is empty",
		}, "application/json")
	}

	_, err = os.Stat(filepath.Join(arg_fold, to))
	if err != nil {
		LogPrefix(c, "500", "'"+filepath.Join(arg_fold, to)+"' not exists "+err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "'" + filepath.Join(arg_fold, to) + "' not exists",
		}, "application/json")
	}

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

			src_file_path := filepath.Join(arg_fold, u_path, name)
			if len(from_path) > 0 {
				src_file_path = filepath.Join(arg_fold, from_path, name)
			}

			target_file_path := filepath.Join(arg_fold, to, name)

			src_stat, err := os.Stat(src_file_path)
			if err != nil {
				LogPrefix(c, "500", "Source '"+src_file_path+"' not exists")
				continue
			}

			_, err = os.Stat(target_file_path)
			if err == nil {

				LogPrefix(c, "200", "Target '"+target_file_path+"' is exists. It will be rewrite.")

				if err := os.RemoveAll(target_file_path); err != nil {

					LogPrefix(c, "500", "'"+target_file_path+"' err "+err.Error())
					continue
				}
			}

			if src_stat.IsDir() {
				err := CopyDir(src_file_path, target_file_path)

				if err != nil {
					LogPrefix(c, "500", "CopyDir '"+src_file_path+"' to '"+target_file_path+"' err "+err.Error())
					continue
				}

				LogPrefix(c, "200", "Copy dir '"+src_file_path+"' to "+target_file_path)

			} else {
				err := CopyFile(src_file_path, target_file_path)

				if err != nil {
					LogPrefix(c, "500", "CopyFile '"+src_file_path+"' to '"+target_file_path+"' err "+err.Error())
					continue
				}

				LogPrefix(c, "200", "Copy file '"+src_file_path+"' to "+target_file_path)
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


