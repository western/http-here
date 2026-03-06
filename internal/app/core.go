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
	//"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"

	"github.com/western/http-here/v2/internal/api"
	"github.com/western/http-here/v2/internal/cert"
	"github.com/western/http-here/v2/internal/conf"
	"github.com/western/http-here/v2/internal/model2"
	"github.com/western/http-here/v2/internal/util"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/template/html/v2"
)

//go:embed view/*
var view_fs embed.FS

//go:embed assets/*
var embedDirStatic embed.FS

var (
	green_clr  = color.New(color.FgGreen).SprintFunc()
	white_clr  = color.New(color.Bold, color.FgWhite).SprintFunc()
	cian_clr   = color.New(color.FgCyan).SprintFunc()
	yellow_clr = color.New(color.FgYellow).SprintFunc()
)

func Core() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	// -------------------------------------------------------------------------------------------------------------------------------------------

	arg_help := flag.Bool("help", false, "Show help")

	arg_port := flag.Int("port", 8000, "Port to use")
	arg_tls := flag.Bool("tls", false, "Start HTTPS")
	arg_tls_debug := flag.Bool("tls-debug", false, "Start HTTPS with verbosity")

	arg_login := flag.String("login", "", "Login for user basic auth")
	arg_password := flag.String("password", "", "Password for user basic auth")
	arg_basic := flag.Bool("basic", false, "Set basic auth and generate several accounts every time")

	arg_upload_disable := flag.Bool("upload-disable", false, "Disable upload API and form controller")
	arg_folder_make_disable := flag.Bool("folder-make-disable", false, "Disable make folder API and form controller")
	arg_share_only := flag.Bool("share-only", false, "Set --upload-disable and --folder-make-disable")
	arg_index_disable := flag.Bool("index-disable", false, "Disable current folder read")

	arg_extend_mode := flag.Bool("extend-mode", false, "Enable delete mechanics. Be very careful. It disabled by default.")
	arg_crypt := flag.Bool("crypt", false, "Enable encrypt file support.")
	arg_spa := flag.Bool("spa", false, "Enable frontend SPA (Single Page Application)")

	arg_prefork := flag.Bool("prefork", false, "Enable spawn multiple processes")

	arg_silence := flag.Bool("silence", false, "Disable all console messages")
	arg_nolog := flag.Bool("nolog", false, "Do not write any data to event_log table")

	arg_usedb := flag.Bool("usedb", false, "Database enable")
	arg_cache_dir := flag.Int("cache-dir", 30, "Cache timeout for readdir, seconds")

	flag.Parse()

	if *arg_tls_debug {
		*arg_tls = *arg_tls_debug
	}

	if *arg_share_only {
		*arg_upload_disable = true
		*arg_folder_make_disable = true
	}

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
			`     --login ` + white_clr(`[str]`) + `             Login for basic authorization`,
			`     --password ` + white_clr(`[str]`) + `          Password for basic authorization`,
			``,
			`     --basic                   Set basic auth and generate several accounts every time`,
			``,
			``,
			`     --index-disable           Disable current folder read`,
			`     --share-only              Set --upload-disable and --folder-make-disable`,
			``,
			``,
			`     --extend-mode             Enable extended mode`,
			``,
			`     --prefork                 Enable spawn multiple processes`,
			``,
			`     --crypt                   Enable encrypt file support`,
			``,
			`     --spa                     Enable frontend SPA (Single Page Application)`,
			``,
			``,
			`     --silence                 Disable all console messages`,
			``,
			`     --usedb                   Database enable`,
			`     --nolog                   Do not write any data to event_log table`,
			``,
			`     --cache-dir ` + white_clr(`[int]`) + `         Cache timeout for readdir, seconds [30]`,
			``,
			``,
			`examples:`,
			``,
			`     The safest run`,
			`                        ` + green_clr(`http-here`) + ` --tls --basic ` + white_clr(`/some/path`),
			``,
			`     Only share`,
			`                        ` + green_clr(`http-here`) + ` --share-only ` + white_clr(`/tmp/fold`),
			``,
			`     For work`,
			`                        ` + green_clr(`http-here`) + ` --tls --login ` + white_clr(`login`+util.RandStringRunes(2)) + ` --password ` + white_clr(util.RandStringRunes(12)) + ` --prefork --extend-mode ` + white_clr(`/tmp/fold`),
			``,
			`     Maximum performance`,
			`                        ` + green_clr(`http-here`) + ` --prefork --silence ` + white_clr(`/tmp`),
			``,
		}

		fmt.Println(strings.Join(inf[:], "\n"))
		return
	}

	// -------------------------------------------------------------------------------------------------------------------------------------------

	if *arg_usedb && *arg_prefork {

		fmt.Println("")
		fmt.Println("You can not run --prefork with --usedb")
		return
	}

	if len(os.Args) > 1 {

		//fmt.Println("os.Args=", os.Args)

		if os.Args[1] == "user" {

			model2.Open()

			cmdSubcommandUser()
		}

		if os.Args[1] == "usermod" {

			model2.Open()

			cmdSubcommandUserMod()
		}

		if os.Args[1] == "log" {

			model2.Open()

			cmdSubcommandLog()
		}

		if os.Args[1] == "file" {

			model2.Open()

			cmdSubcommandFile()
		}
	}

	// -------------------------------------------------------------------------------------------------------------------------------------------

	// if you not changed cache-dir arg from default 30, the app expect operate with it
	if *arg_prefork && *arg_cache_dir == 30 {

		*arg_cache_dir = 5

		if !fiber.IsChild() {
			fmt.Println()
			fmt.Println("  For prefork automatic set cache-dir=", *arg_cache_dir)
		}
	}

	go func() {
		ticker := time.NewTicker(5 * time.Minute)

		defer ticker.Stop()
		for range ticker.C {
			api.CleanupCacheDir(*arg_cache_dir)
		}
	}()

	// -------------------------------------------------------------------------------------------------------------------------------------------

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

	conf.ArgFold = arg_fold

	// -------------------------------------------------------------------------------------------------------------------------------------------

	// -------------------------------------------------------------------------------------------------------------------------------------------

	if *arg_usedb {

		model2.Open()
	}

	if !fiber.IsChild() {

		model2.EventLogAdd(nil, 0, "INIT", "run "+conf.ArgFold)

		go model2.FileChkAsync()
	}

	// -------------------------------------------------------------------------------------------------------------------------------------------
	// FOLDERS

	if !fiber.IsChild() {

		if _, err := os.Stat(path.Join(conf.ConfigRoot)); err != nil {
			if err := os.MkdirAll(path.Join(conf.ConfigRoot), os.ModePerm); err != nil {
				fmt.Println(err)
				return
			}
		}

		if _, err := os.Stat(path.Join(conf.ConfigRoot, "thumb")); err != nil {
			if err := os.MkdirAll(path.Join(conf.ConfigRoot, "thumb"), os.ModePerm); err != nil {
				fmt.Println(err)
				return
			}
		}

		if _, err := os.Stat(path.Join(conf.ConfigRoot, "tls")); os.IsNotExist(err) {
			if err := os.MkdirAll(path.Join(conf.ConfigRoot, "tls"), os.ModePerm); err != nil {
				fmt.Println(err)
				return
			}
		}

		if _, err = os.Stat(path.Join(conf.ConfigRoot, "temp")); err != nil {

			fmt.Println("")
			//fmt.Println("  Make temp folder")

			if err = os.MkdirAll(path.Join(conf.ConfigRoot, "temp"), os.ModePerm); err != nil {
				fmt.Println(err)
				return
			}
		} else {

			fmt.Println("")
			//fmt.Println(yellow_clr("  Clear temp folder"))

			if err = os.RemoveAll(path.Join(conf.ConfigRoot, "temp")); err != nil {
				fmt.Println(err)
				return
			}

			if err = os.MkdirAll(path.Join(conf.ConfigRoot, "temp"), os.ModePerm); err != nil {
				fmt.Println(err)
				return
			}

		}
	}

	if !fiber.IsChild() {

		util.WalkAndClearZeroFile(path.Join(conf.ConfigRoot, "thumb"), 0)
		//go WalkAndClearOld(path.Join(conf.ConfigRoot, "thumb"))
	}

	// -------------------------------------------------------------------------------------------------------------------------------------------

	engine := html.NewFileSystem(http.FS(view_fs), ".html")

	config := fiber.Config{
		Prefork:               *arg_prefork,
		DisableStartupMessage: true,
		ServerHeader:          "",
		Views:                 engine,
		BodyLimit:             conf.FieldSizeMax,

		// Timeouts
		ReadTimeout:  30 * time.Second, // nil, The amount of time allowed to read the full request, including the body. The default timeout is unlimited.
		WriteTimeout: 30 * time.Second, // nil, The maximum duration before timing out writes of the response. The default timeout is unlimited.
		//IdleTimeout:     120 * time.Second, // nil, The maximum amount of time to wait for the next request when keep-alive is enabled. If IdleTimeout is zero, the value of ReadTimeout is used.
		IdleTimeout:     30 * time.Second,
		ReadBufferSize:  8192, // 4096
		WriteBufferSize: 8192, // 4096

		// Enable request/response pooling
		//EnableTrustedProxyCheck: false,
		//ProxyHeader:            "X-Forwarded-For",

		// Optimize for high concurrency
		//Concurrency:           256 * 1024, // 256K concurrent connections

		// Disable features for better performance
		//DisablePreParseMultipartForm: true, // false
		//StreamRequestBody:            true, // false
	}

	app := fiber.New(config)

	app.Use(func(c *fiber.Ctx) error {

		c.Locals("arg_upload_disable", *arg_upload_disable)
		c.Locals("arg_folder_make_disable", *arg_folder_make_disable)
		c.Locals("arg_extend_mode", *arg_extend_mode)
		c.Locals("arg_crypt", *arg_crypt)
		c.Locals("arg_spa", *arg_spa)
		c.Locals("arg_silence", *arg_silence)
		c.Locals("arg_nolog", *arg_nolog)
		c.Locals("arg_usedb", *arg_usedb)
		c.Locals("arg_cache_dir", *arg_cache_dir)

		return c.Next()
	})

	// -------------------------------------------------------------------------------------------------------------------------------------------
	// MIDDLEWARE

	app.Use(favicon.New())

	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed, // 1
	}))

	// -------------------------------------------------------------------------------------------------------------------------------------------
	// ASSETS

	app.Use("/__assets", filesystem.New(filesystem.Config{
		Root:       http.FS(embedDirStatic),
		PathPrefix: "",
		Browse:     false,
	}))

	// -------------------------------------------------------------------------------------------------------------------------------------------
	// BASIC AUTH

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

				model2.EventLogAdd(c, 401, "basicauth", path.Join(conf.ArgFold, c.Path()))

				c.Set(fiber.HeaderWWWAuthenticate, "Basic realm='Restricted'")
				return c.Status(fiber.StatusUnauthorized).Render("view/401", fiber.Map{}, "view/layout/error")
			},
		}))

	}

	if len(*arg_login) > 0 && len(*arg_password) > 0 {

		app.Use(basicauth.New(basicauth.Config{

			Authorizer: func(user, pass string) bool {
				if user == *arg_login && pass == *arg_password {
					return true
				}

				return false
			},
			Unauthorized: func(c *fiber.Ctx) error {

				model2.EventLogAdd(c, 401, "basicauth", path.Join(conf.ArgFold, c.Path()))

				c.Set(fiber.HeaderWWWAuthenticate, "Basic realm='Restricted'")
				return c.Status(fiber.StatusUnauthorized).Render("view/401", fiber.Map{}, "view/layout/error")
			},
		}))

		if !fiber.IsChild() {
			fmt.Println("")
			fmt.Println("  Basic auth set: " + cian_clr(*arg_login) + " " + cian_clr(*arg_password))
		}
	}

	// -------------------------------------------------------------------------------------------------------------------------------------------

	if *arg_extend_mode {

		app.Get("/__thumb/*", GetThumb)
	}

	app.Get("/__temp/*", func(c *fiber.Ctx) error {

		c_path, err := url.QueryUnescape(c.Path())
		c_path = strings.TrimLeft(c_path, "/__temp")

		if err != nil {

			model2.EventLogAdd(c, 500, "__temp", "Error "+path.Join(conf.ConfigRoot, "temp", c_path)+" "+err.Error())
			return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")
		}

		c_path = util.CleanDirtyPath(c_path)
		full_filename := path.Join(conf.ConfigRoot, "temp", c_path)

		if runtime.GOOS == "windows" {
			full_filename = util.RotateSlash(full_filename)
		}

		_, err = os.Stat(full_filename)
		if err != nil {

			model2.EventLogAdd(c, 404, "__temp", "'"+full_filename+"' not found")

			return c.JSON(fiber.Map{
				"code": 404,
				"msg":  path.Join(c_path) + " not found",
			}, "application/json")
		}

		model2.EventLogAdd(c, 200, "__temp", "Temp get "+full_filename)

		return c.SendFile(full_filename)
	})

	// -------------------------------------------------------------------------------------------------------------------------------------------
	// ROUTE

	if !*arg_index_disable {

		if *arg_extend_mode {
			app.Get("/__doc/*", GetEditDoc)
			app.Get("/__code/*", GetEditCode)
			app.Get("/__md/*", GetEditMd)
			app.Get("/__search/", GetSearch)
			app.Get("/api/file/convert/*", api.GetFileConvert)
			app.Post("/api/file/edit", PostFileEdit)
		}

		app.Get("/api/list", api.GetList)
	}

	if !*arg_upload_disable {
		app.Post("/api/file", api.PostFile)
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

	if !*arg_index_disable {

		if *arg_spa {
			app.Get("/", GetSpa)
			app.Get("/*", api.GetAll)
		} else {
			app.Get("/*", api.GetAll)
		}
	}

	// -------------------------------------------------------------------------------------------------------------------------------------------
	// BAD REQUEST

	app.Use(func(c *fiber.Ctx) error {

		model2.EventLogAdd(c, 404, "last_handle", path.Join(conf.ArgFold, c.Path()))

		return c.Status(fiber.StatusNotFound).Render("view/404", fiber.Map{}, "view/layout/error")
	})

	// -------------------------------------------------------------------------------------------------------------------------------------------

	crt_filename := path.Join(conf.ConfigRoot, "tls", "server.pem")
	key_filename := path.Join(conf.ConfigRoot, "tls", "server.key")

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

		cert.Run(path.Join(conf.ConfigRoot, "tls"), "server")

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

		fmt.Println("  Serve folder: " + cian_clr(conf.ArgFold))
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

