package controller

import (
	"errors"
	"fmt"

	"archive/zip"
	"path/filepath"
	"regexp"
	"strconv"
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

		log.Println(err)
		LogPrefix(c, "500", "Error url parse "+referer)
		return c.JSON(fiber.Map{
			"code": 500,
		}, "application/json")
	}

	u_path := CleanDirtyPath(u.Path)

	form, _ := c.MultipartForm()
	files := form.File["fileBlob"]

	for _, file := range files {
		file_ext := filepath.Ext(file.Filename)
		file_ext = strings.Replace(file_ext, ".", "", -1)
		file_ext = strings.ToLower(file_ext)
		if file_ext == "jpeg" {
			file_ext = "jpg"
		}

		originalFileName := strings.TrimSuffix(
			filepath.Base(file.Filename),
			filepath.Ext(file.Filename),
		)

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

			log.Println(err)
			LogPrefix(c, "500", "Error create "+filepath.Join(arg_fold, u_path, filename))
			return c.JSON(fiber.Map{
				"code": 500,
			}, "application/json")
		}
		defer out.Close()

		LogPrefix(c, "200", "Save '"+filepath.Join(arg_fold, u_path, filename)+"'")

		readerFile, _ := file.Open()
		_, err = io.Copy(out, readerFile)
		if err != nil {

			log.Println(err)
			LogPrefix(c, "500", "Error copy "+filepath.Join(arg_fold, u_path, filename))
			return c.JSON(fiber.Map{
				"code": 500,
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

		log.Println(err)
		LogPrefix(c, "500", "Error url parse "+referer)
		return c.JSON(fiber.Map{
			"code": 500,
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
		return c.JSON(fiber.Map{
			"code": 500,
			"msg":  "name is empty",
		}, "application/json")
	}

	if fileInfo, err := os.Stat(filepath.Join(arg_fold, u_path, name)); err == nil {

		if fileInfo.IsDir() {
			LogPrefix(c, "500", "'"+filepath.Join(arg_fold, u_path, name)+"' already exists")
			return c.JSON(fiber.Map{
				"code": 500,
				"msg":  filepath.Join(u_path, name) + " already exists",
			}, "application/json")
		}
	}

	if err := os.Mkdir(filepath.Join(arg_fold, u_path, name), os.ModePerm); err != nil {

		log.Println(err)
		LogPrefix(c, "500", "Error mkdir "+filepath.Join(arg_fold, u_path, name))
		return c.JSON(fiber.Map{
			"code": 500,
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

		log.Println(err)
		LogPrefix(c, "500", "Error url parse "+referer)
		return c.JSON(fiber.Map{
			"code": 500,
		}, "application/json")
	}

	u_path := CleanDirtyPath(u.Path)

	for i := 1; i < 50; i++ {

		key := "name[" + strconv.Itoa(i) + "]"
		val := c.FormValue(key)

		if len(val) > 0 {
			//fmt.Println("val="+val)

			name := strings.ReplaceAll(val, "/", "")
			//re := regexp.MustCompile("\\s+")
			//name = re.ReplaceAllLiteralString(name, " ")

			name = CleanDirtyPath(name)

			if len(name) == 0 {
				LogPrefix(c, "500", "name is empty")
				continue
				/*
					return c.JSON(fiber.Map{
						"code": 500,
						"msg":  "name is empty",
					}, "application/json")
				*/
			}

			fileInfo, err := os.Stat(filepath.Join(arg_fold, u_path, name))

			if err != nil {
				LogPrefix(c, "500", "'"+filepath.Join(arg_fold, u_path, name)+"' not exists")
				continue
				/*
					return c.JSON(fiber.Map{
						"code": 500,
						"msg":  filepath.Join(u_path, name) + " not exists",
					}, "application/json")
				*/
			}

			if fileInfo.IsDir() {

				// remove fold and all inside data

				if err := os.RemoveAll(filepath.Join(arg_fold, u_path, name)); err != nil {

					LogPrefix(c, "500", "'"+filepath.Join(arg_fold, u_path, name)+"' err")
					continue
					/*
						return c.JSON(fiber.Map{
							"code": 500,
							"msg":  filepath.Join(u_path, name) + " err",
						}, "application/json")
					*/
				}

				LogPrefix(c, "200", "Remove fold '"+filepath.Join(arg_fold, u_path, name)+"'")

			} else {

				// remove one file

				if err := os.Remove(filepath.Join(arg_fold, u_path, name)); err != nil {

					LogPrefix(c, "500", "'"+filepath.Join(arg_fold, u_path, name)+"' err")
					continue
					/*
						return c.JSON(fiber.Map{
							"code": 500,
							"msg":  filepath.Join(u_path, name) + " err",
						}, "application/json")
					*/
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

		log.Println(err)
		LogPrefix(c, "500", "Error url parse "+referer)
		return c.JSON(fiber.Map{
			"code": 500,
		}, "application/json")
	}

	u_path := CleanDirtyPath(u.Path)
	
	
	to := c.FormValue("to")
	_, err = os.Stat(filepath.Join(arg_fold, to))
	if err != nil {
		LogPrefix(c, "500", "'"+filepath.Join(arg_fold, to)+"' not exists")
		return c.JSON(fiber.Map{
			"code": 500,
		}, "application/json")
	}
	

	for i := 1; i < 50; i++ {

		key := "name[" + strconv.Itoa(i) + "]"
		val := c.FormValue(key)

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

			
                
				
                
            err = os.Rename( filepath.Join(arg_fold, u_path, name), filepath.Join(arg_fold, to, name) )
            
            if err != nil {
			    //fmt.Errorf("something %s", foo)
			    LogPrefix(c, "500", "Rename error "+fmt.Sprintf("%s", err))
			    
		    }else{
		        
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
		log.Fatal(err2)
	}
	//fmt.Println( homepath )

	if _, err3 := os.Stat(filepath.Join(homepath, ".httphere", "temp")); err3 != nil {

		if err4 := os.MkdirAll(filepath.Join(homepath, ".httphere", "temp"), os.ModePerm); err4 != nil {
			log.Fatal(err4)
		}
	}

	referer := c.Get("Referer")

	// already decoded
	u, err := url.Parse(referer)
	if err != nil {

		log.Println(err)
		LogPrefix(c, "500", "Error url parse "+referer)
		return c.JSON(fiber.Map{
			"code": 500,
		}, "application/json")
	}

	u_path := CleanDirtyPath(u.Path)

	//archive_name := "archive-"+time.Now().Format("2006-01-02T15:04:05")+".zip"
	archive_name := "archive-" + time.Now().Format("20060102-150405") + ".zip"

	//archive, err := os.Create(filepath.Join(arg_fold, u_path, "archive.zip"))
	archive, err := os.Create(filepath.Join(homepath, ".httphere", "temp", archive_name))

	if err != nil {
		panic(err)
	}
	defer archive.Close()

	zipWriter := zip.NewWriter(archive)

	for i := 1; i < 50; i++ {

		key := "name[" + strconv.Itoa(i) + "]"
		val := c.FormValue(key)

		if len(val) > 0 {
			//fmt.Println("val="+val)

			name := strings.ReplaceAll(val, "/", "")
			//re := regexp.MustCompile("\\s+")
			//name = re.ReplaceAllLiteralString(name, " ")

			name = CleanDirtyPath(name)

			if len(name) == 0 {
				LogPrefix(c, "500", "name is empty")
				continue
				/*
					return c.JSON(fiber.Map{
						"code": 500,
						"msg":  "name is empty",
					}, "application/json")
				*/
			}

			fileInfo, err := os.Stat(filepath.Join(arg_fold, u_path, name))

			if err != nil {
				LogPrefix(c, "500", "'"+filepath.Join(arg_fold, u_path, name)+"' not exists")
				continue
				/*
					return c.JSON(fiber.Map{
						"code": 500,
						"msg":  filepath.Join(u_path, name) + " not exists",
					}, "application/json")
				*/
			}

			if fileInfo.IsDir() {

				/*

					if err := os.RemoveAll(filepath.Join(arg_fold, u_path, name)); err != nil {

						LogPrefix(c, "500", "'"+filepath.Join(arg_fold, u_path, name)+"' err")
						continue

					}

					LogPrefix(c, "200", "Remove fold '"+filepath.Join(arg_fold, u_path, name)+"'")
				*/

				addFilesToZip(zipWriter, filepath.Join(arg_fold, u_path, name), name)

			} else {

				/*
				   zipWriter := zip.NewWriter(archive)


				   f1, err := os.Open(filepath.Join(arg_fold, u_path, name))
				   if err != nil {
				       panic(err)
				   }
				   defer f1.Close()


				   w1, err := zipWriter.Create( name)
				   if err != nil {
				       panic(err)
				   }
				   if _, err := io.Copy(w1, f1); err != nil {
				       panic(err)
				   }

				   zipWriter.Close()
				*/

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

				/*
					if err := os.Remove(filepath.Join(arg_fold, u_path, name)); err != nil {

						LogPrefix(c, "500", "'"+filepath.Join(arg_fold, u_path, name)+"' err")
						continue

					}

					LogPrefix(c, "200", "Remove '"+filepath.Join(arg_fold, u_path, name)+"'")
				*/

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
		log.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(homepath, ".httphere", "thumb")); err != nil {

		if err2 := os.MkdirAll(filepath.Join(homepath, ".httphere", "thumb"), os.ModePerm); err2 != nil {
			log.Fatal(err2)
		}
	}

	/*
		referer := c.Get("Referer")

		// already decoded
		u, err := url.Parse(referer)
		if err != nil {

			log.Println(err)
			LogPrefix(c, "500", "Error url parse "+referer)
			return c.JSON(fiber.Map{
				"code": 500,
			}, "application/json")
		}

		u_path := CleanDirtyPath(u.Path)
	*/

	c_path, err := url.QueryUnescape(c.Path())
	if err != nil {

		log.Println(err)
		LogPrefix(c, "500", "Error "+filepath.Join(arg_fold, c_path))
		//return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout")
		return c.JSON(fiber.Map{
			"code": 500,
			"msg":  "Error " + filepath.Join(arg_fold, c_path),
		}, "application/json")
	}

	c_path = CleanDirtyPath(c_path)
	c_path = strings.Replace(c_path, "/__resize", "", 1)

	/*
			c_width := c.Params("width")
			c_height := c.Params("height")

			c_path = strings.Replace(c_path, "/"+c_width, "", 1)
			c_path = strings.Replace(c_path, "/"+c_height, "", 1)


			i_width, err := strconv.Atoi(c_width)
		    if err != nil {
		        log.Fatal(err)
		    }

			i_height, err := strconv.Atoi(c_height)
		    if err != nil {
		        log.Fatal(err)
		    }
	*/

	c_width := "800"
	c_height := "600"

	i_width := 800
	i_height := 600

	//fmt.Println("="+filepath.Join(arg_fold, c_path))

	modtime_human := ""
	size_human := ""

	if fileInfo, err := os.Stat(filepath.Join(arg_fold, c_path)); err == nil {

		modtime := fileInfo.ModTime()
		modtime_human = modtime.Format("2006-01-02 15:04:05")
		//fmt.Println("modtime_human="+modtime_human)

		size := fileInfo.Size()
		size_human = prettyByteSize(size)
		//fmt.Println("size_human="+size_human)

		if fileInfo.IsDir() {
			//log.Println(err)
			LogPrefix(c, "500", "Error "+filepath.Join(arg_fold, c_path)+" It is a folder")
			//return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout")
			return c.JSON(fiber.Map{
				"code": 500,
				"file": filepath.Join("/__resize/", c_path) + " It is a folder",
				"msg":  "It is a folder",
			}, "application/json")
		}

	} else if errors.Is(err, os.ErrNotExist) {

		LogPrefix(c, "404", filepath.Join(arg_fold, c_path))

		/*
			return c.Status(fiber.StatusNotFound).Render("view/404", fiber.Map{
				"File": c_path,
			}, "view/layout")
		*/
		return c.JSON(fiber.Map{
			"code": 404,
			"file": filepath.Join("/__resize/", c_path),
			"msg":  "Not found",
		}, "application/json")
	}

	//orig_path := filepath.Dir(c_path)
	//fmt.Println("orig_path="+orig_path)

	file_ext := filepath.Ext(c_path)
	file_ext = strings.Replace(file_ext, ".", "", -1)
	file_ext = strings.ToLower(file_ext)
	if file_ext == "jpeg" {
		file_ext = "jpg"
	}
	//fmt.Println("file_ext="+file_ext)

	/*
		orig_filename := strings.TrimSuffix(
			filepath.Base(c_path),
			filepath.Ext(c_path),
		)
	*/
	//fmt.Println("orig_filename="+orig_filename)

	is_match, _ := regexp.MatchString("^(jpg|jpeg|png|gif)$", file_ext)
	if !is_match {
		LogPrefix(c, "500", filepath.Join("/__resize/", c_path)+" Only for JPEG, PNG and GIF images")
		return c.JSON(fiber.Map{
			"code": 500,
			"file": filepath.Join("/__resize/", c_path),
			"msg":  "Only for JPEG, PNG and GIF images",
		}, "application/json")
	}

	hash_name := md5.Sum([]byte(filepath.Join(arg_fold, c_path) + modtime_human + size_human + c_width + "x" + c_height))
	hex_name := hex.EncodeToString(hash_name[:])

	//fmt.Println("hex_name="+hex_name)

	if _, err := os.Stat(filepath.Join(homepath, ".httphere", "thumb", hex_name)); err == nil {

		LogPrefix(c, "200", "SendFile from cache "+filepath.Join(c_path))
		return c.SendFile(filepath.Join(homepath, ".httphere", "thumb", hex_name), false)

	} else if errors.Is(err, os.ErrNotExist) {

		//fmt.Println("make new thumb="+filepath.Join( c_path))

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

			src, _, err = exiffix.Decode(input)
			if err != nil {
				log.Fatal(err)
			}
			/*
			       src, err = jpeg.Decode(input)
			   	if err != nil {
			   		log.Fatal(err)
			   	}*/
		}

		if file_ext == "gif" {
			src, err = gif.Decode(input)
			if err != nil {
				log.Fatal(err)
			}
		}

		ratio := (float64)(src.Bounds().Max.Y) / (float64)(src.Bounds().Max.X)
		i_height = int(math.Round(float64(i_width) * ratio))

		// Set the expected size that you want:
		//dst := image.NewRGBA(image.Rect(0, 0, src.Bounds().Max.X/2, src.Bounds().Max.Y/2))

		//dst := image.NewRGBA(image.Rect(0, 0, src.Bounds().Max.X, src.Bounds().Max.Y))
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

		// Encode to `output`:
		//png.Encode(output, dst)

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

		LogPrefix(c, "200", "Resize and SendFile "+filepath.Join(c_path))

		return c.SendFile(filepath.Join(homepath, ".httphere", "thumb", hex_name), false)

	}

	//filepath.Join(homepath, ".httphere", "thumb")

	//return c.SendFile(filepath.Join(arg_fold, c_path), false)

	return c.JSON(fiber.Map{
		"code": 400,
		"file": filepath.Join(c_path),
		"msg":  "Bad request",
	}, "application/json")

}
