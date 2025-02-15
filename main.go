package main

import (
	"embed"
	"errors"
	"flag"
	"fmt"
	_ "html/template"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fatih/color"

	"github.com/western/http-here/conf"
	"github.com/western/http-here/controller"
	"github.com/western/http-here/model"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/template/html/v2"
)

//go:embed view/*
var view_fs embed.FS

//go:embed assets/*
var embedDirStatic embed.FS

func main() {

	arg_help := flag.Bool("help", false, "Show help")

	arg_port := flag.Int("port", 8000, "Port to use")
	arg_tls := flag.Bool("tls", false, "Start HTTPS (need easyrsa linux package)")
	arg_tls_debug := flag.Bool("tls-debug", false, "Start HTTPS with verbosity")

	arg_user := flag.String("user", "", "Login for user basic auth")
	arg_password := flag.String("password", "", "Password for user basic auth")
	arg_basic := flag.Bool("basic", false, "Set basic auth and generate several accounts every time")

	arg_upload_disable := flag.Bool("upload-disable", false, "Disable upload API and form controller")
	arg_folder_make_disable := flag.Bool("folder-make-disable", false, "Disable make folder API and form controller")
	arg_index_disable := flag.Bool("index-disable", false, "Disable current folder read")

	arg_extend_mode := flag.Bool("extend-mode", false, "Enable delete mechanics. Be very careful. It disabled by default.")
	arg_crypt := flag.Bool("crypt", false, "Enable file crypt support.")
	arg_spa := flag.Bool("spa", false, "Enable frontend SPA (Single Page Application)")

	arg_prefork := flag.Bool("prefork", false, "Enable spawn multiple processes")
	//arg_prepare_thumbnails := flag.Bool("prepare-thumbnails", false, "Run and make thumbnails for target folders.")

	flag.Parse()

	if *arg_tls_debug {
		*arg_tls = *arg_tls_debug
	}

	db, err := model.ConnectToSQLite()
	if err != nil {
		panic(err)
	}
	//defer db.Close()

	model.EventLogAdd(db, "INIT", "run")

	go model.FileChkAsync(db)

	green_clr := color.New(color.FgGreen).SprintFunc()
	white_clr := color.New(color.Bold, color.FgWhite).SprintFunc()

	if *arg_help {

		inf := []string{
			``,
			conf.Version,
			``,
			`usage: ` + green_clr(`http-here`) + ` [options] [path]`,
			``,
			`options:`,
			``,
			`     --port                    Port to use [8000]`,
			`     --tls                     Start HTTPS (need easy-rsa linux package).`,
			`     --tls-debug               Start HTTPS with verbosity`,
			``,
			``,
			`     --user                    Login for basic authorization.`,
			`     --password                Password for basic authorization.`,
			``,
			`     --basic                   Set basic auth and generate several accounts every time.`,
			``,
			``,
			`     --upload-disable          Disable upload API and form controller.`,
			`     --folder-make-disable     Disable make folder API and form controller.`,
			`     --index-disable           Disable current folder read.`,
			``,
			``,
			`     --extend-mode             Enable delete mechanics. Be very careful. It disabled by default.`,
			``,
			`     --prefork                 Enable spawn multiple processes.`,
			``,
			`     --crypt                   Enable file crypt support.`,
			``,
			`     --spa                     Enable frontend SPA (Single Page Application).`,
			``,
			``,
			`examples:`,
			``,
			`     The safest run`,
			`                        ` + green_clr(`http-here`) + ` --tls --basic ` + white_clr(`/some/path`),
			``,
			`     Only share`,
			`                        ` + green_clr(`http-here`) + ` --upload-disable --folder-make-disable ` + white_clr(`/tmp/fold`),
			``,
			`     Powerful`,
			`                        ` + green_clr(`http-here`) + ` --tls --user ` + white_clr(`user`+controller.RandStringRunes(2)) + ` --password ` + white_clr(controller.RandStringRunes(12)) + ` --prefork ` + white_clr(`/tmp/fold`),
			``,
		}

		fmt.Println(strings.Join(inf[:], "\n"))
		return
	}

	arg_fold, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		return
	}

	if len(flag.Args()) > 0 {
		arg_fold = flag.Args()[0]
		if arg_fold, err = filepath.Abs(arg_fold); err != nil {
			fmt.Println(err)
			return
		}
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

	homepath, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("User homepath detect error: ", err)
		return
	}

	if _, err := os.Stat(filepath.Join(homepath, ".httphere", "thumb")); err != nil {

		if err2 := os.MkdirAll(filepath.Join(homepath, ".httphere", "thumb"), os.ModePerm); err2 != nil {

			fmt.Println(err2)
			return
		}
	}

	if !fiber.IsChild() {

		controller.WalkAndClearZeroFile(filepath.Join(homepath, ".httphere", "thumb"), 0)
		//go controller.WalkAndClearOld(filepath.Join(homepath, ".httphere", "thumb"))
	}

	/*
		if *arg_prepare_thumbnails && !fiber.IsChild() {

			fmt.Println()
			fmt.Println("  Run and make thumbnails for target folders")

			go controller.WalkAndMakeThumbnail(arg_fold, 0)
		}
	*/

	//engine := html.New("./view", ".html")
	engine := html.NewFileSystem(http.FS(view_fs), ".html")

	/*
			engine.AddFunc(
		        "unescape", func(s string) template.HTML {
		            return template.HTML(s)
		        },
		    )
	*/

	config := fiber.Config{
		Prefork:               *arg_prefork,
		DisableStartupMessage: true,
		ServerHeader:          "",
		Views:                 engine,
		BodyLimit:             14 * 1024 * 1024 * 1024,
	}

	app := fiber.New(config)

	app.Use(func(c *fiber.Ctx) error {

		c.Locals("arg_fold", arg_fold)

		if *arg_upload_disable {

			c.Locals("arg_upload_disable", "1")
		}
		if *arg_folder_make_disable {

			c.Locals("arg_folder_make_disable", "1")
		}
		if *arg_extend_mode {

			c.Locals("arg_extend_mode", "1")
		}
		if *arg_crypt {

			c.Locals("arg_crypt", "1")
		}
		if *arg_spa {

			c.Locals("arg_spa", "1")
		}

		return c.Next()
	})

	app.Use(favicon.New())

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

	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed, // 1
	}))

	cian_clr := color.New(color.FgCyan).SprintFunc()
	yellow_clr := color.New(color.FgYellow).SprintFunc()

	if *arg_basic && *arg_prefork {

		fmt.Println("")
		fmt.Println("You can not run --prefork with --basic. Use --user and --password instead.")
		return
	}

	if *arg_basic && !*arg_prefork {

		fmt.Println("")
		fmt.Println("  Basic auth set: ")

		var password_list = map[string]string{}

		for i := range 10 {

			login := "login" + strconv.Itoa(i) + controller.RandStringRunes(2)
			password := controller.RandStringRunes(16)

			password_list[login] = password

			fmt.Println("         " + login + "    " + password)
		}

		app.Use(basicauth.New(basicauth.Config{

			Users: password_list,

			Unauthorized: func(c *fiber.Ctx) error {

				controller.LogPrefix(c, "401", filepath.Join(arg_fold, c.Path()))

				c.Set(fiber.HeaderWWWAuthenticate, "Basic realm='Restricted'")
				return c.Status(fiber.StatusUnauthorized).Render("view/401", fiber.Map{}, "view/layout")
			},
		}))

	}

	if len(*arg_user) > 0 && len(*arg_password) > 0 {

		app.Use(basicauth.New(basicauth.Config{

			Authorizer: func(user, pass string) bool {
				if user == *arg_user && pass == *arg_password {
					return true
				}

				return false
			},
			Unauthorized: func(c *fiber.Ctx) error {

				controller.LogPrefix(c, "401", filepath.Join(arg_fold, c.Path()))

				c.Set(fiber.HeaderWWWAuthenticate, "Basic realm='Restricted'")
				return c.Status(fiber.StatusUnauthorized).Render("view/401", fiber.Map{}, "view/layout")
			},
		}))

		if !fiber.IsChild() {
			fmt.Println("")
			fmt.Println("  Basic auth set: " + cian_clr(*arg_user) + " " + cian_clr(*arg_password))
		}
	}

	app.Use("/__assets", filesystem.New(filesystem.Config{
		Root:       http.FS(embedDirStatic),
		PathPrefix: "",
		Browse:     false,
	}))

	if !fiber.IsChild() {

		if _, err3 := os.Stat(filepath.Join(homepath, ".httphere", "temp")); err3 != nil {

			fmt.Println("")
			//fmt.Println("  Make temp folder")

			if err4 := os.MkdirAll(filepath.Join(homepath, ".httphere", "temp"), os.ModePerm); err4 != nil {
				fmt.Println(err4)
				return
			}
		} else {

			fmt.Println("")
			//fmt.Println(yellow("  Clear temp folder"))

			if err := os.RemoveAll(filepath.Join(homepath, ".httphere", "temp")); err != nil {
				fmt.Println(err)
				return
			}

			if err4 := os.MkdirAll(filepath.Join(homepath, ".httphere", "temp"), os.ModePerm); err4 != nil {
				fmt.Println(err4)
				return
			}
		}
	}

	//app.Static("/__temp", filepath.Join(homepath, ".httphere", "temp"))

	if *arg_extend_mode {
		//app.Get("/__resize/:width/:height/*", controller.GetResize)
		app.Get("/__resize/*", controller.GetResize)
	}

	app.Get("/__temp/*", func(c *fiber.Ctx) error {

		c_path, err := url.QueryUnescape(c.Path())
		c_path = strings.TrimLeft(c_path, "/__temp")
		if err != nil {

			controller.LogPrefix(c, "500", "Error "+filepath.Join(homepath, ".httphere", "temp", c_path)+" "+err.Error())
			return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout")
		}

		c_path = controller.CleanDirtyPath(c_path)
		//fmt.Println("c_path=" + c_path)

		_, err = os.Stat(filepath.Join(homepath, ".httphere", "temp", c_path))

		if err != nil {
			controller.LogPrefix(c, "404", "'"+filepath.Join(homepath, ".httphere", "temp", c_path)+"' not exists")

			return c.JSON(fiber.Map{
				"code": 404,
				"msg":  filepath.Join(c_path) + " not exists",
			}, "application/json")

		}

		controller.LogPrefix(c, "200", "Temp get "+filepath.Join(homepath, ".httphere", "temp", c_path))

		return c.SendFile(filepath.Join(homepath, ".httphere", "temp", c_path))
	})

	app.Options("/*", controller.OptionsAll)

	if !*arg_index_disable {

		if *arg_extend_mode {
			app.Get("/__doc/*", controller.GetEditDoc)
			app.Get("/__code/*", controller.GetEditCode)
			app.Get("/__search/", controller.GetSearch)
			app.Get("/__convert/*", controller.GetConvert)
			app.Post("/api/edit", controller.PostEdit)
			app.Get("/api/list", controller.GetList)
		}
	}

	if !*arg_upload_disable {
		app.Post("/api/upload", controller.PostUpload)
	}

	if !*arg_folder_make_disable {
		app.Post("/api/folder", controller.PostFolder)
		app.Post("/api/file", controller.PostFile)
	}

	if *arg_extend_mode {
		app.Post("/api/delete", controller.PostDelete)
		app.Post("/api/move", controller.PostMove)
		app.Post("/api/copy", controller.PostCopy)
		app.Post("/api/rename", controller.PostRename)
		app.Post("/api/zip", controller.PostZip)

		app.Get("/api/search", controller.GetApiSearch)
	}

	if !*arg_index_disable {

		if *arg_spa {
			app.Get("/", controller.GetSpa)
			app.Get("/*", controller.GetAll)
		} else {
			app.Get("/*", controller.GetAll)
		}
	}

	app.Use(func(c *fiber.Ctx) error {

		controller.LogPrefix(c, "404", filepath.Join(arg_fold, c.Path()))

		return c.Status(fiber.StatusNotFound).Render("view/404", fiber.Map{}, "view/layout")
	})

	// /home/andrew/.httphere/easyrsa/pki/issued/server1.crt
	// /home/andrew/.httphere/easyrsa/pki/private/server1.key

	crt_filename := filepath.Join(homepath, ".httphere", "easyrsa", "pki", "issued", "server1.crt")
	key_filename := filepath.Join(homepath, ".httphere", "easyrsa", "pki", "private", "server1.key")

	crt_is_exists := false

	if _, err3 := os.Stat(crt_filename); err3 == nil {
		crt_is_exists = true
	}

	if *arg_tls && crt_is_exists && !fiber.IsChild() {

		fmt.Println("")
		fmt.Println("  Crt: " + yellow_clr(crt_filename))
		fmt.Println("  Key: " + yellow_clr(key_filename))
		fmt.Println("")
	}

	if *arg_tls && !crt_is_exists && !fiber.IsChild() {

		_, err1 := exec.Command("bash", "-c", "easyrsa  ").Output()
		if err1 != nil {

			fmt.Println("easyrsa error: " + err1.Error())
			return
		}

		if _, err3 := os.Stat(filepath.Join(homepath, ".httphere", "easyrsa")); err3 != nil {

			if err4 := os.MkdirAll(filepath.Join(homepath, ".httphere", "easyrsa"), os.ModePerm); err4 != nil {
				fmt.Println(err4)
				return
			}
		}

		_, err5 := exec.Command("bash", "-c", "cd "+filepath.Join(homepath, ".httphere", "easyrsa")).Output()
		if err5 != nil {
			fmt.Println(err5)
			return
		}

		cmd := exec.Command("bash", "-c", "easyrsa init-pki")
		cmd.Dir = filepath.Join(homepath, ".httphere", "easyrsa")
		out3, _ := cmd.Output()

		if *arg_tls_debug {
			fmt.Println(yellow_clr("--------------------------------------------------------------------------------------------------"))
			fmt.Printf(" %s\n", out3)
		}

		var_data := `
set_var EASYRSA_DN "cn_only"
set_var EASYRSA_KEY_SIZE 2048
set_var EASYRSA_REQ_CN   "ca@desec.example.com"
set_var EASYRSA_BATCH    "yes"
set_var EASYRSA_CA_EXPIRE 3650
set_var EASYRSA_CERT_EXPIRE 3650
        `

		f, err7 := os.OpenFile(filepath.Join(homepath, ".httphere", "easyrsa", "pki", "vars"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		if err7 != nil {
			fmt.Println(err7)
			return
		}
		defer f.Close()

		f.WriteString(var_data)
		f.Close()

		cmd2 := exec.Command("bash", "-c", "easyrsa build-ca nopass")
		cmd2.Dir = filepath.Join(homepath, ".httphere", "easyrsa")
		out4, _ := cmd2.Output()

		if *arg_tls_debug {
			fmt.Println(yellow_clr("--------------------------------------------------------------------------------------------------"))
			fmt.Printf(" %s\n", out4)
		}

		cmd3 := exec.Command("bash", "-c", "easyrsa --req-cn=ChangeMe build-client-full server1 nopass")
		cmd3.Dir = filepath.Join(homepath, ".httphere", "easyrsa")
		out5, _ := cmd3.Output()

		if *arg_tls_debug {
			fmt.Println(yellow_clr("--------------------------------------------------------------------------------------------------"))
			fmt.Printf(" %s\n", out5)

			fmt.Println(yellow_clr("--------------------------------------------------------------------------------------------------"))
			fmt.Println("")
		}

		fmt.Println(yellow_clr("  Generate new TLS keys"))
		fmt.Println("")
		fmt.Println("  Crt: " + yellow_clr(crt_filename))
		fmt.Println("  Key: " + yellow_clr(key_filename))
		fmt.Println("")
	}

	if !fiber.IsChild() {
		fmt.Println("")
		if *arg_tls {
			fmt.Println(yellow_clr("  TLS Server port " + strconv.Itoa(*arg_port)))
		} else {
			fmt.Println("  Server port " + cian_clr(strconv.Itoa(*arg_port)))
		}

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

					if v.IP.To4() != nil {

						if *arg_tls {
							fmt.Println("     https://" + v.IP.String() + ":" + cian_clr(strconv.Itoa(*arg_port)))
						} else {
							fmt.Println("     http://" + v.IP.String() + ":" + cian_clr(strconv.Itoa(*arg_port)))
						}
					}
				}

			}
		}

		fmt.Println("")

		fmt.Println("  Serve folder: " + cian_clr(arg_fold))
		fmt.Println("")
		fmt.Println(cian_clr("  [ Control + C ] ") + "Break Server")
		fmt.Println("")
	}

	if *arg_tls {

		log.Fatal(app.ListenTLS(":"+strconv.Itoa(*arg_port), crt_filename, key_filename))

	} else {

		log.Fatal(app.Listen(":" + strconv.Itoa(*arg_port)))
	}
}
