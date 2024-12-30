package controller

import (
	_ "errors"
	"fmt"

	"archive/zip"
	"path/filepath"
	"regexp"
	_ "strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	_ "bufio"
	"os"
	_ "os/exec"

	"io"
	_ "log"
	"net/url"
	"time"

	_ "crypto/md5"
	_ "encoding/hex"

	_ "github.com/edwvee/exiffix"
	_ "golang.org/x/image/draw"
	_ "image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	_ "math"
)

func PostUpload(c *fiber.Ctx) error {

	arg_fold := ""
	arg_fold = c.Locals("arg_fold").(string)

	arg_crypt := ""
	if c.Locals("arg_crypt") != nil {
		arg_crypt = c.Locals("arg_crypt").(string)
	}

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

	form, _ := c.MultipartForm()
	files := form.File["fileBlob"]

	for _, file := range files {

		file_ext := GetExtNorm(file.Filename)
		originalFileName := GetFileName(file.Filename)

		originalFileName = strings.ReplaceAll(originalFileName, "/", "")
		re := regexp.MustCompile("\\s+")
		originalFileName = re.ReplaceAllLiteralString(originalFileName, "-")
		re = regexp.MustCompile("[\\-]{2,}")
		originalFileName = re.ReplaceAllLiteralString(originalFileName, "-")

		filename := originalFileName + "." + file_ext
		filename = CleanDirtyPath(filename)

		// -------------------------------------------------------------------------------------------------------------------------

		if fileInfo, err := os.Stat(filepath.Join(arg_fold, u_path, filename)); err == nil {

			if !fileInfo.IsDir() {
				LogPrefix(c, "200", "'"+filepath.Join(arg_fold, u_path, filename)+"' already exists. It will be rewrite.")
			}
		}

		// -------------------------------------------------------------------------------------------------------------------------

		code := c.Cookies("code")

		if arg_crypt == "" && len(code) > 0 {

			cookie := new(fiber.Cookie)
			cookie.Name = "code"
			c.Cookie(cookie)
		}

		if arg_crypt == "1" && len(code) > 0 {

			//panic(code)

			f, err := os.CreateTemp("", "httphere_crypt*")
			if err != nil {
				panic(err)
			}
			//fmt.Println("Crypt Temp file name:", f.Name())
			defer os.Remove(f.Name())

			readerFile, _ := file.Open()
			_, err = io.Copy(f, readerFile)
			if err != nil {
				panic(err)
			}
			f.Close()

			isOk, err := CryptFile(f.Name(), code)
			if !isOk {

				LogPrefix(c, "500", "Error CryptFile "+err.Error())
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code": 500,
					"msg":  "Error CryptFile",
				}, "application/json")
			}

			out, err := os.Create(filepath.Join(arg_fold, u_path, filename+".crypt"))
			if err != nil {

				LogPrefix(c, "500", "Error create "+filepath.Join(arg_fold, u_path, filename+".crypt")+" "+err.Error())
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code": 500,
					"msg":  "Error create " + filepath.Join(arg_fold, u_path, filename+".crypt"),
				}, "application/json")
			}
			defer out.Close()

			readerFile, _ = os.Open(f.Name())

			_, err = io.Copy(out, readerFile)
			if err != nil {
				panic(err)
			}
			f.Close()

			LogPrefix(c, "200", "Save encrypted '"+filepath.Join(arg_fold, u_path, filename+".crypt")+"'")

		} else {

			out, err := os.Create(filepath.Join(arg_fold, u_path, filename))
			if err != nil {

				LogPrefix(c, "500", "Error create "+filepath.Join(arg_fold, u_path, filename)+" "+err.Error())
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code": 500,
					"msg":  "Error create " + filepath.Join(arg_fold, u_path, filename),
				}, "application/json")
			}
			defer out.Close()

			readerFile, _ := file.Open()
			_, err = io.Copy(out, readerFile)
			if err != nil {

				LogPrefix(c, "500", "Error copy "+filepath.Join(arg_fold, u_path, filename)+" "+err.Error())
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code": 500,
					"msg":  "Error copy " + filepath.Join(arg_fold, u_path, filename),
				}, "application/json")
			}

			LogPrefix(c, "200", "Save '"+filepath.Join(arg_fold, u_path, filename)+"'")

		}

		/*
			out, err := os.Create(filepath.Join(arg_fold, u_path, filename))
			if err != nil {

				LogPrefix(c, "500", "Error create "+filepath.Join(arg_fold, u_path, filename)+" "+err.Error())
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code": 500,
					"msg":  "Error create " + filepath.Join(arg_fold, u_path, filename),
				}, "application/json")
			}
			defer out.Close()

			LogPrefix(c, "200", "Save '"+filepath.Join(arg_fold, u_path, filename)+"'")

			readerFile, _ := file.Open()
			_, err = io.Copy(out, readerFile)
			if err != nil {

				LogPrefix(c, "500", "Error copy "+filepath.Join(arg_fold, u_path, filename)+" "+err.Error())
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code": 500,
					"msg":  "Error copy " + filepath.Join(arg_fold, u_path, filename),
				}, "application/json")
			}
		*/

	}

	return c.JSON(fiber.Map{
		"code": 200,
	}, "application/json")
}

