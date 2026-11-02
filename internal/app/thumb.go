package app

import (
	//"bufio"
	"errors"
	//"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path"
	//"path/filepath"
	"regexp"
	"runtime"
	//"strconv"
	"context"
	"strings"
	//"syscall"
	"golang.org/x/image/draw"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"math"
	"time"

	"github.com/western/http-here/v2/internal/conf"
	"github.com/western/http-here/v2/internal/model2"
	//"github.com/western/http-here/v2/internal/safefs"
	"github.com/western/http-here/v2/internal/lopool"
	"github.com/western/http-here/v2/internal/util"

	"github.com/edwvee/exiffix"
	"github.com/gofiber/fiber/v3"
)

var (
	previewImgRegex = regexp.MustCompile("^(jpg|jpeg|png|gif)$")
	previewDocRegex = regexp.MustCompile("^(pdf|rtf|doc|docx|xls|xlsx|odt|ods)$")
	previewAllRegex = regexp.MustCompile("^(jpg|jpeg|png|gif|pdf|rtf|doc|docx|xls|xlsx|odt|ods)$")
)

func GetThumb(c fiber.Ctx) error {

	if _, err := os.Stat(path.Join(conf.ConfigRoot, "thumb")); err != nil {

		if err := os.MkdirAll(path.Join(conf.ConfigRoot, "thumb"), os.ModePerm); err != nil {

			model2.EventLogAdd(c, 500, "GetThumb", "MkdirAll error "+err.Error())

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "Error thumbnails folder create",
			})
		}
	}

	// ------------------------------------------------------------------------------------------------------------------------------

	c_Path, err := url.QueryUnescape(c.Path())
	if err != nil {

		model2.EventLogAdd(c, 500, "GetThumb", "Error "+c_Path+" "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Error " + c_Path,
		})
	}

	c_Path = util.CleanDirtyPath(c_Path)
	c_Path = strings.Replace(c_Path, "/__thumb", "", 1)

	targetPath := path.Join(conf.ArgFold, c_Path)

	// ------------------------------------------------------------------------------------------------------------------------------

	file_ext_norm := util.GetExtNorm(c_Path) // EXT without prefix point and without JPEG exts

	is_preview_match := previewAllRegex.MatchString(file_ext_norm)
	if !is_preview_match {

		model2.EventLogAdd(c, 500, "GetThumb", path.Join("/__thumb/", c_Path)+" Only for JPEG, PNG, GIF and office files")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"file": path.Join("/__thumb/", c_Path),
			"msg":  "Only for JPEG, PNG, GIF and office files",
		})
	}

	// ------------------------------------------------------------------------------------------------------------------------------
	// open root with trust ArgFold

	Root := conf.ArgFoldRootFS

	fileInfoOriginal, err := Root.Stat(strings.TrimPrefix(c_Path, "/"))
	if err != nil {

		model2.EventLogAdd(c, 404, "GetThumb", "stat err "+err.Error())

		/*
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "stat err",
			})
		*/
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code": 404,
			"file": path.Join("/__thumb/", c_Path),
			"msg":  "Not found",
		})
	}

	fileModeOriginal := fileInfoOriginal.Mode()

	if !fileModeOriginal.IsRegular() {

		model2.EventLogAdd(c, 500, "GetThumb", "It is not regular file. "+targetPath)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "It is not regular file",
		})
	}

	if fileModeOriginal.Perm()&0444 == 0444 {

	} else {

		model2.EventLogAdd(c, 403, "GetThumb", "Forbidden for read "+targetPath)

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"code": 403,
			"msg":  "Forbidden",
		})
	}

	// ------------------------------------------------------------------------------------------------------------------------------

	// ------------------------------------------------------------------------------------------------------------------------------

	if fileInfoOriginal.Size() > 20*1024*1024 {

		model2.EventLogAdd(c, 500, "GetThumb", path.Join("/__thumb/", c_Path)+", original file size is more than 20 Megabytes")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"file": path.Join("/__thumb/", c_Path),
		})
	}

	hex_name := util.GetMd5File(targetPath)

	if len(hex_name) == 0 {

		model2.EventLogAdd(c, 500, "GetThumb", path.Join("/__thumb/", c_Path)+", hex_name is empty (2)")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"file": path.Join("/__thumb/", c_Path),
		})
	}

	// ------------------------------------------------------------------------------------------------------------------------------
	// RETURN CACHE file from disk

	if _, err := conf.ThumbRootFS.Stat(hex_name); err == nil {

		model2.EventLogAdd(c, 200, "GetThumb", "SendFile thumb/cache "+c_Path)

		return c.SendFile(path.Join(conf.ConfigRoot, "thumb", hex_name))
	}

	// ------------------------------------------------------------------------------------------------------------------------------

	is_preview_img := previewImgRegex.MatchString(file_ext_norm)
	if is_preview_img {

		if fileInfoOriginal.Size() < 50*1024 {

			model2.EventLogAdd(c, 200, "GetThumb", "SendFile orig file without resize "+c_Path+", size is small")

			return c.SendFile(targetPath)
		}

		return ServeImgFile(c, c_Path, hex_name)
	}

	is_preview_doc := previewDocRegex.MatchString(file_ext_norm)
	if is_preview_doc {

		//return ServeOfficeFile(c, c_Path, hex_name)
		return ServeOfficePool(c, c_Path, hex_name)
	}

	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"code": 400,
		"file": c_Path,
		"msg":  "Bad request",
	})

}

