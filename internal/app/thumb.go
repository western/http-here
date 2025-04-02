package app

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	_ "log"
	"net/url"
	"os"
	"os/exec"
	"path"
	_ "path/filepath"
	"regexp"
	"runtime"
	_ "strconv"
	"strings"
	_ "syscall"
	"time"

	"github.com/western/http-here/internal/model"
	"github.com/western/http-here/internal/util"

	"github.com/gofiber/fiber/v2"

	"github.com/edwvee/exiffix"
	"golang.org/x/image/draw"
	"gorm.io/gorm"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"math"
)

func GetThumb(c *fiber.Ctx) error {

	arg_fold := ""
	arg_fold = c.Locals("arg_fold").(string)

	prefix := c.Locals("prefix").(string)

	db := c.Locals("db").(*gorm.DB)

	if _, err := os.Stat(path.Join(prefix, "thumb")); err != nil {

		if err := os.MkdirAll(path.Join(prefix, "thumb"), os.ModePerm); err != nil {

			model.EventLogAdd(db, c, "500", "GetThumb", "MkdirAll error "+err.Error())

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "Error thumbnails folder create",
			}, "application/json")
		}
	}

	c_path, err := url.QueryUnescape(c.Path())
	if err != nil {

		model.EventLogAdd(db, c, "500", "GetThumb", "Error "+path.Join(arg_fold, c_path)+" "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Error " + path.Join(arg_fold, c_path),
		}, "application/json")
	}

	c_path = util.CleanDirtyPath(c_path)
	c_path = strings.Replace(c_path, "/__thumb", "", 1)

	// ------------------------------------------------------------------------------------------------------------------------------

	i_width := 600

	fileInfo, err := os.Stat(path.Join(arg_fold, c_path))

	if err == nil {

		if fileInfo.IsDir() {

			model.EventLogAdd(db, c, "500", "GetThumb", "Error "+path.Join(arg_fold, c_path)+" It is a folder")

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"file": path.Join("/__thumb/", c_path) + " It is a folder",
				"msg":  "It is a folder",
			}, "application/json")
		}

	} else if errors.Is(err, os.ErrNotExist) {

		model.EventLogAdd(db, c, "404", "GetThumb", path.Join(arg_fold, c_path))

		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code": 404,
			"file": path.Join("/__thumb/", c_path),
			"msg":  "Not found",
		}, "application/json")
	}

	file_ext := util.GetExtNorm(c_path)
	orig_filename := util.GetFileName(c_path)

	is_match, _ := regexp.MatchString("^(jpg|jpeg|png|gif|pdf|rtf|doc|docx|xls|xlsx|odt|ods)$", file_ext)
	if !is_match {

		model.EventLogAdd(db, c, "500", "GetThumb", path.Join("/__thumb/", c_path)+" Only for JPEG, PNG, GIF and office files")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"file": path.Join("/__thumb/", c_path),
			"msg":  "Only for JPEG, PNG, GIF and office files",
		}, "application/json")
	}

	// ------------------------------------------------------------------------------------------------------------------------------

	//hash_name := md5.Sum([]byte(orig_filename + modtime_human + size_human + c_width))
	//hex_name := hex.EncodeToString(hash_name[:])
	hex_name := ""

	if db != nil {

		var row model.File
		result := db.Select("md5, full_path").Where("full_path = ?", path.Join(arg_fold, c_path)).First(&row)

		if result.Error == nil {

			hex_name = row.MD5

			if len(hex_name) == 0 {

				model.EventLogAdd(db, c, "500", "GetThumb", path.Join("/__thumb/", c_path)+", hex_name is empty")

				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code": 500,
					"file": path.Join("/__thumb/", c_path),
				}, "application/json")
			}

			if _, err := os.Stat(path.Join(prefix, "thumb", hex_name)); err == nil {

				model.EventLogAdd(db, c, "200", "GetThumb", "SendFile db thumb/cache "+path.Join(c_path))
				return c.SendFile(path.Join(prefix, "thumb", hex_name), false)
			}
		}
	}

	hex_name = util.GetMd5File(path.Join(arg_fold, c_path))

	if len(hex_name) == 0 {

		model.EventLogAdd(db, c, "500", "GetThumb", path.Join("/__thumb/", c_path)+", hex_name is empty (2)")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"file": path.Join("/__thumb/", c_path),
		}, "application/json")
	}

	// ------------------------------------------------------------------------------------------------------------------------------

	is_img_match, _ := regexp.MatchString("^(jpg|jpeg|png|gif)$", file_ext)

	if is_img_match {

		if _, err := os.Stat(path.Join(prefix, "thumb", hex_name)); err == nil {

			model.EventLogAdd(db, c, "200", "GetThumb", "SendFile thumb/cache "+path.Join(c_path))

			return c.SendFile(path.Join(prefix, "thumb", hex_name), false)

		} else if errors.Is(err, os.ErrNotExist) {

			// ------------------------------------------------------------------------------------------------------------------------------

			if fileInfo.Size() < 50*1024 {

				model.EventLogAdd(db, c, "200", "GetThumb", "SendFile orig without resize "+path.Join(c_path)+", size is small")

				return c.SendFile(path.Join(arg_fold, c_path), false)
			}

			// ------------------------------------------------------------------------------------------------------------------------------

			input, _ := os.Open(path.Join(arg_fold, c_path))
			defer input.Close()

			output, _ := os.Create(path.Join(prefix, "thumb", hex_name))
			defer output.Close()

			var src image.Image

			// Decode the image (from PNG to image.Image):
			if file_ext == "png" {
				src, err = png.Decode(input)
				if err != nil {
					panic(err)
				}
			}

			if file_ext == "jpg" {

				// src, err = jpeg.Decode(input)
				src, _, err = exiffix.Decode(input)
				if err != nil {
					panic(err)
				}
			}

			if file_ext == "gif" {
				src, err = gif.Decode(input)
				if err != nil {
					panic(err)
				}
			}

			ratio := (float64)(src.Bounds().Max.Y) / (float64)(src.Bounds().Max.X)
			i_height := int(math.Round(float64(i_width) * ratio))

			var dst *image.RGBA

			if src.Bounds().Max.X > i_width || src.Bounds().Max.Y > i_height {
				dst = image.NewRGBA(image.Rect(0, 0, i_width, i_height))
			} else {

				model.EventLogAdd(db, c, "200", "GetThumb", "SendFile orig without resize "+path.Join(arg_fold, c_path))

				err := os.Remove(path.Join(prefix, "thumb", hex_name))
				if err != nil {
					panic(err)
				}

				return c.SendFile(path.Join(arg_fold, c_path), false)
			}

			// Resize:
			draw.NearestNeighbor.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)

			if file_ext == "png" {
				err = png.Encode(output, dst)
				if err != nil {
					panic(err)
				}
			}

			if file_ext == "jpg" {
				err = jpeg.Encode(output, dst, nil)
				if err != nil {
					panic(err)
				}
			}

			if file_ext == "gif" {
				err = gif.Encode(output, dst, nil)
				if err != nil {
					panic(err)
				}
			}

			output.Close()
			input.Close()

			src = nil
			dst = nil

			model.EventLogAdd(db, c, "200", "GetThumb", "Resize and SendFile "+path.Join(c_path))

			return c.SendFile(path.Join(prefix, "thumb", hex_name), false)
		}

	}

	is_office_match, _ := regexp.MatchString("^(pdf|rtf|doc|docx|xls|xlsx|odt|ods)$", file_ext)

	if is_office_match {

		if runtime.GOOS == "windows" {

			_, err := os.Stat("C:\\Program Files\\LibreOffice\\program\\soffice.exe")
			if err != nil {

				model.EventLogAdd(db, c, "500", "GetThumb", "Error libreoffice not found "+err.Error())

				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code": 500,
					"msg":  "Error libreoffice not found",
				}, "application/json")
			}

		} else {

			_, err := exec.Command("bash", "-c", "libreoffice --help").Output()
			if err != nil {

				model.EventLogAdd(db, c, "500", "GetThumb", "Error libreoffice not found "+err.Error())

				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code": 500,
					"msg":  "Error libreoffice not found",
				}, "application/json")
			}
		}

		// --------------------------------------------------------------------------------------------------------------------------------

		var FI os.FileInfo

		if FI, err = os.Stat(path.Join(prefix, "thumb", hex_name)); err != nil {

			//panic("Stat error " + path.Join(prefix, "thumb", hex_name) + " " + err.Error())
		}

		if FI != nil {

			if FI.Size() == 0 {

				if err = os.Remove(path.Join(prefix, "thumb", hex_name)); err != nil {
					panic("Problem of remove zero file " + path.Join(prefix, "thumb", hex_name) + " " + err.Error())
				}

			} else {

				model.EventLogAdd(db, c, "200", "GetThumb", "SendFile thumb/cache "+path.Join(c_path))

				return c.SendFile(path.Join(prefix, "thumb", hex_name), false)
			}
		}

		// --------------------------------------------------------------------------------------------------------------------------------

		filepath_tmp := path.Join(prefix, "temp")

		if runtime.GOOS == "windows" {
			filepath_tmp = util.RotateSlash(filepath_tmp)
		}

		var readerFile *os.File

		read_err_cnt := 1
		read_err := errors.New("read error 99")

		for read_err != nil {

			if runtime.GOOS == "windows" {

				arg_fold_path := util.RotateSlash(path.Join(arg_fold, c_path))

				util.RunAnyCommandUnderWin(`"C:/Program Files/LibreOffice/program/soffice.exe" --headless --norestore --nologo --convert-to png --outdir "` + filepath_tmp + `" "` + arg_fold_path + `"`)

			} else {

				cmd := exec.Command("bash", "-c", "libreoffice --headless --norestore --nologo --convert-to png --outdir "+filepath_tmp+" \""+path.Join(arg_fold, c_path)+"\"")
				cmd.Dir = arg_fold
				//out, _ := cmd.Output()
				//fmt.Println("out=", out)

				stderr, _ := cmd.StderrPipe()
				if err := cmd.Start(); err != nil {
					panic(err)
				}

				scanner := bufio.NewScanner(stderr)
				for scanner.Scan() {
					fmt.Println("libreoffice:", scanner.Text())
				}

			}

			readerFile, read_err = os.Open(path.Join(filepath_tmp, orig_filename+".png"))
			if read_err != nil {

				model.EventLogAdd(db, c, "500", "GetThumb", "REPEAT Error libreoffice, open file "+read_err.Error())

				time.Sleep(2 * time.Second)
			}

			if read_err_cnt > 5 {

				model.EventLogAdd(db, c, "500", "GetThumb", "SEVERAL Errors libreoffice, open file "+read_err.Error())

				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code": 500,
					"msg":  "Error libreoffice",
				}, "application/json")
			}

			read_err_cnt++
		}

		output, _ := os.Create(path.Join(prefix, "thumb", hex_name))

		_, err = io.Copy(output, readerFile)
		if err != nil {

			model.EventLogAdd(db, c, "500", "GetThumb", "Error copy after libreoffice convert: "+err.Error())

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "Error copy",
			}, "application/json")

		}
		output.Close()
		readerFile.Close()

		if _, err = os.Stat(path.Join(filepath_tmp, orig_filename+".png")); err == nil {

			os.Remove(path.Join(filepath_tmp, orig_filename+".png"))
		}

		// --------------------------------------------------------------------------------------------------------------------------------

		model.EventLogAdd(db, c, "200", "GetThumb", "Make office thumbnail and SendFile "+path.Join(c_path))

		return c.SendFile(path.Join(prefix, "thumb", hex_name), false)
	}

	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"code": 400,
		"file": path.Join(c_path),
		"msg":  "Bad request",
	}, "application/json")

}
