


package controller

import (
    _ "fmt"
    "os"
    "path/filepath"
    "errors"
    "html/template"
    "log"
    "strings"
    "net/url"


    "github.com/gofiber/fiber/v2"
)






func OptionsAll(c *fiber.Ctx) error {

    return c.JSON(fiber.Map{
        "code": 200,
        "method": "OPTIONS",
    }, "application/json")
}



func GetAll(c *fiber.Ctx) error {

    //fmt.Println("GetAll arg_fold=" + os.Getenv("arg_fold"))

    /*
    ex, err := os.Executable()
    if err != nil {
        panic(err)
    }
    exPath := filepath.Dir(ex)
    exPath = "/home/andrew/Code/http-here/public"
    fmt.Println("exPath="+ exPath)
    */
    arg_fold := os.Getenv("arg_fold")

    //fmt.Println("c.path="+c.Path())


    c_path, err := url.QueryUnescape(c.Path())
    if err != nil {
        log.Fatal(err)
        return c.JSON(fiber.Map{
            "code": 500,
            "err": err,
        }, "application/json")
    }
    //fmt.Println("c.path decoded="+c_path)

    c_path = filepath.Clean(c_path)
    //fmt.Println("c.path clean="+c_path)


    /*
    if( strings.HasPrefix(c.Path(), "/__assets/") ){
        fmt.Println("xxx")

        //return c.Status(fiber.StatusNotFound).SendString("Sorry can't find that!")

        return c.Render("view/assets/index.js", fiber.Map{

        })

    }
    */


    //fmt.Println( "try " + arg_fold + c_path )

    if fileInfo, err := os.Stat(arg_fold + c_path); err == nil {
        // path/to/whatever exists



        if fileInfo.IsDir() {
            //fmt.Println( arg_fold + c.Path() + " is exists and DIR" )

            breadcrumb := ""
            folderlist := ""
            filelist := ""




            res1 := strings.Split(c_path, "/")
            pt := ""
            for indx, el := range res1 {
                if( indx == 0 ){
                    continue;
                }
                pt += "/" + el
                breadcrumb += `<li class="breadcrumb-item"><a class="nodecor" href="`+pt+`">`+el+`</a></li>`;
            }


            entries, err := os.ReadDir( arg_fold + c_path )
            if err != nil {
                log.Fatal(err)
            }

            //fmt.Println(entries);

            if len(entries) == 0 {
                filelist = "Empty folder";
            }

            for _, e := range entries {
                //fmt.Println(e.Name())

                //fmt.Println( "= "+filepath.Join(arg_fold, c_path, e.Name()) )





                if fileInfo2, _ := os.Stat(  filepath.Join(arg_fold, c_path, e.Name())   ); err == nil {

                    if fileInfo2.IsDir() {
                        folderlist += `
                            <a href="`+filepath.Join(c_path, e.Name())+`" class="list-group-item list-group-item-action fold"><i class="bi bi-folder"></i> `+e.Name()+`</a>

                        `;
                    }else{
                        filelist += `
                            <a  href="`+filepath.Join(c_path, e.Name())+`" class="list-group-item list-group-item-action file"><i class="bi bi-file-earmark"></i> `+e.Name()+`</a>

                        `;
                    }
                }

            }






            return c.Render("view/index", fiber.Map{
                "Breadcrumb": template.HTML(breadcrumb),
                "Filelist": template.HTML(folderlist + filelist),
                "files_count_max": 20,
                "fieldSize_max": 7*1024*1024*1024,
                "fieldSize_max_human": "7 Gb",
            }, "view/layout")






        }else{
            //fmt.Println( arg_fold + c_path + " is exists and FILE" )

            return c.SendFile(arg_fold + c_path, false)
        }

    } else if errors.Is(err, os.ErrNotExist) {
        // path/to/whatever does *not* exist
        //return c.Status(fiber.StatusNotFound).SendString("Sorry can't find that!")

        //fmt.Println( arg_fold + c_path + " is not exist" )


        return c.Status(fiber.StatusNotFound).Render("view/404", fiber.Map{
            "File": c_path,
        }, "view/layout")
    }



    return c.Status(fiber.StatusNotFound).Render("view/404", fiber.Map{}, "view/layout")


    //c.JSON(http.StatusOK, gin.H{"list": list})
    //return c.Status(200).JSON(list)
    //return c.Status(200).JSON("str2")
}