func ServeImgFile(c fiber.Ctx, c_Path string, hex_name string) error {

	var err error

	// ------------------------------------------------------------------------------------------------------------------------------

	// ------------------------------------------------------------------------------------------------------------------------------

	input, err := conf.ArgFoldRootFS.Open(strings.TrimPrefix(c_Path, "/"))
	if err != nil {

		model2.EventLogAdd(c, 500, "ServeImgFile", "err1: "+err.Error())
		c.Type("png")
		return c.Send(util.PNGPixel)
	}
	defer input.Close()

	// ------------------------------------------------------------------------------------------------------------------------------

	output, err := conf.ThumbRootFS.Create(hex_name)
	if err != nil {

		model2.EventLogAdd(c, 500, "ServeImgFile", "err2: "+err.Error())
		c.Type("png")
		return c.Send(util.PNGPixel)
	}
	defer output.Close()

	// ------------------------------------------------------------------------------------------------------------------------------

	file_ext := util.GetExtNorm(c_Path)

	var src image.Image

	// Decode the image (from PNG to image.Image):
	if file_ext == "png" {
		src, err = png.Decode(input)
		if err != nil {
			//panic(err)

			model2.EventLogAdd(c, 500, "ServeImgFile", "err3: "+err.Error())
			c.Type("png")
			return c.Send(util.PNGPixel)
		}
	}

	if file_ext == "jpg" {

		// src, err = jpeg.Decode(input)
		src, _, err = exiffix.Decode(input)
		if err != nil {
			//panic(err)

			model2.EventLogAdd(c, 500, "ServeImgFile", "err4: "+err.Error())
			c.Type("png")
			return c.Send(util.PNGPixel)
		}
	}

	if file_ext == "gif" {
		src, err = gif.Decode(input)
		if err != nil {
			//panic(err)

			model2.EventLogAdd(c, 500, "ServeImgFile", "err5: "+err.Error())
			c.Type("png")
			return c.Send(util.PNGPixel)
		}
	}

	i_width := 600

	ratio := (float64)(src.Bounds().Max.Y) / (float64)(src.Bounds().Max.X)
	i_height := int(math.Round(float64(i_width) * ratio))

	var dst *image.RGBA

	if src.Bounds().Max.X > i_width || src.Bounds().Max.Y > i_height {
		dst = image.NewRGBA(image.Rect(0, 0, i_width, i_height))
	} else {

		model2.EventLogAdd(c, 200, "ServeImgFile", "SendFile orig without resize "+path.Join(conf.ArgFold, c_Path))

		err := conf.ThumbRootFS.Remove(hex_name)
		if err != nil {
			//panic(err)

			model2.EventLogAdd(c, 500, "ServeImgFile", "err6: "+err.Error())
			c.Type("png")
			return c.Send(util.PNGPixel)
		}

		return c.SendFile(path.Join(conf.ArgFold, c_Path))
	}

	// Resize:
	draw.NearestNeighbor.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)

	if file_ext == "png" {
		err = png.Encode(output, dst)
		if err != nil {
			//panic(err)

			model2.EventLogAdd(c, 500, "ServeImgFile", "err7: "+err.Error())
			c.Type("png")
			return c.Send(util.PNGPixel)
		}
	}

	if file_ext == "jpg" {
		err = jpeg.Encode(output, dst, nil)
		if err != nil {
			//panic(err)

			model2.EventLogAdd(c, 500, "ServeImgFile", "err8: "+err.Error())
			c.Type("png")
			return c.Send(util.PNGPixel)
		}
	}

	if file_ext == "gif" {
		err = gif.Encode(output, dst, nil)
		if err != nil {
			//panic(err)

			model2.EventLogAdd(c, 500, "ServeImgFile", "err9: "+err.Error())
			c.Type("png")
			return c.Send(util.PNGPixel)
		}
	}

	output.Close()
	input.Close()

	src = nil
	dst = nil

	model2.EventLogAdd(c, 200, "ServeImgFile", "Resize and SendFile "+c_Path)

	return c.SendFile(path.Join(conf.ConfigRoot, "thumb", hex_name))
}

