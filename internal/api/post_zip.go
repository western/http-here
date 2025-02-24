package api

import (
	"archive/zip"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/western/http-here/internal/model"
	"github.com/western/http-here/internal/util"

	"github.com/gofiber/fiber/v2"
)

func PostZip(c *fiber.Ctx) error {

	arg_fold := ""
	arg_fold = c.Locals("arg_fold").(string)

	db, err := model.ConnectToSQLite()
	if err != nil {
		panic(err)
	}

	homepath, err := os.UserHomeDir()
	if err != nil {
		//fmt.Println(err2)
		model.EventLogAdd(db, c, "500", "PostZip", "home detect error: "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Error homedir detect",
		}, "application/json")
	}

	if _, err := os.Stat(filepath.Join(homepath, ".httphere", "temp")); err != nil {

		if err := os.MkdirAll(filepath.Join(homepath, ".httphere", "temp"), os.ModePerm); err != nil {

			//fmt.Println(err4)
			model.EventLogAdd(db, c, "500", "PostZip", "mkdirall error: "+err.Error())

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

		//util.LogPrefix(c, "500", "Error url parse "+referer+" "+err.Error())
		model.EventLogAdd(db, c, "500", "PostZip", "Error url parse "+referer+" "+err.Error())

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

	archive, err := os.Create(filepath.Join(homepath, ".httphere", "temp", archive_name))

	if err != nil {
		//fmt.Println(err)
		model.EventLogAdd(db, c, "500", "PostZip", "create error: "+err.Error())

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

		//util.LogPrefix(c, "500", "form is empty")
		model.EventLogAdd(db, c, "500", "PostZip", "form is empty")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "form is empty",
		}, "application/json")
	}

	for _, val := range names {

		name := strings.ReplaceAll(val, "/", "")

		name = util.CleanDirtyPath(name)

		if len(name) == 0 {
			//util.LogPrefix(c, "500", "name is empty")
			model.EventLogAdd(db, c, "500", "PostZip", "name is empty")
			continue
		}

		fileInfo, err := os.Stat(filepath.Join(arg_fold, u_path, name))
		if err != nil {
			//util.LogPrefix(c, "500", "'"+filepath.Join(arg_fold, u_path, name)+"' not exists")
			model.EventLogAdd(db, c, "500", "PostZip", "'"+filepath.Join(arg_fold, u_path, name)+"' not exists")
			continue
		}

		header, err := zip.FileInfoHeader(fileInfo)
		if err != nil {

			//util.LogPrefix(c, "500", "'"+filepath.Join(arg_fold, u_path, name)+"' err: "+err.Error())
			model.EventLogAdd(db, c, "500", "PostZip", "'"+filepath.Join(arg_fold, u_path, name)+"' err: "+err.Error())
			continue
		}
		header.Method = zip.Store

		if fileInfo.IsDir() {

			addFilesToZip(zipWriter, filepath.Join(arg_fold, u_path, name), name)

		} else {

			f1, err := os.Open(filepath.Join(arg_fold, u_path, name))
			if err != nil {
				panic(err)
			}
			defer f1.Close()

			//w1, err := zipWriter.Create(name)
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

	//return c.SendFile(filepath.Join(arg_fold, u_path, "archive.zip"), false)
	//return c.Download(filepath.Join(arg_fold, u_path, "archive.zip"), "archive.zip");

	//util.LogPrefix(c, "200", "Temp file create "+filepath.Join(homepath, ".httphere", "temp", archive_name))
	model.EventLogAdd(db, c, "200", "PostZip", "Temp file create "+filepath.Join(homepath, ".httphere", "temp", archive_name))

	return c.JSON(fiber.Map{
		"code": 200,
		"file": filepath.Join("/__temp/", archive_name),
	}, "application/json")

}

func addFilesToZip(w *zip.Writer, basePath, baseInZip string) {
	// Open the Directory
	//files, err := ioutil.ReadDir(basePath)
	files, err := os.ReadDir(basePath)
	if err != nil {
		fmt.Println(err)
	}

	for _, file := range files {
		//fmt.Println(  filepath.Join(basePath, file.Name())   )
		if !file.IsDir() {
			//dat, err := ioutil.ReadFile(basePath + file.Name())
			dat, err := os.ReadFile(filepath.Join(basePath, file.Name()))
			if err != nil {
				fmt.Println(err)
			}

			fileInfo, err := os.Stat(filepath.Join(basePath, file.Name()))
			if err != nil {
				fmt.Println(err)
				//LogPrefix(c, "500", "'"+filepath.Join(basePath, file.Name())+"' not exists")
				//continue
			}

			header, err := zip.FileInfoHeader(fileInfo)
			if err != nil {
				fmt.Println(err)
				//LogPrefix(c, "500", "'"+filepath.Join(arg_fold, u_path, name)+"' err: "+err.Error())
				//continue
			}
			header.Method = zip.Store
			header.Name = filepath.Join(baseInZip, file.Name())

			// Add some files to the archive.
			//f, err := w.Create(filepath.Join(baseInZip, file.Name()))
			f, err := w.CreateHeader(header)
			if err != nil {
				fmt.Println(err)
			}
			_, err = f.Write(dat)
			if err != nil {
				fmt.Println(err)
			}
		} else if file.IsDir() {

			// Recurse
			//newBase := basePath + file.Name() + "/"
			newBase := filepath.Join(basePath, file.Name()) + "/"
			//fmt.Println("Recursing and Adding SubDir: " + file.Name())
			//fmt.Println("Recursing and Adding SubDir: " + newBase)

			//addFiles(w, newBase, baseInZip+file.Name()+"/")
			addFilesToZip(w, newBase, filepath.Join(baseInZip, file.Name())+"/")
		}
	}
}