func PostFolder(c *fiber.Ctx) error {

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

	name := c.FormValue("name")

	name = strings.ReplaceAll(name, "/", "")
	re := regexp.MustCompile("\\s+")
	name = re.ReplaceAllLiteralString(name, " ")

	name = CleanDirtyPath(name)

	if len(name) == 0 {
		LogPrefix(c, "500", "name is empty")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "name is empty",
		}, "application/json")
	}

	if fileInfo, err := os.Stat(filepath.Join(arg_fold, u_path, name)); err == nil {

		if fileInfo.IsDir() {
			LogPrefix(c, "500", "'"+filepath.Join(arg_fold, u_path, name)+"' already exists")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  filepath.Join(u_path, name) + " already exists",
			}, "application/json")
		}
	}

	if err := os.Mkdir(filepath.Join(arg_fold, u_path, name), os.ModePerm); err != nil {

		LogPrefix(c, "500", "Error mkdir "+filepath.Join(arg_fold, u_path, name)+" "+err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Error mkdir " + filepath.Join(arg_fold, u_path, name),
		}, "application/json")
	}

	LogPrefix(c, "200", "Mkdir '"+filepath.Join(arg_fold, u_path, name)+"'")

	return c.JSON(fiber.Map{
		"code": 200,
	}, "application/json")
}

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

	return c.JSON(fiber.Map{
		"code": 200,
	}, "application/json")

}

func PostMove(c *fiber.Ctx) error {

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
			target_file_path := filepath.Join(arg_fold, to, name)

			_, err := os.Stat(src_file_path)
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

			err = os.Rename(src_file_path, target_file_path)

			if err != nil {

				LogPrefix(c, "500", "Rename error "+fmt.Sprintf("%s", err))

			} else {

				LogPrefix(c, "200", "Move '"+src_file_path+"' to "+target_file_path)
			}

		}
	}

	return c.JSON(fiber.Map{
		"code": 200,
	}, "application/json")

}

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

			/*
				            out, err := os.Create(target_file_path)
							if err != nil {

								LogPrefix(c, "500", "Error create "+target_file_path+" "+err.Error())
								return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
									"code": 500,
									"msg":  "Error create " + target_file_path,
								}, "application/json")
							}
							defer out.Close()

							readerFile, _ := os.Open(src_file_path)

							_, err = io.Copy(out, readerFile)
							if err != nil {
								panic(err)
							}
							out.Close()
							readerFile.Close()
			*/

		}
	}

	return c.JSON(fiber.Map{
		"code": 200,
	}, "application/json")

}

func PostRename(c *fiber.Ctx) error {

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
	if err == nil {
		LogPrefix(c, "500", "'"+filepath.Join(arg_fold, to)+"' already exists ")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "'" + filepath.Join(arg_fold, to) + "' already exists",
		}, "application/json")
	}

	name := c.FormValue("name")
	name = CleanDirtyPath(name)

	if len(name) == 0 {
		LogPrefix(c, "500", "name is empty")
		return c.JSON(fiber.Map{
			"code": 500,
			"msg":  "name is empty",
		}, "application/json")
	}

	_, err = os.Stat(filepath.Join(arg_fold, u_path, name))
	if err != nil {
		LogPrefix(c, "500", "'"+filepath.Join(arg_fold, u_path, name)+"' not exists "+err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "'" + filepath.Join(arg_fold, u_path, name) + "' not exists",
		}, "application/json")
	}

	src_file_path := filepath.Join(arg_fold, u_path, name)
	target_file_path := filepath.Join(arg_fold, u_path, to)

	err = os.Rename(src_file_path, target_file_path)

	if err != nil {

		LogPrefix(c, "500", "Rename error "+err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Rename error",
		}, "application/json")

	} else {

		LogPrefix(c, "200", "Rename '"+src_file_path+"' to "+target_file_path)
	}

	return c.JSON(fiber.Map{
		"code": 200,
	}, "application/json")

}

func PostZip(c *fiber.Ctx) error {

	arg_fold := ""
	arg_fold = c.Locals("arg_fold").(string)

	homepath, err2 := os.UserHomeDir()
	if err2 != nil {
		fmt.Println(err2)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Error homedir detect",
		}, "application/json")
	}

	if _, err3 := os.Stat(filepath.Join(homepath, ".httphere", "temp")); err3 != nil {

		if err4 := os.MkdirAll(filepath.Join(homepath, ".httphere", "temp"), os.ModePerm); err4 != nil {
			fmt.Println(err4)
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

		LogPrefix(c, "500", "Error url parse "+referer+" "+err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Error url parse " + referer,
		}, "application/json")
	}

	u_path := CleanDirtyPath(u.Path)

	archive_name := "archive-" + time.Now().Format("20060102-150405") + ".zip"

	archive, err := os.Create(filepath.Join(homepath, ".httphere", "temp", archive_name))

	if err != nil {
		fmt.Println(err)
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

		LogPrefix(c, "500", "form is empty")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "form is empty",
		}, "application/json")
	}

	for _, val := range names {

		name := strings.ReplaceAll(val, "/", "")

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

		header, err := zip.FileInfoHeader(fileInfo)
		if err != nil {

			LogPrefix(c, "500", "'"+filepath.Join(arg_fold, u_path, name)+"' err: "+err.Error())
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

	LogPrefix(c, "200", "Temp file create "+filepath.Join(homepath, ".httphere", "temp", archive_name))

	return c.JSON(fiber.Map{
		"code": 200,
		"file": filepath.Join("/__temp/", archive_name),
	}, "application/json")

}