func ServeOfficePool(c fiber.Ctx, c_Path string, hex_name string) error {

	ctx := c.Context()

	targetPath := path.Join(conf.ArgFold, c_Path)

	cctx, cancel := context.WithTimeout(ctx, lopool.ConvertTimeout)
	//_, cancel := context.WithTimeout(ctx, lopool.ConvertTimeout)
	defer cancel()

	pngPath, err := lopool.LibreOfficePool.Convert(cctx, targetPath, "/tmp")
	if err != nil {
		//panic(err)

		model2.EventLogAdd(c, 500, "ServeOfficePool", c_Path+" err1:"+err.Error())

		c.Type("png")
		return c.Send(util.PNGPixel)
	}

	//fmt.Println("pngPath=", pngPath)

	defer os.Remove(pngPath)

	err = util.CopyFile(pngPath, path.Join(conf.ConfigRoot, "thumb", hex_name))
	if err != nil {
		//panic(err)

		model2.EventLogAdd(c, 500, "ServeOfficePool", c_Path+" err2:"+err.Error())

		c.Type("png")
		return c.Send(util.PNGPixel)
	}

	// --------------------------------------------------------------------------------------------------------------------------------

	model2.EventLogAdd(c, 200, "ServeOfficePool", "Make office thumbnail and SendFile "+c_Path)

	//return c.SendFile(path.Join(conf.ConfigRoot, "thumb", hex_name))

	/*
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Some libreoffice error",
		})
	*/

	return c.SendFile(path.Join(conf.ConfigRoot, "thumb", hex_name))
}

