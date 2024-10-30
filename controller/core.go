package controller

import (
	"errors"
	_ "fmt"
	"html/template"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func OptionsAll(c *fiber.Ctx) error {

	return c.JSON(fiber.Map{
		"code":   200,
		"method": "OPTIONS",
	}, "application/json")
}

func GetAll(c *fiber.Ctx) error {

	arg_fold := os.Getenv("arg_fold")
	arg_upload_disable := os.Getenv("arg_upload_disable")
	arg_folder_make_disable := os.Getenv("arg_folder_make_disable")

	//log.Println("arg_upload_disable="+arg_upload_disable)
	//log.Println("arg_folder_make_disable="+arg_folder_make_disable)

	c_path, err := url.QueryUnescape(c.Path())
	if err != nil {
		/*
			log.Fatal(err)
			return c.JSON(fiber.Map{
				"code": 500,
				"err":  err,
			}, "application/json")
		*/
		log.Println(err)
		LogPrefix(c, "500", "Error "+filepath.Join(arg_fold, c_path))
		return c.Status(fiber.StatusNotFound).Render("view/500", fiber.Map{}, "view/layout")
	}

	//fmt.Println("c_path1="+c_path)
	c_path = CleanDirtyPath(c_path)
	//c_path = filepath.Clean(c_path)
	//fmt.Println("c_path2="+c_path)
	//fmt.Println("join="+filepath.Join(arg_fold, c_path))

	if fileInfo, err := os.Stat(filepath.Join(arg_fold, c_path)); err == nil {

		if fileInfo.IsDir() {

			LogPrefix(c, "200", "Dir "+filepath.Join(arg_fold, c_path))

			breadcrumb := ""
			folderlist := ""
			filelist := ""

			res1 := strings.Split(c_path, "/")
			pt := ""
			for indx, el := range res1 {
				if indx == 0 {
					continue
				}
				pt += "/" + el
				breadcrumb += `<li class="breadcrumb-item"><a class="nodecor" href="` + pt + `">` + el + `</a></li>`
			}

			entries, err := os.ReadDir(filepath.Join(arg_fold, c_path))
			if err != nil {
				//log.Fatal(err)
				log.Println(err)
				LogPrefix(c, "500", "Error "+filepath.Join(arg_fold, c_path))
				return c.Status(fiber.StatusNotFound).Render("view/500", fiber.Map{}, "view/layout")
			}

			if len(entries) == 0 {
				filelist = "Empty folder"
			}

			for _, e := range entries {

				if fileInfo2, _ := os.Stat(filepath.Join(arg_fold, c_path, e.Name())); err == nil {

					modtime := fileInfo2.ModTime()
					modtime_human := modtime.Format("2006-01-02 15:04:05")

					size := fileInfo2.Size()
					size_human := prettyByteSize(size)

					if fileInfo2.IsDir() {
						folderlist += `
                            
                            <!--a href="` + filepath.Join(c_path, e.Name()) + `" class="list-group-item list-group-item-action fold"><i class="bi bi-folder"></i> ` + e.Name() + `</a-->
                            
                            <a href="` + filepath.Join(c_path, e.Name()) + `" class="list-group-item list-group-item-action fold">
                                <div class="d-flex w-100 justify-content-between">
                                    <h5 class="mb-1"><i class="bi bi-folder"></i> ` + e.Name() + `</h5>
                                    <!--small class="text-muted">` + size_human + ` </small-->
                                </div>
                                <!--p class="mb-1">Some placeholder content in a paragraph.</p-->
                                <!--small class="text-muted">` + modtime_human + `</small-->
                            </a>

                        `
					} else {
						filelist += `
                            
                            <!--a  href="` + filepath.Join(c_path, e.Name()) + `" class="list-group-item list-group-item-action file"><i class="bi bi-file-earmark"></i> ` + e.Name() + `</a-->
                            
                            <a href="` + filepath.Join(c_path, e.Name()) + `" class="list-group-item list-group-item-action file">
                                <div class="d-flex w-100 justify-content-between">
                                    <h6 class="mb-1"><i class="bi bi-file-earmark"></i> ` + e.Name() + `</h6>
                                    <small class="text-muted">` + size_human + ` </small>
                                </div>
                                <!--p class="mb-1">Some placeholder content in a paragraph.</p-->
                                <small class="text-muted">` + modtime_human + `</small>
                            </a>

                        `
					}
				}

			}

			return c.Render("view/index", fiber.Map{

				"Breadcrumb": template.HTML(breadcrumb),
				"Filelist":   template.HTML(folderlist + filelist),

				"files_count_max":     20,
				"fieldSize_max":       7 * 1024 * 1024 * 1024,
				"fieldSize_max_human": "7 Gb",

				"arg_upload_disable":      arg_upload_disable,
				"arg_folder_make_disable": arg_folder_make_disable,
			}, "view/layout")

		} else {

			LogPrefix(c, "200", "SendFile "+filepath.Join(arg_fold, c_path))

			return c.SendFile(filepath.Join(arg_fold, c_path), false)
		}

	} else if errors.Is(err, os.ErrNotExist) {

		LogPrefix(c, "404", filepath.Join(arg_fold, c_path))

		return c.Status(fiber.StatusNotFound).Render("view/404", fiber.Map{
			"File": c_path,
		}, "view/layout")
	}

	return c.Status(fiber.StatusNotFound).Render("view/404", fiber.Map{}, "view/layout")

}
