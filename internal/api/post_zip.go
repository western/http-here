package api

import (
	"archive/zip"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/western/http-here/v2/internal/conf"
	"github.com/western/http-here/v2/internal/model2"
	"github.com/western/http-here/v2/internal/util"

	"github.com/gofiber/fiber/v2"
)

func PostZip(c *fiber.Ctx) error {

	if _, err := os.Stat(path.Join(conf.ConfigRoot, "temp")); err != nil {

		if err := os.MkdirAll(path.Join(conf.ConfigRoot, "temp"), os.ModePerm); err != nil {

			model2.EventLogAdd(c, 500, "PostZip", "mkdirall error: "+err.Error())

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "Error temp folder create",
			}, "application/json")
		}
	}

	referer := c.Get("Referer")

	// already decoded
	u, err := url.Parse(referer)
	if err != nil {

		model2.EventLogAdd(c, 500, "PostZip", "Error url parse "+referer+" "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Error url parse " + referer,
		}, "application/json")
	}

	u_path := util.CleanDirtyPath(u.Path)

	// -------------------------------------------------------------------------------------------------------------------------

	form_path := c.FormValue("path")
	form_path = util.CleanDirtyPath(form_path)
	if len(form_path) > 0 {
		u_path = form_path
	}

	// -------------------------------------------------------------------------------------------------------------------------

	archive_name := "archive-" + time.Now().Format("20060102-150405") + ".zip"

	zip_full_path := path.Join(conf.ConfigRoot, "temp", archive_name)

	if runtime.GOOS == "windows" {
		zip_full_path = util.RotateSlash(zip_full_path)
	}

	model2.EventLogAdd(c, 200, "PostZip", "Temp file prepare "+zip_full_path)

	archive, err := os.Create(zip_full_path)

	if err != nil {

		model2.EventLogAdd(c, 500, "PostZip", "create error: "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Create file archive error",
		}, "application/json")
	}
	defer archive.Close()

	zipWriter := zip.NewWriter(archive)

	form, _ := c.MultipartForm()
	names := form.Value["name"]

	if len(names) == 0 {

		model2.EventLogAdd(c, 500, "PostZip", "form is empty")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "form is empty",
		}, "application/json")
	}

	for _, val := range names {

		name := strings.ReplaceAll(val, "/", "")

		name = util.CleanDirtyPath(name)

		if len(name) == 0 {

			model2.EventLogAdd(c, 500, "PostZip", "name is empty")
			continue
		}

		readTarget := path.Join(conf.ArgFold, u_path, name)

		fileInfo, err := os.Stat(readTarget)
		if err != nil {

			model2.EventLogAdd(c, 500, "PostZip", "'"+readTarget+"' not exists")
			continue
		}

		header, err := zip.FileInfoHeader(fileInfo)
		if err != nil {

			model2.EventLogAdd(c, 500, "PostZip", "'"+readTarget+"' err: "+err.Error())
			continue
		}
		header.Method = zip.Store

		if fileInfo.IsDir() {

			addFilesToZip(zipWriter, readTarget, name)

		} else {

			f1, err := os.Open(readTarget)
			if err != nil {
				//panic(err)

				if os.IsPermission(err) {

					model2.EventLogAdd(c, 403, "PostZip", "Forbidden for read "+readTarget)
					continue
				}

				model2.EventLogAdd(c, 500, "PostZip", "Read "+readTarget+" err: "+err.Error())
				continue
			}
			defer f1.Close()

			w1, err := zipWriter.CreateHeader(header)
			if err != nil {
				panic(err)
			}
			if _, err := io.Copy(w1, f1); err != nil {
				panic(err)
			}

		}

	}

	zipWriter.Close()
	archive.Close()

	//return c.SendFile(path.Join(conf.ArgFold, u_path, "archive.zip"), false)
	//return c.Download(path.Join(conf.ArgFold, u_path, "archive.zip"), "archive.zip");

	model2.EventLogAdd(c, 200, "PostZip", "Temp file created "+zip_full_path)

	return c.JSON(fiber.Map{
		"code": 200,
		"file": archive_name,
	}, "application/json")

}

func addFilesToZip(w *zip.Writer, basePath, baseInZip string) {

	files, err := os.ReadDir(basePath)
	if err != nil {
		fmt.Println(err)
	}

	for _, file := range files {

		if !file.IsDir() {

			fullPath := path.Join(basePath, file.Name())

			fileInfo, err := os.Stat(fullPath)
			if err != nil {
				fmt.Println(err)
				continue
			}

			header, err := zip.FileInfoHeader(fileInfo)
			if err != nil {
				fmt.Println(err)
				continue
			}
			header.Method = zip.Store
			header.Name = path.Join(baseInZip, file.Name())

			f, err := w.CreateHeader(header)
			if err != nil {
				fmt.Println(err)
				continue
			}

			src, err := os.Open(fullPath)
			defer src.Close()
			if err != nil {
				fmt.Println(err)
				continue
			} else {
				_, err = io.Copy(f, src)
				if err != nil {
					fmt.Println(err)
					continue
				}

			}

		} else if file.IsDir() {

			newBase := path.Join(basePath, file.Name()) + "/"

			addFilesToZip(w, newBase, path.Join(baseInZip, file.Name())+"/")
		}
	}
}
