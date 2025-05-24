package app

import (
	"embed"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	_ "reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/fatih/color"
	//_ "github.com/gofiber/swagger"

	"github.com/western/http-here/internal/api"
	"github.com/western/http-here/internal/cert"
	"github.com/western/http-here/internal/conf"
	//_ "github.com/western/http-here/internal/docs"
	"github.com/western/http-here/internal/model"
	"github.com/western/http-here/internal/util"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/template/html/v2"

	"gorm.io/gorm"
)

//go:embed view/*
var view_fs embed.FS

//go:embed assets/*
var embedDirStatic embed.FS

func Core() {

	arg_help := flag.Bool("help", false, "Show help")

	arg_port := flag.Int("port", 8000, "Port to use")
	arg_tls := flag.Bool("tls", false, "Start HTTPS")
	arg_tls_debug := flag.Bool("tls-debug", false, "Start HTTPS with verbosity")

	arg_user := flag.String("user", "", "Login for user basic auth")
	arg_password := flag.String("password", "", "Password for user basic auth")
	arg_basic := flag.Bool("basic", false, "Set basic auth and generate several accounts every time")

	arg_upload_disable := flag.Bool("upload-disable", false, "Disable upload API and form controller")
	arg_folder_make_disable := flag.Bool("folder-make-disable", false, "Disable make folder API and form controller")
	arg_index_disable := flag.Bool("index-disable", false, "Disable current folder read")

	arg_extend_mode := flag.Bool("extend-mode", false, "Enable delete mechanics. Be very careful. It disabled by default.")
	arg_crypt := flag.Bool("crypt", false, "Enable encrypt file support.")
	arg_spa := flag.Bool("spa", false, "Enable frontend SPA (Single Page Application)")

	arg_prefork := flag.Bool("prefork", false, "Enable spawn multiple processes")

	arg_silence := flag.Bool("silence", false, "Disable all console messages")
	arg_nolog := flag.Bool("nolog", false, "Do not write any data to event_log table")

	arg_nodb := flag.Bool("nodb", false, "Disable database use")

	flag.Parse()

	if *arg_tls_debug {
		*arg_tls = *arg_tls_debug
	}

	green_clr := color.New(color.FgGreen).SprintFunc()
	white_clr := color.New(color.Bold, color.FgWhite).SprintFunc()

	if *arg_help {

		inf := []string{
			``,
			green_clr(conf.Version),
			``,
			`Simple zero-configuration command line http server with lightweight interface to work with files`,
			``,
			`usage: ` + green_clr(`http-here`) + ` [options] [path]`,
			``,
			`options:`,
			``,
			`     --port ` + white_clr(`[int]`) + `              Port to use [8000]`,
			`     --tls                     Start HTTPS`,
			``,
			``,
			`     --user ` + white_clr(`[str]`) + `              Login for basic authorization`,
			`     --password ` + white_clr(`[str]`) + `          Password for basic authorization`,
			``,
			`     --basic                   Set basic auth and generate several accounts every time`,
			``,
			``,
			`     --upload-disable          Disable upload API and form controller`,
			`     --folder-make-disable     Disable make folder API and form controller`,
			`     --index-disable           Disable current folder read`,
			``,
			``,
			`     --extend-mode             Enable delete mechanics. Be very careful. It disabled by default.`,
			``,
			`     --prefork                 Enable spawn multiple processes`,
			``,
			`     --crypt                   Enable encrypt file support`,
			``,
			`     --spa                     Enable frontend SPA (Single Page Application)`,
			``,
			``,
			`     --silence                 Disable all console messages`,
			`     --nolog                   Do not write any data to event_log table`,
			`     --nodb                    Disable database use`,
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
			`                        ` + green_clr(`http-here`) + ` --tls --user ` + white_clr(`user`+util.RandStringRunes(2)) + ` --password ` + white_clr(util.RandStringRunes(12)) + ` --prefork ` + white_clr(`/tmp/fold`),
			``,
			`     Effective`,
			`                        ` + green_clr(`http-here`) + ` --prefork --nodb ` + white_clr(`/tmp`),
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

	if runtime.GOOS == "windows" {
		arg_fold = util.RotateSlash(arg_fold)
	}

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

	prefix := path.Join(homepath, ".httphere")

	/*
		db, err := model.ConnectToSQLite(prefix)
		if err != nil {
			panic(err)
		}*/

	var db *gorm.DB

	if !*arg_nodb {
		db, err = model.ConnectToSQLite(prefix)
		if err != nil {
			panic(err)
		}
	}

	if !fiber.IsChild() {

		model.EventLogAdd(db, nil, "", "INIT", "run "+arg_fold)

		go model.FileChkAsync(db, prefix)
	}

	if _, err := os.Stat(path.Join(prefix, "thumb")); err != nil {
		if err := os.MkdirAll(path.Join(prefix, "thumb"), os.ModePerm); err != nil {

			fmt.Println(err)
			return
		}
	}

	if _, err := os.Stat(path.Join(prefix, "tls")); os.IsNotExist(err) {
		if err := os.MkdirAll(path.Join(prefix, "tls"), os.ModePerm); err != nil {

			fmt.Println(err)
			return
		}
	}

	if !fiber.IsChild() {

		util.WalkAndClearZeroFile(path.Join(prefix, "thumb"), 0)
		//go WalkAndClearOld(path.Join(prefix, "thumb"))
	}

	engine := html.NewFileSystem(http.FS(view_fs), ".html")

	config := fiber.Config{
		Prefork:               *arg_prefork,
		DisableStartupMessage: true,
		ServerHeader:          "",
		Views:                 engine,
		BodyLimit:             conf.FieldSize_max,
	}

	app := fiber.New(config)

	app.Use(func(c *fiber.Ctx) error {

		c.Locals("prefix", prefix)
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
		if *arg_silence {

			c.Locals("arg_silence", "1")
		}
		if *arg_nolog {

			c.Locals("arg_nolog", "1")
		}
		if *arg_nodb {

			c.Locals("arg_nodb", "1")
		}
		c.Locals("db", db)

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

			login := "login" + strconv.Itoa(i) + util.RandStringRunes(2)
			password := util.RandStringRunes(16)

			password_list[login] = password

			fmt.Println("         " + login + "    " + password)
		}

		app.Use(basicauth.New(basicauth.Config{

			Users: password_list,

			Unauthorized: func(c *fiber.Ctx) error {

				model.EventLogAdd(db, c, "401", "basicauth", path.Join(arg_fold, c.Path()))

				c.Set(fiber.HeaderWWWAuthenticate, "Basic realm='Restricted'")
				return c.Status(fiber.StatusUnauthorized).Render("view/401", fiber.Map{}, "view/layout/error")
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

				model.EventLogAdd(db, c, "401", "basicauth", path.Join(arg_fold, c.Path()))

				c.Set(fiber.HeaderWWWAuthenticate, "Basic realm='Restricted'")
				return c.Status(fiber.StatusUnauthorized).Render("view/401", fiber.Map{}, "view/layout/error")
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

		if _, err3 := os.Stat(path.Join(prefix, "temp")); err3 != nil {

			fmt.Println("")
			//fmt.Println("  Make temp folder")

			if err4 := os.MkdirAll(path.Join(prefix, "temp"), os.ModePerm); err4 != nil {
				fmt.Println(err4)
				return
			}
		} else {

			fmt.Println("")
			//fmt.Println(yellow("  Clear temp folder"))

			if err := os.RemoveAll(path.Join(prefix, "temp")); err != nil {
				fmt.Println(err)
				return
			}

			if err4 := os.MkdirAll(path.Join(prefix, "temp"), os.ModePerm); err4 != nil {
				fmt.Println(err4)
				return
			}

		}
	}

	//app.Static("/__temp", path.Join(prefix, "temp"))

	if *arg_extend_mode {
		//app.Get("/__resize/:width/:height/*", controller.GetResize)
		app.Get("/__thumb/*", GetThumb)
	}

	app.Get("/__temp/*", func(c *fiber.Ctx) error {

		c_path, err := url.QueryUnescape(c.Path())
		c_path = strings.TrimLeft(c_path, "/__temp")

		if err != nil {

			model.EventLogAdd(db, c, "500", "__temp", "Error "+path.Join(prefix, "temp", c_path)+" "+err.Error())
			return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")
		}

		c_path = util.CleanDirtyPath(c_path)
		full_filename := path.Join(prefix, "temp", c_path)

		if runtime.GOOS == "windows" {
			full_filename = util.RotateSlash(full_filename)
		}

		_, err = os.Stat(full_filename)
		if err != nil {

			model.EventLogAdd(db, c, "404", "__temp", "'"+full_filename+"' not found")

			return c.JSON(fiber.Map{
				"code": 404,
				"msg":  path.Join(c_path) + " not found",
			}, "application/json")
		}

		model.EventLogAdd(db, c, "200", "__temp", "Temp get "+full_filename)

		return c.SendFile(full_filename)
	})

	app.Options("/*", api.OptionsAll)

	if !*arg_index_disable {

		if *arg_extend_mode {
			app.Get("/__doc/*", GetEditDoc)
			app.Get("/__code/*", GetEditCode)
			app.Get("/__search/", GetSearch)
			app.Get("/api/file/convert/*", api.GetFileConvert)
			app.Post("/api/file/edit", PostFileEdit)
			app.Get("/api/list", api.GetList)
		}
	}

	if !*arg_upload_disable {
		app.Post("/api/file/upload", api.PostFileUpload)
	}

	if !*arg_folder_make_disable {
		app.Post("/api/folder", api.PostFolder)
		app.Post("/api/file/touch", api.PostFileTouch)
	}

	if *arg_extend_mode {
		app.Post("/api/delete", api.PostDelete)
		app.Post("/api/move", api.PostMove)
		app.Post("/api/copy", api.PostCopy)
		app.Post("/api/rename", api.PostRename)
		app.Post("/api/zip", api.PostZip)

		app.Get("/api/search", api.GetSearch)
	}

	if *arg_spa && !*arg_extend_mode {

		fmt.Println("")
		fmt.Println("You can not run --spa without --extend-mode")
		return
	}

	if !*arg_index_disable {

		if *arg_spa {
			app.Get("/", GetSpa)
			app.Get("/*", api.GetAll)
		} else {
			app.Get("/*", api.GetAll)
		}
	}

	app.Use(func(c *fiber.Ctx) error {

		model.EventLogAdd(db, c, "404", "last_handle", path.Join(arg_fold, c.Path()))

		return c.Status(fiber.StatusNotFound).Render("view/404", fiber.Map{}, "view/layout/error")
	})

	crt_filename := path.Join(prefix, "tls", "server.pem")
	key_filename := path.Join(prefix, "tls", "server.key")

	if runtime.GOOS == "windows" {
		crt_filename = util.RotateSlash(crt_filename)
		key_filename = util.RotateSlash(key_filename)
	}

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

		cert.Run(path.Join(prefix, "tls"), "server")

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
