package main

import (
	"embed"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"

	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/western/http-here/controller"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	_ "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html/v2"
)

//go:embed view/*
var view_fs embed.FS

//go:embed assets/*
var embedDirStatic embed.FS

func main() {

	arg_port := flag.Int("port", 8000, "change default listen port")
	arg_user := flag.String("user", "", "login for user basic auth")
	arg_password := flag.String("password", "", "password for user basic auth")
	arg_help := flag.Bool("help", false, "show help")
	flag.Parse()

	if *arg_help {

		inf := []string{
			``,
			`v1.0.3`,
			``,
			`usage: http-here [options] [path]`,
			``,
			`options:`,
			`     --port        Port to use. [8000]`,
			``,
			`     --user        User for basic authorization.`,
			`     --password    Password for basic authorization.`,
			``,
			//`     --basic       Set basic auth and generate several accounts every time.`,
			``,
			//`     --upload-disable`,
			//`     --folder-make-disable`,
			``,
			//`     --tls`,
		}

		fmt.Println(strings.Join(inf[:], "\n"))
		return
	}
    
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	arg_fold := filepath.Dir(ex)

	if len(flag.Args()) > 0 {
		arg_fold = flag.Args()[0]
	}
	arg_fold = filepath.Clean(arg_fold)

	if fileInfo, err := os.Stat(arg_fold); err == nil {

		if !fileInfo.IsDir() {
			fmt.Println(arg_fold + " is not folder")
			return
		}

	} else if errors.Is(err, os.ErrNotExist) {
		fmt.Println(arg_fold + " is not exist")
		return
	}
	
	if strings.HasPrefix(arg_fold, "/etc") {
	    fmt.Println("You can not serve /etc folder")
		return
	}

	os.Setenv("arg_fold", arg_fold)

	//engine := html.New("./view", ".html")
	engine := html.NewFileSystem(http.FS(view_fs), ".html")

	config := fiber.Config{
		Prefork:               false,
		DisableStartupMessage: true,
		ServerHeader:          "",
		Views:                 engine,
		BodyLimit:             7 * 1024 * 1024 * 1024,
	}

	app := fiber.New(config)

	app.Use(cors.New(cors.Config{
		//AllowOrigins: "*",
		AllowCredentials: true,
		AllowOriginsFunc: func(origin string) bool {
			return true
		},
		AllowMethods:  "*",
		AllowHeaders:  "*",
		ExposeHeaders: "*",
	}))

    /*
	app.Use(logger.New(logger.Config{
		Format:     "[${time}] [${ip}] ${status} ${method} ${path}  ${latency}\n",
		TimeFormat: "2006-01-02 15:04:05",
	}))
	*/

	if len(*arg_user) > 0 && len(*arg_password) > 0 {

		app.Use(basicauth.New(basicauth.Config{

			Authorizer: func(user, pass string) bool {
				if user == *arg_user && pass == *arg_password {
					return true
				}

				return false
			},
			//ContextUsername: "_user",
            //ContextPassword: "_pass",
		}))
	}

	app.Use("/__assets", filesystem.New(filesystem.Config{
		Root:       http.FS(embedDirStatic),
		PathPrefix: "",
		Browse:     true,
	}))

	app.Options("/*", controller.OptionsAll)
	app.Get("/*", controller.GetAll)

	app.Post("/api/upload", controller.PostUpload)
	app.Post("/api/folder", controller.PostFolder)

	fmt.Println("")
	fmt.Println("Server port: " + strconv.Itoa(*arg_port))

	fmt.Println("")
	ifaces, err := net.Interfaces()
	if err != nil {
		fmt.Print(fmt.Errorf("localAddresses: %+v\n", err.Error()))
		return
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			fmt.Print(fmt.Errorf("localAddresses: %+v\n", err.Error()))
			continue
		}
		for _, a := range addrs {
			switch v := a.(type) {

			case *net.IPNet:
				//fmt.Printf("%v : %s [%v/%v]\n", i.Name, v, v.IP, v.Mask)
				//fmt.Printf("%v \n", v.IP)
				if v.IP.To4() != nil {
					//fmt.Println( "yes, ipv4" )
					fmt.Println("http://" + v.IP.String() + ":" + strconv.Itoa(*arg_port))
				}
			}

		}
	}
	fmt.Println("")

	fmt.Println("Serve folder: " + arg_fold)
	fmt.Println("")

	log.Fatal(app.Listen(":" + strconv.Itoa(*arg_port)))
}




