package api

import (
	//"bufio"
	//"errors"
	//"fmt"
	//"io"
	"net/url"
	"os"
	//"os/exec"
	//"path"
	"path/filepath"
	//"regexp"
	//"runtime"
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
	//"time"

	"github.com/western/http-here/v2/internal/conf"
	"github.com/western/http-here/v2/internal/model2"
	//"github.com/western/http-here/v2/pkg/safefs"
	"github.com/western/http-here/v2/internal/util"
	"github.com/western/http-here/v2/pkg/lopool"

	"github.com/edwvee/exiffix"
	"github.com/gofiber/fiber/v3"
)

/*
var (
	previewImgRegex = regexp.MustCompile("^(jpg|jpeg|png|gif)$")
	previewDocRegex = regexp.MustCompile("^(pdf|rtf|doc|docx|xls|xlsx|odt|ods)$")
	previewAllRegex = regexp.MustCompile("^(jpg|jpeg|png|gif|pdf|rtf|doc|docx|xls|xlsx|odt|ods)$")
)
*/

func GetThumb(c fiber.Ctx) error {

	/*
		if _, err := os.Stat(filepath.Join(conf.ConfigRoot, "thumb")); err != nil {

			if err := os.MkdirAll(filepath.Join(conf.ConfigRoot, "thumb"), os.ModePerm); err != nil {

				model2.EventLogAdd(c, 500, "GetThumb", "MkdirAll error "+err.Error())

				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code": 500,
					"msg":  "Error thumbnails folder create",
				})
			}
		}
	*/

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

	targetPath := filepath.Join(conf.ArgFold, c_Path)

	// ------------------------------------------------------------------------------------------------------------------------------

	file_ext_norm := util.GetExtNorm(c_Path) // EXT without prefix point and without JPEG exts

	is_preview_match := previewAllRegex.MatchString(file_ext_norm)
	if !is_preview_match {

		model2.EventLogAdd(c, 500, "GetThumb", filepath.Join("/__thumb/", c_Path)+" Only for JPEG, PNG, GIF and office files")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"file": filepath.Join("/__thumb/", c_Path),
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
			"file": filepath.Join("/__thumb/", c_Path),
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

		model2.EventLogAdd(c, 500, "GetThumb", filepath.Join("/__thumb/", c_Path)+", original file size is more than 20 Megabytes")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"file": filepath.Join("/__thumb/", c_Path),
		})
	}

	hex_name := util.GetMd5File(targetPath)

	if len(hex_name) == 0 {

		model2.EventLogAdd(c, 500, "GetThumb", filepath.Join("/__thumb/", c_Path)+", hex_name is empty (2)")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"file": filepath.Join("/__thumb/", c_Path),
		})
	}

	// ------------------------------------------------------------------------------------------------------------------------------
	// RETURN CACHE file from disk

	if _, err := conf.ThumbRootFS.Stat(hex_name); err == nil {

		model2.EventLogAdd(c, 200, "GetThumb", "SendFile thumb/cache "+c_Path)

		return c.SendFile(filepath.Join(conf.ConfigRoot, "thumb", hex_name))
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

	//var err error

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

			model2.EventLogAdd(c, 500, "ServeImgFile", "err3: "+err.Error())
			c.Type("png")
			return c.Send(util.PNGPixel)
		}
	}

	if file_ext == "jpg" {

		// src, err = jpeg.Decode(input)
		src, _, err = exiffix.Decode(input)
		if err != nil {

			model2.EventLogAdd(c, 500, "ServeImgFile", "err4: "+err.Error())
			c.Type("png")
			return c.Send(util.PNGPixel)
		}
	}

	if file_ext == "gif" {
		src, err = gif.Decode(input)
		if err != nil {

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

		model2.EventLogAdd(c, 200, "ServeImgFile", "SendFile orig without resize "+filepath.Join(conf.ArgFold, c_Path))

		err := conf.ThumbRootFS.Remove(hex_name)
		if err != nil {

			model2.EventLogAdd(c, 500, "ServeImgFile", "err6: "+err.Error())
			c.Type("png")
			return c.Send(util.PNGPixel)
		}

		return c.SendFile(filepath.Join(conf.ArgFold, c_Path))
	}

	// Resize:
	draw.NearestNeighbor.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)

	if file_ext == "png" {
		err = png.Encode(output, dst)
		if err != nil {

			model2.EventLogAdd(c, 500, "ServeImgFile", "err7: "+err.Error())
			c.Type("png")
			return c.Send(util.PNGPixel)
		}
	}

	if file_ext == "jpg" {
		err = jpeg.Encode(output, dst, nil)
		if err != nil {

			model2.EventLogAdd(c, 500, "ServeImgFile", "err8: "+err.Error())
			c.Type("png")
			return c.Send(util.PNGPixel)
		}
	}

	if file_ext == "gif" {
		err = gif.Encode(output, dst, nil)
		if err != nil {

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

	return c.SendFile(filepath.Join(conf.ConfigRoot, "thumb", hex_name))
}

func ServeOfficePool(c fiber.Ctx, c_Path string, hex_name string) error {

	ctx := c.Context()

	targetPath := filepath.Join(conf.ArgFold, c_Path)

	cctx, cancel := context.WithTimeout(ctx, lopool.ConvertTimeout)
	defer cancel()

	tmp, err := os.MkdirTemp("", "office_pool_")
	if err != nil {

		model2.EventLogAdd(c, 500, "ServeOfficePool", c_Path+" err0:"+err.Error())

		c.Type("png")
		return c.Send(util.PNGPixel)
	}
	defer os.RemoveAll(tmp)

	pngPath, err := lopool.LibreOfficePool.Convert(cctx, targetPath, tmp)
	if err != nil {

		model2.EventLogAdd(c, 500, "ServeOfficePool", c_Path+" err1:"+err.Error())

		c.Type("png")
		return c.Send(util.PNGPixel)
	}

	defer os.Remove(pngPath)

	err = util.CopyFile(pngPath, filepath.Join(conf.ConfigRoot, "thumb", hex_name))
	if err != nil {

		model2.EventLogAdd(c, 500, "ServeOfficePool", c_Path+" err2:"+err.Error())

		c.Type("png")
		return c.Send(util.PNGPixel)
	}

	// --------------------------------------------------------------------------------------------------------------------------------

	model2.EventLogAdd(c, 200, "ServeOfficePool", "Make office thumbnail and SendFile "+c_Path)

	return c.SendFile(filepath.Join(conf.ConfigRoot, "thumb", hex_name))
}