func ServeOfficeFile(c fiber.Ctx, c_path string, hex_name string) error {

	orig_filename := util.GetFileName(c_path)
	var err error

	if runtime.GOOS == "windows" {

		_, err := os.Stat("C:\\Program Files\\LibreOffice\\program\\soffice.exe")
		if err != nil {

			model2.EventLogAdd(c, 500, "ServeOfficeFile", "Error libreoffice not found "+err.Error())

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "Error libreoffice not found",
			}, "application/json")
		}

	} else {

		//_, err := exec.Command("bash", "-c", "libreoffice --help").Output()
		_, err := exec.LookPath("libreoffice")
		if err != nil {

			model2.EventLogAdd(c, 500, "ServeOfficeFile", "Error libreoffice not found "+err.Error())

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "Error libreoffice not found",
			}, "application/json")
		}
	}

	// --------------------------------------------------------------------------------------------------------------------------------

	var HexNameInfo os.FileInfo

	if HexNameInfo, err = os.Stat(path.Join(conf.ConfigRoot, "thumb", hex_name)); err != nil {

		//panic("Stat error " + path.Join(conf.ConfigRoot, "thumb", hex_name) + " " + err.Error())
	}

	if HexNameInfo != nil {

		if HexNameInfo.Size() == 0 {

			if err = os.Remove(path.Join(conf.ConfigRoot, "thumb", hex_name)); err != nil {
				panic("Problem of remove zero file " + path.Join(conf.ConfigRoot, "thumb", hex_name) + " " + err.Error())
			}

		} else {

			model2.EventLogAdd(c, 200, "ServeOfficeFile", "SendFile thumb/cache "+path.Join(c_path))

			return c.SendFile(path.Join(conf.ConfigRoot, "thumb", hex_name))
		}
	}

	// --------------------------------------------------------------------------------------------------------------------------------

	filepath_tmp := path.Join(conf.ConfigRoot, "temp")
	arg_fold_path := path.Join(conf.ArgFold, c_path)

	if runtime.GOOS == "windows" {
		filepath_tmp = util.RotateSlash(filepath_tmp)
		arg_fold_path = util.RotateSlash(arg_fold_path)
	}

	var readerFile *os.File

	read_err_cnt := 1
	read_err := errors.New("read error 99")

	for read_err != nil {

		MakeLibreofficeThumbnail(filepath_tmp, arg_fold_path)

		/*
			if runtime.GOOS == "windows" {

				util.RunAnyCommandUnderWin(`"C:/Program Files/LibreOffice/program/soffice.exe" --headless --norestore --nologo --convert-to png --outdir "` + filepath_tmp + `" "` + arg_fold_path + `"`)

			} else {



				ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
				defer cancel()

				cmd := exec.CommandContext(ctx, "bash", "-c", "libreoffice --headless --norestore --nologo --convert-to png --outdir "+filepath_tmp+" \""+arg_fold_path+"\"")
				cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

				cmdReader, _ := cmd.StderrPipe()
				//cmd.Stdout = cmd.Stderr   // combine ERR + OUT
				if err := cmd.Start(); err != nil {
					panic(err)
				}

				scanner1 := bufio.NewScanner(cmdReader)
				for scanner1.Scan() {
					fmt.Println("libreoffice:", scanner1.Text())
				}

				done := make(chan error, 1)
				go func() {
					done <- cmd.Wait()
				}()

				select {
				case err := <-done:
					// Command finished on its own
					if err != nil {
						//fmt.Printf("Command finished with error: %v\n", err)
					} else {
						//fmt.Println("Command finished successfully")
					}

				case <-ctx.Done():
					// Context was canceled or timed out
					//fmt.Println("Context done, killing process group...")
					// Get the process group ID (pgid) and kill the entire group using a negative PID
					pgid, err := syscall.Getpgid(cmd.Process.Pid)
					if err == nil {
						// Use -pgid to kill the process group. SIGKILL (9) is forceful.
						if err := syscall.Kill(-pgid, syscall.SIGKILL); err != nil {
							//fmt.Printf("Failed to kill process group: %v\n", err)
						}
					} else {
						//fmt.Printf("Failed to get pgid: %v\n", err)
					}
					// Wait again to reap the process and avoid zombies
					<-done
					//fmt.Printf("Command terminated: %v\n", ctx.Err())
				}

			}
		*/

		readerFile, read_err = os.Open(path.Join(filepath_tmp, orig_filename+".png"))
		if read_err != nil {

			//model2.EventLogAdd(c, 500, "ServeOfficeFile", "REPEAT Error libreoffice, open file "+read_err.Error())

			time.Sleep(2 * time.Second)
		}

		if read_err_cnt > 5 {

			//model2.EventLogAdd(c, 500, "ServeOfficeFile", "SEVERAL Errors libreoffice, open file "+read_err.Error())
			model2.EventLogAdd(c, 500, "ServeOfficeFile", "SEVERAL Errors libreoffice, png file "+orig_filename+" absent")

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "Error libreoffice",
			}, "application/json")
		}

		read_err_cnt++
	}

	output, _ := os.Create(path.Join(conf.ConfigRoot, "thumb", hex_name))

	_, err = io.Copy(output, readerFile)
	if err != nil {

		model2.EventLogAdd(c, 500, "ServeOfficeFile", "Error copy after libreoffice convert: "+err.Error())

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

	model2.EventLogAdd(c, 200, "ServeOfficeFile", "Make office thumbnail and SendFile "+path.Join(c_path))

	return c.SendFile(path.Join(conf.ConfigRoot, "thumb", hex_name))
}
