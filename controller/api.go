package controller

import (
	"errors"
	"fmt"

	"archive/zip"
	"path/filepath"
	"regexp"
	_ "strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"os"

	"io"
	"log"
	"net/url"
	"time"

	"crypto/md5"
	"encoding/hex"

	"github.com/edwvee/exiffix"
	"golang.org/x/image/draw"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"math"
)

func PostUpload(c *fiber.Ctx) error {

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
	files := form.File["fileBlob"]

	for _, file := range files {

		file_ext := GetExtNorm(file.Filename)
		originalFileName := GetFileName(file.Filename)

		originalFileName = strings.ReplaceAll(originalFileName, "/", "")
		re := regexp.MustCompile("\\s+")
		originalFileName = re.ReplaceAllLiteralString(originalFileName, "-")

		filename := originalFileName + "." + file_ext
		filename = CleanDirtyPath(filename)

		if fileInfo, err := os.Stat(filepath.Join(arg_fold, u_path, filename)); err == nil {

			if !fileInfo.IsDir() {
				LogPrefix(c, "200", "'"+filepath.Join(arg_fold, u_path, filename)+"' already exists. It will be rewrite.")
			}
		}

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

			_, err := os.Stat(filepath.Join(arg_fold, u_path, name))
			if err != nil {
				LogPrefix(c, "500", "'"+filepath.Join(arg_fold, u_path, name)+"' not exists")
				continue
			}

			err = os.Rename(filepath.Join(arg_fold, u_path, name), filepath.Join(arg_fold, to, name))

			if err != nil {
				//fmt.Errorf("something %s", foo)
				LogPrefix(c, "500", "Rename error "+fmt.Sprintf("%s", err))

			} else {

				LogPrefix(c, "200", "Move '"+filepath.Join(arg_fold, u_path, name)+"' to "+filepath.Join(arg_fold, to, name))
			}

		}
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

		if fileInfo.IsDir() {

			addFilesToZip(zipWriter, filepath.Join(arg_fold, u_path, name), name)

		} else {

			f1, err := os.Open(filepath.Join(arg_fold, u_path, name))
			if err != nil {
				panic(err)
			}
			defer f1.Close()

			w1, err := zipWriter.Create(name)
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

func GetResize(c *fiber.Ctx) error {

	arg_fold := ""
	arg_fold = c.Locals("arg_fold").(string)

	homepath, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Error homedir detect",
		}, "application/json")
	}

	if _, err := os.Stat(filepath.Join(homepath, ".httphere", "thumb")); err != nil {

		if err2 := os.MkdirAll(filepath.Join(homepath, ".httphere", "thumb"), os.ModePerm); err2 != nil {
			fmt.Println(err2)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "Error thumbnails folder create",
			}, "application/json")
		}
	}

	c_path, err := url.QueryUnescape(c.Path())
	if err != nil {

		LogPrefix(c, "500", "Error "+filepath.Join(arg_fold, c_path)+" "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Error " + filepath.Join(arg_fold, c_path),
		}, "application/json")
	}

	c_path = CleanDirtyPath(c_path)
	c_path = strings.Replace(c_path, "/__resize", "", 1)

	c_width := "600"

	i_width := 600

	modtime_human := ""
	size_human := ""

	if fileInfo, err := os.Stat(filepath.Join(arg_fold, c_path)); err == nil {

		modtime := fileInfo.ModTime()
		modtime_human = modtime.Format("2006-01-02 15:04:05")

		size := fileInfo.Size()
		size_human = PrettyByteSize(size)

		if fileInfo.IsDir() {

			LogPrefix(c, "500", "Error "+filepath.Join(arg_fold, c_path)+" It is a folder")

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"file": filepath.Join("/__resize/", c_path) + " It is a folder",
				"msg":  "It is a folder",
			}, "application/json")
		}

	} else if errors.Is(err, os.ErrNotExist) {

		LogPrefix(c, "404", filepath.Join(arg_fold, c_path))

		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code": 404,
			"file": filepath.Join("/__resize/", c_path),
			"msg":  "Not found",
		}, "application/json")
	}

	file_ext := GetExtNorm(c_path)
	orig_filename := GetFileName(c_path)

	is_match, _ := regexp.MatchString("^(jpg|jpeg|png|gif)$", file_ext)
	if !is_match {
		LogPrefix(c, "500", filepath.Join("/__resize/", c_path)+" Only for JPEG, PNG and GIF images")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"file": filepath.Join("/__resize/", c_path),
			"msg":  "Only for JPEG, PNG and GIF images",
		}, "application/json")
	}

	hash_name := md5.Sum([]byte(orig_filename + modtime_human + size_human + c_width))
	hex_name := hex.EncodeToString(hash_name[:])

	if _, err := os.Stat(filepath.Join(homepath, ".httphere", "thumb", hex_name)); err == nil {

		LogPrefix(c, "200", "SendFile from cache "+filepath.Join(c_path))
		return c.SendFile(filepath.Join(homepath, ".httphere", "thumb", hex_name), false)

	} else if errors.Is(err, os.ErrNotExist) {

		input, _ := os.Open(filepath.Join(arg_fold, c_path))
		defer input.Close()

		output, _ := os.Create(filepath.Join(homepath, ".httphere", "thumb", hex_name))
		defer output.Close()

		var src image.Image

		// Decode the image (from PNG to image.Image):
		if file_ext == "png" {
			src, err = png.Decode(input)
			if err != nil {
				log.Fatal(err)
			}
		}

		if file_ext == "jpg" {

			// src, err = jpeg.Decode(input)
			src, _, err = exiffix.Decode(input)
			if err != nil {
				log.Fatal(err)
			}
		}

		if file_ext == "gif" {
			src, err = gif.Decode(input)
			if err != nil {
				log.Fatal(err)
			}
		}

		ratio := (float64)(src.Bounds().Max.Y) / (float64)(src.Bounds().Max.X)
		i_height := int(math.Round(float64(i_width) * ratio))

		var dst *image.RGBA

		if src.Bounds().Max.X > i_width || src.Bounds().Max.Y > i_height {
			dst = image.NewRGBA(image.Rect(0, 0, i_width, i_height))
		} else {

			LogPrefix(c, "200", "SendFile original without resize "+filepath.Join(arg_fold, c_path))

			err := os.Remove(filepath.Join(homepath, ".httphere", "thumb", hex_name))
			if err != nil {
				log.Fatal(err)
			}

			return c.SendFile(filepath.Join(arg_fold, c_path), false)
		}

		// Resize:
		draw.NearestNeighbor.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)

		if file_ext == "png" {
			err = png.Encode(output, dst)
			if err != nil {
				log.Fatal(err)
			}
		}

		if file_ext == "jpg" {
			err = jpeg.Encode(output, dst, nil)
			if err != nil {
				log.Fatal(err)
			}
		}

		if file_ext == "gif" {
			err = gif.Encode(output, dst, nil)
			if err != nil {
				log.Fatal(err)
			}
		}

		output.Close()
		input.Close()

		src = nil
		dst = nil

		LogPrefix(c, "200", "Resize and SendFile "+filepath.Join(c_path))

		return c.SendFile(filepath.Join(homepath, ".httphere", "thumb", hex_name), false)
	}

	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"code": 400,
		"file": filepath.Join(c_path),
		"msg":  "Bad request",
	}, "application/json")

}
