package controller

import (
	"errors"
	_ "fmt"
	"html/template"
	"log"
	"net/url"
	"os"
	"path/filepath"
	_ "reflect"
	"regexp"
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

	arg_fold := ""
	arg_fold = c.Locals("arg_fold").(string)

	arg_upload_disable := ""
	if c.Locals("arg_upload_disable") != nil {
		arg_upload_disable = c.Locals("arg_upload_disable").(string)
	}

	arg_folder_make_disable := ""
	if c.Locals("arg_folder_make_disable") != nil {
		arg_folder_make_disable = c.Locals("arg_folder_make_disable").(string)
	}

	arg_extend_mode := ""
	if c.Locals("arg_extend_mode") != nil {
		arg_extend_mode = c.Locals("arg_extend_mode").(string)
	}

	//arg_fold := "/tmp"
	//arg_upload_disable := ""
	//arg_folder_make_disable := ""

	c_path, err := url.QueryUnescape(c.Path())
	if err != nil {

		log.Println(err)
		LogPrefix(c, "500", "Error "+filepath.Join(arg_fold, c_path))
		return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout")
	}

	c_path = CleanDirtyPath(c_path)

	if fileInfo, err := os.Stat(filepath.Join(arg_fold, c_path)); err == nil {

		if fileInfo.IsDir() {

			// check index.html inside
			if _, err := os.Stat(filepath.Join(arg_fold, c_path, "index.html")); err == nil {

				LogPrefix(c, "200", "Index file found for path '"+c_path+"', SendFile "+filepath.Join(arg_fold, c_path, "index.html"))
				return c.SendFile(filepath.Join(arg_fold, c_path, "index.html"), false)
			}

			LogPrefix(c, "200", "Dir "+filepath.Join(arg_fold, c_path))

			breadcrumb := ""

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

				log.Println(err)
				LogPrefix(c, "500", "Error "+filepath.Join(arg_fold, c_path))
				return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout")
			}

			//folderlist := ""
			//filelist := ""

			fl := template.HTML("")
			//var fl template.HTML

			//fmt.Println(reflect.TypeOf(fl))
			//fmt.Println(reflect.TypeOf(entries))

			if arg_extend_mode == "1" {

				mode := c.Cookies("mode")

				q_mode := c.Query("mode")
				if len(q_mode) > 0 {
					mode = q_mode
				}

				if mode == "thumbnail" {
					fl = listGenerateViewThumbnails(arg_fold, c_path, entries)
				} else {
					fl = listGenerateViewExtended(arg_fold, c_path, entries)
				}

				cookie := new(fiber.Cookie)
				cookie.Name = "mode"
				cookie.Value = mode
				c.Cookie(cookie)

			} else {

				fl = listGenerateView(arg_fold, c_path, entries)

			}

			return c.Render("view/index", fiber.Map{

				"Breadcrumb": template.HTML(breadcrumb),
				"Filelist":   fl,

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

func listGenerateView(arg_fold string, c_path string, entries []os.DirEntry) template.HTML {

	folderlist := ""
	filelist := ""

	if len(entries) == 0 {
		filelist = "Empty folder"
	}

	for _, e := range entries {

		if fileInfo2, err := os.Stat(filepath.Join(arg_fold, c_path, e.Name())); err == nil {

			modtime := fileInfo2.ModTime()
			modtime_human := modtime.Format("2006-01-02 15:04:05")

			size := fileInfo2.Size()
			size_human := prettyByteSize(size)

			if fileInfo2.IsDir() {
				folderlist += `
                    
                    <a href="` + filepath.Join(c_path, e.Name()) + `" class="list-group-item list-group-item-action fold">
                        <div class="d-flex w-100 justify-content-between">
                            <h5 class="mb-1"><i class="bi bi-folder"></i> ` + e.Name() + `</h5>
                            <!--small class="text-muted">` + size_human + ` </small-->
                        </div>
                    </a>
                    
                `
			} else {
				filelist += `
                    
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

	fl := template.HTML("<div class='list-group'>" + folderlist + filelist + "</div>")

	return fl
}

func listGenerateViewExtended(arg_fold string, c_path string, entries []os.DirEntry) template.HTML {

	folderlist := ""
	filelist := ""

	if len(entries) == 0 {
		filelist += `
            
            <tr>
              <td colspan="5">Empty folder</td>
            </tr>
        `
	}

	for _, e := range entries {

		if fileInfo2, err := os.Stat(filepath.Join(arg_fold, c_path, e.Name())); err == nil {

			modtime := fileInfo2.ModTime()
			modtime_human := modtime.Format("2006-01-02 15:04:05")

			size := fileInfo2.Size()
			size_human := prettyByteSize(size)

			if fileInfo2.IsDir() {
				folderlist += `
                    
                    <tr>
                      <th scope="row" class="text-center"><input class="form-check-input" type="checkbox" name="fold" value="` + e.Name() + `" ></th>
                      <td ><i class="bi bi-folder"></i> <a class="nodecor" href="` + filepath.Join(c_path, e.Name()) + `">` + e.Name() + `</a></td>
                      
                      <td class="d-none d-sm-none d-md-none d-lg-table-cell d-xl-table-cell"></td>
                      <td class="d-none d-sm-none d-md-none d-lg-table-cell d-xl-table-cell"></td>
                      <td class="text-center">
                        <a class="del" href="javascript:void(0)" data-name="` + e.Name() + `"><i class="bi bi-x-lg"></i></a>
                      </td>
                    </tr>

                    
                `
			} else {
				filelist += `
                    
                    <tr>
                      <th scope="row" class="text-center"><input class="form-check-input" type="checkbox" name="file" value="` + e.Name() + `" ></th>
                      <td ><a class="nodecor" href="` + filepath.Join(c_path, e.Name()) + `">` + e.Name() + `</a></td>
                      
                      <td class="d-none d-sm-none d-md-none d-lg-table-cell d-xl-table-cell">` + size_human + `</td>
                      <td class="d-none d-sm-none d-md-none d-lg-table-cell d-xl-table-cell">` + modtime_human + `</td>
                      <td class="text-center">
                        <a class="del" href="javascript:void(0)" data-name="` + e.Name() + `"><i class="bi bi-x-lg"></i></a>
                      </td>
                    </tr>
                    
                `
			}
		}

	}

	fl := template.HTML(`
	    
	    
	    <div class="container">
	    
	    
	        <div class="row " style="padding:10px 0;">

                <div class="col-4" >

                </div>

                <div class="col-8 text-end">
                    
                    <div class="btn-group btn-group-sm" role="group" aria-label="">
                        <a  class="btn btn-outline-secondary " href="?mode=thumbnail" title="Folder and thumb"><i class="bi bi-image"></i> Thumbnails</a>
                        <a  class="btn btn-outline-secondary active" href="?mode=list" title="List layout"><i class="bi bi-list"></i> List layout</a>
                    </div>
                    
                </div>
                
            </div>
	    
	    
            <table class="table file-table table-hover ">
              <thead>
                <tr>
                  <th class="text-center"><input class="form-check-input head-chk" type="checkbox" value="" ></th>
                  <th >name</th>
                  
                  <th class="col d-none d-sm-none d-md-none d-lg-table-cell d-xl-table-cell"></th>
                  <th class="col d-none d-sm-none d-md-none d-lg-table-cell d-xl-table-cell"></th>
                  <th class="text-center">del</th>
                </tr>
              </thead>
              <tbody>
              
                ` + folderlist + `
                
                ` + filelist + `
                
                
              </tbody>
            </table>
            
            <div class="d-grid gap-2 d-md-block">
                <button type="button" class="btn btn-outline-danger btn-sm" id="group_del">Delete group</button>
                <button type="button" class="btn btn-outline-secondary btn-sm" id="group_zip">Zip and download group</button>
            </div>
            
        </div>
	
	`)

	return fl

}

func listGenerateViewThumbnails(arg_fold string, c_path string, entries []os.DirEntry) template.HTML {

	folderlist := ""
	filelist := ""

	if len(entries) == 0 {
		filelist += `
            
            <tr>
              <td colspan="5">Empty folder</td>
            </tr>
        `
	}

	for _, e := range entries {

		if fileInfo2, err := os.Stat(filepath.Join(arg_fold, c_path, e.Name())); err == nil {

			modtime := fileInfo2.ModTime()
			modtime_human := modtime.Format("2006-01-02 15:04:05")

			size := fileInfo2.Size()
			size_human := prettyByteSize(size)

			if fileInfo2.IsDir() {
				folderlist += `
                    
                    <!--
                    <tr>
                      <th scope="row" class="text-center"><input class="form-check-input" type="checkbox" name="fold" value="` + e.Name() + `" ></th>
                      <td ><i class="bi bi-folder"></i> <a class="nodecor" href="` + filepath.Join(c_path, e.Name()) + `">` + e.Name() + `</a></td>
                      
                      <td class="d-none d-sm-none d-md-none d-lg-table-cell d-xl-table-cell"></td>
                      <td class="d-none d-sm-none d-md-none d-lg-table-cell d-xl-table-cell"></td>
                      <td class="text-center">
                        <a class="del" href="javascript:void(0)" data-name="` + e.Name() + `"><i class="bi bi-x-lg"></i></a>
                      </td>
                    </tr>
                    -->
                    
                    <div class="col">
                        <div class="card shadow-sm">
                            <a href="` + filepath.Join(c_path, e.Name()) + `">
                            <svg class="bd-placeholder-img card-img-top" width="100%" height="225" xmlns="http://www.w3.org/2000/svg" role="img"  preserveAspectRatio="xMidYMid slice" focusable="false">
                                <title></title>
                                <rect width="100%" height="100%" fill="#55595c"></rect><text x="50%" y="50%" fill="#eceeef" dy=".3em">Folder</text>
                            </svg>
                            </a>
                            <div class="card-body">
                                <p class="card-text text-truncate">
                                    <b>` + e.Name() + `</b><br>
                                    <span class="fw-lighter">` + modtime_human + `</span>
                                </p>
                                <div class="d-flex justify-content-between align-items-center">
                                    <div class="btn-group">
                                        
                                        <div class="input-group-text">
                                            <input class="form-check-input mt-0" type="checkbox" name="fold" value="` + e.Name() + `" >
                                        </div>
                                        
                                        
                                        <!--a type="button" class="btn btn-sm btn-outline-secondary" href="` + filepath.Join(c_path, e.Name()) + `">View</a-->
                                        <!--a type="button" class="btn btn-sm btn-outline-danger">Del</a-->
                                    </div>
                                    <small class="text-body-secondary">
                                        <div class="btn-group">
                                            
                                            <a type="button" class="del btn btn-sm btn-outline-danger" href="javascript:void(0)" data-name="` + e.Name() + `" ><i class="bi bi-x-lg"></i></a>
                                            
                                            
                                            
                                        </div>
                                        
                                    </small>
                                </div>
                            </div>
                        </div>
                    </div>
                    
                `
			} else {

				ext := filepath.Ext(e.Name())
				ext = strings.ToLower(ext)
				ext = strings.Replace(ext, ".", "", -1)

				//fmt.Println("ext="+ext)

				img_preview := `
		            <svg class="bd-placeholder-img card-img-top" width="100%" height="225" xmlns="http://www.w3.org/2000/svg" role="img"  preserveAspectRatio="xMidYMid slice" focusable="false">
                        <title></title>
                        <rect width="100%" height="100%" fill="#55595c"></rect><text x="50%" y="50%" fill="#eceeef" dy=".3em">File</text>
                    </svg>
				`

				//r, _ := regexp.Compile("^(jpg|jpeg|png|gif)$")
				is_match, _ := regexp.MatchString("^(jpg|jpeg|png|gif)$", ext)

				//if ext == "jpg" {
				if is_match {

					img_preview = `
    		            <div class="bd-placeholder-img card-img-top" width="100%" height="225" style="height:225px; background: no-repeat center/80% url(` + filepath.Join(c_path, e.Name()) + `); background-size: cover;  " >
    		            </div>
    				`
				}

				filelist += `
                    
                    <!--
                    <tr>
                      <th scope="row" class="text-center"><input class="form-check-input" type="checkbox" name="file" value="` + e.Name() + `" ></th>
                      <td ><a class="nodecor" href="` + filepath.Join(c_path, e.Name()) + `">` + e.Name() + `</a></td>
                      
                      <td class="d-none d-sm-none d-md-none d-lg-table-cell d-xl-table-cell">` + size_human + `</td>
                      <td class="d-none d-sm-none d-md-none d-lg-table-cell d-xl-table-cell">` + modtime_human + `</td>
                      <td class="text-center">
                        <a class="del" href="javascript:void(0)" data-name="` + e.Name() + `"><i class="bi bi-x-lg"></i></a>
                      </td>
                    </tr>
                    -->
                    
                    <div class="col">
                        <div class="card shadow-sm">
                            <a href="` + filepath.Join(c_path, e.Name()) + `">
                                ` + img_preview + `
                            </a>
                            <div class="card-body">
                                <p class="card-text text-truncate" >
                                    <b>` + e.Name() + `</b><br>
                                    <span class="fw-lighter">` + modtime_human + `</span>
                                </p>
                                <div class="d-flex justify-content-between align-items-center">
                                    <div class="btn-group">
                                        
                                        
                                        <div class="input-group-text">
                                            <input class="form-check-input mt-0" type="checkbox" name="file" value="` + e.Name() + `" >
                                        </div>
                                        
                                        
                                        
                                        
                                        <!--a type="button" class="btn btn-sm btn-outline-secondary" href="` + filepath.Join(c_path, e.Name()) + `">View</a-->
                                        <!--a type="button" class="btn btn-sm btn-outline-danger">Del</a-->
                                    </div>
                                    <small class="text-body-secondary">
                                        <div class="btn-group">
                                            <div class="input-group-text font-reset" >
                                                ` + size_human + `
                                            </div>
                                            
                                            
                                            
                                            <a type="button" class="del btn btn-sm btn-outline-danger" href="javascript:void(0)" data-name="` + e.Name() + `" ><i class="bi bi-x-lg"></i></a>
                                            
                                        </div>
                                    </small>
                                </div>
                            </div>
                        </div>
                    </div>
                    
                    
                    
                `
			}
		}

	}

	fl := template.HTML(`
	
	    
        
        <div class="container">
            
            <div class="row " style="padding:0 0 20px 0;">

                <div class="col-4" >
                    
                    <div class="btn-group">
                        
                        
                        <div class="input-group-text">
                            <input class="form-check-input mt-0 head-chk" type="checkbox"  >
                        </div>
                        
                    </div>
                    
                    
                    
                </div>

                <div class="col-8 text-end">
                    
                    <div class="btn-group btn-group-sm" role="group" aria-label="">
                        <a  class="btn btn-outline-secondary active" href="?mode=thumbnail" title="Folder and thumb"><i class="bi bi-image"></i> Thumbnails</a>
                        <a  class="btn btn-outline-secondary " href="?mode=list" title="List layout"><i class="bi bi-list"></i> List layout</a>
                    </div>
                    
                </div>
                
            </div>
            
            
            <div class="row row-cols-1 row-cols-sm-2 row-cols-md-3 g-3">
                
                <!--
                <div class="col">
                    <div class="card shadow-sm">
                        <svg class="bd-placeholder-img card-img-top" width="100%" height="225" xmlns="http://www.w3.org/2000/svg" role="img" aria-label="Placeholder: Thumbnail" preserveAspectRatio="xMidYMid slice" focusable="false">
                            <title>Placeholder</title>
                            <rect width="100%" height="100%" fill="#55595c"></rect><text x="50%" y="50%" fill="#eceeef" dy=".3em">Thumbnail</text>
                        </svg>
                        <div class="card-body">
                            <p class="card-text">This is a wider card with supporting text below as a natural lead-in to additional content. This content is a little bit longer.</p>
                            <div class="d-flex justify-content-between align-items-center">
                                <div class="btn-group">
                                    <button type="button" class="btn btn-sm btn-outline-secondary">View</button>
                                    <button type="button" class="btn btn-sm btn-outline-secondary">Edit</button>
                                </div>
                                <small class="text-body-secondary">9 mins</small>
                            </div>
                        </div>
                    </div>
                </div>
                -->
                
                ` + folderlist + `
            
                ` + filelist + `

                
            </div>
        </div>
        
        <div class="d-grid gap-2 d-md-block" style="padding: 50px 0;">
            <button type="button" class="btn btn-outline-danger btn-sm" id="group_del">Delete group</button>
            <button type="button" class="btn btn-outline-secondary btn-sm" id="group_zip">Zip and download group</button>
        </div>
        
	
	`)

	return fl

}
