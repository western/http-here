package api

import (
	"archive/zip"
	"io"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/western/http-here/v2/internal/conf"
	"github.com/western/http-here/v2/internal/model2"
	"github.com/western/http-here/v2/internal/safefs"
	"github.com/western/http-here/v2/internal/util"

	"github.com/gofiber/fiber/v3"
)

func PostZip(c fiber.Ctx) error {

	if _, err := os.Stat(path.Join(conf.ConfigRoot, "temp")); err != nil {

		if err := os.MkdirAll(path.Join(conf.ConfigRoot, "temp"), os.ModePerm); err != nil {

			model2.EventLogAdd(c, 500, "PostZip", "mkdirall error: "+err.Error())

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "Error temp folder create",
			})
		}
	}

	// -------------------------------------------------------------------------------------------------------------------------

	url_Path, err := util.RefererOrPath(c)
	if err != nil {

		model2.EventLogAdd(c, 500, "PostZip", "RefererOrPath error: "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "RefererOrPath error",
		})
	}

	targetPath := path.Join(conf.ArgFold, url_Path)

	// -------------------------------------------------------------------------------------------------------------------------
	// check root with trust ArgFold

	Root := conf.ArgFoldRootFS

	if url_Path != "/" {

		fileInfo, err := Root.Stat(strings.TrimPrefix(url_Path, "/"))
		if err != nil {

			model2.EventLogAdd(c, 500, "PostZip", "stat err "+err.Error())

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "stat err",
			})
		}

		if !fileInfo.IsDir() {

			model2.EventLogAdd(c, 500, "PostZip", "It is not directory. "+targetPath)

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "It is not directory",
			})
		}
	}

	// -------------------------------------------------------------------------------------------------------------------------
	// reopen targetRoot with full targetPath

	targetRoot, err := safefs.NewSafeFS(targetPath)
	if err != nil {

		model2.EventLogAdd(c, 500, "PostZip", "open storage targetRoot err1. "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "open storage targetRoot err1",
		})
	}

	defer targetRoot.Close()

	// -------------------------------------------------------------------------------------------------------------------------

	archive_name := "archive-" + time.Now().Format("20060102-150405") + ".zip"

	zip_full_path := path.Join(conf.ConfigRoot, "temp", archive_name)

	if runtime.GOOS == "windows" {
		zip_full_path = util.RotateSlash(zip_full_path)
	}

	model2.EventLogAdd(c, 200, "PostZip", "Temp file prepare "+zip_full_path)

	// reopen tempRoot with full ConfigRoot/temp for archive
	/*
		tempRoot, err := safefs.NewSafeFS(path.Join(conf.ConfigRoot, "temp"))
		if err != nil {

			model2.EventLogAdd(c, 500, "PostZip", "open storage tempRoot err1. "+err.Error())

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "open storage tempRoot err1",
			})
		}

		defer tempRoot.Close()
	*/

	tempRoot := conf.TempRootFS

	archive, err := tempRoot.Create(archive_name)

	if err != nil {

		model2.EventLogAdd(c, 500, "PostZip", "create error: "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Create file archive error",
		})
	}
	defer archive.Close()

	zipWriter := zip.NewWriter(archive)

	form, _ := c.MultipartForm()

	if form == nil {

		model2.EventLogAdd(c, 500, "PostZip", "form is empty")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "form is empty",
		})
	}

	names := form.Value["name"]

	if len(names) == 0 {

		model2.EventLogAdd(c, 500, "PostZip", "form is empty")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "form is empty",
		})
	}

	for _, val := range names {

		name := strings.ReplaceAll(val, "/", "")

		name = util.CleanDirtyPath(name)

		if len(name) == 0 {

			model2.EventLogAdd(c, 500, "PostZip", "name is empty")
			continue
		}

		fileInfo, err := targetRoot.Stat(name)
		if err != nil {

			model2.EventLogAdd(c, 500, "PostZip", "'"+path.Join(targetPath, name)+"' not exists")
			continue
		}

		header, err := zip.FileInfoHeader(fileInfo)
		if err != nil {

			model2.EventLogAdd(c, 500, "PostZip", "'"+path.Join(targetPath, name)+"' err: "+err.Error())
			continue
		}
		header.Method = zip.Store

		if fileInfo.IsDir() {

			addFilesToZip(zipWriter, targetRoot, name, name)

		} else {

			f1, err := targetRoot.Open(name)
			if err != nil {

				model2.EventLogAdd(c, 500, "PostZip", "Read "+path.Join(targetPath, name)+" err: "+err.Error())
				continue
			}

			w1, err := zipWriter.CreateHeader(header)
			if err != nil {

				model2.EventLogAdd(c, 500, "PostZip", "CreateHeader err: "+err.Error())
				f1.Close()
				continue
			}

			if _, err := io.Copy(w1, f1); err != nil {

				model2.EventLogAdd(c, 500, "PostZip", "Copy "+path.Join(targetPath, name)+" err: "+err.Error())
			}

			f1.Close()
		}

	}

	zipWriter.Close()
	archive.Close()

	model2.EventLogAdd(c, 200, "PostZip", "Temp file created "+zip_full_path)

	return c.JSON(fiber.Map{
		"code": 200,
		"file": archive_name,
	})
}

func addFilesToZip(w *zip.Writer, targetRoot *safefs.RootFS, basePath, baseInZip string) {

	files, err := targetRoot.ReadDir(basePath)
	if err != nil {
		return
	}

	for _, file := range files {

		if !file.IsDir() {

			fileInfo, err := file.Info()
			if err != nil {
				continue
			}

			header, err := zip.FileInfoHeader(fileInfo)
			if err != nil {
				continue
			}
			header.Method = zip.Store
			header.Name = path.Join(baseInZip, file.Name())

			f, err := w.CreateHeader(header)
			if err != nil {
				continue
			}

			src, err := targetRoot.Open(path.Join(basePath, file.Name()))
			if err != nil {
				continue
			}

			if _, err := io.Copy(f, src); err != nil {
				src.Close()
				continue
			}

			src.Close()

		} else {

			newBase := path.Join(basePath, file.Name())

			addFilesToZip(w, targetRoot, newBase, path.Join(baseInZip, file.Name()))
		}
	}
}
