
package app

import (
	_ "archive/zip"
	"errors"
	"fmt"
	"path/filepath"
	_ "reflect"
	"regexp"
	_ "strconv"
	"strings"

	"bufio"
	"os"
	"os/exec"

	"io"
	"log"
	"net/url"
	"time"

	_ "crypto/md5"
	_ "encoding/hex"

	"github.com/western/http-here/internal/model"

	"github.com/gofiber/fiber/v2"

	"github.com/edwvee/exiffix"
	"golang.org/x/image/draw"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"math"
)

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

	//c_width := "600"
	i_width := 600

	//modtime_human := ""
	//size_human := ""

	if fileInfo, err := os.Stat(filepath.Join(arg_fold, c_path)); err == nil {

		//modtime := fileInfo.ModTime()
		//modtime_human = modtime.Format("2006-01-02 15:04:05")

		//size := fileInfo.Size()
		//size_human = PrettyByteSize(size)

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

	//is_match, _ := regexp.MatchString("^(jpg|jpeg|png|gif)$", file_ext)
	is_match, _ := regexp.MatchString("^(jpg|jpeg|png|gif|pdf|rtf|doc|docx|xls|xlsx|odt|ods)$", file_ext)
	if !is_match {
		LogPrefix(c, "500", filepath.Join("/__resize/", c_path)+" Only for JPEG, PNG, GIF and office files")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"file": filepath.Join("/__resize/", c_path),
			"msg":  "Only for JPEG, PNG, GIF and office files",
		}, "application/json")
	}

	//hash_name := md5.Sum([]byte(orig_filename + modtime_human + size_human + c_width))
	//hex_name := hex.EncodeToString(hash_name[:])
	hex_name := ""

	db, err := model.ConnectToSQLite()
	if err != nil {
		panic(err)
	}
	//defer db.Close()

	var row model.File

	if result := db.Where("full_path = ?", filepath.Join(arg_fold, c_path)).First(&row); result.Error == nil {

		hex_name = row.MD5

		if _, err := os.Stat(filepath.Join(homepath, ".httphere", "thumb", hex_name)); err == nil {

			LogPrefix(c, "200", "SendFile db thumb/cache "+filepath.Join(c_path))
			return c.SendFile(filepath.Join(homepath, ".httphere", "thumb", hex_name), false)
		}
	} else {

		hex_name = GetMd5File(filepath.Join(arg_fold, c_path))
	}

	is_img_match, _ := regexp.MatchString("^(jpg|jpeg|png|gif)$", file_ext)

	if is_img_match {

		if _, err := os.Stat(filepath.Join(homepath, ".httphere", "thumb", hex_name)); err == nil {

			LogPrefix(c, "200", "SendFile thumb/cache "+filepath.Join(c_path))
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

	}

	is_office_match, _ := regexp.MatchString("^(pdf|rtf|doc|docx|xls|xlsx|odt|ods)$", file_ext)

	if is_office_match {

		_, err := exec.Command("bash", "-c", "libreoffice --help").Output()
		if err != nil {

			LogPrefix(c, "500", "Error libreoffice not found "+err.Error())

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "Error libreoffice not found",
			}, "application/json")
		}

		// --------------------------------------------------------------------------------------------------------------------------------

		var FI os.FileInfo

		if FI, err = os.Stat(filepath.Join(homepath, ".httphere", "thumb", hex_name)); err != nil {

			//panic("Stat error " + filepath.Join(homepath, ".httphere", "thumb", hex_name) + " " + err.Error())
		}

		if FI != nil {

			if FI.Size() == 0 {

				if err = os.Remove(filepath.Join(homepath, ".httphere", "thumb", hex_name)); err != nil {
					panic("Problem of remove zero file " + filepath.Join(homepath, ".httphere", "thumb", hex_name) + " " + err.Error())
				}

			} else {

				LogPrefix(c, "200", "SendFile thumb/cache "+filepath.Join(c_path))
				return c.SendFile(filepath.Join(homepath, ".httphere", "thumb", hex_name), false)
			}
		}

		// --------------------------------------------------------------------------------------------------------------------------------

		//input, _ := os.Open(filepath.Join(arg_fold, c_path))
		//defer input.Close()

		output, _ := os.Create(filepath.Join(homepath, ".httphere", "thumb", hex_name))
		defer output.Close()

		//filepath_dir := filepath.Dir(c_path)
		filepath_tmp := filepath.Join(homepath, ".httphere", "temp")

		//var readerFile *os.File
		readerFile, read_err := os.Open(filepath.Join(filepath_tmp, orig_filename+".png"))
		read_err_cnt := 1

		for read_err != nil {

			// libreoffice --headless --convert-to png --outdir /tmp "000_RR_fff ddd ttt.docx"
			// --accept='socket,host=localhost,port=8103;urp;StarOffice.ComponentContext'
			cmd := exec.Command("bash", "-c", "libreoffice --headless --norestore --nologo --convert-to png --outdir "+filepath_tmp+" \""+filepath.Join(arg_fold, c_path)+"\"")
			cmd.Dir = arg_fold
			//out, _ := cmd.Output()
			//fmt.Println("out=", out)

			/*
				stderr, err := cmd.StderrPipe()
				if err != nil {
					panic(err)
				}

				if err := cmd.Start(); err != nil {
					panic(err)
				}

				slurp, _ := io.ReadAll(stderr)
				fmt.Printf("%s\n", slurp)

				if err := cmd.Wait(); err != nil {

					LogPrefix(c, "500", "Error libreoffice "+err.Error())

					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"code": 500,
						"msg":  "Error libreoffice",
					}, "application/json")

				}
			*/

			stderr, _ := cmd.StderrPipe()
			if err := cmd.Start(); err != nil {
				panic(err)
			}

			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				//fmt.Println("libreoffice:", scanner.Text())
			}

			readerFile, read_err = os.Open(filepath.Join(filepath_tmp, orig_filename+".png"))
			if read_err != nil {
				//panic(err)

				LogPrefix(c, "500", "REPEAT Error libreoffice, open file "+err.Error())

				/*
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"code": 500,
						"msg":  "Error libreoffice",
					}, "application/json")
				*/

				time.Sleep(2 * time.Second)
			}

			if read_err_cnt > 5 {

				LogPrefix(c, "500", "SEVERAL Errors libreoffice, open file "+err.Error())

				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code": 500,
					"msg":  "Error libreoffice",
				}, "application/json")
			}

			read_err_cnt++
		}

		//defer os.Remove( filepath.Join(filepath_tmp, orig_filename+".png") )
		defer readerFile.Close()

		_, err = io.Copy(output, readerFile)
		if err != nil {
			panic(err)
		}
		output.Close()

		// --------------------------------------------------------------------------------------------------------------------------------

		LogPrefix(c, "200", "Make office thumbnail and SendFile "+filepath.Join(c_path))

		return c.SendFile(filepath.Join(homepath, ".httphere", "thumb", hex_name), false)
	}

	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"code": 400,
		"file": filepath.Join(c_path),
		"msg":  "Bad request",
	}, "application/json")

}
